package app

import (
	"fmt"
	"strings"
	"time"

	"translatego/internal/utils"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppState int

const (
	SetupState AppState = iota
	LoadingState
	CheckingState
	MainState
	ConfigState
)

type SpinnerState int

const (
	SpinnerLoading SpinnerState = iota
	SpinnerRetrying
	SpinnerError
)

type SetupModel struct {
	SelectedIndex int
	Languages     []string
}

type ConfigModel struct {
	SelectedProvider string
	SelectedIndex    int
	APIKeyInput      *textinput.Model
	Providers        map[string]bool
	CurrentStep      int
}

type Model struct {
	State               AppState
	Setup               SetupModel
	Config              ConfigModel
	TextInput           *textinput.Model
	AvailableServices   []utils.ServiceConfig
	Translations        map[string]string
	Spinners            map[string]spinner.Model
	SpinnerStates       map[string]SpinnerState
	Done                bool
	CheckedCount        int
	Width               int
	Height              int
	TargetLang          string
	IsTranslating       bool
	TranslatingCount    int
	TranslationCache    map[string]map[string]string
	RateLimits          map[string]*RateLimiter
	RetryAttempts       map[string]int
	MaxRetries          int
	TranslationProgress map[string]float64
	Viewports           map[string]viewport.Model
	Columns             int
	Progress            progress.Model
	CheckProgress       float64
	StatusMessage       string
	app                 *App
}

type ResultMsg utils.Result

type TranslationMsg struct {
	Service string
	Text    string
	Err     error
}

type RetryMsg struct {
	Service string
	Text    string
	Source  string
	Target  string
	Attempt int
	Delay   time.Duration
}

type RateLimiter struct {
	Requests    int
	LastReset   time.Time
	MaxRequests int
	Window      time.Duration
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		Requests:    0,
		LastReset:   time.Now(),
		MaxRequests: maxRequests,
		Window:      window,
	}
}

func (rl *RateLimiter) Allow() bool {
	now := time.Now()
	if now.Sub(rl.LastReset) > rl.Window {
		rl.Requests = 0
		rl.LastReset = now
	}
	return rl.Requests < rl.MaxRequests
}

func (rl *RateLimiter) RecordRequest() {
	rl.Requests++
}

type CacheManager struct {
	cache map[string]map[string]string
}

func NewCacheManager() *CacheManager {
	return &CacheManager{
		cache: make(map[string]map[string]string),
	}
}

func (cm *CacheManager) GetCacheKey(text, source, target string) string {
	return text + "|" + source + "|" + target
}

func (cm *CacheManager) Get(service, text, source, target string) (string, bool) {
	if serviceCache, exists := cm.cache[service]; exists {
		cacheKey := cm.GetCacheKey(text, source, target)
		if translation, cached := serviceCache[cacheKey]; cached {
			return translation, true
		}
	}
	return "", false
}

