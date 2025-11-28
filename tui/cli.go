package tui

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/M1chlCZ/asciicharm-go/pkg/ascii"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	modePick mode = iota
	modeView
	modePickPathInput
	modeViewSaveName
)

type field int

const (
	fieldResolution field = iota
	fieldContrast
	fieldBrightness
	fieldDither
	fieldCharSet
	fieldColor
	fieldInvert
	fieldCount
)

type Model struct {
	// global
	mode   mode
	w, h   int
	status string
	ready  bool

	// picker
	files        []string
	selectedFile int
	Dir          string
	pathInput    string

	// viewer
	imgPath string
	img     image.Image
	cfg     ascii.ConvertConfig
	res     *ascii.AsciiResult
	err     error
	focused field

	art string

	// save input
	saveKind string
	saveName string
}

func NewPickerModel(dir string, files []string) *Model {
	return &Model{
		mode:         modePick,
		Dir:          dir,
		files:        files,
		selectedFile: 0,
		status:       "Select image with ↑/↓ and press Enter. q to quit.",
	}
}

func NewViewerModel(img image.Image, path string) *Model {
	cfg := ascii.DefaultConfig()
	cfg.Colored = false
	cfg.Dithering = ascii.DitheringFloydSteinberg

	m := Model{
		mode:    modeView,
		img:     img,
		imgPath: path,
		cfg:     cfg,
		focused: fieldResolution,
		status:  "Use arrows to tweak, q to quit, s to save HTML, o to pick another image.",
	}
	m.recompute()
	return &m
}

func (m *Model) recompute() {
	if m.img == nil {
		m.res = nil
		m.err = fmt.Errorf("no image loaded")
		m.art = ""
		return
	}
	res, err := ascii.ConvertImage(m.img, m.cfg)
	if err != nil {
		m.err = err
		m.res = nil
		m.art = ""
		m.status = fmt.Sprintf("error: %v", err)
		return
	}
	m.res = res
	m.err = nil
	m.status = fmt.Sprintf(
		"%s | res=%.2f  ctr=%.2f  brt=%.2f  dither=%s  color=%v invert=%v",
		m.imgPath,
		m.cfg.Resolution,
		m.cfg.Contrast,
		m.cfg.Brightness,
		ditherName(m.cfg.Dithering),
		m.cfg.Colored,
		m.cfg.Inverted,
	)

	m.status = fmt.Sprintf("Editing %s – use arrows to tweak parameters", m.imgPath)

	m.updateArtString()
}

func (m *Model) updateArtString() {
	if m.res == nil {
		m.art = ""
		return
	}
	if m.cfg.Colored {
		m.art = m.res.ToANSI()
	} else {
		m.art = m.res.ToPlainText()
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) adjustCurrent(dir int) {
	step := func(amount float64) float64 {
		if dir < 0 {
			return -amount
		}
		return amount
	}

	switch m.focused {
	case fieldResolution:
		m.cfg.Resolution = clamp(m.cfg.Resolution+step(0.02), 0.05, 1.0)
	case fieldContrast:
		m.cfg.Contrast = clamp(m.cfg.Contrast+step(0.05), 0.1, 3.0)
	case fieldBrightness:
		m.cfg.Brightness = clamp(m.cfg.Brightness+step(0.05), 0.1, 3.0)
	case fieldDither:
		m.cfg.Dithering = cycleDither(m.cfg.Dithering)
	case fieldColor:
		m.cfg.Colored = !m.cfg.Colored
	case fieldInvert:
		m.cfg.Inverted = !m.cfg.Inverted
	case fieldCharSet:
		m.cfg.Charset = (m.cfg.Charset + ascii.CharSet(1)) % ascii.CharSet(4)
	default:
		// do nothing
	}
	m.recompute()
}

func clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modePick:
			return m.updatePickerList(msg)
		case modePickPathInput:
			return m.updatePickerPathInput(msg)
		case modeView:
			return m.updateViewer(msg)
		case modeViewSaveName:
			return m.updateSaveName(msg)
		}
	}
	return m, nil
}

