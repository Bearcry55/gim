# 🚀 gim - A Modern Terminal Text Editor

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=for-the-badge)

A beautiful, lightweight, and feature-rich terminal text editor built with Go and Bubble Tea.

[Features](#-features) • [Installation](#-installation) • [Usage](#-usage) • [Configuration](#-configuration) • [Keyboard Shortcuts](#%EF%B8%8F-keyboard-shortcuts)

</div>

---

## ✨ Features

### 🎨 **Beautiful Themes**
- **7 Built-in Themes**: Monokai, Dracula, Nord, Gruvbox, Solarized Dark, Ocean, and Minimal
- Easy theme switching with `Ctrl+T`
- Persistent theme configuration

### 📁 **Integrated File Browser**
- Navigate your file system without leaving the editor
- Create new files and directories on the fly
- Delete files and directories with confirmation prompts
- Visual directory tree with icons

### 💾 **Auto-Save**
- Automatic saving after 2 seconds of inactivity (configurable)
- Visual indicators for save status
- Manual save with `Ctrl+S`
- Auto-save when switching files or quitting

### 🔧 **Highly Configurable**
- TOML-based configuration file (`~/.gim.toml`)
- Customize line numbers, tab width, and auto-save delay
- Settings persist across sessions

### 🎯 **Developer-Friendly**
- Syntax-highlighted line numbers
- Configurable tab width
- Clean, distraction-free interface
- Vim-style navigation in menus

## 📦 Installation

### Prerequisites
- Go 1.16 or higher
- Git

### Option 1: Install from Source

```bash
# Clone the repository
git clone https://github.com/Bearcry55/gim.git
cd gim

# Build and install
go build -o gim
sudo mv gim /usr/local/bin/

# Or install directly with go install
go install github.com/Bearcry55/gim@latest
```

### Option 2: Quick Install Script

```bash
# Download and install in one command
curl -sSL https://github.com/Bearcry55/gim/raw/main/install.sh | bash
```

### Verify Installation

```bash
gim --version
```

## 🚀 Usage

### Opening Files

```bash
# Create or open a file
gim filename.txt

# Start with file browser
gim
```

### First Run
On first launch, gim creates a default configuration file at `~/.gim.toml`. You can edit this file to customize your experience.

## ⚙️ Configuration

The configuration file is located at `~/.gim.toml`:

```toml
# gim Configuration File
# Edit this file to customize your editor

[editor]
show_line_numbers = true  # Show/hide line numbers
tab_width = 4             # Spaces per tab
auto_save_delay = 2       # Auto-save delay in seconds (0 to disable)

# Available themes: monokai, dracula, nord, gruvbox, solarized-dark, ocean, minimal
theme = "monokai"
```

### Theme Options
- **monokai** - Classic dark theme with vibrant colors
- **dracula** - Popular purple-tinted dark theme
- **nord** - Arctic, north-bluish color palette
- **gruvbox** - Retro groove color scheme
- **solarized-dark** - Precision colors for machines and people
- **ocean** - Deep blue oceanic theme
- **minimal** - Clean, distraction-free black and white

## ⌨️ Keyboard Shortcuts

### Editor Mode
| Shortcut | Action |
|----------|--------|
| `Ctrl+B` | Open file browser |
| `Ctrl+T` | Open theme picker |
| `Ctrl+S` | Manual save |
| `Ctrl+Q` | Quit (auto-saves unsaved changes) |
| `Ctrl+C` | Force quit |

### File Browser Mode
| Shortcut | Action |
|----------|--------|
| `↑` / `k` | Navigate up |
| `↓` / `j` | Navigate down |
| `Enter` | Open file or enter directory |
| `n` | Create new file |
| `m` | Create new directory (mkdir) |
| `d` | Delete file/directory (with confirmation) |
| `Ctrl+B` | Return to editor |
| `Ctrl+T` | Open theme picker |
| `Esc` | Cancel current operation |

### Theme Picker Mode
| Shortcut | Action |
|----------|--------|
| `↑` / `k` | Previous theme |
| `↓` / `j` | Next theme |
| `Enter` | Apply selected theme |
| `Esc` | Cancel and return to editor |
| `Ctrl+T` | Toggle theme picker |

### Input Mode (Creating Files/Directories)
| Shortcut | Action |
|----------|--------|
| `Enter` | Confirm creation |
| `Esc` | Cancel |
| `Backspace` | Delete character |

## 🎯 Key Highlights

### What Makes gim Special?

1. **🔄 Smart Auto-Save**: Never lose your work with intelligent auto-save that activates after you stop typing
2. **🎨 Beautiful UI**: Carefully crafted themes that are easy on the eyes during long coding sessions
3. **⚡ Lightning Fast**: Built with Go for exceptional performance
4. **🧩 Zero Dependencies**: Single binary with everything built-in
5. **💼 Professional**: Perfect for quick edits, note-taking, or full development sessions
6. **🔒 Safe Deletes**: Confirmation dialogs prevent accidental file deletion
7. **📊 Status Indicators**: Always know your file status with visual feedback

## 🤝 Contributing

Contributions are welcome! Feel free to:
- Report bugs
- Suggest new features
- Submit pull requests
- Improve documentation

## 📝 License

MIT License - feel free to use gim for any purpose!

## 👨‍💻 Author

**Deep Narayan Banerjee**
- GitHub: [@Bearcry55](https://github.com/Bearcry55)

## 🌟 Show Your Support

If you find gim useful, please consider giving it a ⭐ on GitHub!

---

<div align="center">

**Made with ❤️ and Go**

[Report Bug](https://github.com/Bearcry55/gim/issues) • [Request Feature](https://github.com/Bearcry55/gim/issues)

</div>
