package executor

import (
	"log"

	"github.com/talatmursalin/ekshunno-executor/xcore/compilers"
	"github.com/talatmursalin/ekshunno-executor/xcore/utils"
)

func GetExecutor(langId utils.LangEnum, src string, limits utils.Limit) Executor {
	var compilerSettings compilers.Compiler
	switch langId {
	case utils.C11:
		compilerSettings = compilers.C11CompilerSettings{}
	case utils.CPP11:
		compilerSettings = compilers.Cpp11CompilerSettings{}
	case utils.CPP14:
		compilerSettings = compilers.Cpp14CompilerSettings{}
	case utils.CPP17:
		compilerSettings = compilers.Cpp17CompilerSettings{}
	case utils.JAVA11:
		compilerSettings = compilers.Java11CompilerSettings{}
	case utils.KOTLIN1:
		compilerSettings = compilers.Kotlin1CompilerSettings{}
	case utils.GO1:
		compilerSettings = compilers.Go1CompilerSettings{}
	case utils.RUST1:
		compilerSettings = compilers.Rust1CompilerSettings{}
	case utils.PYTHON3:
		compilerSettings = compilers.Python3CompilerSettings{}
	default:
		log.Panicf("Executor not found %s", langId)
	}
	return NewSandboxExecutor(src, compilerSettings, limits)
}
