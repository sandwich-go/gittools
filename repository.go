package gittools

import (
	"bufio"
	"context"
	"fmt"
	"github.com/dastoori/higgs"
	"github.com/go-git/go-billy/v5"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type repository struct {
	*git.Repository
	h              *cloner
	headHash       plumbing.Hash
	currentRefName plumbing.ReferenceName
}

func newRepository(h *cloner, r *git.Repository) Repository {
	repo := &repository{h: h, Repository: r}
	_ = repo.updateHeadHash()
	return repo
}

func (r *repository) cleanWorkTree() (*git.Worktree, error) {
	worktree, err := r.Repository.Worktree()
	if err != nil {
		return nil, err
	}
	var is bool
	is, err = r.IsClean()
	if !is {
		return nil, fmt.Errorf("current worktree not clean")
	}
	return worktree, nil
}

func (r *repository) IsClean() (bool, error) {
	worktree, err := r.Repository.Worktree()
	if err != nil {
		return false, err
	}
	status, err0 := worktree.Status()
	if err0 != nil {
		return false, err0
	}
	for k, v := range status {
		if v.Worktree != git.Untracked && v.Staging != git.Untracked {
			continue
		}
		if is, _ := higgs.IsHidden(k); is {
			delete(status, k)
		}
	}
	return status.IsClean(), nil
}

func (r *repository) UserName() string {
	c, _ := r.Config()
	if c == nil {
		return ""
	}
	return c.User.Name
}

func (r *repository) UserEmail() string {
	c, _ := r.Config()
	if c == nil {
		return ""
	}
	return c.User.Email
}

func (r *repository) Root() string {
	wt, _ := r.Worktree()
	if wt == nil || wt.Filesystem == nil {
		return ""
	}
	return wt.Filesystem.Root()
}

func (r *repository) RemoveAll() error {
	return os.RemoveAll(r.Root())
}

func (r *repository) print(err error, v ...interface{}) {
	if r.currentRefName.IsTag() {
		r.h.print(err, append(v, fmt.Sprintf("tag: %s,", r.currentRefName.Short()), fmt.Sprintf("hash: %s,", r.headHash))...)
	} else {
		r.h.print(err, append(v, fmt.Sprintf("branch: %s,", r.currentRefName.Short()), fmt.Sprintf("hash: %s,", r.headHash))...)
	}
}

func (r *repository) updateHeadHash() error {
	head, err := r.Head()
	if err != nil {
		return err
	}
	r.headHash = head.Hash()
	r.currentRefName = head.Name()
	return nil
}

func (r *repository) checkout(ref plumbing.ReferenceName) (err error) {
	defer func() { r.print(err, fmt.Sprintf("checkout, want: %s,", ref.String())) }()
	var workTree *git.Worktree
	workTree, err = r.cleanWorkTree()
	if err != nil {
		return
	}
	if err = workTree.Checkout(&git.CheckoutOptions{
		Branch: ref,
	}); err != nil {
		return
	}
	err = r.updateHeadHash()
	return
}

func (r *repository) Pull(ctx context.Context) (err error) {
	defer func() { r.print(err, "pull,") }()
	var publicKeys *ssh.PublicKeys
	if publicKeys, err = r.h.auth(); err != nil {
		return
	}
	var workTree *git.Worktree
	if workTree, err = r.cleanWorkTree(); err != nil {
		return err
	}
	err = workTree.PullContext(ctx, &git.PullOptions{
		Depth:    r.h.GetDepth(),
		Auth:     publicKeys,
		Progress: r.h.getProgress(),
	})
	if err = checkErr(err); err == nil {
		err = r.updateHeadHash()
	}
	return
}

const (
	ignoreFile    = ".gitignore"
	commentPrefix = "#"
)

func (r *repository) isIgnore(_ context.Context, fileOrDir []string, isDir bool) (is bool, err error) {
	defer func() { r.print(err, fmt.Sprintf("is ignore fileOrDir: %v, isDir: %v", fileOrDir, isDir)) }()
	var workTree *git.Worktree
	if workTree, err = r.Worktree(); err != nil {
		return
	}
	var ps, ps1 []gitignore.Pattern
	ps, _ = gitignore.LoadGlobalPatterns(workTree.Filesystem)
	ps1, _ = gitignore.LoadSystemPatterns(workTree.Filesystem)
	ps = append(ps, ps1...)
	ps1, _ = gitignore.ReadPatterns(workTree.Filesystem, nil)
	ps = append(ps, ps1...)

	for _, i := range ps {
		if i.Match(fileOrDir, isDir) == gitignore.Exclude {
			return true, nil
		}
	}
	return false, err
}

func (r *repository) IsIgnoreDir(ctx context.Context, dirs ...string) (bool, error) {
	return r.isIgnore(ctx, dirs, true)
}

func (r *repository) IsIgnoreFile(ctx context.Context, files ...string) (bool, error) {
	return r.isIgnore(ctx, files, false)
}

func (r *repository) Ignore(_ context.Context, patterns ...string) (err error) {
	defer func() { r.print(err, fmt.Sprintf("ignore pattern: %v", patterns)) }()
	var workTree *git.Worktree
	if workTree, err = r.Worktree(); err != nil {
		return err
	}
	var file billy.File
	var content = make(map[string]struct{})
	var bs []byte

	_, err = workTree.Filesystem.Stat(ignoreFile)
	if err == nil {
		file, err = workTree.Filesystem.OpenFile(ignoreFile, os.O_RDONLY, 0666)
		if err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				s := scanner.Text()
				if !strings.HasPrefix(s, commentPrefix) && len(strings.TrimSpace(s)) > 0 {
					content[s] = struct{}{}
				}
			}
			if _, err = file.Seek(0, io.SeekStart); err == nil {
				bs, err = ioutil.ReadAll(file)
			}
		}
		if file != nil {
			_ = file.Close()
		}
	}
	if err != nil {
		return
	}
	file, err = workTree.Filesystem.Create(ignoreFile)
	if err != nil {
		return
	}
	var cnt = string(bs)
	for _, pattern := range patterns {
		if _, ok := content[pattern]; !ok {
			cnt += "\n"
			cnt += pattern
		}
	}
	cnt += "\n"
	_, err = file.Write([]byte(cnt))
	_ = file.Close()
	return
}