func (m *Model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up":
		if len(m.files) == 0 {
			return m, nil
		}
		m.selectedFile--
		if m.selectedFile < 0 {
			m.selectedFile = len(m.files) - 1
		}
	case "down":
		if len(m.files) == 0 {
			return m, nil
		}
		m.selectedFile++
		if m.selectedFile >= len(m.files) {
			m.selectedFile = 0
		}
	case "enter":
		if len(m.files) == 0 {
			return m, nil
		}
		name := m.files[m.selectedFile]
		path := filepath.Join(m.Dir, name)
		img, err := LoadImage(path)
		if err != nil {
			m.status = fmt.Sprintf("failed to open %s: %v", name, err)
			return m, nil
		}
		vm := NewViewerModel(img, name)
		vm.w, vm.h = m.w, m.h
		vm.ready = m.ready
		return vm, nil
	}
	return m, nil
}

func (m *Model) updatePickerList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up":
		if len(m.files) == 0 {
			return m, nil
		}
		m.selectedFile--
		if m.selectedFile < 0 {
			m.selectedFile = len(m.files) - 1
		}

	case "down":
		if len(m.files) == 0 {
			return m, nil
		}
		m.selectedFile++
		if m.selectedFile >= len(m.files) {
			m.selectedFile = 0
		}

	case "enter":
		if len(m.files) == 0 {
			return m, nil
		}
		name := m.files[m.selectedFile]
		path := filepath.Join(m.Dir, name)
		img, err := LoadImage(path)
		if err != nil {
			m.status = fmt.Sprintf("failed to open %s: %v", name, err)
			return m, nil
		}
		vm := NewViewerModel(img, name)
		vm.Dir = m.Dir
		vm.w, vm.h = m.w, m.h
		vm.ready = m.ready
		return vm, nil

	case "p":
		m.mode = modePickPathInput
		m.pathInput = ""
		m.status = "Type image path and press Enter (Esc to cancel)"
	}
	return m, nil
}

