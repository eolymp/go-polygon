package polygon

import "fmt"

func LocaleFromLanguage(lang string) (string, error) {
	switch lang {
	case "ukrainian", "russian", "english", "hungarian", "azerbaijani", "french", "arabic", "uzbek":
		return lang[:2], nil
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
