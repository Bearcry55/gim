package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Name       string
	Text       string
	LineNumber string
	StatusBar  string
	Background string
	CursorLine string
	Border     string
}

var themes = map[string]Theme{
	"monokai": {
		Name:       "Monokai",
		Text:       "248",
		LineNumber: "102",
		StatusBar:  "197",
		Background: "235",
		CursorLine: "237",
		Border:     "102",
	},
	"dracula": {
		Name:       "Dracula",
		Text:       "253",
		LineNumber: "61",
		StatusBar:  "141",
		Background: "235",
		CursorLine: "236",
		Border:     "61",
	},
	"nord": {
		Name:       "Nord",
		Text:       "254",
		LineNumber: "67",
		StatusBar:  "109",
		Background: "233",
		CursorLine: "234",
		Border:     "67",
	},
	"gruvbox": {
		Name:       "Gruvbox",
		Text:       "223",
		LineNumber: "245",
		StatusBar:  "214",
		Background: "235",
		CursorLine: "237",
		Border:     "245",
	},
	"solarized-dark": {
		Name:       "Solarized Dark",
		Text:       "254",
		LineNumber: "240",
		StatusBar:  "33",
		Background: "234",
		CursorLine: "235",
		Border:     "240",
	},
	"ocean": {
		Name:       "Ocean",
		Text:       "159",
		LineNumber: "24",
		StatusBar:  "39",
		Background: "233",
		CursorLine: "234",
		Border:     "24",
	},
	"minimal": {
		Name:       "Minimal",
		Text:       "15",
		LineNumber: "240",
		StatusBar:  "252",
		Background: "0",
		CursorLine: "235",
		Border:     "240",
	},
}

type Config struct {
	Editor EditorConfig `toml:"editor"`
	Theme  string       `toml:"theme"`
}

type EditorConfig struct {
	ShowLineNumbers bool `toml:"show_line_numbers"`
	TabWidth        int  `toml:"tab_width"`
	AutoSaveDelay   int  `toml:"auto_save_delay"` // seconds
}

type fileItem struct {
	name  string
	path  string
	isDir bool
}

type autoSaveMsg struct{}
type saveCompletedMsg struct {
	success bool
	err     error
}

type model struct {
	textarea          textarea.Model
	filename          string
	currentDir        string
	files             []fileItem
	selectedFile      int
	showBrowser       bool
	err               error
	config            Config
	theme             Theme
	width             int
	height            int
	availableThemes   []string
	selectedTheme     int
	showThemePicker   bool
	inputMode         bool
	inputPrompt       string
	inputValue        string
	createDirectory   bool
	showDeleteConfirm bool
	deleteTarget      fileItem
	lastSavedContent  string
	lastModified      time.Time
	autoSaveTimer     *time.Timer
	saving            bool
}

func loadConfig() Config {
	defaultConfig := Config{
		Editor: EditorConfig{
			ShowLineNumbers: true,
			TabWidth:        4,
			AutoSaveDelay:   2,
		},
		Theme: "monokai",
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return defaultConfig
	}

	configPath := filepath.Join(home, ".gim.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		createDefaultConfig(configPath, defaultConfig)
		fmt.Printf("Created default config at: %s\n", configPath)
		return defaultConfig
	}

	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		fmt.Printf("Warning: Could not parse config file, using defaults: %v\n", err)
		return defaultConfig
	}

	// Set default auto-save delay if not specified
	if config.Editor.AutoSaveDelay == 0 {
		config.Editor.AutoSaveDelay = 2
	}

	return config
}

func createDefaultConfig(path string, config Config) {
	file, err := os.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	configContent := `# gim Configuration File
# Edit this file to customize your editor

[editor]
show_line_numbers = true  # Show/hide line numbers
tab_width = 4             # Spaces per tab
auto_save_delay = 2       # Auto-save delay in seconds (0 to disable)

# Available themes: monokai, dracula, nord, gruvbox, solarized-dark, ocean, minimal
theme = "monokai"
`
	file.WriteString(configContent)
}

