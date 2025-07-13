package tui

import (
	"fmt"
	"os"

	//"path/filepath"
	"github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/textinput"
	//"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	testPage = iota
	resultsPage
)

type Model struct {
	page         int
	fileText     string
	previousText textinput.Model
	inputText    textinput.Model
	viewText     string

	totalCorrect            int
	numAttempts             int
	totalLengthCorrectWords int
	stopBackspaceIdx        int
	totalTimeSeconds        int
	currentWord             string

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
	m.page = testPage
	m.testPageInit()
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("ttype")
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var keyCmd tea.Cmd
	var cmd tea.Cmd
	var inputCmd tea.Cmd
	switch m.page {
	case testPage:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			m.previousText = m.inputText
			m.inputText, inputCmd = m.inputText.Update(msg)
			m, keyCmd = m.testPageKeyHandler(msg.String())
		}
	case resultsPage:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.page = testPage
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

func (m Model) testPageKeyHandler(msg string) (Model, tea.Cmd) {
	switch msg {
	case "ctrl+c":
		return m, tea.Quit
	case "backspace":
		if len(m.previousText.Value()) == m.stopBackspaceIdx {
			m.inputText = m.previousText
			return m, nil
		}
		if m.currentWord != "" {
			m.currentWord = m.currentWord[:len(m.currentWord)-1]
		} else {
			// get previous word
			m.currentWord = m.inputText.Value()[getPreviousWordStartIdx(m.inputText.Value()):]
		}
	case " ":
		//check := m.currentWord
		m.totalLengthCorrectWords += len(m.currentWord) + 1
		m.currentWord = ""
		m.stopBackspaceIdx = len(m.inputText.Value())
	}
	var inputCmd tea.Cmd
	if len(m.inputText.Value()) > len(m.previousText.Value()) {
		input := m.inputText.Value()
		newLetter := string(input[len(input)-1])
		if newLetter != " " || m.currentWord != "" {
			m.currentWord += newLetter
		}
		if newLetter == string(m.fileText[len(input)-1]) {
			m.addAccuracyStats(1, 1)
		} else {
			m.addAccuracyStats(1, 0)
		}
	}
	if len(m.inputText.Value()) == len(m.fileText) {
		m.page = resultsPage
		m.inputText.Blur()
	}
	m.viewText = m.updateViewText()
	return m, inputCmd
}

func (m *Model) testPageInit() tea.Cmd {
	text, err := os.ReadFile("./testFile.md")
	if err != nil {
		panic("could not read file")
	}
	m.fileText = string(text)
	m.viewText = untypedStyle.Render(string(text))
	m.numAttempts = 0
	m.totalCorrect = 0
	m.totalLengthCorrectWords = 0
	m.totalTimeSeconds = 0
	m.currentWord = ""
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
	return viewText + "\n" + "visible input length: " + fmt.Sprint(len(inputText)) + "\n\ncurrent word: " +
		m.currentWord + "\n\ntotal length of correct words: " + fmt.Sprint(m.totalLengthCorrectWords) + "\n\ntotal input length: " +
		fmt.Sprint(len(viewText)) + "\n\naccuracy: " + fmt.Sprint(m.getAccuracy()) + "%"
}

func (m *Model) addAccuracyStats(numAttemptsDelta int, totalCorrect int) {
	m.numAttempts += numAttemptsDelta
	m.totalCorrect += totalCorrect
}

func (m Model) getAccuracy() int {
	if m.numAttempts == 0 {
		return 0
	}
	return (m.totalCorrect * 100) / m.numAttempts
}

func (m Model) getSpeed() int {
	if m.totalTimeSeconds == 0 {
		return 0
	}
	return (m.totalLengthCorrectWords / 5) * (60 / m.totalTimeSeconds)
}

// on space check:
// was the last word correct?
// if so prevent backspace beyond this point
// calculate speed
// on backspace: check backspace index; bounce if is equal to
// recalculate word: if word is not empty word = word[:len(word) - 1]
// else word = last word (if last word does not cross stopbackspaceidx)
func getPreviousWordStartIdx(text string) int {
	i := len(text) - 1
	if text[i] == ' ' { // ignore first space
		i--
	}
	for i > 0 {
		if text[i] == ' ' {
			return i + 1
		}
		i--
	}
	return i
}
