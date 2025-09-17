package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Manager struct {
	copyCmd   string
	pasteCmd  string
	copyArgs  []string
	pasteArgs []string
}

func NewManager() *Manager {
	manager := &Manager{}
	manager.detectClipboardTools()
	return manager
}

func (m *Manager) detectClipboardTools() {
	switch runtime.GOOS {
	case "linux":
		m.detectLinuxClipboard()
	case "darwin":
		m.copyCmd = "pbcopy"
		m.pasteCmd = "pbpaste"
	case "windows":
		m.copyCmd = "clip"
		m.pasteCmd = "powershell"
		m.pasteArgs = []string{"-command", "Get-Clipboard"}
	default:
		m.copyCmd = "xclip"
		m.copyArgs = []string{"-selection", "clipboard"}
		m.pasteCmd = "xclip"
		m.pasteArgs = []string{"-selection", "clipboard", "-o"}
	}
}

func (m *Manager) detectLinuxClipboard() {

	if m.isCommandAvailable("wl-copy") && m.isCommandAvailable("wl-paste") {
		m.copyCmd = "wl-copy"
		m.pasteCmd = "wl-paste"
		return
	}

	if m.isCommandAvailable("xclip") {
		m.copyCmd = "xclip"
		m.copyArgs = []string{"-selection", "clipboard"}
		m.pasteCmd = "xclip"
		m.pasteArgs = []string{"-selection", "clipboard", "-o"}
		return
	}

	if m.isCommandAvailable("xsel") {
		m.copyCmd = "xsel"
		m.copyArgs = []string{"-b"}
		m.pasteCmd = "xsel"
		m.pasteArgs = []string{"-b", "-o"}
		return
	}

	m.copyCmd = "xclip"
	m.copyArgs = []string{"-selection", "clipboard"}
	m.pasteCmd = "xclip"
	m.pasteArgs = []string{"-selection", "clipboard", "-o"}
}

func (m *Manager) isCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (m *Manager) CopyToClipboard(text string) error {
	if m.copyCmd == "" {
		return fmt.Errorf("no clipboard copy command available")
	}

	var cmd *exec.Cmd
	if len(m.copyArgs) > 0 {
		args := append(m.copyArgs, "")
		cmd = exec.Command(m.copyCmd, args...)
	} else {
		cmd = exec.Command(m.copyCmd)
	}

	cmd.Stdin = strings.NewReader(text)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard using %s: %w", m.copyCmd, err)
	}
	return nil
}

func (m *Manager) PasteFromClipboard() (string, error) {
	if m.pasteCmd == "" {
		return "", fmt.Errorf("no clipboard paste command available")
	}

	var cmd *exec.Cmd
	if len(m.pasteArgs) > 0 {
		cmd = exec.Command(m.pasteCmd, m.pasteArgs...)
	} else {
		cmd = exec.Command(m.pasteCmd)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to paste from clipboard using %s: %w", m.pasteCmd, err)
	}

	result := strings.TrimSpace(string(output))
	return result, nil
}

func (m *Manager) GetClipboardInfo() string {
	return fmt.Sprintf("Copy: %s %v, Paste: %s %v",
		m.copyCmd, m.copyArgs, m.pasteCmd, m.pasteArgs)
}
