package executor

import (
	"github.com/talatmursalin/ekshunno-executor/models"
)

type Executor interface {
	Compile() models.Result
	Execute(io string) models.Result
}
