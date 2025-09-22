package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type ServiceError struct {
	Service     string
	StatusCode  int
	ErrorType   string
	Message     string
	Suggestion  string
	IsRetryable bool
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Service, e.Message)
}

const (
	ErrorTypeTimeout       = "TIMEOUT"
	ErrorTypeRateLimit     = "RATE_LIMIT"
	ErrorTypeUnauthorized  = "UNAUTHORIZED"
	ErrorTypeForbidden     = "FORBIDDEN"
	ErrorTypeNotFound      = "NOT_FOUND"
	ErrorTypeServerError   = "SERVER_ERROR"
	ErrorTypeServiceDown   = "SERVICE_DOWN"
	ErrorTypeNetworkError  = "NETWORK_ERROR"
	ErrorTypeLanguageError = "LANGUAGE_ERROR"
	ErrorTypeUnknown       = "UNKNOWN"
)

func CreateServiceError(serviceName string, err error, statusCode int) *ServiceError {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())
	serviceErr := &ServiceError{
		Service:    serviceName,
		StatusCode: statusCode,
	}

	switch {
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"):
		serviceErr.ErrorType = ErrorTypeTimeout
		serviceErr.Message = "Request timed out"
		serviceErr.Suggestion = "Check your internet connection or try again"
		serviceErr.IsRetryable = true

	case statusCode == 429 || strings.Contains(errStr, "rate limit"):
		serviceErr.ErrorType = ErrorTypeRateLimit
		serviceErr.Message = "Rate limit exceeded"
		serviceErr.Suggestion = "Wait a moment before trying again"
		serviceErr.IsRetryable = true

	case statusCode == 401 || strings.Contains(errStr, "unauthorized"):
		serviceErr.ErrorType = ErrorTypeUnauthorized
		serviceErr.Message = "Invalid or missing API key"
		serviceErr.Suggestion = "Check your API key configuration"
		serviceErr.IsRetryable = false

	case statusCode == 403 || strings.Contains(errStr, "forbidden"):
		serviceErr.ErrorType = ErrorTypeForbidden
		serviceErr.Message = "Access forbidden"
		serviceErr.Suggestion = "API key may be invalid or service unavailable"
		serviceErr.IsRetryable = false

	case statusCode == 404:
		serviceErr.ErrorType = ErrorTypeNotFound
		serviceErr.Message = "Service endpoint not found"
		serviceErr.Suggestion = "Service may be temporarily unavailable"
		serviceErr.IsRetryable = true

	case statusCode >= 500 && statusCode < 600:
		serviceErr.ErrorType = ErrorTypeServerError
		serviceErr.Message = fmt.Sprintf("Server error (HTTP %d)", statusCode)
		serviceErr.Suggestion = "Service is experiencing issues, try again later"
		serviceErr.IsRetryable = true

	case statusCode == 503:
		serviceErr.ErrorType = ErrorTypeServiceDown
		serviceErr.Message = "Service temporarily unavailable"
		serviceErr.Suggestion = "Service is down for maintenance, try again later"
		serviceErr.IsRetryable = true

	case strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "network"):
		serviceErr.ErrorType = ErrorTypeNetworkError
		serviceErr.Message = "Network connection failed"
		serviceErr.Suggestion = "Check your internet connection"
		serviceErr.IsRetryable = true

	case strings.Contains(errStr, "distinct languages") || strings.Contains(errStr, "unsupported language"):
		serviceErr.ErrorType = ErrorTypeLanguageError
		serviceErr.Message = "Language configuration issue"
		serviceErr.Suggestion = "Check source and target language settings"
		serviceErr.IsRetryable = false

	default:
		serviceErr.ErrorType = ErrorTypeUnknown
		serviceErr.Message = err.Error()
		serviceErr.Suggestion = "Try a different service or check input text"
		serviceErr.IsRetryable = false
	}

	return serviceErr
}

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

	if serviceErr, ok := err.(*ServiceError); ok {
		return FormatServiceError(serviceErr)
	}

	serviceErr := CreateServiceError(serviceName, err, 0)
	return FormatServiceError(serviceErr)
}

func FormatServiceError(err *ServiceError) string {
	if err == nil {
		return "Unknown error"
	}

	var icon string
	switch err.ErrorType {
	case ErrorTypeTimeout:
		icon = "⏱️"
	case ErrorTypeRateLimit:
		icon = "🚫"
	case ErrorTypeUnauthorized:
		icon = "🔑"
	case ErrorTypeForbidden:
		icon = "⛔"
	case ErrorTypeNotFound:
		icon = "❓"
	case ErrorTypeServerError:
		icon = "🔧"
	case ErrorTypeServiceDown:
		icon = "🔄"
	case ErrorTypeNetworkError:
		icon = "🌐"
	case ErrorTypeLanguageError:
		icon = "🗣️"
	default:
		icon = "❌"
	}

	var retryInfo string
	if err.IsRetryable {
		retryInfo = "\n🔄 Will retry automatically"
	} else {
		retryInfo = "\n⚠️  Manual intervention required"
	}

	return fmt.Sprintf("%s %s Error\n"+
		"Service: %s\n"+
		"Issue: %s\n"+
		"Suggestion: %s%s",
		icon, err.ErrorType, err.Service, err.Message, err.Suggestion, retryInfo)
}

func GetServiceSpecificErrorMessage(serviceName string, err error, statusCode int) string {
	serviceErr := CreateServiceError(serviceName, err, statusCode)

	switch serviceName {
	case "OPENAI":
		if serviceErr.ErrorType == ErrorTypeUnauthorized {
			serviceErr.Suggestion = "Get your API key from https://platform.openai.com/account/api-keys"
		} else if serviceErr.ErrorType == ErrorTypeRateLimit {
			serviceErr.Suggestion = "OpenAI has usage limits. Check your account quota or upgrade plan"
		}
	case "OPENROUTER":
		if serviceErr.ErrorType == ErrorTypeUnauthorized {
			serviceErr.Suggestion = "Get your API key from https://openrouter.ai/keys"
		} else if serviceErr.ErrorType == ErrorTypeRateLimit {
			serviceErr.Suggestion = "OpenRouter rate limits apply. Wait or upgrade your plan"
		} else if serviceErr.ErrorType == ErrorTypeNetworkError && strings.Contains(serviceErr.Message, "response") {
			serviceErr.Suggestion = "Large text may cause issues. Try shorter text or check network"
		}
	case "GOOGLE":
		if serviceErr.ErrorType == ErrorTypeRateLimit {
			serviceErr.Suggestion = "Google Translate has rate limits. Try again in a few minutes"
		}
	case "DEEPL":
		if serviceErr.ErrorType == ErrorTypeUnauthorized {
			serviceErr.Suggestion = "DeepL requires a valid API key for advanced features"
		}
	case "REVERSO", "REVERSO2":
		if serviceErr.ErrorType == ErrorTypeRateLimit {
			serviceErr.Suggestion = "Reverso limits requests. Wait before trying again"
		}
	case "MYMEMORY":
		if serviceErr.ErrorType == ErrorTypeRateLimit {
			serviceErr.Suggestion = "MyMemory has daily limits for anonymous users"
		}
	case "LINGVA":
		if serviceErr.ErrorType == ErrorTypeServerError {
			serviceErr.Suggestion = "Lingva is a community service. Try again later"
		}
	}

	return FormatServiceError(serviceErr)
}
