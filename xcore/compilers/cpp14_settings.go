package compilers

import "fmt"

type Cpp14CompilerSettings struct {
}

func (s Cpp14CompilerSettings) GetImageName() string {
	return "gcc:5"
}

func (s Cpp14CompilerSettings) GetSourceName() string {
	return "a.cpp"
}

func (s Cpp14CompilerSettings) GetCompileCommand(src_dir, out_dir string) string {
	return fmt.Sprintf("g++ -O2 -std=gnu++14 -static %s/a.cpp -o %s/a.out", src_dir, out_dir)
}

func (s Cpp14CompilerSettings) GetExecuteCommand(out_dir string) string {
	return fmt.Sprintf("%s/a.out", out_dir)
}

func (s Cpp14CompilerSettings) IsInterpreter() bool {
	return false
}
