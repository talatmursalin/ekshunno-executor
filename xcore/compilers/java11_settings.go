package compilers

import "fmt"

type Java11CompilerSettings struct {
}

func (s Java11CompilerSettings) GetImageName() string {
	return "openjdk:11"
}

func (s Java11CompilerSettings) GetSourceName() string {
	return "Main.java"
}

func (s Java11CompilerSettings) GetCompileCommand(src_dir, out_dir string) string {
	return fmt.Sprintf("javac -encoding UTF-8 -d %s %s/Main.java", out_dir, src_dir)
}

func (s Java11CompilerSettings) GetExecuteCommand(out_dir string) string {
	return fmt.Sprintf("java -cp %s Main", out_dir)
}

func (s Java11CompilerSettings) IsInterpreter() bool {
	return false
}
