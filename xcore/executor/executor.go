package executor

import (
	"github.com/talatmursalin/ekshunno-executor/xcore/utils"
)

type Executor interface {
	Compile() utils.Result
	Execute(io string) utils.Result
}
