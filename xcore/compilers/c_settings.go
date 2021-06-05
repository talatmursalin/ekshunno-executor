package compilers

import "fmt"

type CCompilerSettings struct {
}

func (s CCompilerSettings) GetImageName() string {
	return "gcc:5"
}

func (s CCompilerSettings) GetSourceName() string {
	return "a.c"
}

func (s CCompilerSettings) GetCompileCommand(src_dir, out_dir string) string {
	return fmt.Sprintf("gcc -O2 -std=gnu11 -static %s/a.c -lm -o %s/a.out", src_dir, out_dir)
}

func (s CCompilerSettings) GetExecuteCommand(out_dir string) string {
	return fmt.Sprintf("%s/a.out", out_dir)
}

func (s CCompilerSettings) IsInterpreter() bool {
	return false
}

func NewCCompilerSettings() *CCompilerSettings {
	return &CCompilerSettings{}
}
