package app

import (
	"fmt"
	"strings"

	"github.com/common-nighthawk/go-figure"
)

func (m *Model) SetupView() string {
	figure := figure.NewFigure("Translatego", "big", true)
	asciiArt := figure.String()

	borderWidth := 85
	lines := strings.Split(asciiArt, "\n")
	var borderedLines []string

	borderedLines = append(borderedLines, strings.Repeat("═", borderWidth))

	for _, line := range lines {
		if len(line) > 0 {
			padded := fmt.Sprintf("║ %-*s ║", borderWidth-4, line)
			borderedLines = append(borderedLines, padded)
		}
	}

	borderedLines = append(borderedLines, strings.Repeat("═", borderWidth))

	finalArt := strings.Join(borderedLines, "\n")

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

	return finalArt + "\n\n" + TitleStyle.
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

func WrapText(text string, width int) string {
	if width <= 0 {
		width = 20
	}
	if text == "" {
		return text
	}

	if width < 8 {
		width = 8
	}

	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if len(line) <= width {
			result = append(result, line)
			continue
		}

		words := strings.Fields(line)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}

		var currentLine []string
		lineLength := 0

		for _, word := range words {
			wordLength := len(word)

			if wordLength > width {
				if len(currentLine) > 0 {
					result = append(result, strings.Join(currentLine, " "))
					currentLine = []string{}
					lineLength = 0
				}

				for len(word) > width {
					result = append(result, word[:width-1]+"-")
					word = word[width-1:]
				}

				if len(word) > 0 {
					currentLine = append(currentLine, word)
					lineLength = len(word)
				}
				continue
			}

			if lineLength == 0 {
				currentLine = append(currentLine, word)
				lineLength = wordLength
			} else if lineLength+1+wordLength <= width {
				currentLine = append(currentLine, word)
				lineLength += 1 + wordLength
			} else {
				result = append(result, strings.Join(currentLine, " "))
				currentLine = []string{word}
				lineLength = wordLength
			}
		}

		if len(currentLine) > 0 {
			result = append(result, strings.Join(currentLine, " "))
		}
	}

	return strings.Join(result, "\n")
}
