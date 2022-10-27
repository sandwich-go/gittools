package gittools

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

type base struct {
	r           *repository
	isDeleted   bool
	getRefSpecs func() []config.RefSpec
}

func (b *base) Push(ctx context.Context) error {
	publicKeys, err := b.r.h.auth()
	if err != nil {
		return err
	}
	return checkErr(b.r.Repository.PushContext(ctx, &git.PushOptions{
		Auth:     publicKeys,
		Progress: b.r.h.getProgress(),
		RefSpecs: b.getRefSpecs(),
	}))
}

func (b *base) Delete(_ context.Context) error {
	b.isDeleted = true
	return nil
}

type tag struct {
	base
	*plumbing.Reference
}

func newTag(r *repository, ref *plumbing.Reference) Tag {
	t := &tag{Reference: ref}
	t.base = base{r: r, getRefSpecs: t.getRefSpecs}
	return t
}

func (t *tag) getRefSpecs() []config.RefSpec {
	tn := getTagReferenceName(t.Name().String()).String()
	if t.isDeleted {
		return []config.RefSpec{config.RefSpec(fmt.Sprintf(":%s", tn))}
	}
	return []config.RefSpec{config.RefSpec(fmt.Sprintf("+%s:%s", tn, tn))}
}

func (t *tag) Delete(ctx context.Context) error {
	err := t.r.Repository.DeleteTag(t.Name().Short())
	if err == nil {
		err = t.base.Delete(ctx)
	}
	return err
}
