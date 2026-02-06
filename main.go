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

// Styles
var (
	// Header / separators
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#005c2e"))
	sepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#003d1a"))

	// Type column
	dirTag  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff66")).Bold(true).Render("DIR")
	fileTag = lipgloss.NewStyle().Foreground(lipgloss.Color("#006633")).Render("FILE")
	symTag  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffaa")).Render("LINK")

	// Name
	dirNameStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff66")).Bold(true)
	fileNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00cc55"))
	dotNameStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#006633"))
	symNameStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffaa"))

	// Size
	sizeNumStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00e756"))
	sizeUnitStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#005c2e"))
	sizeDashStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#003d1a"))

	// Error
	errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff3334"))

	// Count
	countStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#005c2e"))
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

		ext := ""
		if !isDir {
			ext = strings.TrimPrefix(filepath.Ext(name), ".")
		}

		items = append(items, entry{
			name:  name,
			isDir: isDir,
			isSym: isSym,
			size:  info.Size(),
			dot:   isDot,
			ext:   ext,
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
		fmt.Println(countStyle.Render("  empty"))
		return
	}

	// Column widths
	maxName := 4 // minimum "NAME"
	maxExt := 3  // minimum "EXT"
	for _, it := range items {
		if len(it.name) > maxName {
			maxName = len(it.name)
		}
		if len(it.ext) > maxExt {
			maxExt = len(it.ext)
		}
	}

	// Header
	hType := headerStyle.Render(pad("TYPE", 4))
	hName := headerStyle.Render(pad("NAME", maxName))
	hExt := headerStyle.Render(pad("EXT", maxExt))
	hSize := headerStyle.Render(padLeft("SIZE", 7))
	sep := sepStyle.Render("  ")

	fmt.Println()
	fmt.Println("  " + hType + sep + hName + sep + hExt + sep + hSize)
	fmt.Println("  " + sepStyle.Render(strings.Repeat("─", 4)) + sep +
		sepStyle.Render(strings.Repeat("─", maxName)) + sep +
		sepStyle.Render(strings.Repeat("─", maxExt)) + sep +
		sepStyle.Render(strings.Repeat("─", 7)))

	dirCount := 0
	fileCount := 0

	for _, it := range items {
		// Type tag
		tag := fileTag
		if it.isDir {
			tag = dirTag
			dirCount++
		} else {
			fileCount++
		}
		if it.isSym {
			tag = symTag
		}
		tCol := pad(tag, 4)

		// Name
		nameStr := pad(it.name, maxName)
		switch {
		case it.isSym:
			nameStr = symNameStyle.Render(nameStr)
		case it.isDir && it.dot:
			nameStr = dotNameStyle.Render(nameStr)
		case it.isDir:
			nameStr = dirNameStyle.Render(nameStr)
		case it.dot:
			nameStr = dotNameStyle.Render(nameStr)
		default:
			nameStr = fileNameStyle.Render(nameStr)
		}

		// Ext
		extStr := pad(it.ext, maxExt)
		if it.ext == "" {
			extStr = dotNameStyle.Render(pad("—", maxExt))
		} else {
			extStr = sizeUnitStyle.Render(extStr)
		}

		// Size
		var sizeStr string
		if it.isDir {
			sizeStr = sizeDashStyle.Render(padLeft("—", 7))
		} else {
			num, unit := humanSizeParts(it.size)
			sizeStr = sizeNumStyle.Render(padLeft(num, 5)) + sizeUnitStyle.Render(padLeft(unit, 2))
		}

		fmt.Println("  " + tCol + sep + nameStr + sep + extStr + sep + sizeStr)
	}

	// Footer
	fmt.Println()
	parts := []string{}
	if dirCount > 0 {
		parts = append(parts, fmt.Sprintf("%d dir", dirCount))
		if dirCount > 1 {
			parts[len(parts)-1] += "s"
		}
	}
	if fileCount > 0 {
		parts = append(parts, fmt.Sprintf("%d file", fileCount))
		if fileCount > 1 {
			parts[len(parts)-1] += "s"
		}
	}
	fmt.Println("  " + countStyle.Render(strings.Join(parts, ", ")))
	fmt.Println()
}

func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

func humanSizeParts(b int64) (string, string) {
	if b == 0 {
		return "0", "B"
	}
	units := []string{"B", "K", "M", "G", "T"}
	i := int(math.Log(float64(b)) / math.Log(1024))
	if i >= len(units) {
		i = len(units) - 1
	}
	val := float64(b) / math.Pow(1024, float64(i))
	if i == 0 {
		return fmt.Sprintf("%d", b), "B"
	}
	if val >= 10 {
		return fmt.Sprintf("%d", int(val)), units[i]
	}
	return fmt.Sprintf("%.1f", val), units[i]
}
