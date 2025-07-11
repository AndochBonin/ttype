package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/textinput"
	//"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	fileMenu = iota
	testPage
)

type Model struct {
	page         int
	filePicker   filepicker.Model
	selectedFile string
	fileText     string
	inputText    textinput.Model
	viewText     string
	//timer timer.Model
}

var (
	untypedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	correctStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	wrongStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	fitStyle     = lipgloss.NewStyle().Width(100)
)

func initialModel() (Model, error) {
	m := Model{}
	m.page = fileMenu
	m.filePicker = filepicker.New()
	m.filePicker.SetHeight(0)
	m.filePicker.ShowHidden = false
	m.filePicker.ShowSize = true
	m.filePicker.ShowPermissions = false
	m.filePicker.AllowedTypes = []string{".txt", ".md"}
	//exePath, err := os.Executable()
	// if err != nil {
	// 	return m, err
	// }
	//m.filePicker.CurrentDirectory = filepath.Dir(exePath)
	currentDirectory, err := filepath.Abs("./")
	if err != nil {
		return Model{}, nil
	}
	m.filePicker.CurrentDirectory = currentDirectory
	return m, nil
}

func (m Model) Init() tea.Cmd {

	return tea.Batch(m.filePicker.Init(), tea.SetWindowTitle("ttype"))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var keyCmd tea.Cmd
	var cmd tea.Cmd

	switch m.page {
	case fileMenu:
		m.filePicker, cmd = m.filePicker.Update(msg)
		didSelect, path := m.filePicker.DidSelectFile(msg)
		if didSelect {
			m.selectedFile = path
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			keyCmd = m.fileMenuKeyHandler(msg.String(), didSelect)
		}
	case testPage:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			keyCmd = m.testPageKeyHandler(msg.String())
		}
	}
	var inputCmd tea.Cmd
	m.inputText, inputCmd = m.inputText.Update(msg)
	m.viewText = m.updateViewText()
	cmds = append(cmds, cmd, keyCmd, inputCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	title := "ttype"
	var view string
	switch m.page {
	case fileMenu:
		view = "selected: "
		view += m.selectedFile + "\n\n"
		view += m.filePicker.View()
	case testPage:
		view = "typing test page" + "\n\n"
		view += fitStyle.Render(m.viewText) + "\n"
	}
	return title + "\n\n" + view
}

func Run() error {
	m, initErr := initialModel()
	if initErr != nil {
		return initErr
	}
	p := tea.NewProgram(m)
	_, runErr := p.Run()
	return runErr
}

func (m *Model) fileMenuKeyHandler(msg string, didSelect bool) tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.filePicker, cmd = m.filePicker.Update(msg)
	cmds = append(cmds, cmd)
	switch msg {
	case "ctrl+c":
		return tea.Quit
	case "enter":
		if didSelect {
			cmds = append(cmds, m.testPageInit())
			m.page = testPage
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) testPageKeyHandler(msg string) tea.Cmd {
	switch msg {
	case "ctrl+c":
		return tea.Quit
	case "esc":
		m.page = fileMenu
	}
	return nil
}

func (m *Model) testPageInit() tea.Cmd {
	text, err := os.ReadFile(m.selectedFile)
	if err != nil {
		m.fileText = ""
		m.viewText = ""
	} else {
		m.fileText = string(text)
		m.viewText = untypedStyle.Render(string(text))
	}
	inputText := textinput.New()
	inputText.CharLimit = len(m.fileText)
	m.inputText = inputText
	return m.inputText.Focus()
}

func (m Model) updateViewText() string {
	inputText := m.inputText.Value()
	fileText := m.fileText
	viewText := ""
	currentSection := ""
	currentStyle := lipgloss.NewStyle().Foreground(lipgloss.NoColor{})
	var updateSection = func(s lipgloss.Style, idx int) {
		if s.GetForeground() == currentStyle.GetForeground() {
			currentSection += string(fileText[idx])
		} else {
			viewText += currentStyle.Render(currentSection)
			currentSection = string(fileText[idx])
			currentStyle = s
		}
	}

	for i := range inputText {
		if inputText[i] == fileText[i] {
			updateSection(correctStyle, i)
		} else {
			updateSection(wrongStyle, i)
		}
	}
	viewText += currentStyle.Render(currentSection)
	viewText += untypedStyle.Render(fileText[len(inputText):])
	return viewText + "\n" + "file length: " + fmt.Sprint(len(fileText)) + "\n\nvisible input length: " +
		fmt.Sprint(len(inputText)) + "\n\ntotal input length: " + fmt.Sprint(len(viewText))
}
