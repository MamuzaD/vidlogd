package views

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
	"github.com/charmbracelet/lipgloss"
	"github.com/mamuzad/vidlogd/internal/models"
	"github.com/mamuzad/vidlogd/internal/services"
	"github.com/mamuzad/vidlogd/internal/ui"
)

const (
	url = iota
	title
	channel
	release
	logDate
	rating
	rewatch
	review
	button
)

type FieldType int

const (
	FormFieldDate FieldType = iota
	FormFieldDateHour
	FormFieldURL
	FormFieldText
	FormFieldRating
	FormFieldCheckbox
)

type FormField struct {
	Placeholder string
	Label       string
	Required    bool
	Value       string
	CharLimit   int
	Width       int
	Type        FieldType
	SideBySide  bool
}

type FormModel struct {
	title          string
	inputs         []textinput.Model
	fields         []FormField
	focused        int
	fieldErrors    []string
	touched        []bool
	buttonText     string
	onSave         func(FormModel) tea.Cmd
	onCancel       func() tea.Cmd
	lastURL        string
	ratingValue    float64 // current rating value for the rating field
	help           help.Model
	renderedFields map[int]bool // track which fields have been rendered (for side by side)
	// vim mode support
	vimMode string
}

// FormKeyMap implements help.KeyMap for the form
type FormKeyMap struct {
	onRating bool
	vimMode  string
}

func (k FormKeyMap) ShortHelp() []key.Binding {
	keys := []key.Binding{
		ui.GlobalKeyMap.NextField,
		ui.GlobalKeyMap.PrevField,
		ui.GlobalKeyMap.Select,
		ui.GlobalKeyMap.Help,
	}

	return keys
}