func (cm *CacheManager) Set(service, text, source, target, translation string) {
	if _, exists := cm.cache[service]; !exists {
		cm.cache[service] = make(map[string]string)
	}
	cacheKey := cm.GetCacheKey(text, source, target)
	cm.cache[service][cacheKey] = translation
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.State == SetupState {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.Width = msg.Width
			m.Height = msg.Height
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				m.TargetLang = m.Setup.Languages[m.Setup.SelectedIndex]
				m.State = LoadingState
				return m, m.Init()
			case "up", "k":
				if m.Setup.SelectedIndex > 0 {
					m.Setup.SelectedIndex--
				}
			case "down", "j":
				if m.Setup.SelectedIndex < len(m.Setup.Languages)-1 {
					m.Setup.SelectedIndex++
				}
			}
		}
		return m, nil
	}

	if m.State == ConfigState {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.Width = msg.Width
			m.Height = msg.Height
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				if m.Config.CurrentStep == 0 {
					allProviders := []string{"OPENAI", "GOOGLE", "DEEPL", "REVERSO", "MYMEMORY"}
					var providersRequiringKeys []string

					for _, provider := range allProviders {
						if m.app.config.IsAPIKeyRequired(provider) {
							providersRequiringKeys = append(providersRequiringKeys, provider)
						}
					}

					if len(providersRequiringKeys) == 0 {
						m.State = SetupState
						return m, nil
					}

					if m.Config.SelectedIndex < len(providersRequiringKeys) {
						m.Config.SelectedProvider = providersRequiringKeys[m.Config.SelectedIndex]
						m.Config.CurrentStep = 1
						if m.Config.APIKeyInput == nil {
							ti := textinput.New()
							ti.Placeholder = "Enter your API key"
							ti.CharLimit = 200
							ti.Width = 50
							ti.EchoMode = textinput.EchoPassword
							ti.EchoCharacter = '•'
							m.Config.APIKeyInput = &ti
						}
						m.Config.APIKeyInput.Focus()
					}
				} else {
					if m.Config.APIKeyInput != nil {
						apiKey := m.Config.APIKeyInput.Value()
						if apiKey != "" {
							m.app.config.SetAPIKey(m.Config.SelectedProvider, apiKey)
						}
					}
					m.State = SetupState
				}
				return m, nil
			case "up", "k":
				if m.Config.CurrentStep == 0 {
					allProviders := []string{"OPENAI", "GOOGLE", "DEEPL", "REVERSO", "MYMEMORY"}
					var providersRequiringKeys []string

					for _, provider := range allProviders {
						if m.app.config.IsAPIKeyRequired(provider) {
							providersRequiringKeys = append(providersRequiringKeys, provider)
						}
					}

					if len(providersRequiringKeys) > 0 && m.Config.SelectedIndex > 0 {
						m.Config.SelectedIndex--
						m.Config.SelectedProvider = providersRequiringKeys[m.Config.SelectedIndex]
					}
				}
			case "down", "j":
				if m.Config.CurrentStep == 0 {
					allProviders := []string{"OPENAI", "GOOGLE", "DEEPL", "REVERSO", "MYMEMORY"}
					var providersRequiringKeys []string

					for _, provider := range allProviders {
						if m.app.config.IsAPIKeyRequired(provider) {
							providersRequiringKeys = append(providersRequiringKeys, provider)
						}
					}

					if len(providersRequiringKeys) > 0 && m.Config.SelectedIndex < len(providersRequiringKeys)-1 {
						m.Config.SelectedIndex++
						m.Config.SelectedProvider = providersRequiringKeys[m.Config.SelectedIndex]
					}
				}
			}
		}

		if m.Config.APIKeyInput != nil && m.Config.CurrentStep == 1 {
			newModel, cmd := m.Config.APIKeyInput.Update(msg)
			*m.Config.APIKeyInput = newModel
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)
	}

	if m.State == LoadingState {
		switch msg.(type) {
		case struct{}:
			m.State = CheckingState
			return m, m.Init()
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Progress.Width = msg.Width - 4

		numServices := len(m.AvailableServices)
		if numServices == 0 {
			return m, nil
		}

		var cols int
		switch numServices {
		case 1:
			cols = 1
		case 2:
			cols = 2
		case 3:
			cols = 3
		case 4:
			cols = 2
		case 5, 6:
			cols = 3
		default:
			cols = 3
		}

		boxWidth := (m.Width - 2) / cols
		rows := (numServices + cols - 1) / cols
		boxHeight := (m.Height - 10) / rows

		for name, vp := range m.Viewports {
			vp.Width = boxWidth - 2
			vp.Height = boxHeight - 2
			m.Viewports[name] = vp
		}
	case ResultMsg:
		m.CheckedCount++
		m.CheckProgress = float64(m.CheckedCount) / float64(len(m.app.services))
		if msg.Err == nil && msg.Status == 200 {
			for _, svc := range m.app.services {
				if svc.Name == msg.Name {
					m.AvailableServices = append(m.AvailableServices, svc)
					break
				}
			}
			m.Translations[msg.Name] = ""
			sp := spinner.New()
			sp.Spinner = spinner.Dot
			sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
			m.Spinners[msg.Name] = sp
			m.SpinnerStates[msg.Name] = SpinnerLoading
			// Calculate initial viewport dimensions based on current terminal size
			numSvcs := len(m.AvailableServices) + 1 // +1 for the new service being added
			var initialCols int
			switch numSvcs {
			case 1:
				initialCols = 1
			case 2:
				initialCols = 2
			case 3, 4:
				initialCols = 2
			case 5, 6:
				initialCols = 3
			default:
				initialCols = 3
			}
			initialBoxWidth := (m.Width - 2) / initialCols
			initialRows := (numSvcs + initialCols - 1) / initialCols
			initialBoxHeight := (m.Height - 10) / initialRows

			vp := viewport.New(initialBoxWidth-2, initialBoxHeight-2)
			vp.SetContent("")
			m.Viewports[msg.Name] = vp
			cmds = append(cmds, sp.Tick)
		}
		if m.CheckedCount == len(m.app.services) {
			m.Done = true
			m.State = MainState
			if m.TextInput != nil {
				m.TextInput.Focus()
			}
		}
	case TranslationMsg:
		m.handleTranslationResult(msg)
	case RetryMsg:
		m.handleRetry(msg, &cmds)
	case spinner.TickMsg:
		for name, sp := range m.Spinners {
			sp, cmd = sp.Update(msg)
			m.Spinners[name] = sp
			cmds = append(cmds, cmd)
		}
	case tea.KeyMsg:
		m.handleKeyPress(msg, &cmds)
	}

	if m.TextInput != nil {
		newModel, cmd := m.TextInput.Update(msg)
		*m.TextInput = newModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) Init() tea.Cmd {
	if m.State == SetupState {
		return nil
	}

	if m.State == LoadingState {
		return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return struct{}{}
		})
	}

	if m.State == CheckingState {
		cmds := make([]tea.Cmd, 0, len(m.app.services))
		for i, svc := range m.app.services {
			cfg := svc
			index := i
			cmds = append(cmds, tea.Tick(time.Duration(index)*300*time.Millisecond, func(t time.Time) tea.Msg {
				return ResultMsg(utils.CheckService(cfg))
			}))
		}
		return tea.Batch(cmds...)
	}

	return nil
}

