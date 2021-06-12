package utils

type LangEnum string

const (
	C11     LangEnum = "c11"
	CPP11   LangEnum = "cpp11"
	CPP14   LangEnum = "cpp14"
	CPP17   LangEnum = "cpp17"
	JAVA11  LangEnum = "java11"
	KOTLIN1 LangEnum = "kotlin1"
	PYTHON3 LangEnum = "python3"
	GO1     LangEnum = "go1"
	RUST1   LangEnum = "rust1"
)

func StringToLangId(langId string) LangEnum {
	switch langId {
	case "c11":
		return C11
	case "cpp11":
		return CPP11
	case "cpp14":
		return CPP14
	case "cpp17":
		return CPP17
	case "java11":
		return JAVA11
	case "python3":
		return PYTHON3
	case "go1":
		return GO1
	case "rust1":
		return RUST1
	case "kotlin1":
		return KOTLIN1
	default:
		panic("could not find a langId match")
	}
}
