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

var (
	dirStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff66")).Bold(true)
	fileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00e756"))
	sizeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#005c2e"))
	dotStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#004d26"))
	symStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffaa"))
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff3334"))
)

type entry struct {
	name  string
	isDir bool
	isSym bool
	size  int64
	dot   bool
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
		fmt.Fprintln(os.Stderr, errStyle.Render("error: "+err.Error()))
		os.Exit(1)
	}

	var items []entry
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

		if filesOnly && isDir {
			continue
		}

		items = append(items, entry{
			name:  name,
			isDir: isDir,
			isSym: isSym,
			size:  info.Size(),
			dot:   isDot,
		})
	}

	// Sort: dirs first, then files, alphabetical within each
	sort.Slice(items, func(i, j int) bool {
		if items[i].isDir != items[j].isDir {
			return items[i].isDir
		}
		return strings.ToLower(items[i].name) < strings.ToLower(items[j].name)
	})

	if len(items) == 0 {
		return
	}

	// Calculate column widths
	maxName := 0
	for _, it := range items {
		display := it.name
		if it.isDir {
			display += "/"
		}
		if len(display) > maxName {
			maxName = len(display)
		}
	}

	for _, it := range items {
		name := it.name
		sizeStr := ""

		if it.isDir {
			name += "/"
		}

		if !it.isDir {
			sizeStr = humanSize(it.size)
		} else {
			sizeStr = "   -"
		}

		padded := name + strings.Repeat(" ", maxName-len(name)+2)

		var line string
		switch {
		case it.isSym:
			line = symStyle.Render(padded) + sizeStyle.Render(sizeStr)
		case it.isDir && it.dot:
			line = dotStyle.Render(padded) + sizeStyle.Render(sizeStr)
		case it.isDir:
			line = dirStyle.Render(padded) + sizeStyle.Render(sizeStr)
		case it.dot:
			line = dotStyle.Render(padded) + sizeStyle.Render(sizeStr)
		default:
			line = fileStyle.Render(padded) + sizeStyle.Render(sizeStr)
		}

		fmt.Println(line)
	}
}

func humanSize(b int64) string {
	if b == 0 {
		return "   0B"
	}
	units := []string{"B", "K", "M", "G", "T"}
	i := int(math.Log(float64(b)) / math.Log(1024))
	if i >= len(units) {
		i = len(units) - 1
	}
	val := float64(b) / math.Pow(1024, float64(i))
	if i == 0 {
		return fmt.Sprintf("%4dB", b)
	}
	if val >= 10 {
		return fmt.Sprintf("%4d%s", int(val), units[i])
	}
	return fmt.Sprintf("%4.1f%s", val, units[i])
}
