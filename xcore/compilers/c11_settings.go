package compilers

import "fmt"

type C11CompilerSettings struct {
}

func (s C11CompilerSettings) GetImageName() string {
	return "gcc:5"
}

func (s C11CompilerSettings) GetSourceName() string {
	return "a.c"
}

func (s C11CompilerSettings) GetCompileCommand(src_dir, out_dir string) string {
	return fmt.Sprintf("gcc -O2 -std=gnu11 -static %s/a.c -lm -o %s/a.out", src_dir, out_dir)
}

func (s C11CompilerSettings) GetExecuteCommand(out_dir string) string {
	return fmt.Sprintf("%s/a.out", out_dir)
}

func (s C11CompilerSettings) IsInterpreter() bool {
	return false
}
