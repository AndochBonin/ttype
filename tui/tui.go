package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	testPage = iota
	resultsPage
)

type Model struct {
	page                    int
	width                   int
	height                  int
	textLength              int
	viewText                string
	testText                []string
	userText                []string
	currentWordInput        textinput.Model
	currentInputIdx         int
	cursorIdx               int
	totalCorrect            int
	numAttempts             int
	totalLengthCorrectWords int
	totalTimeSeconds        int
	timer                   timer.Model
}

var (
	untypedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	correctStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	wrongStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
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
			if msg.String() != " " {
				m.currentWordInput, inputCmd = m.currentWordInput.Update(msg)
			}
			m, keyCmd = m.testPageKeyHandler(msg.String())
			m.viewText = m.getViewText()
		case timer.TickMsg:
			var timerCmd tea.Cmd
			m.timer, timerCmd = m.timer.Update(msg)
			m.viewText = m.getViewText()
			return m, timerCmd
		case timer.StartStopMsg:
			var timerCmd tea.Cmd
			m.timer, timerCmd = m.timer.Update(msg)
			m.viewText = m.getViewText()
			return m, timerCmd
		case timer.TimeoutMsg:
			m.page = resultsPage
			m.currentWordInput.Blur()
			m.viewText = m.getViewText()
			return m, nil
		}
	case resultsPage:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc", "tab":
				m.viewText = ""
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
	fitStyle := lipgloss.NewStyle().Width(int(float64(m.width) * float64(0.8)))
	switch m.page {
	case testPage:
		space := "   "
		stats := headerStyle.Render("wpm: "+fmt.Sprint(m.getSpeed())+space+"accuracy: "+fmt.Sprint(m.getAccuracy())+"%"+space+
			"time remaining: "+m.timer.View()) + "\n\n"
		view += stats + fitStyle.Render(m.viewText) + "\n"
	case resultsPage:
		view = headerStyle.Render("wpm: "+fmt.Sprint(m.getSpeed())) + "\n\n"
		view += headerStyle.Render("accuracy: "+fmt.Sprint(m.getAccuracy())+"%") + "\n\n"
	}
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, headerStyle.Render(title) + "\n\n" + view)
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
	case "esc", "tab":
		return m, m.testPageInit()
	case "backspace":
		if m.canEditPreviousWord() {
			m.currentInputIdx--
			m.currentWordInput.SetValue(m.userText[m.currentInputIdx])
		}
	case " ":
		if m.currentInputIdx < len(m.testText)-1 {
			if m.isWordCorrect(m.currentInputIdx) {
				m.totalLengthCorrectWords += len(m.testText[m.currentInputIdx]) + 1 // space should be included in character count
			}
			m.currentInputIdx++
			m.currentWordInput.SetValue("")
		}
	}
	m.userText[m.currentInputIdx] = m.currentWordInput.Value()
	m.cursorIdx = len(m.userText[m.currentInputIdx])
	if msg != "backspace" {
		m.updateAccuracystats()
	}
	return m, nil
}

func (m *Model) testPageInit() tea.Cmd {
	m.testText = []string{}
	m.textLength = m.totalTimeSeconds * 4
	for range m.textLength {
		randomWord := strings.ToLower(words[rand.Intn(len(words))])
		m.testText = append(m.testText, randomWord)
	}

	for i, word := range m.testText {
		m.viewText += word
		if i < len(m.testText)-1 {
			m.viewText += " "
		}
	}
	m.viewText = untypedStyle.Render(m.viewText)
	m.userText = make([]string, m.textLength)
	m.currentInputIdx = 0
	m.cursorIdx = 0
	m.numAttempts = 0
	m.totalCorrect = 0
	m.totalLengthCorrectWords = 0
	m.timer = timer.New(time.Second * time.Duration(m.totalTimeSeconds))
	m.currentWordInput = textinput.New()
	return tea.Batch(m.currentWordInput.Focus(), m.timer.Init())
}

func (m *Model) getViewText() string {
	viewText := ""
	for i := range m.testText {
		viewText += m.getStyledWord(i)
		if i == m.currentInputIdx && len(m.userText[i]) >= len(m.testText[i]) {
			viewText += untypedStyle.Background(lipgloss.Color("15")).Render(" ")
		} else {
			viewText += " "
		}
	}
	return viewText
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
	inputWordLength := len(m.userText[m.currentInputIdx])
	testWordLength := len(m.testText[m.currentInputIdx])
	if inputWordLength <= testWordLength && m.testText[m.currentInputIdx][:inputWordLength] == m.userText[m.currentInputIdx] {
		currentCorrectWordLength = inputWordLength
	}
	return ((m.totalLengthCorrectWords + currentCorrectWordLength) / 5) * (60 / secondsPassed)
}

func (m Model) getStyledWord(index int) string {
	testWord := m.testText[index]
	inputWord := m.userText[index]
	if testWord == inputWord {
		return correctStyle.Render(testWord)
	}

	coloredWord := ""
	var shortWord string
	var longWord string
	var leftoverStyle lipgloss.Style
	var underline = false
	underline = index < m.currentInputIdx

	if len(inputWord) >= len(testWord) {
		shortWord = testWord
		longWord = inputWord
		leftoverStyle = wrongStyle
	} else {
		shortWord = inputWord
		longWord = testWord
		leftoverStyle = untypedStyle
	}

	if shortWord == longWord[:len(shortWord)] {
		coloredWord += correctStyle.Underline(underline).Render(shortWord)
	} else {
		for i := range shortWord {
			if testWord[i] == inputWord[i] {
				coloredWord += correctStyle.Underline(underline).Render(string(testWord[i]))
			} else {
				coloredWord += wrongStyle.Underline(underline).Render(string(testWord[i]))
			}
		}
	}

	if index == m.currentInputIdx && shortWord == inputWord {
		coloredWord += untypedStyle.Background(lipgloss.Color("15")).Render(string(longWord[len(shortWord)]))
		if len(shortWord)+1 < len(longWord) {
			coloredWord += untypedStyle.Render(longWord[len(shortWord)+1:])
		}
	} else {
		coloredWord += leftoverStyle.Underline(underline).Render(longWord[len(shortWord):])
	}
	return coloredWord
}

func (m Model) canEditPreviousWord() bool {
	idx := m.currentInputIdx
	if m.userText[idx] == "" && idx > 0 && m.userText[idx-1] != m.testText[idx-1] {
		return true
	}
	return false
}

func (m Model) isWordCorrect(idx int) bool {
	if idx < 0 || idx > len(m.testText) {
		return false
	}
	return m.testText[idx] == m.userText[idx]
}

func (m *Model) updateAccuracystats() {
	ttLen := len(m.testText[m.currentInputIdx])
	utLen := len(m.userText[m.currentInputIdx])
	if utLen < 1 || ttLen < 1 {
		return
	}
	m.numAttempts++
	if utLen <= ttLen && m.testText[m.currentInputIdx][utLen-1] == m.userText[m.currentInputIdx][utLen-1] {
		m.totalCorrect++
	}
}
