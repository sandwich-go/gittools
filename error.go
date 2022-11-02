package gittools

import (
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var (
	NoErrAlreadyUpToDate         = git.NoErrAlreadyUpToDate
	ErrDeleteRefNotSupported     = git.ErrDeleteRefNotSupported
	ErrForceNeeded               = git.ErrForceNeeded
	ErrExactSHA1NotSupported     = git.ErrExactSHA1NotSupported
	ErrBranchExists              = git.ErrBranchExists
	ErrBranchNotFound            = git.ErrBranchNotFound
	ErrTagExists                 = git.ErrTagExists
	ErrTagNotFound               = git.ErrTagNotFound
	ErrFetching                  = git.ErrFetching
	ErrInvalidReference          = git.ErrInvalidReference
	ErrRepositoryNotExists       = git.ErrRepositoryNotExists
	ErrRepositoryIncomplete      = git.ErrRepositoryIncomplete
	ErrRepositoryAlreadyExists   = git.ErrRepositoryAlreadyExists
	ErrRemoteNotFound            = git.ErrRemoteNotFound
	ErrRemoteExists              = git.ErrRemoteExists
	ErrAnonymousRemoteName       = git.ErrAnonymousRemoteName
	ErrWorktreeNotProvided       = git.ErrWorktreeNotProvided
	ErrIsBareRepository          = git.ErrIsBareRepository
	ErrUnableToResolveCommit     = git.ErrUnableToResolveCommit
	ErrPackedObjectsNotSupported = git.ErrPackedObjectsNotSupported
	ErrWorktreeNotClean          = git.ErrWorktreeNotClean
	ErrSubmoduleNotFound         = git.ErrSubmoduleNotFound
	ErrUnstagedChanges           = git.ErrUnstagedChanges
	ErrGitModulesSymlink         = git.ErrGitModulesSymlink
	ErrNonFastForwardUpdate      = git.ErrNonFastForwardUpdate
	ErrObjectNotFound            = plumbing.ErrObjectNotFound
	ErrInvalidType               = plumbing.ErrInvalidType
	ErrReferenceNotFound         = plumbing.ErrReferenceNotFound
)

func checkErr(err error) error {
	if err == NoErrAlreadyUpToDate || err == ErrNonFastForwardUpdate {
		err = nil
	}
	return err
}
