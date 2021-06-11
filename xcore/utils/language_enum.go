package utils

type LangEnum string

const (
	C      LangEnum = "c"
	CPP    LangEnum = "cpp"
	JAVA   LangEnum = "java"
	PYTHON LangEnum = "python"
	GO     LangEnum = "go"
	RUST   LangEnum = "rust"
)

func StringToLangId(langId string) LangEnum {
	switch langId {
	case "c":
		return C
	case "cpp":
		return CPP
	case "java":
		return JAVA
	case "python":
		return PYTHON
	case "go":
		return GO
	case "rust":
		return RUST
	default:
		panic("could not find a langId match")
	}
}