func (m *Model) handleTranslationResult(msg TranslationMsg) {
	if msg.Err == nil {
		m.Translations[msg.Service] = msg.Text
		delete(m.RetryAttempts, msg.Service)
		m.TranslationProgress[msg.Service] = 1.0
		m.app.cache.Set(msg.Service, "", "", "", msg.Text)
	} else {
		m.handleTranslationError(msg)
	}
	m.TranslatingCount--
	if m.TranslatingCount <= 0 {
		m.IsTranslating = false
	}
}

func (m *Model) handleTranslationError(msg TranslationMsg) {
	attempts := m.RetryAttempts[msg.Service]
	if attempts < m.MaxRetries {
		m.RetryAttempts[msg.Service] = attempts + 1

		retryText := fmt.Sprintf("Retrying... (attempt %d/%d)", attempts+1, m.MaxRetries)
		m.Translations[msg.Service] = retryText
		m.TranslationProgress[msg.Service] = 0.5

		if sp, exists := m.Spinners[msg.Service]; exists {
			sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			m.Spinners[msg.Service] = sp
			m.SpinnerStates[msg.Service] = SpinnerRetrying
		}
	} else {
		detailedError := utils.GetDetailedErrorMessage(msg.Service, msg.Err)
		m.Translations[msg.Service] = detailedError
		delete(m.RetryAttempts, msg.Service)
		m.TranslationProgress[msg.Service] = 0.0
	}
}

