package compilers

import "fmt"

type Python3CompilerSettings struct {
}

func (s Python3CompilerSettings) GetImageName() string {
	return "python:3.6"
}

func (s Python3CompilerSettings) GetSourceName() string {
	return "__init__.py"
}

func (s Python3CompilerSettings) GetCompileCommand(src_dir, out_dir string) string {
	return fmt.Sprintf("python -m compileall -q %s/__init__.py", src_dir)
}

func (s Python3CompilerSettings) GetExecuteCommand(out_dir string) string {
	return fmt.Sprintf("python %s/__init__.py", out_dir)
}

func (s Python3CompilerSettings) IsInterpreter() bool {
	return true
}
