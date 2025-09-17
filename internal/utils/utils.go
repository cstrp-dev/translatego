package utils

import (
	"regexp"
	"unicode"
)

func DetectFromLanguage(text string) string {
	nonLangPattern := regexp.MustCompile(`[\s\n\r.,;:!?()\-\"']`)
	cleanText := nonLangPattern.ReplaceAllString(text, "")

	if len(cleanText) == 0 {
		return "auto"
	}

	var latinCount, cyrillicCount, otherCount int

	for _, r := range cleanText {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			latinCount++
		} else if (r >= 'а' && r <= 'я') || (r >= 'А' && r <= 'Я') || r == 'ё' || r == 'Ё' {
			cyrillicCount++
		} else if unicode.IsLetter(r) {
			otherCount++
		}
	}

	total := latinCount + cyrillicCount + otherCount
	if total == 0 {
		return "auto"
	}

	if cyrillicCount > latinCount && cyrillicCount > otherCount {
		return "ru"
	} else if latinCount > cyrillicCount && latinCount > otherCount {
		return "en"
	} else if otherCount > latinCount && otherCount > cyrillicCount {
		return "auto"
	} else {
		return "auto"
	}
}

func DetectToLanguage(lang, selectedLanguage string) string {
	switch lang {
	case "en":
		return selectedLanguage
	case selectedLanguage:
		return "en"
	case "auto":
		return selectedLanguage
	default:
		return selectedLanguage
	}
}

func GetDetailedErrorMessage(serviceName string, err error) string {
	if err == nil {
		return "Unknown error"
	}

	errMsg := err.Error()

	switch serviceName {
	case "GOOGLE":
		if errMsg == "status 429" {
			return "❌ Google: Rate limit exceeded. Please try again later."
		} else if errMsg == "status 403" {
			return "❌ Google: Access forbidden. Service may be unavailable."
		} else if errMsg == "request timeout after 10 seconds" {
			return "❌ Google: Request timeout. Service may be slow or unavailable."
		}
		return "❌ Google: " + errMsg
	case "DEEPL":
		if errMsg == "status 429" {
			return "❌ DeepL: Rate limit exceeded. Please try again later."
		} else if errMsg == "status 403" {
			return "❌ DeepL: Access forbidden. API key may be invalid."
		} else if errMsg == "request timeout after 10 seconds" {
			return "❌ DeepL: Request timeout. Service may be slow or unavailable."
		}
		return "❌ DeepL: " + errMsg
	case "REVERSO":
		if errMsg == "status 429" {
			return "❌ Reverso: Rate limit exceeded. Please try again later."
		} else if errMsg == "status 403" {
			return "❌ Reverso: Access forbidden. Service may be unavailable."
		} else if errMsg == "request timeout after 10 seconds" {
			return "❌ Reverso: Request timeout. Service may be slow or unavailable."
		}
		return "❌ Reverso: " + errMsg
	case "MYMEMORY":
		if errMsg == "status 429" {
			return "❌ MyMemory: Rate limit exceeded. Please try again later."
		} else if errMsg == "status 403" {
			return "❌ MyMemory: Access forbidden. Service may be unavailable."
		} else if errMsg == "request timeout after 10 seconds" {
			return "❌ MyMemory: Request timeout. Service may be slow or unavailable."
		}
		return "❌ MyMemory: " + errMsg
	case "LINGVA":
		if errMsg == "status 429" {
			return "❌ Lingva: Rate limit exceeded. Please try again later."
		} else if errMsg == "status 403" {
			return "❌ Lingva: Access forbidden. Service may be unavailable."
		} else if errMsg == "request timeout after 10 seconds" {
			return "❌ Lingva: Request timeout. Service may be slow or unavailable."
		}
		return "❌ Lingva: " + errMsg
	case "OPENAI":
		if errMsg == "status 401" {
			return "❌ OpenAI: Invalid API key. Please check your configuration."
		} else if errMsg == "status 429" {
			return "❌ OpenAI: Rate limit exceeded. Please try again later."
		} else if errMsg == "status 403" {
			return "❌ OpenAI: Access forbidden. API key may be invalid."
		} else if errMsg == "request timeout after 10 seconds" {
			return "❌ OpenAI: Request timeout. Service may be slow or unavailable."
		}
		return "❌ OpenAI: " + errMsg
	default:
		if errMsg == "request timeout after 10 seconds" {
			return "❌ " + serviceName + ": Request timeout. Service may be slow or unavailable."
		}
		return "❌ " + serviceName + ": " + errMsg
	}
}
