package utils

import "errors"

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

func StringToLangId(langId string) (LangEnum, error) {
	switch langId {
	case "c11":
		return C11, nil
	case "cpp11":
		return CPP11, nil
	case "cpp14":
		return CPP14, nil
	case "cpp17":
		return CPP17, nil
	case "java11":
		return JAVA11, nil
	case "python3":
		return PYTHON3, nil
	case "go1":
		return GO1, nil
	case "rust1":
		return RUST1, nil
	case "kotlin1":
		return KOTLIN1, nil
	default:
		return C11, errors.New("invalid language id")
	}
}
