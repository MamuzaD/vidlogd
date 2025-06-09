package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	url = iota
	title
	channel
	release
	logDate
	rating
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
	ratingValue float64 // current rating value for the rating field
	help        help.Model
	// vim mode support
	vimEnabled bool
	vimMode    string
}

// FormKeyMap implements help.KeyMap for the form
type FormKeyMap struct {
	onRating   bool
	vimEnabled bool
	vimMode    string
}

func (k FormKeyMap) ShortHelp() []key.Binding {
	keys := []key.Binding{
		GlobalKeyMap.NextField,
		GlobalKeyMap.PrevField,
		GlobalKeyMap.Select,
		GlobalKeyMap.Help,
	}

	return keys
}

func (k FormKeyMap) FullHelp() [][]key.Binding {
	baseKeys := [][]key.Binding{
		{
			GlobalKeyMap.NextField,
			GlobalKeyMap.PrevField,
			GlobalKeyMap.Select,
		},
		{
			GlobalKeyMap.Save,
			GlobalKeyMap.Cancel,
			GlobalKeyMap.Help,
		},
	}

	// add vim column if vim is enabled
	if k.vimEnabled {
		vimKeys := []key.Binding{
			GlobalKeyMap.VisualMode,
			GlobalKeyMap.Paste,
			GlobalKeyMap.Yank,
		}
		if k.vimMode == "normal" {
			vimKeys = append(vimKeys, GlobalKeyMap.InsertMode)
		} else {
			vimKeys = append(vimKeys, GlobalKeyMap.NormalMode)
		}
		baseKeys = append(baseKeys, vimKeys)
	}

	if k.onRating {
		baseKeys = append(baseKeys, []key.Binding{GlobalKeyMap.Rating})
	}

	return baseKeys
}

func NewForm(title string, fields []FormField, saveText string, vimEnabled bool) FormModel {
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

		// input.Cursor.SetMode(cursor.CursorHide)
		// // set cursor mode based on vim mode
		if vimEnabled {
			input.Cursor.SetMode(cursor.CursorStatic) // start in normal mode
		} else {
			input.Cursor.SetMode(cursor.CursorBlink) // always blink in non-vim mode
		}

		if i == 0 {
			input.Focus()
		}
		inputs[i] = input
	}

	h := help.New()
	h.ShowAll = false // start with compact help

	// start in insert mode for non-vim, normal mode for vim
	vimMode := "insert"
	if vimEnabled {
		vimMode = "normal"
	}

	return FormModel{
		title:       title,
		inputs:      inputs,
		fields:      fields,
		focused:     0,
		fieldErrors: fieldErrors,
		touched:     touched,
		buttonText:  saveText,
		help:        h,
		vimEnabled:  vimEnabled,
		vimMode:     vimMode,
	}
}

