package compilers

type Compiler interface {
	GetImageName() string
	GetSourceName() string
	GetCompileCommand(src_dir, out_dir string) string
	GetExecuteCommand(out_dir string) string
	IsInterpreter() bool
}
