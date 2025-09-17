package config

import (
	"multitrago/internal/utils"
	"net/http"
)

type ServiceConfig = utils.ServiceConfig

func GetAvailableServices() []ServiceConfig {
	return []ServiceConfig{
		{
			Name:   "GOOGLE",
			URL:    "https://translate-serverless.vercel.app/api/translate",
			Method: http.MethodPost,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(`{"q":"hello","source":"en","target":"de"}`),
		},
		{
			Name:   "DEEPL",
			URL:    "https://deeplx-vercel-phi.vercel.app/api/translate",
			Method: http.MethodPost,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(`{"text":"hello","source_lang":"EN","target_lang":"DE"}`),
		},
		{
			Name:   "REVERSO",
			URL:    "https://api.reverso.net/translate/v1/translation",
			Method: http.MethodPost,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(`{"format":"text","from":"eng","to":"ger","input":"hello"}`),
		},
		{
			Name:   "REVERSO2",
			URL:    "https://api.reverso.net/translate/v1/translation",
			Method: http.MethodPost,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(`{"format":"text","from":"eng","to":"ger","input":"hello"}`),
		},
		{
			Name:   "LINGVA",
			URL:    "https://lingva.ml/api/v1/{source}/{target}/{text}",
			Method: http.MethodGet,
		},
		{
			Name:   "MYMEMORY",
			URL:    "https://api.mymemory.translated.net/get?q=hello&langpair=en|de",
			Method: http.MethodGet,
		},
		{
			Name:   "OPENAI",
			URL:    "https://api.openai.com/v1/chat/completions",
			Method: http.MethodPost,
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer YOUR_OPENAI_KEY",
			},
			Body: []byte(`{
				"model": "gpt-3.5-turbo",
				"messages": [{"role":"user","content":"Translate hello to German"}]
			}`),
		},
	}
}

func GetSupportedLanguages() []string {
	return []string{"ru", "en", "de", "fr", "es", "it", "ja", "zh", "ko", "ar"}
}

func GetLanguageNames() map[string]string {
	return map[string]string{
		"ru": "Русский",
		"en": "English",
		"de": "Deutsch",
		"fr": "Français",
		"es": "Español",
		"it": "Italiano",
		"ja": "日本語",
		"zh": "中文",
		"ko": "한국어",
		"ar": "العربية",
	}
}
