# gittools

git工具

## 例子

```golang
package main

import (
    "context"
    "fmt"
    "github.com/sandwich-go/gittools"
    "time"
)

func initGitTools()  {
    userName  := "batman"
    userEmail := "batman@123.com"
    // 初始化
    gittools.Default(gittools.WithUserName(userName), gittools.WithUserEmail(userEmail))
}

func main() {
    initGitTools()
	
    // 由于初始化了Default,因此可以直接使用Default()函数
    url := "git@github.com:sandwich-go/gittools.git"
    dir := "" // 目录为空，则会clone到临时目录中
    repo, err := gittools.Default().Clone(context.Background(), url, dir)
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
	
    // modify file
    fileName := "modify.txt"
    fileContent := []byte(fmt.Sprintf("%s write it!", time.Now().String()))
    err = repo.RewriteFile(context.Background(), fileName, fileContent)
    if err != nil {
        fmt.Println(err)
        return
    }

    // commit and push  
    commitMsg := "commit by git tools"
    err = repo.Commit(context.Background(), commitMsg)
    if err != nil {
        fmt.Println(err)
        return
    }
    err = repo.Push(context.Background())
    if err != nil {
        fmt.Println(err)
        return
    }
	
    fmt.Println("ok")
}
```