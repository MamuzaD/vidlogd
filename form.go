package main

import (
	"regexp"
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
	button
)

type FieldType int

const (
	FormFieldDate FieldType = iota
	FormFieldURL
	FormFieldText
	FormFieldRating
)

type FormField struct {
	Placeholder string
	Label       string
	Required    bool
	Value       string
	CharLimit   int
	Width       int
	Type        FieldType
}

type FormModel struct {
	title       string
	inputs      []textinput.Model
	fields      []FormField
	focused     int
	fieldErrors []string
	touched     []bool
	buttonText  string
	onSave      func(FormModel) tea.Cmd
	onCancel    func() tea.Cmd
	lastURL     string
}

func NewForm(title string, fields []FormField, saveText string) FormModel {
	inputs := make([]textinput.Model, len(fields))
	fieldErrors := make([]string, button+1) // +1 to include space for button error
	touched := make([]bool, len(fields))

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
		title:       title,
		inputs:      inputs,
		fields:      fields,
		focused:     0,
		fieldErrors: fieldErrors,
		touched:     touched,
		buttonText:  saveText,
	}
}

func NewVideoLogForm(editing bool, existingVideo *Video) FormModel {
	fields := []FormField{
		{Placeholder: "https://youtube.com/watch?v=...", Label: "YouTube URL:", Required: true, CharLimit: 200, Width: 60, Type: FormFieldURL},
		{Placeholder: "video title", Label: "Title:", Required: true, CharLimit: 100, Width: 50, Type: FormFieldText},
		{Placeholder: "channel name", Label: "Channel:", Required: true, CharLimit: 50, Width: 50, Type: FormFieldText},
		{Placeholder: "YYYY-MM-DD", Label: "Video Release Date:", Required: true, CharLimit: 10, Width: 12, Type: FormFieldDate},
		{Placeholder: "YYYY-MM-DD", Label: "Log Date:", Required: true, CharLimit: 10, Width: 12, Type: FormFieldDate},
		{Placeholder: "write your review...", Label: "Review:", Required: false, CharLimit: 500, Width: 60, Type: FormFieldText},
	}

	// pre-fill fields if editing
	if editing && existingVideo != nil {
		fields[url].Value = existingVideo.URL
		fields[title].Value = existingVideo.Title
		fields[channel].Value = existingVideo.Channel
		fields[release].Value = existingVideo.ReleaseDate
		fields[logDate].Value = existingVideo.LogDate
		fields[review].Value = existingVideo.Review
	}

	var formTitle string
	var buttonText string

	if editing {
		formTitle = "edit video log"
		buttonText = "update video"
	} else {
		formTitle = "log a video"
		buttonText = "save video"
	}

	return NewForm(formTitle, fields, buttonText)
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
			m.touched[url] = true // mark URL as touched so error shows
			m.fieldErrors[url] = msg.Error
		} else {
			// remove any prev url error
			m.fieldErrors[url] = ""

			// prefill and remove errors for metadata fields
			if msg.Metadata.Title != "" && len(m.inputs) > title {
				m.inputs[title].SetValue(msg.Metadata.Title)
				m.fieldErrors[title] = ""
				// ensure title is visible by moving cursor to start
				m.inputs[title].CursorStart()
			}
			if msg.Metadata.Creator != "" && len(m.inputs) > channel {
				m.inputs[channel].SetValue(msg.Metadata.Creator)
				m.fieldErrors[channel] = ""
				// ensure channel is visible by moving cursor to start
				m.inputs[channel].CursorStart()
			}
			if msg.Metadata.ReleaseDate != "" && len(m.inputs) > release {
				m.inputs[release].SetValue(msg.Metadata.ReleaseDate)
				m.fieldErrors[release] = ""
			}
			if len(m.inputs) > logDate {
				currentDate := time.Now().Format("2006-01-02")
				m.inputs[logDate].SetValue(currentDate)
				m.fieldErrors[logDate] = ""
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
			m.validateCurrentField()
			m.nextInput()
		case "shift+tab", "up":
			m.validateCurrentField()
			m.prevInput()
		case "enter":
			if m.focused == len(m.inputs) {
				return m.handleSave()
			} else {
				m.validateCurrentField()
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

	// render all input fields
	for i, field := range m.fields {
		s.WriteString(field.Label + "\n")
		s.WriteString(m.inputs[i].View())

		// show field-specific error if field has been touched and has an error
		if m.touched[i] && m.fieldErrors[i] != "" {
			s.WriteString("\n  ⚠ " + m.fieldErrors[i])
		}
		s.WriteString("\n\n")
	}

	// save button
	if m.focused == len(m.inputs) {
		s.WriteString("> [" + m.buttonText + "] <\n\n")
	} else {
		s.WriteString("  [" + m.buttonText + "]\n\n")
	}

	if m.fieldErrors[button] != "" {
		s.WriteString(" ⚠ " + m.fieldErrors[button])
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

// validateCurrentField validates the currently focused field
func (m *FormModel) validateCurrentField() string {
	if m.focused >= len(m.inputs) {
		return ""
	}

	// mark field as touched
	m.touched[m.focused] = true

	return m.validateFieldByIndex(m.focused)
}

// validateFieldByIndex validates a specific field by index
func (m *FormModel) validateFieldByIndex(index int) string {
	if index >= len(m.inputs) || index >= len(m.fields) {
		return ""
	}

	field := m.fields[index]
	value := strings.TrimSpace(m.inputs[index].Value())

	// clear previous error
	m.fieldErrors[index] = ""

	// skip validation if field is empty and not required
	if value == "" && !field.Required {
		return ""
	}

	var errorMsg string

	// check if required
	if field.Required && value == "" {
		errorMsg = "field is required"
	}

	// validate based on field type
	switch field.Type {
	case FormFieldDate:
		if !isValidDate(value) {
			errorMsg = "invalid date (YYYY-MM-DD)"
		}
	case FormFieldURL:
		if !isValidYouTubeURL(value) {
			errorMsg = "invalid youtube url"
		}
	}
	// update field error
	if errorMsg != "" {
		m.fieldErrors[index] = errorMsg
	}
	return errorMsg
}

// isValidDate validates if the string is a valid date in YYYY-MM-DD format
func isValidDate(dateStr string) bool {
	if dateStr == "" {
		return false
	}

	// check format with regex first
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(dateStr) {
		return false
	}

	// try to parse the date to ensure it's actually valid
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func (m FormModel) handleSave() (FormModel, tea.Cmd) {
	hasErrors := false

	// validate all fields first
	for i := range m.fields {
		m.touched[i] = true // mark all fields as touched when saving

		errorMsg := m.validateFieldByIndex(i)
		if errorMsg != "" {
			hasErrors = true
		}
	}

	if hasErrors {
		m.fieldErrors[button] = "errors in the form"
		return m, nil
	} else {
		m.fieldErrors[button] = ""
	}

	if m.onSave != nil {
		return m, m.onSave(m)
	}

	return m, nil
}
