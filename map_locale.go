package polygon

import "fmt"

func mapLanguageToLocale(lang string) (string, error) {
	switch lang {
	case "ukrainian", "russian", "english", "hungarian":
		return lang[:2], nil
	case "kazakh":
		return "kk", nil
	default:
		return lang, fmt.Errorf("unknown language %#v", lang)
	}
}
