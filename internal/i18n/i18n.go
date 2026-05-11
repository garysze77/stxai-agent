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
	cache   = map[string]map[string]string{}
	cacheMu sync.RWMutex
)

const defaultLang = "en"

func T(key, lang string, args ...interface{}) string {
	if lang != "en" && lang != "zh-HK" && lang != "zh-CN" {
		lang = defaultLang
	}

	cacheMu.RLock()
	trans, ok := cache[lang]
	cacheMu.RUnlock()

	if !ok {
		trans = load(lang)
		cacheMu.Lock()
		cache[lang] = trans
		cacheMu.Unlock()
	}

	text, ok := trans[key]
	if !ok && lang != defaultLang {
		enTrans := load(defaultLang)
		cacheMu.Lock()
		cache[defaultLang] = enTrans
		cacheMu.Unlock()
		text = enTrans[key]
	}
	if text == "" {
		return key
	}
	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}
	return text
}

func load(lang string) map[string]string {
	data, err := translationsFS.ReadFile("translations/" + lang + ".json")
	if err != nil {
		return map[string]string{}
	}
	var nested map[string]interface{}
	if err := json.Unmarshal(data, &nested); err != nil {
		return map[string]string{}
	}
	return flatten(nested)
}

func flatten(nested map[string]interface{}) map[string]string {
	out := map[string]string{}
	flattenInto("", nested, out)
	return out
}

func flattenInto(prefix string, nested map[string]interface{}, out map[string]string) {
	for k, v := range nested {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			out[key] = val
		case map[string]interface{}:
			flattenInto(key, val, out)
		}
	}
}

func Supported(lang string) bool {
	return lang == "en" || lang == "zh-HK" || lang == "zh-CN"
}
