# TranslateGo

A modern, terminal-based multi-service translation tool written in Go. TranslateGo provides a beautiful TUI (Text User Interface) for translating text using various online translation services like Google Translate, DeepL, Reverso, MyMemory, Lingva, and OpenAI.

## Features

- 🌍 **Multi-service support**: Choose from multiple translation providers
- 🎨 **Beautiful TUI**: Modern terminal interface with Bubbletea
- 📋 **Clipboard integration**: Paste from clipboard (Ctrl+V) and copy translations (Alt+1/2/3)
- 💾 **Caching**: Avoid redundant API calls with intelligent caching
- ⚡ **Rate limiting**: Respect API limits with built-in rate limiting
- 🔄 **Retry mechanism**: Automatic retries on failures
- 🌐 **Multi-language**: Support for 10+ languages
- 🔑 **API key management**: Secure storage of API keys (e.g., for OpenAI)

## Installation

### From AUR (Arch Linux)

```bash
yay -S translatego
```

### Manual Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/cstrp/translatego.git
   cd translatego
   ```

2. Build the binary:

   ```bash
   go build -o translatego ./cmd/main.go
   ```

3. Install:

   ```bash
   sudo cp translatego /usr/local/bin/
   ```

## Usage

Run the application:

```bash
translatego
```

### Interface Guide

1. **Language Selection**: Choose your target language from the list
2. **API Configuration**: Set up API keys for services that require them (OpenAI)
3. **Translation**: Enter text to translate and press Enter
4. **Copy Results**: Use Alt+1/2/3 to copy specific translations to clipboard

### Keyboard Shortcuts

- `↑/↓` or `k/j`: Navigate menus
- `Enter`: Select/Translate
- `Ctrl+V`: Paste from clipboard
- `Alt+1/2/3`: Copy translation to clipboard
- `Alt+5`: Cycle layout
- `q` or `Ctrl+C`: Quit

## Supported Languages

- English (en)
- Russian (ru)
- German (de)
- French (fr)
- Spanish (es)
- Italian (it)
- Japanese (ja)
- Chinese (zh)
- Korean (ko)
- Arabic (ar)

## Supported Services

- Google Translate
- DeepL
- Reverso
- MyMemory
- Lingva
- OpenAI (requires API key)

## Configuration

Configuration is stored in `~/.config/translatego/config.json`. API keys are securely stored and only required for services like OpenAI.

## Dependencies

- Go 1.19+
- Internet connection for translations

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Author

@cstrp-dev
