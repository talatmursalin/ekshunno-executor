package models

import (
	"github.com/talatmursalin/ekshunno-executor/customenums"
)

type Result struct {
	Verdict customenums.VerdictEnum
	Time    float32
	Memory  float32
	Output  string
}
