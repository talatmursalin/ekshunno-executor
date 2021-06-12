package compilers

import "fmt"

type Go1CompilerSettings struct {
}

func (s Go1CompilerSettings) GetImageName() string {
	return "golang:1.12"
}

func (s Go1CompilerSettings) GetSourceName() string {
	return "main.go"
}

func (s Go1CompilerSettings) GetCompileCommand(src_dir, out_dir string) string {
	return fmt.Sprintf("go build -o %s/main %s/main.go", out_dir, src_dir)
}

func (s Go1CompilerSettings) GetExecuteCommand(out_dir string) string {
	return fmt.Sprintf("%s/main", out_dir)
}

func (s Go1CompilerSettings) IsInterpreter() bool {
	return false
}
