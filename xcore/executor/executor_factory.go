package executor

import (
	"log"

	"github.com/talatmursalin/ekshunno-executor/customenums"
	"github.com/talatmursalin/ekshunno-executor/models"
	"github.com/talatmursalin/ekshunno-executor/xcore/compilers"
)

func GetExecutor(langId customenums.LangEnum, src string, limits models.Limit) Executor {
	var compilerSettings compilers.Compiler
	switch langId {
	case customenums.C11:
		compilerSettings = compilers.C11CompilerSettings{}
	case customenums.CPP11:
		compilerSettings = compilers.Cpp11CompilerSettings{}
	case customenums.CPP14:
		compilerSettings = compilers.Cpp14CompilerSettings{}
	case customenums.CPP17:
		compilerSettings = compilers.Cpp17CompilerSettings{}
	case customenums.JAVA11:
		compilerSettings = compilers.Java11CompilerSettings{}
	case customenums.KOTLIN1:
		compilerSettings = compilers.Kotlin1CompilerSettings{}
	case customenums.GO1:
		compilerSettings = compilers.Go1CompilerSettings{}
	case customenums.RUST1:
		compilerSettings = compilers.Rust1CompilerSettings{}
	case customenums.PYTHON3:
		compilerSettings = compilers.Python3CompilerSettings{}
	default:
		log.Panicf("Executor not found %s", langId)
	}
	return NewSandboxExecutor(src, compilerSettings, limits)
}
