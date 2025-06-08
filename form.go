package main

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	url = iota
	title
	channel
	release
	logDate
	review
)

type FormField struct {
	Placeholder string
	Label       string
	Required    bool
	Value       string
	CharLimit   int
	Width       int
}

type FormModel struct {
	title      string
	inputs     []textinput.Model
	fields     []FormField
	focused    int
	errorMsg   string
	buttonText string
	onSave     func(FormModel) tea.Cmd
	onCancel   func() tea.Cmd
	lastURL    string // track last URL to detect changes
}

func NewForm(title string, fields []FormField, saveText string) FormModel {
	inputs := make([]textinput.Model, len(fields))

	for i, field := range fields {
		input := textinput.New()

		input.CharLimit = field.CharLimit
		input.Width = field.Width

		input.Placeholder = field.Placeholder
		if field.Value != "" {
			input.SetValue(field.Value)
		}
		if i == 0 {
			input.Focus()
		}
		inputs[i] = input
	}

	return FormModel{
		title:      title,
		inputs:     inputs,
		fields:     fields,
		focused:    0,
		buttonText: saveText,
	}
}

func (m *FormModel) SetHandlers(onSave func(FormModel) tea.Cmd, onCancel func() tea.Cmd) {
	m.onSave = onSave
	m.onCancel = onCancel
}

func (m FormModel) GetValue(index int) string {
	if index < len(m.inputs) {
		return m.inputs[index].Value()
	}
	return ""
}

func (m FormModel) GetAllValues() []string {
	values := make([]string, len(m.inputs))
	for i := range m.inputs {
		values[i] = m.inputs[i].Value()
	}
	return values
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case MetadataFetchedMsg:
		// auto-fill form fields with YouTube metadata
		if msg.Error != "" {
			m.errorMsg = msg.Error
		} else {
			if msg.Metadata.Title != "" && len(m.inputs) > title {
				m.inputs[title].SetValue(msg.Metadata.Title)
			}
			if msg.Metadata.Creator != "" && len(m.inputs) > channel {
				m.inputs[channel].SetValue(msg.Metadata.Creator)
			}
			if msg.Metadata.ReleaseDate != "" && len(m.inputs) > release {
				m.inputs[release].SetValue(msg.Metadata.ReleaseDate)
			}
			if len(m.inputs) > logDate {
				currentDate := time.Now().Format("2006-01-02")
				m.inputs[logDate].SetValue(currentDate)
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.onCancel != nil {
				return m, m.onCancel()
			}
			return m, nil
		case "tab", "down":
			m.nextInput()
		case "shift+tab", "up":
			m.prevInput()
		case "enter":
			if m.focused == len(m.inputs) {
				return m.handleSave()
			} else {
				m.nextInput()
			}
		}
	}

	// update the focused input
	if m.focused < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		cmds = append(cmds, cmd)

		// check if URL and auto-fill metadata
		if m.focused == url && len(m.inputs) > url {
			currentURL := m.inputs[url].Value()
			if currentURL != m.lastURL && isValidYouTubeURL(currentURL) {
				m.lastURL = currentURL
				// auto-fill metadata in background
				cmds = append(cmds, fetchYouTubeMetadata(currentURL))
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m FormModel) View() string {
	var s strings.Builder

	s.WriteString(m.title + "\n\n")

	if m.errorMsg != "" {
		s.WriteString("Error: " + m.errorMsg + "\n\n")
	}

	// render all input fields
	for i, field := range m.fields {
		s.WriteString(field.Label + "\n")
		s.WriteString(m.inputs[i].View() + "\n\n")
	}

	// save button
	if m.focused == len(m.inputs) {
		s.WriteString("> [" + m.buttonText + "] <\n\n")
	} else {
		s.WriteString("  [" + m.buttonText + "]\n\n")
	}

	return s.String()
}

// nextInput moves focus to the next input
func (m *FormModel) nextInput() {
	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Blur()
	}

	m.focused++
	if m.focused > len(m.inputs) {
		m.focused = 0
	}

	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Focus()
	}
}

// prevInput moves focus to the previous input
func (m *FormModel) prevInput() {
	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Blur()
	}

	m.focused--
	if m.focused < 0 {
		m.focused = len(m.inputs)
	}

	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Focus()
	}
}

func (m FormModel) handleSave() (FormModel, tea.Cmd) {
	// validate required fields
	for i, field := range m.fields {
		if field.Required && strings.TrimSpace(m.inputs[i].Value()) == "" {
			m.errorMsg = m.fields[i].Label + " required"
			return m, nil
		}
	}

	if m.onSave != nil {
		return m, m.onSave(m)
	}

	return m, nil
}
