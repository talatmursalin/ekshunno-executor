package compilers

import "fmt"

type Kotlin1CompilerSettings struct {
}

func (s Kotlin1CompilerSettings) GetImageName() string {
	return "zenika/kotlin"
}

func (s Kotlin1CompilerSettings) GetSourceName() string {
	return "Main.kt"
}

func (s Kotlin1CompilerSettings) GetCompileCommand(src_dir, out_dir string) string {
	return fmt.Sprintf("kotlinc -d %s %s/Main.kt", out_dir, src_dir)
}

func (s Kotlin1CompilerSettings) GetExecuteCommand(out_dir string) string {
	return fmt.Sprintf("kotlin -Dfile.encoding=UTF-8 -J-XX:+UseSerialGC -J-Xss64m -J-Xms1920m -J-Xmx1920m -cp %s MainKt", out_dir)
}

func (s Kotlin1CompilerSettings) IsInterpreter() bool {
	return false
}