func NewVideoLogForm(editing bool, existingVideo *Video, vimEnabled bool) FormModel {
	fields := []FormField{
		{Placeholder: "https://youtube.com/watch?v=...", Label: "YouTube URL:", Required: true, CharLimit: 200, Width: 60, Type: FormFieldURL},
		{Placeholder: "video title", Label: "Title:", Required: true, CharLimit: 100, Width: 50, Type: FormFieldText},
		{Placeholder: "channel name", Label: "Channel:", Required: true, CharLimit: 50, Width: 50, Type: FormFieldText},
		{Placeholder: "YYYY-MM-DD", Label: "Video Release Date:", Required: true, CharLimit: 10, Width: 12, Type: FormFieldDate},
		{Placeholder: "YYYY-MM-DD", Label: "Log Date:", Required: true, CharLimit: 10, Width: 12, Type: FormFieldDate},
		{Placeholder: "", Label: "Rating:", Required: false, CharLimit: 1, Width: 20, Type: FormFieldRating},
		{Placeholder: "write your review...", Label: "Review:", Required: false, CharLimit: 500, Width: 60, Type: FormFieldText},
	}

	var ratingValue float64
	// pre-fill fields if editing
	if editing && existingVideo != nil {
		fields[url].Value = existingVideo.URL
		fields[title].Value = existingVideo.Title
		fields[channel].Value = existingVideo.Channel
		fields[release].Value = existingVideo.ReleaseDate
		fields[logDate].Value = existingVideo.LogDate
		ratingValue = existingVideo.Rating
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

	form := NewForm(formTitle, fields, buttonText, vimEnabled)
	form.ratingValue = ratingValue

	// store original URL when editing to prevent auto-fill for same video
	if editing && existingVideo != nil {
		form.lastURL = existingVideo.URL
	}

	return form
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

func (m FormModel) GetRating() float64 {
	return m.ratingValue
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
		switch {
		case key.Matches(msg, GlobalKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, GlobalKeyMap.Exit), key.Matches(msg, GlobalKeyMap.Cancel):
			if m.onCancel != nil {
				return m, m.onCancel()
			}
			return m, nil
		// vim keys
		case key.Matches(msg, GlobalKeyMap.InsertMode):
			// only handle if vim is enabled and we're in normal mode
			if m.vimEnabled && m.vimMode == "normal" {
				m.vimMode = "insert"
				// enable cursor blinking in insert mode
				if m.focused < len(m.inputs) {
					m.inputs[m.focused].Cursor.SetMode(cursor.CursorBlink)
				}
				return m, textinput.Blink
			}
		case key.Matches(msg, GlobalKeyMap.NormalMode):
			// only handle if vim is enabled and we're in insert mode
			if m.vimEnabled && m.vimMode == "insert" {
				m.vimMode = "normal"
				// disable cursor blinking in normal mode
				if m.focused < len(m.inputs) {
					m.inputs[m.focused].Cursor.SetMode(cursor.CursorStatic)
				}
				return m, nil
			}
		case key.Matches(msg, GlobalKeyMap.Paste):
			if m.vimEnabled && m.vimMode == "normal" {
				// get clipboard content
				clipboardContent, err := clipboard.ReadAll()
				if err != nil {
					return m, nil
				}
				if m.focused < len(m.inputs) {
					m.inputs[m.focused].SetValue(clipboardContent)
				}
			}
		case key.Matches(msg, GlobalKeyMap.NextField):
			if !m.vimEnabled || m.vimMode == "normal" {
				m.validateCurrentField()
				m.nextInput()
			}
		case key.Matches(msg, GlobalKeyMap.PrevField):
			if !m.vimEnabled || m.vimMode == "normal" {
				m.validateCurrentField()
				m.prevInput()
			}
		case key.Matches(msg, GlobalKeyMap.Save):
			return m.handleSave()
		case key.Matches(msg, GlobalKeyMap.Select):
			if m.focused == len(m.inputs) {
				return m.handleSave()
			} else {
				m.validateCurrentField()
				m.nextInput()
			}
		case key.Matches(msg, GlobalKeyMap.RatingDown):
			if m.focused == rating {
				if m.ratingValue > 0 {
					m.ratingValue -= 0.5
				}
				return m, nil
			}
		case key.Matches(msg, GlobalKeyMap.RatingUp):
			if m.focused == rating {
				if m.ratingValue < 5 {
					m.ratingValue += 0.5
				}
				return m, nil
			}
		case key.Matches(msg, GlobalKeyMap.Rating):
			if m.focused == rating {
				ratingStr := msg.String()
				rating, _ := strconv.ParseFloat(ratingStr, 64)
				m.ratingValue = rating
				return m, nil
			}
		case key.Matches(msg, GlobalKeyMap.RatingHalf):
			if m.focused == rating {
				// add 0.5 to current rating if it's a whole number
				if m.ratingValue == float64(int(m.ratingValue)) && m.ratingValue < 5 {
					m.ratingValue += 0.5
				}
				return m, nil
			}
		}
	}

	// only update text inputs if not in vim normal mode
	shouldUpdateInput := !m.vimEnabled || m.vimMode == "insert"

	// update the focused input
	if m.focused < len(m.inputs) && shouldUpdateInput {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		cmds = append(cmds, cmd)
	}

	// check if URL and auto-fill metadata - regardless of vim mode
	if m.focused == url && len(m.inputs) > url {
		currentURL := m.inputs[url].Value()
		if currentURL != m.lastURL && isValidYouTubeURL(currentURL) {
			m.lastURL = currentURL
			// auto-fill metadata in background
			cmds = append(cmds, fetchYouTubeMetadata(currentURL))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m FormModel) View() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render(m.title))

	// show vim mode status if vim is enabled
	if m.vimEnabled {
		modeStr := strings.ToUpper(m.vimMode)
		s.WriteString(modeStyle.Render(modeStr))
	}

	s.WriteString("\n\n")

	// render all input fields
	for i, field := range m.fields {
		s.WriteString(field.Label + "\n")

		if field.Type == FormFieldRating {
			// render rating stars
			stars := m.renderRatingStars(i == m.focused)
			s.WriteString(stars)
		} else {
			var styledInput string
			if m.focused == i {
				styledInput = formFieldFocusedStyle.Render(m.inputs[i].View())
			} else {
				styledInput = formFieldStyle.Render(m.inputs[i].View())
			}
			s.WriteString(styledInput)
		}

		// show field-specific error if field has been touched and has an error
		if m.touched[i] && m.fieldErrors[i] != "" {
			s.WriteString("\n  ⚠ " + m.fieldErrors[i])
		}
		s.WriteString("\n\n")
	}

	// save button
	if m.focused == len(m.inputs) {
		s.WriteString(buttonStyleFocused.Render(m.buttonText))
	} else {
		s.WriteString(buttonStyle.Render(m.buttonText))
	}

	if m.fieldErrors[button] != "" {
		s.WriteString(" ⚠ " + m.fieldErrors[button])
	}

	keymap := FormKeyMap{onRating: m.focused == rating, vimEnabled: m.vimEnabled, vimMode: m.vimMode}
	s.WriteString("\n\n" + m.help.View(keymap))

	return s.String()
}

// renderRatingStars renders the star rating display
func (m FormModel) renderRatingStars(focused bool) string {
	var s strings.Builder
	s.WriteString(" ")
	// render 5 stars
	for i := 1; i <= 5; i++ {
		starValue := float64(i)
		if m.ratingValue >= starValue {
			s.WriteString("★") // filled star
		} else if m.ratingValue >= starValue-0.5 {
			s.WriteString("⯨") // half star
		} else {
			s.WriteString("☆") // empty star
		}
		if i < 5 {
			s.WriteString(" ")
		}
	}

	// show current rating text
	if m.ratingValue == 0 {
		s.WriteString("  (no rating) ")
	} else {
		// only show decimal if not a whole number
		if m.ratingValue == float64(int(m.ratingValue)) {
			s.WriteString(fmt.Sprintf("  (%d/5) ", int(m.ratingValue)))
		} else {
			s.WriteString(fmt.Sprintf("  (%.1f/5) ", m.ratingValue))
		}
	}

	var styledRating string
	if focused {
		styledRating = formFieldFocusedStyle.Render(s.String())
	} else {
		styledRating = formFieldStyle.Render(s.String())
	}
	return styledRating
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
