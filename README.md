# gittools

git工具

## 例子

```golang
package main

import (
	"context"
	"fmt"
	"github.com/sandwich-go/gittools"
)

func main() {
	userName := "batman"
	userEmail := "batman@123.com"
	// 初始化
	gittools.Default(gittools.WithUserName(userName), gittools.WithUserEmail(userEmail))

	// 由于初始化了Default,因此可以直接使用Default()函数
	url := "git@github.com:sandwich-go/gittools.git"
	repo, err := gittools.Default().CloneToMemory(context.Background(), url)
	if err != nil {
		fmt.Println(err)
		return
    }
	// checkout master
	err = repo.CheckoutBranch(context.Background(), "")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ok")
}
```