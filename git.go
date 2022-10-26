package gittools

import (
	"context"
)

type Repository interface {
	// UserName 获取.git/config的user.name
	UserName() string
	// UserEmail 获取.git/config的user.email
	UserEmail() string

	// Root 若Repository是克隆在本地，显示克隆的根目录
	Root() string
	// RemoveAll 若Repository是克隆在本地，删除克隆的根目录
	RemoveAll() error

	// CheckoutBranch checkout分支，若branch为空，则checkout master分支
	CheckoutBranch(ctx context.Context, branch string) error
	// CheckoutTag checkout标签，若tag为空，则checkout master分支
	CheckoutTag(ctx context.Context, tag string) error

	// Pull git pull
	Pull(ctx context.Context) error
	// Push git push
	Push(ctx context.Context) error
	// Commit git commit -m ""
	Commit(ctx context.Context, msg string) error

	// RewriteFile 重写文件内容，不存在则创建
	RewriteFile(ctx context.Context, file string, data []byte) error
}

type Cloner interface {
	// Clone 克隆指定的url的Repository到本地dir目录，若dir为空，则为临时目录（临时目录可以通过Repository.Root()获取）
	Clone(ctx context.Context, url, dir string) (Repository, error)
	// CloneToMemory 克隆指定的url的Repository到缓存中
	CloneToMemory(ctx context.Context, url string) (Repository, error)
}
