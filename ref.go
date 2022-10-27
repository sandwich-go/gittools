package gittools

import (
	"github.com/go-git/go-git/v5/plumbing"
)

func getBranchReferenceName(branch string) plumbing.ReferenceName {
	if plumbing.ReferenceName(branch).IsBranch() {
		return plumbing.ReferenceName(branch)
	}
	return plumbing.NewBranchReferenceName(branch)
}

func getBranchRemoteReferenceName(branch string) plumbing.ReferenceName {
	if plumbing.ReferenceName(branch).IsRemote() {
		return plumbing.ReferenceName(branch)
	}
	return plumbing.NewRemoteReferenceName("origin", branch)
}

func (r *repository) getBranchReferenceName(branch string) (plumbing.ReferenceName, error) {
	if len(branch) == 0 {
		head, err := r.Head()
		if err != nil {
			return "", err
		}
		return getBranchReferenceName(head.Name().Short()), nil
	}
	return getBranchReferenceName(branch), nil
}

func getTagReferenceName(tag string) plumbing.ReferenceName {
	if plumbing.ReferenceName(tag).IsTag() {
		return plumbing.ReferenceName(tag)
	}
	return plumbing.NewTagReferenceName(tag)
}

func (r *repository) getTagReferenceName(tag string) (plumbing.ReferenceName, error) {
	if len(tag) == 0 {
		head, err := r.Head()
		if err != nil {
			return "", err
		}
		return getTagReferenceName(head.Name().Short()), nil
	}
	return getTagReferenceName(tag), nil
}
