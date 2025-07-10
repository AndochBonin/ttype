package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/filepicker"

	//"github.com/charmbracelet/bubbles/textinput"
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
	err          error
	//inputText textinput.Model
	//timer timer.Model
}

func initialModel() (Model, error) {
	m := Model{}
	m.page = fileMenu
	m.filePicker = filepicker.New()
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
		fmt.Println(err.Error())
	} else {
		m.filePicker.CurrentDirectory = currentDirectory
		fmt.Println(m.filePicker.CurrentDirectory)
	}

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
	cmds = append(cmds, cmd, keyCmd)
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
		view += m.fileText
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
	m.filePicker, cmd = m.filePicker.Update(msg)
	switch msg {
	case "ctrl+c":
		return tea.Quit
	case "enter":
		if didSelect {
			m.err = m.testPageInit()
			m.page = testPage
		}
	}
	return cmd
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

func (m *Model) testPageInit() error {
	text, err := os.ReadFile(m.selectedFile)
	if err != nil {
		return err
	}
	m.fileText = string(text)
	return nil
}

// fileMenu:
// show file picker of all folders and text files
// on enter go to testPage with file text content

// testPage:
// screen always displays the same text -> each individual characters color is decided independently
// not typed yet: grey, typed incorrectly: red, typed correctly: green

// when len(inputText) == len(fileText) -> end the test -> display stats