func (r *repository) Add(_ context.Context, fileOrDirs ...string) (err error) {
	defer func() { r.print(err, fmt.Sprintf("add file/dir: %v", fileOrDirs)) }()
	var workTree *git.Worktree
	if workTree, err = r.Worktree(); err != nil {
		return err
	}
	for _, f := range fileOrDirs {
		_, err = workTree.Add(f)
		if err != nil {
			break
		}
	}
	return
}

func (r *repository) RewriteFile(_ context.Context, file string, data []byte) (err error) {
	defer func() { r.print(err, "rewrite file,") }()
	var workTree *git.Worktree
	if workTree, err = r.Worktree(); err != nil {
		return err
	}
	var f billy.File
	if f, err = workTree.Filesystem.Create(file); err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err = f.Write(data); err == nil {
		_, err = workTree.Add(file)
	}
	return
}

func (r *repository) Commit(_ context.Context, comment string) (err error) {
	defer func() { r.print(err, fmt.Sprintf("commit, comment: %s", comment)) }()
	var workTree *git.Worktree
	if workTree, err = r.Worktree(); err != nil {
		return err
	}
	var status git.Status
	if status, err = workTree.Status(); err != nil {
		return err
	}
	if status.IsClean() {
		return nil
	}
	if _, err = workTree.Commit(comment, &git.CommitOptions{}); err != nil {
		return err
	}
	err = r.updateHeadHash()
	return
}

func (r *repository) Push(ctx context.Context) (err error) {
	defer func() { r.print(err, "push,") }()
	var publicKeys *ssh.PublicKeys
	if publicKeys, err = r.h.auth(); err != nil {
		return
	}
	err = r.Repository.PushContext(ctx, &git.PushOptions{
		Auth:     publicKeys,
		Progress: r.h.getProgress(),
	})
	err = checkErr(err)
	return
}

func (r *repository) CheckoutBranch(_ context.Context, branch string) error {
	if len(branch) == 0 {
		return r.checkout(plumbing.Master)
	}
	return r.checkout(getBranchRemoteReferenceName(branch))
}

func (r *repository) Branch(_ context.Context, branch string) (bc Branch, err error) {
	defer func() { r.print(err, fmt.Sprintf("branch, name: %s", branch)) }()
	var brn plumbing.ReferenceName
	brn, err = r.getBranchReferenceName(branch)
	if err != nil {
		return
	}
	var b *config.Branch
	b, err = r.Repository.Branch(brn.Short())
	if err == nil {
		bc = newBranch(r, b)
	}
	return
}

func (r *repository) getHash(hash string) (plumbing.Hash, error) {
	var hh plumbing.Hash
	if len(hash) == 0 {
		ref, err := r.Head()
		if err != nil {
			return plumbing.ZeroHash, nil
		}
		hh = ref.Hash()
	} else {
		hh = plumbing.NewHash(hash)
	}
	return hh, nil
}