func (k FormKeyMap) FullHelp() [][]key.Binding {
	baseKeys := [][]key.Binding{
		{
			ui.GlobalKeyMap.NextField,
			ui.GlobalKeyMap.PrevField,
		},
		{
			ui.GlobalKeyMap.Select,
			ui.GlobalKeyMap.Save,
		},
		{
			ui.GlobalKeyMap.Cancel,
			ui.GlobalKeyMap.Help,
		},
	}

	// add vim column if vim is enabled
	if Settings.VimMotions {
		vimKeys := []key.Binding{
			ui.GlobalKeyMap.VisualMode,
			ui.GlobalKeyMap.Paste,
			ui.GlobalKeyMap.Yank,
		}
		if k.vimMode == "normal" {
			vimKeys = append(vimKeys, ui.GlobalKeyMap.InsertMode)
		} else {
			vimKeys = append(vimKeys, ui.GlobalKeyMap.NormalMode)
		}
		baseKeys = append(baseKeys, vimKeys)
	}

	if k.onRating {
		baseKeys = append(baseKeys, []key.Binding{ui.GlobalKeyMap.Rating})
	}

	return baseKeys
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

		// set cursor mode based on vim mode
		if Settings.VimMotions {
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
	if Settings.VimMotions {
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
		vimMode:     vimMode,
	}
}

func NewVideoLogForm(editing bool, existingVideo *models.Video) FormModel {
	fields := []FormField{
		{Placeholder: "https://youtube.com/watch?v=...", Label: "YouTube URL:", Required: true, CharLimit: 200, Width: 60, Type: FormFieldURL},
		{Placeholder: "video title", Label: "Title:", Required: true, CharLimit: 100, Width: 60, Type: FormFieldText},
		{Placeholder: "channel name", Label: "Channel:", Required: true, CharLimit: 50, Width: 52, Type: FormFieldText},
		{Placeholder: "YYYY-MM-DD", Label: "Video Release Date:", Required: true, CharLimit: 16, Width: 17, Type: FormFieldDate, SideBySide: true},
		{Placeholder: "YYYY-MM-DD HH:MM AM/PM", Label: "Log Date:", Required: true, CharLimit: 19, Width: 22, Type: FormFieldDateHour, SideBySide: true},
		{Placeholder: "", Label: "Rating:", Required: false, CharLimit: 1, Width: 20, Type: FormFieldRating, SideBySide: true},
		{Placeholder: "", Label: "Rewatched:", Required: false, Width: 10, Type: FormFieldCheckbox, SideBySide: true},
		{Placeholder: "write your review...", Label: "Review:", Required: false, CharLimit: 500, Width: 60, Type: FormFieldText},
	}

	var ratingValue float64
	// pre-fill fields if editing
	if editing && existingVideo != nil {
		fields[url].Value = existingVideo.URL
		fields[title].Value = existingVideo.Title
		fields[channel].Value = existingVideo.Channel
		fields[release].Value = existingVideo.ReleaseDate
		fields[logDate].Value = existingVideo.LogDate.Format(models.DateTimeFormat)
		ratingValue = existingVideo.Rating
		if existingVideo.Rewatched {
			fields[rewatch].Value = "true"
		} else {
			fields[rewatch].Value = "false"
		}
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

	form := NewForm(formTitle, fields, buttonText)
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

// updates the vim mode setting for the form
func (m *FormModel) UpdateVimMode() {
	if Settings.VimMotions {
		m.vimMode = "normal"
		// update cursor mode for all inputs
		for i := range m.inputs {
			if i == m.focused {
				m.inputs[i].Cursor.SetMode(cursor.CursorStatic) // focused input gets static cursor in normal mode
			} else {
				m.inputs[i].Cursor.SetMode(cursor.CursorStatic)
			}
		}
	} else {
		m.vimMode = "insert"
		// set cursor mode for all inputs
		for i := range m.inputs {
			if i == m.focused {
				m.inputs[i].Cursor.SetMode(cursor.CursorBlink) // focused input gets blinking cursor in non-vim mode
			} else {
				m.inputs[i].Cursor.SetMode(cursor.CursorHide) // unfocused inputs hide cursor
			}
		}
	}
}

func (m FormModel) Value(index int) string {
	if index < len(m.inputs) {
		return m.inputs[index].Value()
	}
	return ""
}

func (m FormModel) AllValues() []string {
	values := make([]string, len(m.inputs))
	for i := range m.inputs {
		values[i] = m.inputs[i].Value()
	}
	return values
}

func (m FormModel) Rating() float64 {
	return m.ratingValue
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case services.MetadataFetchedMsg:
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
				if len(m.inputs[title].Value()) > m.fields[title].Width {
					// ensure title is visible by moving cursor to start
					m.inputs[title].CursorStart()
				}
			}
			if msg.Metadata.Creator != "" && len(m.inputs) > channel {
				m.inputs[channel].SetValue(msg.Metadata.Creator)
				m.fieldErrors[channel] = ""
				if len(m.inputs[channel].Value()) > m.fields[channel].Width {
					// ensure channel is visible by moving cursor to start
					m.inputs[channel].CursorStart()
				}
			}
			if msg.Metadata.ReleaseDate != "" && len(m.inputs) > release {
				m.inputs[release].SetValue(msg.Metadata.ReleaseDate)
				m.fieldErrors[release] = ""
			}
			if len(m.inputs) > logDate {
				currentDate := time.Now().Format(models.DateTimeFormat)
				m.inputs[logDate].SetValue(currentDate)
				m.fieldErrors[logDate] = ""
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, ui.GlobalKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, ui.GlobalKeyMap.Back, ui.GlobalKeyMap.Cancel):
			shouldCancel := false
			if !Settings.VimMotions {
				shouldCancel = key.Matches(msg, ui.GlobalKeyMap.Cancel)
			} else {
				shouldCancel = m.vimMode == "normal"
			}
			if shouldCancel {
				if m.onCancel != nil {
					return m, m.onCancel()
				}
				return m, nil
			}
		// vim keys
		case key.Matches(msg, ui.GlobalKeyMap.InsertMode):
			// only handle if vim is enabled and we're in normal mode
			if Settings.VimMotions && m.vimMode == "normal" {
				m.vimMode = "insert"
				// enable cursor blinking in insert mode
				if m.focused < len(m.inputs) {
					m.inputs[m.focused].Cursor.SetMode(cursor.CursorBlink)
				}
				return m, textinput.Blink
			}
		case key.Matches(msg, ui.GlobalKeyMap.NormalMode):
			// only handle if vim is enabled and we're in insert mode
			if Settings.VimMotions && m.vimMode == "insert" {
				m.vimMode = "normal"
				// disable cursor blinking in normal mode
				if m.focused < len(m.inputs) {
					m.inputs[m.focused].Cursor.SetMode(cursor.CursorStatic)
				}
				return m, nil
			}
		case key.Matches(msg, ui.GlobalKeyMap.Paste):
			if Settings.VimMotions && m.vimMode == "normal" {
				// get clipboard content
				clipboardContent, err := clipboard.ReadAll()
				if err != nil {
					return m, nil
				}
				if m.focused < len(m.inputs) {
					m.inputs[m.focused].SetValue(clipboardContent)
				}
			}
		case key.Matches(msg, ui.GlobalKeyMap.NextField):
			if !Settings.VimMotions || m.vimMode == "normal" {
				m.validateCurrentField()
				m.nextInput()
			}
		case key.Matches(msg, ui.GlobalKeyMap.PrevField):
			if !Settings.VimMotions || m.vimMode == "normal" {
				m.validateCurrentField()
				m.prevInput()
			}
		case key.Matches(msg, ui.GlobalKeyMap.Save):
			return m.handleSave()
		case key.Matches(msg, ui.GlobalKeyMap.Select):
			if m.focused == len(m.inputs) {
				return m.handleSave()
			} else {
				// toggle checkbox
				if m.focused < len(m.fields) && m.fields[m.focused].Type == FormFieldCheckbox {
					currentValue := m.inputs[m.focused].Value()
					if currentValue == "true" {
						m.inputs[m.focused].SetValue("false")
					} else {
						m.inputs[m.focused].SetValue("true")
					}
					return m, nil
				}
				m.validateCurrentField()
				m.nextInput()
			}
		case key.Matches(msg, ui.GlobalKeyMap.RatingDown):
			if m.focused == rating {
				if m.ratingValue > 0 {
					m.ratingValue -= 0.5
				}
				return m, nil
			}
		case key.Matches(msg, ui.GlobalKeyMap.RatingUp):
			if m.focused == rating {
				if m.ratingValue < 5 {
					m.ratingValue += 0.5
				}
				return m, nil
			}
		case key.Matches(msg, ui.GlobalKeyMap.Rating):
			if m.focused == rating {
				ratingStr := msg.String()
				rating, _ := strconv.ParseFloat(ratingStr, 64)
				m.ratingValue = rating
				return m, nil
			}
		case key.Matches(msg, ui.GlobalKeyMap.RatingHalf):
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
	shouldUpdateInput := !Settings.VimMotions || m.vimMode == "insert"

	// update the focused input
	if m.focused < len(m.inputs) && shouldUpdateInput {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		cmds = append(cmds, cmd)
	}

	// check if URL and auto-fill metadata - regardless of vim mode
	if m.focused == url && len(m.inputs) > url {
		currentURL := m.inputs[url].Value()
		if currentURL != m.lastURL && services.IsValidYouTubeURL(currentURL) {
			m.lastURL = currentURL
			// auto-fill metadata in background
			cmds = append(cmds, services.FetchYouTubeMetadata(currentURL))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m FormModel) View() string {
	var s strings.Builder

	// clear rendered fields tracking for this render
	m.renderedFields = make(map[int]bool)

	s.WriteString(ui.HeaderStyle.Render(m.title))

	// show vim mode status if vim is enabled
	if Settings.VimMotions {
		modeStr := strings.ToUpper(m.vimMode)
		s.WriteString(ui.ModeStyle.Render(modeStr))
	}

	s.WriteString("\n\n")

	// render all input fields
	for i, field := range m.fields {
		// handle side by side rendering
		if field.SideBySide && !m.isFieldAlreadyRendered(i) {
			// find the next field to pair with
			pairedIndex := -1
			for j := i + 1; j < len(m.fields); j++ {
				if m.fields[j].SideBySide {
					pairedIndex = j
					break
				}
			}

			if pairedIndex != -1 {
				// render both fields side by side
				leftSection := m.renderField(i)
				rightSection := m.renderField(pairedIndex)

				sideBySide := lipgloss.JoinHorizontal(
					lipgloss.Top,
					leftSection,
					strings.Repeat(" ", 8), // spacing
					rightSection,
				)
				s.WriteString(sideBySide)
				s.WriteString("\n")

				// mark both fields as rendered
				m.markFieldAsRendered(i)
				m.markFieldAsRendered(pairedIndex)
				continue
			}
		}

		// skip if already rendered as part of side by side
		if m.isFieldAlreadyRendered(i) {
			continue
		}

		s.WriteString(m.renderField(i))
		s.WriteString("\n")
	}

	// save button
	if m.focused == len(m.inputs) {
		s.WriteString(ui.ButtonStyleFocused.Render(m.buttonText))
	} else {
		s.WriteString(ui.ButtonStyle.Render(m.buttonText))
	}

	if m.fieldErrors[button] != "" {
		s.WriteString("\n   " + m.fieldErrors[button])
	}

	keymap := FormKeyMap{onRating: m.focused == rating, vimMode: m.vimMode}
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
		var star string
		if m.ratingValue >= starValue {
			star = ui.StarStyle.Render("★") // filled star
		} else if m.ratingValue >= starValue-0.5 {
			star = ui.StarStyle.Render("⯨") // half star
		} else {
			star = "☆" // empty star
		}
		s.WriteString(star)
		if i < 5 {
			s.WriteString(" ")
		}
	}

	// show current rating text
	if m.ratingValue == 0 {
		s.WriteString("          ")
	} else {
		// only show decimal if not a whole number
		if m.ratingValue == float64(int(m.ratingValue)) {
			s.WriteString(fmt.Sprintf("  (%d/5)   ", int(m.ratingValue)))
		} else {
			s.WriteString(fmt.Sprintf("  (%.1f/5) ", m.ratingValue))
		}
	}

	var styledRating string
	if focused {
		styledRating = ui.FormFieldFocusedStyle.Render(s.String())
	} else {
		styledRating = ui.FormFieldStyle.Render(s.String())
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
			errorMsg = "invalid date"
		}
	case FormFieldDateHour:
		if !isValidDateHour(value) {
			errorMsg = "invalid datetime"
		}
	case FormFieldURL:
		if !services.IsValidYouTubeURL(value) {
			errorMsg = "invalid youtube url"
		}
	}
	// update field error
	if errorMsg != "" {
		m.fieldErrors[index] = errorMsg
	}
	return errorMsg
}

// isValidDateHour validates if the string is a valid date in YYYY-MM-DD HH:MM AM/PM format
func isValidDateHour(dateStr string) bool {
	if dateStr == "" {
		return false
	}

	// check format with regex first - supports both 12:34 AM and 1:23 PM formats
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{1,2}:\d{2} (AM|PM)$`)
	if !dateRegex.MatchString(dateStr) {
		return false
	}

	// try to parse the date to ensure it's actually valid
	_, err := time.Parse(models.DateTimeFormat, dateStr)
	return err == nil
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
	_, err := time.Parse(models.ISODateFormat, dateStr)
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

func (m FormModel) isFieldAlreadyRendered(index int) bool {
	if m.renderedFields == nil {
		return false
	}
	return m.renderedFields[index]
}

func (m *FormModel) markFieldAsRendered(index int) {
	if m.renderedFields == nil {
		m.renderedFields = make(map[int]bool)
	}
	m.renderedFields[index] = true
}

func (m FormModel) renderField(i int) string {
	if i >= len(m.fields) {
		return ""
	}

	var s strings.Builder
	field := m.fields[i]

	// render label
	s.WriteString(field.Label + "\n")

	// render field content based on type
	switch field.Type {
	case FormFieldRating:
		// render rating stars
		stars := m.renderRatingStars(i == m.focused)
		s.WriteString(stars)
	case FormFieldCheckbox:
		// render checkbox
		checked := m.inputs[i].Value() == "true"
		var checkbox string
		if checked {
			checkbox = " ✓ "
		} else {
			checkbox = "   "
		}
		if m.focused == i {
			checkbox = ui.FormFieldFocusedStyle.Render(checkbox)
		} else {
			checkbox = ui.FormFieldStyle.Render(checkbox)
		}
		s.WriteString(checkbox)
	default:
		var styledInput string
		if m.focused == i {
			styledInput = ui.FormFieldFocusedStyle.Render(m.inputs[i].View())
		} else {
			styledInput = ui.FormFieldStyle.Render(m.inputs[i].View())
		}
		s.WriteString(styledInput)
	}

	// show field-specific error if field has been touched and has an error
	if m.touched[i] && m.fieldErrors[i] != "" {
		s.WriteString("\n   " + m.fieldErrors[i])
	}

	return s.String()
}

func (form FormModel) Video() models.Video {
	return models.CreateVideo(
		form.Value(url),
		form.Value(title),
		form.Value(channel),
		form.Value(release),
		form.Value(logDate),
		form.Value(review),
		form.Value(rewatch) == "true",
		form.Rating(),
	)
}
