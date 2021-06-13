package compilers

import "fmt"

type Cpp17CompilerSettings struct {
}

func (s Cpp17CompilerSettings) GetImageName() string {
	return "gcc:5"
}

func (s Cpp17CompilerSettings) GetSourceName() string {
	return "a.cpp"
}

func (s Cpp17CompilerSettings) GetCompileCommand(src_dir, out_dir string) string {
	return fmt.Sprintf("g++ -O2 -std=gnu++17 -static %s/a.cpp -o %s/a.out", src_dir, out_dir)
}

func (s Cpp17CompilerSettings) GetExecuteCommand(out_dir string) string {
	return fmt.Sprintf("%s/a.out", out_dir)
}

func (s Cpp17CompilerSettings) IsInterpreter() bool {
	return false
}