func (r *repository) CreateBranch(ctx context.Context, branch string, hash string) (bc Branch, err error) {
	defer func() { r.print(err, fmt.Sprintf("create branch, name: %s", branch)) }()
	var brn plumbing.ReferenceName
	brn, err = r.getBranchReferenceName(branch)
	if err != nil {
		return
	}
	var hh plumbing.Hash
	if hh, err = r.getHash(hash); err != nil {
		return
	}
	if err = r.Repository.CreateBranch(&config.Branch{
		Name:  brn.Short(),
		Merge: brn,
	}); err != nil {
		return
	}
	err = r.Storer.SetReference(plumbing.NewHashReference(brn, hh))
	if err == nil {
		bc, err = r.Branch(ctx, brn.Short())
	}
	return
}

func (r *repository) DeleteLocalBranch(ctx context.Context, branch string) (err error) {
	defer func() { r.print(err, fmt.Sprintf("delete local branch, name: %s", branch)) }()
	var brn plumbing.ReferenceName
	brn, err = r.getBranchReferenceName(branch)
	if err != nil {
		return
	}
	var bc Branch
	if bc, err = r.Branch(ctx, brn.Short()); err == nil {
		err = bc.Delete(ctx)
	}
	return
}

func (r *repository) DeleteBranch(ctx context.Context, branch string) (err error) {
	defer func() { r.print(err, fmt.Sprintf("delete branch, name: %s", branch)) }()
	var brn plumbing.ReferenceName
	brn, err = r.getBranchReferenceName(branch)
	if err != nil {
		return
	}
	bc := newBranch(r, &config.Branch{
		Name:  brn.Short(),
		Merge: brn,
	})
	_ = bc.Delete(ctx)
	err = bc.Push(ctx)
	return
}

func (r *repository) CheckoutTag(_ context.Context, tag string) error {
	if len(tag) == 0 {
		return r.checkout(plumbing.Master)
	}
	return r.checkout(getTagReferenceName(tag))
}

func (r *repository) Tag(_ context.Context, tag string) (t Tag, err error) {
	defer func() { r.print(err, fmt.Sprintf("tag, name: %s", tag)) }()
	var trn plumbing.ReferenceName
	trn, err = r.getTagReferenceName(tag)
	if err != nil {
		return
	}
	var ref *plumbing.Reference
	ref, err = r.Repository.Tag(trn.Short())
	if err == nil {
		t = newTag(r, ref)
	}
	return
}

func (r *repository) CreateTag(_ context.Context, tag, comment, hash string) (t Tag, err error) {
	defer func() { r.print(err, fmt.Sprintf("create tag, name: %s, comment:%s", tag, comment)) }()
	var trn plumbing.ReferenceName
	trn, err = r.getTagReferenceName(tag)
	if err != nil {
		return
	}
	var hh plumbing.Hash
	hh, err = r.getHash(hash)
	if err != nil {
		return
	}
	var ref *plumbing.Reference
	ref, err = r.Repository.CreateTag(trn.Short(), hh, &git.CreateTagOptions{Message: comment})
	if err == nil {
		t = newTag(r, ref)
	}
	return
}

func (r *repository) DeleteLocalTag(ctx context.Context, tag string) (err error) {
	defer func() { r.print(err, fmt.Sprintf("delete local tag, name: %s", tag)) }()
	var trn plumbing.ReferenceName
	trn, err = r.getTagReferenceName(tag)
	if err != nil {
		return
	}
	var t Tag
	if t, err = r.Tag(ctx, trn.Short()); err == nil {
		err = t.Delete(ctx)
	}
	return
}

func (r *repository) DeleteTag(ctx context.Context, tag string) (err error) {
	defer func() { r.print(err, fmt.Sprintf("delete tag, name: %s", tag)) }()
	var trn plumbing.ReferenceName
	trn, err = r.getTagReferenceName(tag)
	if err != nil {
		return
	}
	t := newTag(r, plumbing.NewHashReference(trn, plumbing.ZeroHash))
	_ = t.Delete(ctx)
	err = t.Push(ctx)
	return
}

func (r *repository) Fetch(ctx context.Context) (err error) {
	defer func() { r.print(err, "fetch,") }()
	var publicKeys *ssh.PublicKeys
	if publicKeys, err = r.h.auth(); err != nil {
		return
	}
	err = checkErr(r.Repository.FetchContext(ctx, &git.FetchOptions{
		Auth:     publicKeys,
		Progress: r.h.getProgress(),
		Depth:    r.h.GetDepth(),
	}))
	return
}
