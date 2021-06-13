package compilers

import "fmt"

type Rust1CompilerSettings struct {
}

func (s Rust1CompilerSettings) GetImageName() string {
	return "rust:1.43"
}

func (s Rust1CompilerSettings) GetSourceName() string {
	return "main.rs"
}

func (s Rust1CompilerSettings) GetCompileCommand(src_dir, out_dir string) string {
	return fmt.Sprintf("rustc %s/main.rs -o %s/main", src_dir, out_dir)
}

func (s Rust1CompilerSettings) GetExecuteCommand(out_dir string) string {
	return fmt.Sprintf("%s/main", out_dir)
}

func (s Rust1CompilerSettings) IsInterpreter() bool {
	return false
}
