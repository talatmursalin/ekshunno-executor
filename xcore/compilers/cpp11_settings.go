package compilers

import "fmt"

type Cpp11CompilerSettings struct {
}

func (s Cpp11CompilerSettings) GetImageName() string {
	return "gcc:5"
}

func (s Cpp11CompilerSettings) GetSourceName() string {
	return "a.cpp"
}

func (s Cpp11CompilerSettings) GetCompileCommand(src_dir, out_dir string) string {
	return fmt.Sprintf("g++ -O2 -std=gnu++11 -static %s/a.cpp -o %s/a.out", src_dir, out_dir)
}

func (s Cpp11CompilerSettings) GetExecuteCommand(out_dir string) string {
	return fmt.Sprintf("%s/a.out", out_dir)
}

func (s Cpp11CompilerSettings) IsInterpreter() bool {
	return false
}