func (m *Model) createTranslationCommand(svc utils.ServiceConfig, text, targetLang string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		cfg := svc
		source := utils.DetectFromLanguage(text)
		target := utils.DetectToLanguage(source, targetLang)

		if cached, exists := m.app.cache.Get(cfg.Name, text, source, target); exists {
			return TranslationMsg{Service: cfg.Name, Text: cached, Err: nil}
		}

		if !m.app.rateLimit.Allow(cfg.Name) {
			return TranslationMsg{Service: cfg.Name, Text: "", Err: fmt.Errorf("rate limit exceeded")}
		}

		if cfg.Name == "OPENAI" && m.app.config.IsAPIKeyRequired("OPENAI") {
			apiKey := m.app.config.GetAPIKey("OPENAI")
			if apiKey != "" {

				newHeaders := make(map[string]string)
				for k, v := range cfg.Headers {
					if k == "Authorization" {
						newHeaders[k] = "Bearer " + apiKey
					} else {
						newHeaders[k] = v
					}
				}
				cfg.Headers = newHeaders
			}
		}

		trans, err := utils.TranslateService(cfg, text, source, target)
		if err == nil {
			m.app.cache.Set(cfg.Name, text, source, target, trans)
			m.app.rateLimit.RecordRequest(cfg.Name)
		}

		return TranslationMsg{Service: cfg.Name, Text: trans, Err: err}
	})
}

func (m *Model) handleRetry(msg RetryMsg, cmds *[]tea.Cmd) {

	cfg := utils.ServiceConfig{}
	for _, svc := range m.AvailableServices {
		if svc.Name == msg.Service {
			cfg = svc
			break
		}
	}

	if cfg.Name == "" {
		return
	}

	cmd := func() tea.Msg {
		source := utils.DetectFromLanguage(msg.Text)
		target := utils.DetectToLanguage(source, msg.Target)

		finalCfg := cfg
		if cfg.Name == "OPENAI" && m.app.config.IsAPIKeyRequired("OPENAI") {
			apiKey := m.app.config.GetAPIKey("OPENAI")
			if apiKey != "" {
				if finalCfg.Headers == nil {
					finalCfg.Headers = make(map[string]string)
				}
				newHeaders := make(map[string]string)
				for k, v := range finalCfg.Headers {
					if k == "Authorization" {
						newHeaders[k] = "Bearer " + apiKey
					} else {
						newHeaders[k] = v
					}
				}
				finalCfg.Headers = newHeaders
			}
		}

		if !m.app.rateLimit.Allow(cfg.Name) {
			return TranslationMsg{Service: cfg.Name, Text: "", Err: fmt.Errorf("rate limit exceeded")}
		}

		trans, err := utils.TranslateService(finalCfg, msg.Text, source, target)
		if err == nil {
			m.app.cache.Set(cfg.Name, msg.Text, source, target, trans)
			m.app.rateLimit.RecordRequest(cfg.Name)
		}
		return TranslationMsg{Service: cfg.Name, Text: trans, Err: err}
	}
	*cmds = append(*cmds, cmd)
}

func (m *Model) handleKeyPress(msg tea.KeyMsg, cmds *[]tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		*cmds = append(*cmds, tea.Quit)
	case "ctrl+v":
		if text, err := m.app.clipboard.PasteFromClipboard(); err == nil && text != "" {
			if m.TextInput != nil {
				m.TextInput.SetValue(text)
			}
		}
	case "ctrl+l":
		if m.TextInput != nil {
			m.TextInput.SetValue("")
		}
	case "alt+1":
		m.copyTranslationToClipboard(0)
	case "alt+2":
		m.copyTranslationToClipboard(1)
	case "alt+3":
		m.copyTranslationToClipboard(2)
	case "alt+4":
		m.copyTranslationToClipboard(3)
	case "alt+5":
		m.cycleLayout()
	case "alt+c":
	case "enter":
		m.handleEnterKey(cmds)
	}
}