func (m *Model) updatePickerPathInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		// append typed characters
		m.pathInput += string(msg.Runes)

	case tea.KeyBackspace:
		if len(m.pathInput) > 0 {
			m.pathInput = m.pathInput[:len(m.pathInput)-1]
		}

	default:
		switch msg.String() {
		case "enter":
			path := strings.TrimSpace(m.pathInput)
			if path == "" {
				m.status = "Path cannot be empty"
				return m, nil
			}

			img, err := LoadImage(path)
			if err != nil {
				m.status = fmt.Sprintf("failed to open %s: %v", path, err)
				return m, nil
			}

			// update dir based on given path
			dir := filepath.Dir(path)
			base := filepath.Base(path)

			vm := NewViewerModel(img, base)
			vm.Dir = dir
			vm.w, vm.h = m.w, m.h
			vm.ready = m.ready
			return vm, nil

		case "esc":
			m.mode = modePick
			m.pathInput = ""
			m.status = "Select image with ↑/↓ and press Enter. Press p for manual path."
			return m, nil

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) updateSaveName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		m.saveName += string(msg.Runes)

	case tea.KeyBackspace:
		if len(m.saveName) > 0 {
			m.saveName = m.saveName[:len(m.saveName)-1]
		}

	default:
		switch msg.String() {
		case "enter":
			name := strings.TrimSpace(m.saveName)
			if name == "" {
				m.status = "Filename cannot be empty"
				return m, nil
			}

			lower := strings.ToLower(name)
			if m.saveKind == "html" && !strings.HasSuffix(lower, ".html") && !strings.HasSuffix(lower, ".htm") {
				name += ".html"
			}
			if m.saveKind == "md" && !strings.HasSuffix(lower, ".md") && !strings.HasSuffix(lower, ".markdown") {
				name += ".md"
			}

			if m.res != nil {
				var data []byte
				if m.saveKind == "html" {
					data = []byte(m.res.ToHTML())
				} else {
					if m.cfg.Colored {
						data = []byte(m.res.ToMarkdownColored())
					} else {
						data = []byte(m.res.ToMarkdown())
					}
				}

				if err := os.WriteFile(name, data, 0o644); err != nil {
					m.status = fmt.Sprintf("failed to write %s: %v", name, err)
				} else {
					m.status = fmt.Sprintf("saved %s", name)
				}
			}

			m.mode = modeView
			return m, nil

		case "esc":
			m.mode = modeView
			m.status = "Save cancelled"
			return m, nil

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) updateViewer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "o":
		files, err := ListImageFiles(m.Dir)
		if err != nil {
			m.status = fmt.Sprintf("failed to list images: %v", err)
			return m, nil
		}
		pm := NewPickerModel(m.Dir, files)
		pm.w, pm.h = m.w, m.h
		pm.ready = m.ready
		return pm, nil

	// LEFT / RIGHT: select field (move highlight)
	case "left":
		m.focused--
		if m.focused < 0 {
			m.focused = fieldCount - 1
		}
	case "right":
		m.focused++
		if m.focused >= fieldCount {
			m.focused = 0
		}

	// UP / DOWN: adjust current field value
	case "up":
		m.adjustCurrent(+1)
	case "down":
		m.adjustCurrent(-1)

	case "tab":
		m.focused++
		if m.focused >= fieldCount {
			m.focused = 0
		}
	case "shift+tab":
		m.focused--
		if m.focused < 0 {
			m.focused = fieldCount - 1
		}

	case "c":
		m.cfg.Colored = !m.cfg.Colored
		m.updateArtString()
	case "i":
		m.cfg.Inverted = !m.cfg.Inverted
		m.recompute()
	case "d":
		m.cfg.Dithering = cycleDither(m.cfg.Dithering)
		m.recompute()
	case "s":
		if m.res != nil {
			m.mode = modeViewSaveName
			m.saveKind = "html"
			if strings.TrimSpace(m.saveName) == "" {
				base := strings.TrimSuffix(m.imgPath, filepath.Ext(m.imgPath))
				m.saveName = base + "_ascii.html"
			}
			m.status = "Editing HTML filename – type to change, Enter to save, Esc to cancel"
		}

	case "m":
		if m.res != nil {
			m.mode = modeViewSaveName
			m.saveKind = "md"
			if strings.TrimSpace(m.saveName) == "" {
				base := strings.TrimSuffix(m.imgPath, filepath.Ext(m.imgPath))
				m.saveName = base + "_ascii.md"
			}
			m.status = "Editing Markdown filename – type to change, Enter to save, Esc to cancel"
		}
	}

	return m, nil
}

func (m *Model) View() string {
	if !m.ready {
		return "Loading…"
	}

	switch m.mode {
	case modePick, modePickPathInput:
		return m.viewPicker()
	case modeView, modeViewSaveName:
		return m.viewViewer()
	default:
		return "invalid mode"
	}
}

