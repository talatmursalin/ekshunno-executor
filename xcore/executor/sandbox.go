package executor

import (
	"github.com/talatmursalin/ekshunno-executor/xcore/compilers"
	"github.com/talatmursalin/ekshunno-executor/xcore/utils"
)

type SandboxExecutor struct {
	compilerSettings compilers.Compiler
	limits           utils.Limit
	src              string
}

func (sb SandboxExecutor) Compile() utils.Result {

	return utils.Result{
		Verdict: utils.OK,
		Time:    0.5,
		Memory:  232.07,
		Output:  "hello aorld",
	}
}

func (sb SandboxExecutor) Execute(io string) utils.Result {
	return utils.Result{
		Verdict: utils.OK,
		Time:    0.5,
		Memory:  232.07,
		Output:  "hello aorld",
	}
}

func NewSandboxExecutor(src string, sett compilers.Compiler, limits utils.Limit) *SandboxExecutor {
	return &SandboxExecutor{
		compilerSettings: sett,
		limits:           limits,
		src:              src,
	}
}
