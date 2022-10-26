package gittools

import (
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestGit(t *testing.T) {
	Convey("git ", t, func() {
		userName := "botman"
		g := Default(WithUserName(userName))
		r, err := g.CloneToMemory(context.Background(), "git@github.com:sandwich-go/redisson.git")
		So(err, ShouldBeNil)
		So(r, ShouldNotBeNil)
		So(r.UserName(), ShouldEqual, userName)

		err = r.CheckoutBranch(context.Background(), "")
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

		fileName := "a.txt"
		fileContent := []byte(fmt.Sprintf("%s write it!", time.Now().String()))
		commitMsg := "commit by test git"
		g = Default()
		r, err = g.Clone(context.Background(), "git@github.com:sandwich-go/go-redis-client-benchmark.git", "")
		So(err, ShouldBeNil)
		So(r, ShouldNotBeNil)
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
		cloneDir := r.Root()
		t.Log("clone to dir:", cloneDir)
		err = r.RemoveAll()
		So(err, ShouldBeNil)
	})
}
