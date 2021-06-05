package main

import (
	"github.com/talatmursalin/ekshunno-executor/xcore/compilers"
)

func getImage(s compilers.CCompilerSettings) string {
	return s.GetImageName()
}

func main() {
	// var c compilers.Compiler
	c := compilers.NewCCompilerSettings()
	println(getImage(*c))

}
