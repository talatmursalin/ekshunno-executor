package executor

import (
	"github.com/talatmursalin/ekshunno-executor/xcore/compilers"
	"github.com/talatmursalin/ekshunno-executor/xcore/utils"
)

func GetExecutor(langId utils.LangEnum, src string, limits utils.Limit) Executor {
	var compilerSettings compilers.Compiler
	switch langId {
	case utils.C:
		compilerSettings = compilers.NewCCompilerSettings()
	default:
		println("raise error")
	}
	return NewSandboxExecutor(src, compilerSettings, limits)
}