func (m *Model) viewViewer() string {
	isSaving := m.mode == modeViewSaveName

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("69"))

	helpStyle := lipgloss.NewStyle().
		Faint(false).
		Foreground(lipgloss.Color("249"))

	focusBadgeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("11")).
		Padding(0, 1)

	focusedControlStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("15")).
		Padding(0, 1)

	inactiveControlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Faint(true).
		Padding(0, 1)

	saveBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		Padding(0, 1).
		MarginTop(1)

	currentFieldName := func() string {
		switch m.focused {
		case fieldResolution:
			return "Resolution"
		case fieldContrast:
			return "Contrast"
		case fieldBrightness:
			return "Brightness"
		case fieldDither:
			return "Dithering"
		case fieldColor:
			return "Color"
		case fieldInvert:
			return "Invert"
		default:
			return "Unknown"
		}
	}()

	focusBadge := focusBadgeStyle.Render(" ACTIVE: " + currentFieldName + " ")

	controlChip := func(f field, label, value string) string {
		text := label + ": " + value
		if m.focused == f {
			return focusedControlStyle.Render("▶ [" + text + "] ◀")
		}
		return inactiveControlStyle.Render("  " + text + "  ")
	}

	controlsRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		controlChip(fieldResolution, "Res", fmt.Sprintf("%.2f", m.cfg.Resolution)),
		controlChip(fieldContrast, "Ctr", fmt.Sprintf("%.2f", m.cfg.Contrast)),
		controlChip(fieldBrightness, "Brt", fmt.Sprintf("%.2f", m.cfg.Brightness)),
		controlChip(fieldDither, "Dither", ditherName(m.cfg.Dithering)),
		controlChip(fieldCharSet, "Charset", charsetName(m.cfg.Charset)),
		controlChip(fieldColor, "Color", fmt.Sprintf("%v", m.cfg.Colored)),
		controlChip(fieldInvert, "Invert", fmt.Sprintf("%v", m.cfg.Inverted)),
	)

	controlsBlock := lipgloss.JoinVertical(
		lipgloss.Center,
		focusBadge,
		controlsRow,
	)

	art := m.art
	if art == "" {
		if m.err != nil {
			art = fmt.Sprintf("error: %v", m.err)
		} else {
			art = "no result yet"
		}
	}

	maxWidth := m.w - 4
	if maxWidth < 20 {
		maxWidth = 20
	}

	artFrame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		MarginTop(1).
		MaxWidth(maxWidth).
		Render(art)

	var help string
	if isSaving {
		help = helpStyle.Render(
			"Saving " + strings.ToUpper(m.saveKind) + " – type filename, Enter save, Esc cancel",
		)
	} else {
		help = helpStyle.Render(
			"←/→ select control   ↑/↓ change value   c color   i invert   d dither   s save html   m save markdown   o open image   q quit",
		)
	}

	rows := []string{
		titleStyle.Render("ASCII Image Tuner – " + m.imgPath),
		artFrame,
		controlsBlock,
		help,
	}

	if isSaving {
		cursor := "_"
		saveInfo := fmt.Sprintf("Save %s file", strings.ToUpper(m.saveKind))
		nameLine := "Name: " + m.saveName + cursor
		saveBox := saveBoxStyle.Render(saveInfo + "\n" + nameLine)
		rows = append(rows, saveBox)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		rows...,
	)
}

func (m *Model) viewPicker() string {
	isPathInput := m.mode == modePickPathInput

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212"))

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	pathBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1).
		MarginTop(1)

	var b strings.Builder

	b.WriteString(titleStyle.Render("Select Image"))
	b.WriteByte('\n')

	help := "↑/↓ move  Enter open  p manual path  q quit"
	b.WriteString(help)
	b.WriteByte('\n')
	b.WriteByte('\n')

	if len(m.files) == 0 {
		b.WriteString("No image files found in " + m.Dir)
		b.WriteByte('\n')
	} else {
		maxRows := m.h - 5
		if maxRows < 3 {
			maxRows = 3
		}
		start := 0
		if m.selectedFile >= maxRows {
			start = m.selectedFile - maxRows + 1
		}
		end := start + maxRows
		if end > len(m.files) {
			end = len(m.files)
		}

		for i := start; i < end; i++ {
			name := m.files[i]
			if i == m.selectedFile {
				b.WriteString(selectedStyle.Render("> " + name))
			} else {
				b.WriteString(itemStyle.Render("  " + name))
			}
			b.WriteByte('\n')
		}
	}

	if isPathInput {
		cursor := "_"
		pathDisplay := "Path: " + m.pathInput + cursor
		info := "Manual path input (Enter: open, Esc: cancel)"
		box := pathBoxStyle.Render(info + "\n" + pathDisplay)
		b.WriteString(box)
		b.WriteByte('\n')
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		MarginTop(1)
	b.WriteString(statusStyle.Render(m.status))
	b.WriteByte('\n')

	return b.String()
}
