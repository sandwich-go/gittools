package gittools

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestGit(t *testing.T) {
	Convey("git ", t, func() {
		userName := "botman"
		g := Default(WithUserName(userName))
		var r Repository
		var err error

		var a = gitignore.ParsePattern("*.pyc", nil)
		fmt.Println(a.Match([]string{"gen/meta/migration/__pycache__/aaaaa.pyc"}, false))
		return
		r, err = g.CloneToMemory(context.Background(), "git@github.com:sandwich-go/redisson.git")
		So(err, ShouldBeNil)
		So(r, ShouldNotBeNil)
		So(r.UserName(), ShouldEqual, userName)

		err = r.CheckoutBranch(context.Background(), "")
		So(err, ShouldBeNil)
		err = r.Fetch(context.Background())
		So(err, ShouldBeNil)
		err = r.CheckoutBranch(context.Background(), "version/1.0")
		So(err, ShouldBeNil)
		err = r.CheckoutBranch(context.Background(), "version/1.1")
		So(err, ShouldBeNil)

		err = r.CheckoutTag(context.Background(), "")
		So(err, ShouldBeNil)
		err = r.CheckoutTag(context.Background(), "v1.0.11")
		So(err, ShouldBeNil)
		err = r.CheckoutTag(context.Background(), "v1.1.16")
		So(err, ShouldBeNil)

		var b Branch
		newBranchName := "new_branch"
		fileName := "a.txt"
		fileContent := []byte(fmt.Sprintf("%s write it!", time.Now().String()))
		commitMsg := "commit by test git"
		g = Default()
		r, err = g.Clone(context.Background(), "git@github.com:sandwich-go/go-redis-client-benchmark.git", "")
		So(err, ShouldBeNil)
		So(r, ShouldNotBeNil)
		err = r.Fetch(context.Background())
		So(err, ShouldBeNil)
		err = r.Pull(context.Background())
		So(err, ShouldBeNil)
		err = r.Push(context.Background())
		So(err, ShouldBeNil)
		err = r.RewriteFile(context.Background(), fileName, fileContent)
		So(err, ShouldBeNil)
		err = r.Commit(context.Background(), commitMsg)
		So(err, ShouldBeNil)
		err = r.Push(context.Background())
		So(err, ShouldBeNil)
		b, err = r.CreateBranch(context.Background(), newBranchName, "")
		So(err, ShouldBeNil)
		So(b, ShouldNotBeNil)
		err = b.Push(context.Background())
		So(err, ShouldBeNil)
		b, err = r.Branch(context.Background(), newBranchName)
		So(err, ShouldBeNil)
		So(b, ShouldNotBeNil)
		err = b.Delete(context.Background())
		So(err, ShouldBeNil)
		err = b.Push(context.Background())
		So(err, ShouldBeNil)

		var tg Tag
		newTagName := "new_tag"
		newTagComment := "this is comment"
		tg, err = r.CreateTag(context.Background(), newTagName, newTagComment, "")
		So(err, ShouldBeNil)
		So(tg, ShouldNotBeNil)
		err = tg.Push(context.Background())
		So(err, ShouldBeNil)
		tg, err = r.Tag(context.Background(), newTagName)
		So(err, ShouldBeNil)
		So(tg, ShouldNotBeNil)
		err = tg.Delete(context.Background())
		So(err, ShouldBeNil)
		err = tg.Push(context.Background())
		So(err, ShouldBeNil)

		cloneDir := r.Root()
		t.Log("clone to dir:", cloneDir)
		err = r.RemoveAll()
		So(err, ShouldBeNil)
	})
}