func loadDirectory(path string) []fileItem {
	entries, err := os.ReadDir(path)
	if err != nil {
		return []fileItem{}
	}

	var items []fileItem

	items = append(items, fileItem{
		name:  "..",
		path:  filepath.Dir(path),
		isDir: true,
	})

	for _, entry := range entries {
		items = append(items, fileItem{
			name:  entry.Name(),
			path:  filepath.Join(path, entry.Name()),
			isDir: entry.IsDir(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].isDir != items[j].isDir {
			return items[i].isDir
		}
		return items[i].name < items[j].name
	})

	return items
}

func getThemesList() []string {
	keys := make([]string, 0, len(themes))
	for k := range themes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func initialModel(filename string, config Config) model {
	ti := textarea.New()
	ti.Placeholder = "Start typing... (Ctrl+B: Browser | Ctrl+T: Themes)"
	ti.Focus()
	ti.ShowLineNumbers = config.Editor.ShowLineNumbers
	ti.CharLimit = 0
	ti.SetHeight(20)
	ti.SetWidth(80)

	// Load theme
	theme, exists := themes[config.Theme]
	if !exists {
		theme = themes["monokai"]
	}

	ti.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color(theme.CursorLine))
	ti.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
	ti.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.LineNumber))
	ti.FocusedStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color(theme.Background))

	var currentDir string
	var lastSavedContent string
	if filename != "" {
		absPath, _ := filepath.Abs(filename)
		currentDir = filepath.Dir(absPath)

		content, err := os.ReadFile(filename)
		if err == nil {
			ti.SetValue(string(content))
			lastSavedContent = string(content)
		}
	} else {
		currentDir, _ = os.Getwd()
	}

	themeList := getThemesList()
	selectedThemeIdx := 0
	for i, t := range themeList {
		if t == config.Theme {
			selectedThemeIdx = i
			break
		}
	}

	m := model{
		textarea:         ti,
		filename:         filename,
		currentDir:       currentDir,
		files:            loadDirectory(currentDir),
		selectedFile:     0,
		showBrowser:      false,
		config:           config,
		theme:            theme,
		err:              nil,
		availableThemes:  themeList,
		selectedTheme:    selectedThemeIdx,
		showThemePicker:  false,
		inputMode:        false,
		inputPrompt:      "",
		inputValue:       "",
		createDirectory:  false,
		showDeleteConfirm: false,
		deleteTarget:      fileItem{},
		lastSavedContent: lastSavedContent,
		lastModified:     time.Now(),
		autoSaveTimer:    nil,
		saving:           false,
	}

	// Force input mode for new files when no filename provided
	if filename == "" {
		m.showBrowser = true
		m.inputMode = true
		m.inputPrompt = "New file: "
		m.createDirectory = false
		m.files = loadDirectory(currentDir)
		m.selectedFile = 0
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tea.EnterAltScreen)
}

func (m *model) applyTheme(themeName string) {
	theme, exists := themes[themeName]
	if !exists {
		return
	}

	m.theme = theme
	m.config.Theme = themeName

	m.textarea.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color(theme.CursorLine))
	m.textarea.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
	m.textarea.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.LineNumber))
	m.textarea.FocusedStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color(theme.Background))

	// Save theme to config
	home, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(home, ".gim.toml")
		file, err := os.Create(configPath)
		if err == nil {
			defer file.Close()
			encoder := toml.NewEncoder(file)
			encoder.Encode(m.config)
		}
	}
}

func (m *model) hasUnsavedChanges() bool {
	return m.filename != "" && m.textarea.Value() != m.lastSavedContent
}

func (m *model) saveFile() tea.Cmd {
	if m.filename == "" || m.saving {
		return nil
	}

	m.saving = true
	content := m.textarea.Value()
	filename := m.filename

	return func() tea.Msg {
		// Create directory if it doesn't exist
		dir := filepath.Dir(filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return saveCompletedMsg{success: false, err: err}
		}

		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			return saveCompletedMsg{success: false, err: err}
		}
		return saveCompletedMsg{success: true, err: nil}
	}
}

func (m *model) scheduleAutoSave() tea.Cmd {
	if m.config.Editor.AutoSaveDelay == 0 || m.filename == "" {
		return nil
	}

	// Cancel existing timer
	if m.autoSaveTimer != nil {
		m.autoSaveTimer.Stop()
	}

	duration := time.Duration(m.config.Editor.AutoSaveDelay) * time.Second
	m.autoSaveTimer = time.NewTimer(duration)

	return func() tea.Msg {
		<-m.autoSaveTimer.C
		return autoSaveMsg{}
	}
}