func (m *Model) copyTranslationToClipboard(index int) {
	if index < len(m.AvailableServices) {
		serviceName := m.AvailableServices[index].Name
		if trans := m.Translations[serviceName]; trans != "" && !strings.HasPrefix(trans, "Error:") {
			m.app.clipboard.CopyToClipboard(trans)
		}
	}
}

func (m *Model) cycleLayout() {
	numServices := len(m.AvailableServices)
	if numServices == 0 {
		return
	}

	var optimalCols int
	switch numServices {
	case 1:
		optimalCols = 1
	case 2:
		optimalCols = 2
	case 3, 4:
		optimalCols = 2
	case 5, 6:
		optimalCols = 3
	default:
		optimalCols = 3
	}

	m.Columns = optimalCols

	if m.Width > 0 && m.Height > 0 {
		boxWidth := (m.Width - 2) / m.Columns
		rows := (numServices + m.Columns - 1) / m.Columns
		boxHeight := (m.Height - 10) / rows

		for name, vp := range m.Viewports {
			vp.Width = boxWidth - 2
			vp.Height = boxHeight - 2
			m.Viewports[name] = vp
		}
	}
}

func (m *Model) handleEnterKey(cmds *[]tea.Cmd) {
	if m.TextInput == nil {
		return
	}

	text := m.TextInput.Value()
	if !m.IsTranslating && text != "" {
		m.IsTranslating = true
		m.TranslatingCount = len(m.AvailableServices)

		var validServices []utils.ServiceConfig
		for _, svc := range m.AvailableServices {
			if svc.Name == "OPENAI" && m.app.config.IsAPIKeyRequired("OPENAI") {
				apiKey := m.app.config.GetAPIKey("OPENAI")
				if apiKey == "" {
					m.TranslatingCount--
					m.Translations[svc.Name] = "⚠️ OpenAI API key required"
					m.TranslationProgress[svc.Name] = 0.0
					continue
				}
			}
			validServices = append(validServices, svc)
		}

		m.TranslatingCount = len(validServices)

		for _, svc := range validServices {
			m.Translations[svc.Name] = ""
			m.TranslationProgress[svc.Name] = 0.0
			sp := spinner.New()
			sp.Spinner = spinner.Dot
			sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
			m.Spinners[svc.Name] = sp
			m.SpinnerStates[svc.Name] = SpinnerLoading

			*cmds = append(*cmds, sp.Tick)
		}

		for _, svc := range validServices {
			translationCmd := m.createTranslationCommand(svc, text, m.TargetLang)
			*cmds = append(*cmds, translationCmd)
		}
	}
}

func (m *Model) ConfigView() string {
	var content strings.Builder

	title := TitleStyle.Render("🔧 Translatego Configuration")
	content.WriteString(title + "\n\n")

	if m.Config.CurrentStep == 0 {
		content.WriteString("Select a provider that requires API key setup:\n\n")

		allProviders := []string{"OPENAI", "GOOGLE", "DEEPL", "REVERSO", "MYMEMORY"}
		var providersRequiringKeys []string

		for _, provider := range allProviders {
			if m.app.config.IsAPIKeyRequired(provider) {
				providersRequiringKeys = append(providersRequiringKeys, provider)
			}
		}

		if len(providersRequiringKeys) == 0 {
			content.WriteString("✅ All providers are configured and ready to use!\n\n")
			content.WriteString("Press Enter to continue or q to quit\n")
		} else {
			for i, provider := range providersRequiringKeys {
				var marker string
				if i == m.Config.SelectedIndex {
					marker = "▶ "
				} else {
					marker = "  "
				}

				var status string
				if m.app.config.GetAPIKey(provider) != "" {
					status = " ✅ Configured"
				} else {
					status = " ⚠️  Requires API key"
				}

				content.WriteString(fmt.Sprintf("%s%d. %s%s\n", marker, i+1, provider, status))
			}

			content.WriteString("\nUse ↑/↓ to navigate, Enter to select, q to quit\n")
		}
	} else {
		content.WriteString(fmt.Sprintf("Enter API key for %s:\n\n", m.Config.SelectedProvider))

		if m.Config.APIKeyInput != nil {
			inputStyle := InputStyle.Width(m.Width - 4)
			content.WriteString(inputStyle.Render("API Key:\n"+m.Config.APIKeyInput.View()) + "\n\n")
		}

		content.WriteString("Press Enter to save, q to quit\n")
		content.WriteString("💡 Your API key will be stored securely in ~/.config/translatego/config.json\n")
	}

	return content.String()
}

