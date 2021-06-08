package main

import (
	"io/ioutil"

	"github.com/talatmursalin/ekshunno-executor/xcore/compilers"
	"github.com/talatmursalin/ekshunno-executor/xcore/executor"
	"github.com/talatmursalin/ekshunno-executor/xcore/utils"
)

func main() {
	dat, err := ioutil.ReadFile("src.c")
	if err != nil {
		panic(err)
	}
	src := string(dat)
	// fmt.Println(src)
	sb := executor.NewSandboxExecutor(
		src,
		*compilers.NewCCompilerSettings(),
		*utils.NewLimit(2, 256, 100),
	)
	sb.Compile()
	sb.Execute("")
	sb.Clear()
}
