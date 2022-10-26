package gittools

import (
	"io"
	"log"
	"os"
)

type Logger interface {
	Println(v ...interface{})
	Fatalln(v ...interface{})
	Writer() io.Writer
}

//go:generate optiongen --option_with_struct_name=false --new_func=NewConfig --xconf=true --empty_composite_nil=true --usage_tag_name=usage
func ConfigOptionDeclareWithDefault() interface{} {
	return map[string]interface{}{
		"RsaPath":   ".ssh/id_rsa",                                 // @MethodComment(rsa 绝对路径或者home目录下相对路径)
		"Logger":    Logger(log.New(os.Stdout, "", log.LstdFlags)), // @MethodComment(日志输出)
		"UserName":  "",                                            // @MethodComment(config user.name)
		"UserEmail": "",                                            // @MethodComment(config user.email)
		"Depth":     1,                                             // @MethodComment(git depth)
	}
}
