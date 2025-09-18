package main

import (
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"translatego/internal/app"
)

func main() {
	ti := textinput.New()
	ti.Placeholder = "Enter text to translate"
	ti.CharLimit = 500
	ti.Width = 50

	application := app.NewApp()

	model := application.GetModel()
	model.TextInput = &ti

	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
