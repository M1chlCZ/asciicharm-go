package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/M1chlCZ/asciicharm-go/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var pathFlag string
	flag.StringVar(&pathFlag, "i", "", "input image path (optional, otherwise pick in TUI)")
	flag.Parse()

	var m *tui.Model

	if strings.TrimSpace(pathFlag) != "" {
		img, err := tui.LoadImage(pathFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		m = tui.NewViewerModel(img, filepath.Base(pathFlag))
		m.Dir = filepath.Dir(pathFlag)
	} else {
		// start in picker
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		files, err := tui.ListImageFiles(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		m = tui.NewPickerModel(dir, files)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui error:", err)
		os.Exit(1)
	}
}
