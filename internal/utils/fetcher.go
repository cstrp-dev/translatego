package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ServiceConfig struct {
	Name    string
	URL     string
	Method  string
	Headers map[string]string
	Body    []byte
}

type Result struct {
	Name   string
	URL    string
	Status int
	Err    error
}

var client = &http.Client{Timeout: 5 * time.Second}

func CheckService(cfg ServiceConfig) Result {
	var req *http.Request
	var err error

	var body []byte
	var checkURL string

	switch cfg.Name {
	case "REVERSO", "REVERSO2":
		body = []byte(`{"format":"text","from":"en","to":"de","input":"test"}`)
		checkURL = cfg.URL
	case "LINGVA":
		checkURL = "https://lingva.thedaviddelta.com/api/v1/en/de/test"
	case "MYMEMORY":
		checkURL = "https://api.mymemory.translated.net/get?q=test&langpair=en|de"
	case "GOOGLE":
		body = []byte(`{"q":"test","source":"en","target":"de"}`)
		checkURL = cfg.URL
	case "DEEPL":
		body = []byte(`{"text":"test","source_lang":"EN","target_lang":"DE"}`)
		checkURL = cfg.URL
	case "OPENAI":
		body = []byte(`{
			"model": "gpt-3.5-turbo",
			"messages": [{"role":"user","content":"Translate test to German"}]
		}`)
		checkURL = cfg.URL
	default:
		body = cfg.Body
		checkURL = cfg.URL
	}

	if len(body) > 0 {
		req, err = http.NewRequest(cfg.Method, checkURL, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(cfg.Method, checkURL, nil)
	}
	if err != nil {
		return Result{Name: cfg.Name, URL: checkURL, Err: err}
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return Result{Name: cfg.Name, URL: checkURL, Err: err}
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	return Result{Name: cfg.Name, URL: checkURL, Status: res.StatusCode}
}

func TranslateService(cfg ServiceConfig, text, source, target string) (string, error) {
	if source == target {
		if source == "en" {
			target = "ru"
		} else {
			target = "en"
		}
	}

	sourceCode := source
	targetCode := target

	if cfg.Name == "REVERSO" || cfg.Name == "REVERSO2" {
		sourceCode = convertToReversoLangCode(source)
		targetCode = convertToReversoLangCode(target)
	}

	var body []byte
	var finalURL string = cfg.URL
	switch cfg.Name {
	case "GOOGLE":
		body = []byte(fmt.Sprintf(`{"message":"%s","from":"%s","to":"%s"}`, text, sourceCode, targetCode))
	case "DEEPL":
		body = []byte(fmt.Sprintf(`{"text":"%s","source_lang":"%s","target_lang":"%s"}`, text, strings.ToUpper(sourceCode), strings.ToUpper(targetCode)))
	case "REVERSO":
		body = []byte(fmt.Sprintf(`{"format":"text","from":"%s","to":"%s","input":"%s"}`, sourceCode, targetCode, text))
	case "REVERSO2":
		body = []byte(fmt.Sprintf(`{"format":"text","from":"%s","to":"%s","input":"%s","options":{"sentenceSplitter":true,"origin":"translation.web","contextResults":false,"languageDetection":false}}`, sourceCode, targetCode, text))
	case "MYMEMORY":
		finalURL = fmt.Sprintf("https://api.mymemory.translated.net/get?q=%s&langpair=%s|%s", url.QueryEscape(text), sourceCode, targetCode)
	case "LINGVA":
		finalURL = fmt.Sprintf("https://lingva.thedaviddelta.com/api/v1/%s/%s/%s", sourceCode, targetCode, url.QueryEscape(text))
	case "OPENAI":
		body = []byte(fmt.Sprintf(`{
			"model": "gpt-3.5-turbo",
			"messages": [{"role":"user","content":"Translate '%s' from %s to %s. Return only the translation, no additional text."}]
		}`, text, getLanguageName(sourceCode), getLanguageName(targetCode)))
	}

	var req *http.Request
	var err error
	if len(body) > 0 {
		req, err = http.NewRequest(cfg.Method, finalURL, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(cfg.Method, finalURL, nil)
	}
	if err != nil {
		return "", err
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	res, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("request timeout after 10 seconds")
		}
		return "", fmt.Errorf("request failed: %w", err)
	}

	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", res.StatusCode)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	responseBody := buf.String()

	switch cfg.Name {
	case "GOOGLE":
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(responseBody), &data); err != nil {
			return "", err
		}
		if trans, ok := data["translation"].(map[string]interface{}); ok {
			if res, ok := trans["trans_result"].(map[string]interface{}); ok {
				if dst, ok := res["dst"].(string); ok {
					return dst, nil
				}
			}
		}
		return responseBody, nil
	case "DEEPL":
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(responseBody), &data); err != nil {
			return "", err
		}
		if text, ok := data["data"].(string); ok {
			return text, nil
		}
		return responseBody, nil
	case "REVERSO":
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(responseBody), &data); err != nil {
			return "", err
		}
		if trans, ok := data["translation"].([]interface{}); ok && len(trans) > 0 {
			if t, ok := trans[0].(string); ok {
				return t, nil
			}
		}
		return responseBody, nil
	case "REVERSO2":
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(responseBody), &data); err != nil {
			return "", err
		}
		if trans, ok := data["translation"].([]interface{}); ok && len(trans) > 0 {
			if t, ok := trans[0].(string); ok {
				return t, nil
			}
		}
		return responseBody, nil
	case "MYMEMORY":
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(responseBody), &data); err != nil {
			return "", err
		}
		if resp, ok := data["responseData"].(map[string]interface{}); ok {
			if trans, ok := resp["translatedText"].(string); ok {
				return trans, nil
			}
		}
		return responseBody, nil
	case "LINGVA":
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(responseBody), &data); err != nil {
			return "", err
		}
		if trans, ok := data["translation"].(string); ok {
			decoded := strings.ReplaceAll(trans, "+", " ")
			decoded = strings.Join(strings.Fields(decoded), " ")
			return strings.TrimSpace(decoded), nil
		}
		return responseBody, nil
	case "OPENAI":
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(responseBody), &data); err != nil {
			return "", err
		}
		if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if msg, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := msg["content"].(string); ok {
						return content, nil
					}
				}
			}
		}
		return responseBody, nil
	default:
		return responseBody, nil
	}
}

func convertToReversoLangCode(langCode string) string {
	switch langCode {
	case "en":
		return "eng"
	case "ru":
		return "rus"
	case "de":
		return "ger"
	case "fr":
		return "fra"
	case "es":
		return "spa"
	case "it":
		return "ita"
	case "ja":
		return "jpn"
	case "zh":
		return "chi"
	case "ko":
		return "kor"
	case "ar":
		return "ara"
	default:
		return langCode
	}
}

func getLanguageName(langCode string) string {
	switch langCode {
	case "en":
		return "English"
	case "ru":
		return "Russian"
	case "de":
		return "German"
	case "fr":
		return "French"
	case "es":
		return "Spanish"
	case "it":
		return "Italian"
	case "ja":
		return "Japanese"
	case "zh":
		return "Chinese"
	case "ko":
		return "Korean"
	case "ar":
		return "Arabic"
	default:
		return langCode
	}
}
