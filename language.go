package polygon

import "fmt"

func LocaleFromLanguage(lang string) (string, error) {
	switch lang {
	case "ukrainian", "russian", "english", "hungarian", "azerbaijani", "french", "arabic", "uzbek", "slovene":
		return lang[:2], nil
	case "armenian":
		return "hy", nil
	case "lithuanian":
		return "lt", nil
	case "serbian":
		return "sr", nil
	case "kazakh":
		return "kk", nil
	case "spanish":
		return "es", nil
	case "polish":
		return "pl", nil
	case "german":
		return "de", nil
	case "turkish":
		return "tr", nil
	default:
		return lang, fmt.Errorf("unknown language %#v", lang)
	}
}
