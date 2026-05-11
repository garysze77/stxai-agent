package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed translations/*.json
var translationsFS embed.FS

var (
	translations   = make(map[string]map[string]string)
	loadOnce       sync.Once
	defaultLang    = "en"
	validLangs     = map[string]bool{"en": true, "zh-HK": true, "zh-CN": true}
)

func load() {
	loadOnce.Do(func() {
		for _, lang := range []string{"en", "zh-HK", "zh-CN"} {
			data, err := translationsFS.ReadFile("translations/" + lang + ".json")
			if err != nil {
				continue
			}
			flat := make(map[string]string)
			var raw map[string]interface{}
			if json.Unmarshal(data, &raw) == nil {
				flatten(raw, "", flat)
			}
			translations[lang] = flat
		}
	})
}

func flatten(obj map[string]interface{}, prefix string, out map[string]string) {
	for k, v := range obj {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			out[key] = val
		case map[string]interface{}:
			flatten(val, key, out)
		}
	}
}

// T returns the translation for key in the given language.
// Falls back to English if the key is missing.
func T(key, lang string, args ...interface{}) string {
	load()

	if !validLangs[lang] {
		lang = defaultLang
	}

	text := ""
	if m, ok := translations[lang]; ok {
		text = m[key]
	}
	if text == "" && lang != defaultLang {
		if m, ok := translations[defaultLang]; ok {
			text = m[key]
		}
	}
	if text == "" {
		text = key
	}

	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}
	return text
}

// ValidLang returns whether the given language code is supported.
func ValidLang(lang string) bool {
	return validLangs[lang]
}
