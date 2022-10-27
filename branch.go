package gittools

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/config"
)

type branch struct {
	base
	*config.Branch
}

func newBranch(r *repository, b *config.Branch) Branch {
	t := &branch{Branch: b}
	t.base = base{r: r, getRefSpecs: t.getRefSpecs}
	return t
}

func (b *branch) getRefSpecs() []config.RefSpec {
	tn := getBranchReferenceName(b.Name).String()
	if b.isDeleted {
		return []config.RefSpec{config.RefSpec(fmt.Sprintf(":%s", tn))}
	}
	return []config.RefSpec{}
}

func (b *branch) Delete(ctx context.Context) error {
	err := b.r.Repository.DeleteBranch(b.Name)
	if err == nil {
		err = b.r.Storer.RemoveReference(b.Merge)
	}
	if err == nil {
		err = b.base.Delete(ctx)
	}
	return err
}
