package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const maxNameLen = 50

var (
	// Section headers
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#005c2e"))
	sepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#003d1a"))

	// Dir names
	dirNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff66")).Bold(true)
	dotDirStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#006633")).Bold(true)

	// File names
	fileNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00cc55"))
	dotFileStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#006633"))

	// Subtitle: size · ext
	subStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#004d26")).Italic(true)

	// Symlinks
	symNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffaa"))

	// Count / footer
	countStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#003d1a"))

	// Error
	errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff3334"))
)

type entry struct {
	name  string
	isDir bool
	isSym bool
	size  int64
	dot   bool
	ext   string
}

func main() {
	showAll := false
	filesOnly := false
	target := "."

	for _, arg := range os.Args[1:] {
		switch arg {
		case "-a", "--all":
			showAll = true
		case "-f", "--files":
			filesOnly = true
		case "-h", "--help":
			fmt.Println("Usage: ls [options] [path]")
			fmt.Println("  -a, --all     show hidden files")
			fmt.Println("  -f, --files   files only")
			fmt.Println("  -h, --help    this message")
			return
		default:
			if !strings.HasPrefix(arg, "-") {
				target = arg
			}
		}
	}

	entries, err := os.ReadDir(target)
	if err != nil {
		fmt.Fprintln(os.Stderr, errStyle.Render("  error: "+err.Error()))
		os.Exit(1)
	}

	var dirs, files []entry
	for _, e := range entries {
		name := e.Name()
		isDot := strings.HasPrefix(name, ".")

		if isDot && !showAll {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		isDir := e.IsDir()
		isSym := e.Type()&os.ModeSymlink != 0

		if isSym {
			resolved, err := filepath.EvalSymlinks(filepath.Join(target, name))
			if err == nil {
				ri, err := os.Stat(resolved)
				if err == nil {
					isDir = ri.IsDir()
				}
			}
		}

		ext := ""
		if !isDir {
			ext = strings.TrimPrefix(filepath.Ext(name), ".")
		}

		it := entry{
			name:  name,
			isDir: isDir,
			isSym: isSym,
			size:  info.Size(),
			dot:   isDot,
			ext:   ext,
		}

		if isDir && !filesOnly {
			dirs = append(dirs, it)
		} else if !isDir {
			files = append(files, it)
		}
	}

	sortEntries := func(items []entry) {
		sort.Slice(items, func(i, j int) bool {
			return strings.ToLower(items[i].name) < strings.ToLower(items[j].name)
		})
	}
	sortEntries(dirs)
	sortEntries(files)

	if len(dirs) == 0 && len(files) == 0 {
		fmt.Println(countStyle.Render("  empty"))
		return
	}

	fmt.Println()

	// Dirs section
	if len(dirs) > 0 {
		fmt.Println("  " + headerStyle.Render("dirs") + " " + sepStyle.Render(strings.Repeat("─", 40)))
		fmt.Println()
		for _, d := range dirs {
			name := truncate(d.name)
			switch {
			case d.isSym:
				fmt.Println("  " + symNameStyle.Render(name))
			case d.dot:
				fmt.Println("  " + dotDirStyle.Render(name))
			default:
				fmt.Println("  " + dirNameStyle.Render(name))
			}
		}
		fmt.Println()
	}

	// Files section
	if len(files) > 0 {
		fmt.Println("  " + headerStyle.Render("files") + " " + sepStyle.Render(strings.Repeat("─", 39)))
		fmt.Println()
		for _, f := range files {
			name := truncate(f.name)
			switch {
			case f.isSym:
				fmt.Println("  " + symNameStyle.Render(name))
			case f.dot:
				fmt.Println("  " + dotFileStyle.Render(name))
			default:
				fmt.Println("  " + fileNameStyle.Render(name))
			}

			// Subtitle: size · ext
			parts := []string{}
			parts = append(parts, humanSize(f.size))
			if f.ext != "" {
				parts = append(parts, f.ext)
			}
			fmt.Println("  " + subStyle.Render("  "+strings.Join(parts, " · ")))
		}
		fmt.Println()
	}

	// Footer
	parts := []string{}
	if len(dirs) > 0 {
		s := fmt.Sprintf("%d dir", len(dirs))
		if len(dirs) > 1 {
			s += "s"
		}
		parts = append(parts, s)
	}
	if len(files) > 0 {
		s := fmt.Sprintf("%d file", len(files))
		if len(files) > 1 {
			s += "s"
		}
		parts = append(parts, s)
	}
	fmt.Println("  " + countStyle.Render(strings.Join(parts, ", ")))
	fmt.Println()
}

func truncate(s string) string {
	if len(s) <= maxNameLen {
		return s
	}
	return s[:maxNameLen-1] + "…"
}

func humanSize(b int64) string {
	if b == 0 {
		return "0 B"
	}
	units := []string{"B", "K", "M", "G", "T"}
	i := int(math.Log(float64(b)) / math.Log(1024))
	if i >= len(units) {
		i = len(units) - 1
	}
	val := float64(b) / math.Pow(1024, float64(i))
	if i == 0 {
		return fmt.Sprintf("%d B", b)
	}
	if val >= 10 {
		return fmt.Sprintf("%d %s", int(val), units[i])
	}
	return fmt.Sprintf("%.1f %s", val, units[i])
}
