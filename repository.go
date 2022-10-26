package gittools

import (
	"context"
	"fmt"
	"github.com/dastoori/higgs"
	"github.com/go-git/go-billy/v5"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
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
	repo.updateCurrentRefName(plumbing.Master)
	return repo
}

func (r *repository) cleanWorkTree() (*git.Worktree, error) {
	worktree, err := r.Repository.Worktree()
	if err != nil {
		return nil, err
	}
	status, err0 := worktree.Status()
	if err0 != nil {
		return nil, err0
	}
	for k, v := range status {
		if v.Worktree != git.Untracked && v.Staging != git.Untracked {
			continue
		}
		if is, _ := higgs.IsHidden(k); is {
			delete(status, k)
		}
	}
	if !status.IsClean() {
		return nil, fmt.Errorf("current worktree not clean")
	}
	return worktree, nil
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

func getBranchName(branch string) plumbing.ReferenceName {
	if len(branch) == 0 {
		return plumbing.Master
	}
	return plumbing.NewRemoteReferenceName("origin", strings.TrimPrefix(branch, "origin"))
}

func getTagName(tag string) plumbing.ReferenceName {
	if len(tag) == 0 {
		return plumbing.Master
	}
	return plumbing.NewTagReferenceName(tag)
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
	return nil
}

func (r *repository) updateCurrentRefName(ref plumbing.ReferenceName) {
	r.currentRefName = ref
}

func (r *repository) checkout(ref plumbing.ReferenceName) (err error) {
	defer func() { r.print(err, "checkout,") }()
	var workTree *git.Worktree
	workTree, err = r.cleanWorkTree()
	if err != nil {
		return err
	}
	if err = workTree.Checkout(&git.CheckoutOptions{
		Branch: ref,
	}); err != nil {
		return err
	}
	if err = r.updateHeadHash(); err != nil {
		return err
	}
	r.updateCurrentRefName(ref)
	return nil
}

func (r *repository) CheckoutBranch(_ context.Context, branch string) error {
	return r.checkout(getBranchName(branch))
}

func (r *repository) CheckoutTag(_ context.Context, tag string) error {
	return r.checkout(getTagName(tag))
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
	if err = workTree.PullContext(ctx, &git.PullOptions{
		Depth:    r.h.spec.GetDepth(),
		Auth:     publicKeys,
		Progress: r.h.getProgress(),
	}); err != nil {
		if err == git.NoErrAlreadyUpToDate || err == git.ErrNonFastForwardUpdate {
			// if err is that the repo is up-to-date or a commit is done already we ignore it
			err = nil
		} else {
			return err
		}
	}
	if err = r.updateHeadHash(); err != nil {
		return err
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
	defer func() {
		_ = f.Close()
	}()
	_, err = f.Write(data)
	if err == nil {
		_, err = workTree.Add(file)
	}
	return
}

func (r *repository) Commit(_ context.Context, msg string) (err error) {
	defer func() { r.print(err, "commit,") }()
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
	if _, err = workTree.Commit(msg, &git.CommitOptions{}); err != nil {
		return err
	}
	if err = r.updateHeadHash(); err != nil {
		return err
	}
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
	if err == git.NoErrAlreadyUpToDate || err == git.ErrNonFastForwardUpdate {
		// if err is that the repo is up-to-date or a commit is done already we ignore it
		err = nil
	}
	return err
}
