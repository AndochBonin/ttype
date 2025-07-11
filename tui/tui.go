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
	resultsPage
)

type Model struct {
	page          int
	filePicker    filepicker.Model
	selectedFile  string
	fileText      string
	inputText     textinput.Model
	viewText      string
	testPageStats struct {
		totalCorrect            int
		numAttempts             int
		totalLengthCorrectWords int
		numCorrectWords         int
		totalTimeSeconds        int
	}
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
	var inputCmd tea.Cmd

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
			old := m.inputText
			m.inputText, inputCmd = m.inputText.Update(msg)
			if len(m.inputText.Value()) > len(old.Value()) {
				if input := m.inputText.Value(); input[len(input)-1] == m.fileText[len(input)-1] {
					m.addAccuracyStats(1, 1)
				} else {
					m.addAccuracyStats(1, 0)
				}
			}
			if len(m.inputText.Value()) == len(m.fileText) {
				m.page = resultsPage
				m.inputText.Blur()
				// calculate speed
			}
			m.viewText = m.updateViewText()
		}
	case resultsPage:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.page = fileMenu
			}
		}
	}
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
	case resultsPage:
		view = "accuracy: " + fmt.Sprint(m.getAccuracy()) + "%"
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
		panic("could not read file")
	}
	m.fileText = string(text)
	m.viewText = untypedStyle.Render(string(text))
	m.testPageStats.numAttempts = 0
	m.testPageStats.numCorrectWords = 0
	m.testPageStats.totalCorrect = 0
	m.testPageStats.totalLengthCorrectWords = 0
	m.testPageStats.totalTimeSeconds = 0

	inputText := textinput.New()
	inputText.CharLimit = len(m.fileText)
	m.inputText = inputText
	return m.inputText.Focus()
}

func (m *Model) updateViewText() string {
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
		fmt.Sprint(len(inputText)) + "\n\ntotal input length: " + fmt.Sprint(len(viewText)) +
		"\n\naccuracy: " + fmt.Sprint(m.getAccuracy()) + "%"
}

func (m *Model) addAccuracyStats(numAttemptsDelta int, totalCorrect int) {
	m.testPageStats.numAttempts += numAttemptsDelta
	m.testPageStats.totalCorrect += totalCorrect
}

func (m Model) getAccuracy() int {
	return (m.testPageStats.totalCorrect * 100) / m.testPageStats.numAttempts
}

// speed: ((number of characters in correctly typed words + 1 for space after correct word) / 5) -> normalise to 60seconds
// 	m.testPageStats.numCorrectWords = 0
// 	m.testPageStats.totalLengthCorrectWords = 0
// 	m.testPageStats.totalTimeSeconds = 0