func (m *Model) View() string {
	if m.State == SetupState {
		return m.SetupView()
	}

	if m.State == ConfigState {
		return m.ConfigView()
	}

	if m.State == LoadingState {
		return "\n\n🌍 Preparing Translatego...\n\n"
	}

	if m.State == CheckingState || (m.State == MainState && m.CheckedCount < 5) {
		pad := strings.Repeat(" ", 2)
		return "\n" +
			pad + m.Progress.ViewAs(m.CheckProgress) + "\n\n" +
			pad + fmt.Sprintf("Checking %d/%d services...", m.CheckedCount, 6) + "\n"
	}

	inputStyle := InputStyle.
		Width(m.Width - 4).
		Height(5)

	inputBox := inputStyle.Render("Text to translate:\n" + m.TextInput.View())

	var translationBoxes []string
	numServices := len(m.AvailableServices)
	if numServices == 0 {
		return inputBox + "\n\nNo services available."
	}

	var cols int
	switch numServices {
	case 1:
		cols = 1
	case 2:
		cols = 2
	case 3, 4:
		cols = 2
	case 5, 6:
		cols = 3
	default:
		cols = 3
	}

	rows := (numServices + cols - 1) / cols
	boxWidth := (m.Width - 2) / cols // Reduced margin from 4 to 2
	boxHeight := (m.Height - 10) / rows

	for _, svc := range m.AvailableServices {
		trans := m.Translations[svc.Name]
		if trans == "" && m.IsTranslating {
			trans = m.Spinners[svc.Name].View() + " Translating..."
		} else if trans == "" {
			trans = "Ready for translation"
		}

		progress := m.TranslationProgress[svc.Name]
		var progressBar string
		if progress > 0 && progress < 1.0 {
			progressBar = "\n" + CreateProgressBar(progress, boxWidth-2)
		}

		wrappedTrans := WrapText(trans, boxWidth-2)

		boxStyle := BoxStyle.
			Width(boxWidth - 2).
			Height(boxHeight - 2)

		displayContent := wrappedTrans + progressBar
		box := boxStyle.Render(fmt.Sprintf("[%s]\n%s", svc.Name, displayContent))
		translationBoxes = append(translationBoxes, box)
	}

	var gridRows []string
	for i := 0; i < len(translationBoxes); i += cols {
		end := i + cols
		if end > len(translationBoxes) {
			end = len(translationBoxes)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, translationBoxes[i:end]...)
		gridRows = append(gridRows, row)
	}
	translationsView := lipgloss.JoinVertical(lipgloss.Left, gridRows...)

	layout := lipgloss.JoinVertical(lipgloss.Left, inputBox, translationsView)

	help := fmt.Sprintf("\nPress Enter to translate | Target language: %s | Ctrl+V paste | Ctrl+L clear | Alt+1/2/3 copy | q to quit.", m.TargetLang)

	return layout + help
}

var (
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("69")).
			Padding(0, 0). // Reduced padding to maximize text space
			Align(lipgloss.Left)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	InstructionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)
)
