package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// REWRITE NEEDED - move text into array of trimmed words, model behavior after monkeytype.com

var textLength = 100

var randomWordFunction = []func() string{randomdata.Noun, randomdata.Adjective, randomdata.City, randomdata.Day,
	randomdata.City, randomdata.Month}

const (
	testPage = iota
	resultsPage
)

type Model struct {
	page                    int
	width                   int
	height                  int
	fileText                string
	previousText            textinput.Model
	inputText               textinput.Model
	viewText                string
	totalCorrect            int
	numAttempts             int
	totalLengthCorrectWords int
	stopBackspaceIdx        int
	totalTimeSeconds        int
	currentWord             string
	timer                   timer.Model
}

var (
	untypedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	correctStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	wrongStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	fitStyle     = lipgloss.NewStyle().Width(100)
	headerStyle  = lipgloss.NewStyle().Bold(true)
)

func initialModel(testDurationSeconds int) (Model, error) {
	m := Model{}
	m.page = testPage
	m.totalTimeSeconds = testDurationSeconds
	m.testPageInit()
	return m, nil
}

func (m Model) Init() tea.Cmd {
	rand.Seed(time.Now().UnixNano())
	return tea.Batch(tea.SetWindowTitle("ttype"), m.timer.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var keyCmd tea.Cmd
	var cmd tea.Cmd
	var inputCmd tea.Cmd
	switch m.page {
	case testPage:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
		case tea.KeyMsg:
			m.previousText = m.inputText
			m.inputText, inputCmd = m.inputText.Update(msg)
			m, keyCmd = m.testPageKeyHandler(msg.String())
			m.viewText = m.updateViewText()
		case timer.TickMsg:
			var timerCmd tea.Cmd
			m.timer, timerCmd = m.timer.Update(msg)
			m.viewText = m.updateViewText()
			return m, timerCmd
		case timer.StartStopMsg:
			var timerCmd tea.Cmd
			m.timer, timerCmd = m.timer.Update(msg)
			m.viewText = m.updateViewText()
			return m, timerCmd
		case timer.TimeoutMsg:
			m.page = resultsPage
			m.inputText.Blur()
			m.viewText = m.updateViewText()
			return m, nil
		}
	case resultsPage:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.page = testPage
				return m, m.testPageInit()
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
		view += fitStyle.Render(m.viewText) + "\n"
	case resultsPage:
		view = "wpm: " + fmt.Sprint(m.getSpeed()) + "\n\n"
		view += "accuracy: " + fmt.Sprint(m.getAccuracy()) + "%" + "\n\n"
	}
	return title + "\n\n" + view
}

func Run(testDurationSeconds int) error {
	m, initErr := initialModel(testDurationSeconds)
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
	case "tab":
		return m, m.testPageInit()
	case "backspace":
		if len(m.previousText.Value()) == m.stopBackspaceIdx {
			m.inputText = m.previousText
			return m, nil
		}
		if m.currentWord != "" {
			m.currentWord = m.currentWord[:len(m.currentWord)-1]
		} else {
			m.currentWord = m.inputText.Value()[getPreviousWordStartIdx(m.inputText.Value()):]
		}
	case " ":
		wordLength := len(m.currentWord)
		totalInputLength := len(m.inputText.Value())
		if m.isWordCorrect(totalInputLength-wordLength-1, totalInputLength) {
			m.totalLengthCorrectWords += wordLength + 1
			m.stopBackspaceIdx = len(m.inputText.Value())
		}
		m.currentWord = ""
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
		wordLength := len(m.currentWord)
		totalInputLength := len(m.inputText.Value())
		if m.isWordCorrect(totalInputLength-wordLength-1, totalInputLength) {
			m.totalLengthCorrectWords += wordLength
			m.page = resultsPage
			m.inputText.Blur()
		}
	}
	return m, inputCmd
}

func (m *Model) testPageInit() tea.Cmd {
	m.fileText = ""
	for i := 0; i < textLength; i++ {
		m.fileText += strings.ToLower(randomWordFunction[rand.Intn(len(randomWordFunction))]())
		if i < textLength-1 {
			m.fileText += " "
		}
	}

	m.viewText = untypedStyle.Render(m.fileText)
	m.numAttempts = 0
	m.totalCorrect = 0
	m.totalLengthCorrectWords = 0
	m.currentWord = ""
	inputText := textinput.New()
	inputText.CharLimit = len(m.fileText)
	m.inputText = inputText
	m.previousText = inputText
	m.stopBackspaceIdx = 0
	m.timer = timer.New(time.Second * time.Duration(m.totalTimeSeconds))
	return tea.Batch(m.inputText.Focus(), m.timer.Init())
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
	wpm := headerStyle.Render("wpm: " + fmt.Sprint(m.getSpeed()))
	accuracy := headerStyle.Render("accuracy: " + fmt.Sprint(m.getAccuracy()) + "%")
	timeLeft := headerStyle.Render("time remaining: " + m.timer.View())
	space := "   "
	return wpm + space + accuracy + space + timeLeft + "\n\n" + viewText
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
	secondsPassed := m.totalTimeSeconds - int(m.timer.Timeout.Seconds())
	if secondsPassed == 0 {
		return 0
	}
	currentCorrectWordLength := 0
	wordLength := len(m.currentWord)
	totalInputLength := len(m.inputText.Value())
	if m.isWordCorrect(totalInputLength-wordLength, totalInputLength) {
		currentCorrectWordLength = wordLength
	}
	return ((m.totalLengthCorrectWords + currentCorrectWordLength) / 5) * (60 / secondsPassed)
}

func getPreviousWordStartIdx(text string) int {
	i := len(text) - 1
	if i < 0 {
		return 0
	}
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

func (m Model) isWordCorrect(startIdx int, endIdx int) bool {
	if startIdx < 0 {
		return false
	}
	for i := startIdx; i < endIdx; i++ {
		if m.fileText[i] != m.inputText.Value()[i] {
			return false
		}
	}
	return true
}
