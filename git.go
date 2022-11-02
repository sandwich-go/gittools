package gittools

import (
	"context"
)

type Branch interface {
	// Delete 删除本地分支
	Delete(ctx context.Context) error
	// Push git push，如果本地被删除，push后，远端也将删除
	Push(ctx context.Context) error
}

type Tag interface {
	// Delete 删除本地标签
	Delete(ctx context.Context) error
	// Push git push，如果本地被标签，push后，远端也将删除
	Push(ctx context.Context) error
}

type Repository interface {
	// UserName 获取.git/config的user.name
	UserName() string
	// UserEmail 获取.git/config的user.email
	UserEmail() string

	// Root 若Repository是克隆在本地，显示克隆的根目录
	Root() string
	// RemoveAll 若Repository是克隆在本地，删除克隆的根目录
	RemoveAll() error

	// Pull git pull
	Pull(ctx context.Context) error
	// Push git push
	Push(ctx context.Context) error
	// Commit git commit -m ""
	Commit(ctx context.Context, comment string) error

	// RewriteFile 重写文件内容，不存在则创建
	RewriteFile(ctx context.Context, file string, data []byte) error

	// CheckoutBranch checkout分支，若branch为空，则checkout master分支
	CheckoutBranch(ctx context.Context, branch string) error
	// Branch 获取分支
	Branch(ctx context.Context, branch string) (Branch, error)
	// CreateBranch 根据hash创建分支，若不指定hash，则为当前head hash
	CreateBranch(ctx context.Context, branch string, hash string) (Branch, error)
	// DeleteLocalBranch 删除本地分支
	DeleteLocalBranch(ctx context.Context, branch string) error
	// DeleteBranch 删除本地和远程分支
	DeleteBranch(ctx context.Context, branch string) error

	// CheckoutTag checkout标签，若tag为空，则checkout master分支
	CheckoutTag(ctx context.Context, tag string) error
	// Tag 获取标签
	Tag(ctx context.Context, tagName string) (Tag, error)
	// CreateTag 根据hash创建标签，若不指定hash，则为当前head hash
	CreateTag(ctx context.Context, tag, comment, hash string) (Tag, error)
	// DeleteLocalTag 删除本地标签
	DeleteLocalTag(ctx context.Context, tag string) error
	// DeleteTag 删除本地和远程标签
	DeleteTag(ctx context.Context, tag string) error

	// Fetch git fetch
	Fetch(ctx context.Context) error
}

type Cloner interface {
	// Clone 克隆指定的url的Repository到本地dir目录，若dir为空，则为临时目录（临时目录可以通过Repository.Root()获取）
	Clone(ctx context.Context, url, dir string) (Repository, error)
	// CloneToMemory 克隆指定的url的Repository到缓存中
	CloneToMemory(ctx context.Context, url string) (Repository, error)
	// ConfigInterface visitor + ApplyOption interface for Config
	ConfigInterface
}
