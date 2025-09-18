package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"translatego/internal/utils"
)

type Config struct {
	Version   string                    `json:"version"`
	Providers map[string]ProviderConfig `json:"providers"`
	Settings  Settings                  `json:"settings"`
}

type ProviderConfig struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body,omitempty"`
	Enabled bool              `json:"enabled"`
	APIKey  string            `json:"api_key,omitempty"`
}

type Settings struct {
	DefaultTargetLang string `json:"default_target_lang"`
	MaxRetries        int    `json:"max_retries"`
	TimeoutSeconds    int    `json:"timeout_seconds"`
	CacheEnabled      bool   `json:"cache_enabled"`
}

type Manager struct {
	configDir  string
	configFile string
	config     *Config
}

func NewManager() *Manager {
	var configDir string

	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, "translatego")
	} else {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".config", "translatego")
	}

	return &Manager{
		configDir:  configDir,
		configFile: filepath.Join(configDir, "config.json"),
		config:     nil,
	}
}

func (m *Manager) Initialize() error {
	if err := os.MkdirAll(m.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if _, err := os.Stat(m.configFile); os.IsNotExist(err) {
		if err := m.createDefaultConfig(); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	}

	return m.Load()
}

func (m *Manager) createDefaultConfig() error {
	services := GetAvailableServices()

	providers := make(map[string]ProviderConfig)
	for _, service := range services {
		provider := ProviderConfig{
			Name:    service.Name,
			URL:     service.URL,
			Method:  service.Method,
			Headers: service.Headers,
			Enabled: true,
		}

		if service.Name == "OPENAI" {
			provider.APIKey = ""
		}

		if len(service.Body) > 0 {
			provider.Body = string(service.Body)
		}

		providers[service.Name] = provider
	}

	m.config = &Config{
		Version:   "1.0.0",
		Providers: providers,
		Settings: Settings{
			DefaultTargetLang: "ru",
			MaxRetries:        3,
			TimeoutSeconds:    5,
			CacheEnabled:      true,
		},
	}

	return m.Save()
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	m.config = &Config{}
	if err := json.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func (m *Manager) Save() error {
	if m.config == nil {
		return fmt.Errorf("config is not initialized")
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (m *Manager) GetConfig() *Config {
	return m.config
}

func (m *Manager) GetProviders() map[string]ProviderConfig {
	if m.config == nil {
		return nil
	}
	return m.config.Providers
}

func (m *Manager) GetEnabledProviders() []utils.ServiceConfig {
	if m.config == nil {
		return nil
	}

	var services []utils.ServiceConfig
	for _, provider := range m.config.Providers {
		if provider.Enabled {
			service := utils.ServiceConfig{
				Name:    provider.Name,
				URL:     provider.URL,
				Method:  provider.Method,
				Headers: provider.Headers,
			}

			if provider.APIKey != "" {
				for key, value := range provider.Headers {
					if value == "Bearer YOUR_OPENAI_KEY" || value == "Bearer YOUR_API_KEY" {
						service.Headers[key] = "Bearer " + provider.APIKey
					}
				}
			}

			if provider.Body != "" {
				service.Body = []byte(provider.Body)
			}

			services = append(services, service)
		}
	}

	return services
}

func (m *Manager) SetAPIKey(providerName, apiKey string) error {
	if m.config == nil {
		return fmt.Errorf("config is not initialized")
	}

	if provider, exists := m.config.Providers[providerName]; exists {
		provider.APIKey = apiKey
		m.config.Providers[providerName] = provider
		return m.Save()
	}

	return fmt.Errorf("provider %s not found", providerName)
}

func (m *Manager) GetAPIKey(providerName string) string {
	if m.config == nil {
		return ""
	}

	if provider, exists := m.config.Providers[providerName]; exists {
		return provider.APIKey
	}

	return ""
}

func (m *Manager) IsAPIKeyRequired(providerName string) bool {
	return providerName == "OPENAI"
}

func (m *Manager) GetConfigDir() string {
	return m.configDir
}

func (m *Manager) GetConfigFile() string {
	return m.configFile
}
