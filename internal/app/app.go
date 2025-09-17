package app

import (
	"multitrago/internal/cache"
	"multitrago/internal/clipboard"
	"multitrago/internal/config"
	"multitrago/internal/ratelimit"
	"multitrago/internal/utils"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
)

type App struct {
	cache     *cache.Manager
	clipboard *clipboard.Manager
	rateLimit *ratelimit.Manager
	config    *config.Manager
	services  []utils.ServiceConfig
}

func NewApp() *App {
	configManager := config.NewManager()
	if err := configManager.Initialize(); err != nil {
		_ = err
	}

	var services []utils.ServiceConfig
	if configServices := configManager.GetEnabledProviders(); len(configServices) > 0 {
		services = configServices
	} else {
		services = config.GetAvailableServices()
	}

	return &App{
		cache:     cache.NewManager(),
		clipboard: clipboard.NewManager(),
		rateLimit: ratelimit.NewManager(),
		config:    configManager,
		services:  services,
	}
}

func (a *App) GetModel() *Model {
	languages := config.GetSupportedLanguages()

	initialState := SetupState
	if a.config.IsAPIKeyRequired("OPENAI") && a.config.GetAPIKey("OPENAI") == "" {
		initialState = ConfigState
	}

	model := &Model{
		State: initialState,
		Setup: SetupModel{
			SelectedIndex: 0,
			Languages:     languages,
		},
		Config: ConfigModel{
			SelectedProvider: "OPENAI",
			SelectedIndex:    0,
			APIKeyInput:      nil,
			Providers:        make(map[string]bool),
			CurrentStep:      0,
		},
		AvailableServices:   []utils.ServiceConfig{},
		Translations:        make(map[string]string),
		Spinners:            make(map[string]spinner.Model),
		SpinnerStates:       make(map[string]SpinnerState),
		Done:                false,
		CheckedCount:        0,
		Width:               80,
		Height:              24,
		TargetLang:          "ru",
		IsTranslating:       false,
		TranslatingCount:    0,
		TranslationCache:    make(map[string]map[string]string),
		RateLimits:          make(map[string]*RateLimiter),
		RetryAttempts:       make(map[string]int),
		MaxRetries:          3,
		TranslationProgress: make(map[string]float64),
		Viewports:           make(map[string]viewport.Model),
		Columns:             2,
		Progress:            progress.New(progress.WithScaledGradient("#000000", "#FFFFFF")),
		CheckProgress:       0.0,
		StatusMessage:       "",
		app:                 a,
	}

	return model
}
