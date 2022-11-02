package gittools

import (
	"context"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/sideband"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/sandwich-go/boost/xos"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	successLogFlag = "☕️☕️☕️, success!"
	failedLogFlag  = "🚫🚫🚫, failed!"
	logPrefix      = "[git]"
)

var defaultCloner Cloner

type cloner struct {
	publicKeys *ssh.PublicKeys
	ConfigInterface
}

func New(opts ...ConfigOption) Cloner { return &cloner{ConfigInterface: NewConfig(opts...)} }

func Default(opts ...ConfigOption) Cloner {
	if defaultCloner == nil {
		defaultCloner = New(opts...)
	}
	return defaultCloner
}

func (h *cloner) print(err error, v ...interface{}) {
	if err != nil {
		h.GetLogger().Fatalln(append(append([]interface{}{logPrefix, failedLogFlag}, v...), "Error:", err)...)
	} else {
		// 若最后一位是字符串，并且以','结尾，则移除','
		// 例如:
		// 2022/10/25 19:07:05 [git] ☕️☕️☕️, success! checkout, branch: master, hash: 086e42373a2433101b52bf35ecf84e1df9445c3f,
		// to:
		// 2022/10/25 19:07:05 [git] ☕️☕️☕️, success! checkout, branch: master, hash: 086e42373a2433101b52bf35ecf84e1df9445c3f
		v = append([]interface{}{logPrefix, successLogFlag}, v...)
		if l := len(v); l > 0 {
			if s, ok := v[l-1].(string); ok {
				v[l-1] = strings.TrimSuffix(s, ",")
			}
		}
		h.GetLogger().Println(v...)
	}
}

func (h *cloner) auth() (*ssh.PublicKeys, error) {
	if h.publicKeys != nil {
		return h.publicKeys, nil
	}
	var err error
	var rsaPath string
	defer func() { h.print(err, fmt.Sprintf("auth, path: %s", rsaPath)) }()
	if !filepath.IsAbs(h.GetRsaPath()) {
		var home string
		if home, err = os.UserHomeDir(); err == nil {
			rsaPath = filepath.Join(home, h.GetRsaPath())
		}
	} else {
		rsaPath = h.GetRsaPath()
	}
	if err != nil {
		return nil, err
	}
	exists := xos.ExistsFile(rsaPath)
	if !exists {
		err = fmt.Errorf("not found rsa path")
		return nil, err
	}
	var buf []byte
	buf, err = ioutil.ReadFile(rsaPath)
	if err != nil {
		return nil, err
	}
	h.publicKeys, err = ssh.NewPublicKeys("git", buf, "")
	return h.publicKeys, err
}

func (h *cloner) checkConfig(r *git.Repository) error {
	c, err := r.Config()
	if err != nil {
		return err
	}
	if c == nil {
		c = config.NewConfig()
	}
	c.User.Name = h.GetUserName()
	c.User.Email = h.GetUserEmail()
	return r.SetConfig(c)
}

func (h *cloner) getProgress() sideband.Progress {
	return h.GetLogger().Writer()
}

func (h *cloner) clone(ctx context.Context, url, dir string) (Repository, error) {
	publicKeys, err := h.auth()
	if err != nil {
		return nil, err
	}
	var r *git.Repository
	var opts = &git.CloneOptions{
		URL:      url,
		Auth:     publicKeys,
		Depth:    h.GetDepth(),
		Progress: h.getProgress(),
	}
	if len(dir) == 0 {
		r, err = git.CloneContext(ctx, memory.NewStorage(), memfs.New(), opts)
	} else {
		r, err = git.PlainCloneContext(ctx, dir, false, opts)
	}
	if err != nil {
		return nil, err
	}
	if err = h.checkConfig(r); err != nil {
		return nil, err
	}
	return newRepository(h, r), nil
}

func (h *cloner) Clone(ctx context.Context, url, dir string) (Repository, error) {
	if len(dir) == 0 {
		var err error
		if dir, err = ioutil.TempDir(dir, ""); err != nil {
			return nil, err
		}
	}
	repo, err := h.clone(ctx, url, dir)
	if repo != nil {
		dir = repo.Root()
	}
	h.print(err, fmt.Sprintf("clone to dir, url: %s, dir: %s", url, dir))
	return repo, err
}

func (h *cloner) CloneToMemory(ctx context.Context, url string) (Repository, error) {
	repo, err := h.clone(ctx, url, "")
	h.print(err, fmt.Sprintf("clone to memory, url: %s", url))
	return repo, err
}