func (m *model) openFile(filePath string) {
	// Auto-save current file before switching
	if m.hasUnsavedChanges() && m.filename != "" {
		os.WriteFile(m.filename, []byte(m.textarea.Value()), 0644)
		m.lastSavedContent = m.textarea.Value()
	}

	content, err := os.ReadFile(filePath)
	if err == nil {
		m.textarea.SetValue(string(content))
		m.filename = filePath
		m.lastSavedContent = string(content)
		m.err = nil
	} else {
		m.err = err
	}
	m.showBrowser = false
	m.textarea.Focus()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(msg.Height - 4)
		return m, nil

	case autoSaveMsg:
		if m.hasUnsavedChanges() {
			return m, m.saveFile()
		}
		return m, nil

	case saveCompletedMsg:
		m.saving = false
		if msg.success {
			m.lastSavedContent = m.textarea.Value()
			m.lastModified = time.Now()
			m.err = nil
			// Reload files in browser
			if m.currentDir != "" {
				m.files = loadDirectory(m.currentDir)
			}
		} else {
			m.err = msg.err
		}
		return m, nil

	case tea.KeyMsg:
		// Handle theme picker first
		if m.showThemePicker {
			switch msg.String() {
			case "ctrl+c", "ctrl+q":
				// Auto-save before quitting
				if m.hasUnsavedChanges() && m.filename != "" {
					os.WriteFile(m.filename, []byte(m.textarea.Value()), 0644)
				}
				return m, tea.Quit
			case "esc", "ctrl+t":
				m.showThemePicker = false
				m.textarea.Focus()
			case "up", "k":
				if m.selectedTheme > 0 {
					m.selectedTheme--
				}
			case "down", "j":
				if m.selectedTheme < len(m.availableThemes)-1 {
					m.selectedTheme++
				}
			case "enter":
				m.applyTheme(m.availableThemes[m.selectedTheme])
				m.showThemePicker = false
				m.textarea.Focus()
			}
			return m, nil
		}

		// Handle file browser
		if m.showBrowser {
			// Delete confirmation dialog
			if m.showDeleteConfirm {
				switch msg.String() {
				case "ctrl+c", "ctrl+q":
					// Auto-save before quitting
					if m.hasUnsavedChanges() && m.filename != "" {
						os.WriteFile(m.filename, []byte(m.textarea.Value()), 0644)
					}
					return m, tea.Quit
				case "y", "Y":
					var err error
					if m.deleteTarget.isDir {
						err = os.RemoveAll(m.deleteTarget.path)
					} else {
						err = os.Remove(m.deleteTarget.path)
					}
					if err != nil {
						m.err = err
					} else {
						m.files = loadDirectory(m.currentDir)
						if m.selectedFile >= len(m.files) {
							m.selectedFile = len(m.files) - 1
						}
						if m.selectedFile < 0 {
							m.selectedFile = 0
						}
						m.err = nil
					}
					m.showDeleteConfirm = false
					m.deleteTarget = fileItem{}
				case "n", "N", "esc":
					m.showDeleteConfirm = false
					m.deleteTarget = fileItem{}
				}
				return m, nil
			}

			// Input mode for creating files/directories
			if m.inputMode {
				switch msg.String() {
				case "ctrl+c", "ctrl+q":
					// Auto-save before quitting
					if m.hasUnsavedChanges() && m.filename != "" {
						os.WriteFile(m.filename, []byte(m.textarea.Value()), 0644)
					}
					return m, tea.Quit
				case "esc":
					m.inputMode = false
					m.inputValue = ""
					m.inputPrompt = ""
					m.createDirectory = false
				case "enter":
					if m.inputValue == "" {
						m.inputMode = false
						m.createDirectory = false
						return m, nil
					}
					newPath := filepath.Join(m.currentDir, m.inputValue)
					if _, statErr := os.Stat(newPath); statErr == nil {
						m.err = fmt.Errorf("file or directory already exists")
						m.inputMode = false
						m.inputValue = ""
						m.createDirectory = false
						return m, nil
					}

					var err error
					if m.createDirectory {
						err = os.MkdirAll(newPath, 0755)
					} else {
						err = os.WriteFile(newPath, []byte{}, 0644)
					}

					if err != nil {
						m.err = err
					} else {
						m.files = loadDirectory(m.currentDir)
						m.selectedFile = 0
						for i, f := range m.files {
							if f.path == newPath {
								m.selectedFile = i
								break
							}
						}
						m.err = nil
						// Auto-open the new file if it's a file (not directory)
						if !m.createDirectory {
							m.openFile(newPath)
						}
					}
					m.inputMode = false
					m.inputValue = ""
					m.inputPrompt = ""
					m.createDirectory = false
				case "backspace", "\b":
					if len(m.inputValue) > 0 {
						m.inputValue = m.inputValue[:len(m.inputValue)-1]
					}
				default:
					if len(msg.String()) == 1 && msg.String()[0] >= ' ' && msg.String()[0] <= '~' {
						m.inputValue += msg.String()
					}
				}
				return m, nil
			}

			// Normal browser navigation
			switch msg.String() {
			case "ctrl+c", "ctrl+q":
				// Auto-save before quitting
				if m.hasUnsavedChanges() && m.filename != "" {
					os.WriteFile(m.filename, []byte(m.textarea.Value()), 0644)
				}
				return m, tea.Quit
			case "ctrl+b":
				m.showBrowser = false
				m.textarea.Focus()
			case "ctrl+t":
				m.showBrowser = false
				m.showThemePicker = true
				m.textarea.Blur()
			case "n":
				m.inputPrompt = "New file: "
				m.inputValue = ""
				m.inputMode = true
				m.createDirectory = false
			case "m":
				m.inputPrompt = "New directory: "
				m.inputValue = ""
				m.inputMode = true
				m.createDirectory = true
			case "d":
				selected := m.files[m.selectedFile]
				if selected.name == ".." {
					m.err = fmt.Errorf("cannot delete parent directory")
				} else {
					m.deleteTarget = selected
					m.showDeleteConfirm = true
				}
			case "up", "k":
				if m.selectedFile > 0 {
					m.selectedFile--
				}
			case "down", "j":
				if m.selectedFile < len(m.files)-1 {
					m.selectedFile++
				}
			case "enter":
				selected := m.files[m.selectedFile]
				if selected.isDir {
					m.currentDir = selected.path
					m.files = loadDirectory(m.currentDir)
					m.selectedFile = 0
				} else {
					m.openFile(selected.path)
				}
			}
			return m, nil
		}

		// Editor mode
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			// Auto-save before quitting
			if m.hasUnsavedChanges() && m.filename != "" {
				os.WriteFile(m.filename, []byte(m.textarea.Value()), 0644)
			}
			return m, tea.Quit
		case "ctrl+b":
			m.showBrowser = true
			m.textarea.Blur()
			return m, nil
		case "ctrl+t":
			m.showThemePicker = true
			m.textarea.Blur()
			return m, nil
		case "ctrl+s":
			// Manual save
			return m, m.saveFile()
		}

		// Track changes for auto-save
		oldValue := m.textarea.Value()
		m.textarea, _ = m.textarea.Update(msg)
		newValue := m.textarea.Value()

		// If content changed, schedule auto-save
		if oldValue != newValue && m.filename != "" {
			cmds = append(cmds, m.scheduleAutoSave())
		}

		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m model) View() string {
	lines := strings.Count(m.textarea.Value(), "\n") + 1

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.StatusBar)).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Bold(true)

	var status string
	if m.err != nil {
		status = fmt.Sprintf("❌ Error: %v", m.err)
	} else {
		autoSaveStatus := ""
		if m.hasUnsavedChanges() {
			if m.saving {
				autoSaveStatus = " 💾 Saving..."
			} else {
				autoSaveStatus = " ⏱️ Auto-save pending"
			}
		} else if m.filename != "" {
			timeSinceSave := time.Since(m.lastModified)
			if timeSinceSave < 3*time.Second {
				autoSaveStatus = " ✅ Saved"
			}
		}

		if m.showBrowser {
			if m.inputMode {
				status = fmt.Sprintf("📁 %s | %s%s | Enter: Create | Esc: Cancel", filepath.Base(m.currentDir), m.inputPrompt, m.inputValue)
			} else {
				status = fmt.Sprintf("📁 %s%s | Lines: %d | ↑↓: Navigate | Enter: Open | N: New | M: New Dir | D: Delete", filepath.Base(m.currentDir), autoSaveStatus, lines)
			}
		} else if m.showThemePicker {
			status = fmt.Sprintf("🎨 Theme: %s%s | Lines: %d | ↑↓: Select | Enter: Apply | Esc: Cancel", m.theme.Name, autoSaveStatus, lines)
		} else {
			status = fmt.Sprintf("📝 %s%s | Lines: %d | Theme: %s | Ctrl+B: Browser | Ctrl+T: Themes | Ctrl+Q: Quit", filepath.Base(m.filename), autoSaveStatus, lines, m.theme.Name)
		}
	}

	statusBar := statusStyle.Render(status)

	// Theme picker view
	if m.showThemePicker {
		pickerStyle := lipgloss.NewStyle().
			Width(50).
			Height(m.height - 4).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(m.theme.Border)).
			Padding(1, 2).
			Align(lipgloss.Center)

		var pickerContent strings.Builder
		pickerContent.WriteString("🎨 Select Theme\n\n")

		for i, themeName := range m.availableThemes {
			theme := themes[themeName]
			prefix := "  "
			if i == m.selectedTheme {
				prefix = "▶ "
			}

			line := fmt.Sprintf("%s%s", prefix, theme.Name)
			if i == m.selectedTheme {
				line = lipgloss.NewStyle().
					Foreground(lipgloss.Color(m.theme.StatusBar)).
					Bold(true).
					Render(line)
			}
			pickerContent.WriteString(line + "\n")
		}

		picker := pickerStyle.Render(pickerContent.String())

		centered := lipgloss.Place(
			m.width,
			m.height-4,
			lipgloss.Center,
			lipgloss.Center,
			picker,
		)

		return fmt.Sprintf("%s\n%s", centered, statusBar)
	}

	// File browser view
	if m.showBrowser {
		// Delete confirmation dialog
		if m.showDeleteConfirm {
			confirmStyle := lipgloss.NewStyle().
				Width(60).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("196")).
				Background(lipgloss.Color("235")).
				Padding(2, 4).
				Align(lipgloss.Center)

			itemType := "file"
			if m.deleteTarget.isDir {
				itemType = "directory"
			}

			confirmText := fmt.Sprintf(
				"⚠️  Delete Confirmation\n\n"+
					"Are you sure you want to delete this %s?\n\n"+
					"%s: %s\n\n"+
					"This action cannot be undone!\n\n"+
					"[Y] Yes, delete   [N] No, cancel",
				itemType,
				lipgloss.NewStyle().Bold(true).Render(strings.ToUpper(itemType)),
				lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.StatusBar)).Bold(true).Render(m.deleteTarget.name),
			)

			confirm := confirmStyle.Render(confirmText)

			centered := lipgloss.Place(
				m.width,
				m.height-4,
				lipgloss.Center,
				lipgloss.Center,
				confirm,
			)

			return fmt.Sprintf("%s\n%s", centered, statusBar)
		}

		// Browser file list
		browserStyle := lipgloss.NewStyle().
			Width(50).
			Height(m.height - 4).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(m.theme.Border)).
			Padding(1, 1)

		var browserContent strings.Builder
		browserContent.WriteString(fmt.Sprintf("📁 %s\n\n", filepath.Base(m.currentDir)))

		if m.inputMode {
			cursor := "█"
			inputLine := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%s%s%s", m.inputPrompt, m.inputValue, cursor))
			browserContent.WriteString(inputLine + "\n")
			browserContent.WriteString("\n")
		}

		maxDisplay := m.height - 8
		if m.inputMode {
			maxDisplay -= 3
		}
		if maxDisplay < 1 {
			maxDisplay = 5
		}

		for i, file := range m.files {
			if i >= maxDisplay {
				break
			}
			prefix := "  "
			if i == m.selectedFile {
				prefix = "▶ "
			}

			icon := "📄"
			if file.isDir {
				icon = "📁"
			}

			line := fmt.Sprintf("%s%s %s", prefix, icon, file.name)
			if i == m.selectedFile {
				line = lipgloss.NewStyle().
					Foreground(lipgloss.Color(m.theme.StatusBar)).
					Bold(true).
					Render(line)
			}
			browserContent.WriteString(line + "\n")
		}

		browser := browserStyle.Render(browserContent.String())

		centered := lipgloss.Place(
			m.width,
			m.height-4,
			lipgloss.Center,
			lipgloss.Center,
			browser,
		)

		return fmt.Sprintf("%s\n%s", centered, statusBar)
	}

	// Editor view
	return fmt.Sprintf("%s\n%s", m.textarea.View(), statusBar)
}

func main() {
	var filename string
	if len(os.Args) >= 2 {
		filename = os.Args[1]
	}

	config := loadConfig()
	m := initialModel(filename, config)

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
