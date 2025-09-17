package app

import (
	"fmt"
	"strings"
)

func (m *Model) SetupView() string {
	asciiArt := `
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║            ███╗   ███╗██╗   ██╗██╗  ████████╗██████╗  █████╗  ██████╗        ║
║            ████╗ ████║██║   ██║██║  ╚══██╔══╝██╔══██╗██╔══██╗██╔════╝        ║
║            ██╔████╔██║██║   ██║██║     ██║   ██████╔╝███████║██║  ███╗       ║
║            ██║╚██╔╝██║██║   ██║██║     ██║   ██╔══██╗██╔══██║██║   ██║       ║
║            ██║ ╚═╝ ██║╚██████╔╝███████╗██║   ██║  ██║██║  ██║╚██████╔╝       ║
║            ╚═╝     ╚═╝ ╚═════╝ ╚══════╝╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝        ║
║                                                                              ║
║                    🌍 Advanced Translation Terminal 🌍                       ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝`

	langNames := map[string]string{
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

	var menuItems []string
	for i, lang := range m.Setup.Languages {
		cursor := "  "
		if i == m.Setup.SelectedIndex {
			cursor = "▶ "
		}
		name := langNames[lang]
		menuItems = append(menuItems, fmt.Sprintf("%s%s (%s)", cursor, name, lang))
	}

	menu := strings.Join(menuItems, "\n")

	instructions := "\n\n" + InstructionStyle.
		Render("Use ↑↓ arrows to navigate • Press Enter to select • Press Ctrl+C to quit")

	return asciiArt + "\n\n" + TitleStyle.
		Render("Choose your target language:") + "\n\n" + menu + instructions
}

func CreateProgressBar(progress float64, width int) string {
	if width < 10 {
		width = 10
	}

	filled := int(progress * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	percentage := int(progress * 100)

	return fmt.Sprintf("[%s] %d%%", bar, percentage)
}

func GetDetailedErrorMessage(service string, err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "rate limit"):
		return fmt.Sprintf("⚠️  Rate Limit Exceeded\nService: %s\nSuggestion: Wait a moment and try again", service)
	case strings.Contains(errStr, "timeout"):
		return fmt.Sprintf("⏱️  Request Timeout\nService: %s\nSuggestion: Check your internet connection or try again", service)
	case strings.Contains(errStr, "connection refused"):
		return fmt.Sprintf("🌐 Connection Failed\nService: %s\nSuggestion: Service may be temporarily unavailable", service)
	case strings.Contains(errStr, "status 429"):
		return fmt.Sprintf("🚫 Too Many Requests\nService: %s\nSuggestion: Service is busy, try again in a few minutes", service)
	case strings.Contains(errStr, "status 500"):
		return fmt.Sprintf("🔧 Server Error\nService: %s\nSuggestion: Service is experiencing issues, try again later", service)
	case strings.Contains(errStr, "status 503"):
		return fmt.Sprintf("🔄 Service Unavailable\nService: %s\nSuggestion: Service is temporarily down, try again later", service)
	case strings.Contains(errStr, "distinct languages"):
		return fmt.Sprintf("🌍 Language Conflict\nService: %s\nSuggestion: Source and target languages are the same", service)
	case strings.Contains(errStr, "unsupported language"):
		return fmt.Sprintf("🗣️ Unsupported Language\nService: %s\nSuggestion: Try a different language pair", service)
	default:
		return fmt.Sprintf("❌ Translation Failed\nService: %s\nError: %v\nSuggestion: Check input text or try a different service", service, err)
	}
}
