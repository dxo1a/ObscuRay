package backend

import "os"

var Translations = map[string]map[string]string{
	"en": {
		"tooltip":  "ObscuRay VLESS Proxy",
		"show":     "Show",
		"showDesc": "Show application",
		"quit":     "Quit",
		"quitDesc": "Quit application",
	},
	"ru": {
		"tooltip":  "ObscuRay VLESS прокси",
		"show":     "Показать",
		"showDesc": "Показать приложение",
		"quit":     "Выйти",
		"quitDesc": "Выйти из приложения",
	},
}

func GetLang() string {
	lang := os.Getenv("LANG")
	if len(lang) >= 2 {
		if lang[:2] == "ru" {
			return "ru"
		}
	}
	return "en"
}
