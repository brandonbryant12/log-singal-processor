package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	itemStyle         = lipgloss.NewStyle().MarginLeft(4)
	selectedItemStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("170"))
	activeItemStyle   = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("86"))
	helpStyle         = lipgloss.NewStyle().MarginLeft(4).Foreground(lipgloss.Color("241"))
)

// Step represents the current step in the configuration process
type Step int

const (
	DBSelectionStep Step = iota
	FieldSelectionStep
	SignalSelectionStep
	RowCountStep
	FinishedStep
)

// SignalType represents a type of signal generator
type SignalType string

const (
	SignalTypeAll         SignalType = "All"
	SignalTypeLevenshtein SignalType = "Levenshtein"
	SignalTypeEntropy     SignalType = "Entropy"
)

// Config holds the user's configuration choices
type Config struct {
	DBType          string
	SelectedFields  []string
	SelectedSignals []SignalType
	RowCount        int
}

// Model represents the application state
type Model struct {
	step          Step
	dbOptions     []string
	dbCursor      int
	fieldOptions  []string
	fieldCursors  map[int]struct{} // Selected fields
	fieldCursor   int              // Current cursor position
	signalOptions []SignalType
	signalCursors map[int]struct{} // Selected signals
	signalCursor  int              // Current signal cursor position
	textInput     textinput.Model
	config        Config
	err           error
}

func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter a number"
	ti.Focus()
	ti.CharLimit = 5
	ti.Width = 20

	return Model{
		step:          DBSelectionStep,
		dbOptions:     []string{"oracle", "postgres"},
		dbCursor:      0,
		fieldOptions:  []string{"bio", "email", "phone", "address"},
		fieldCursors:  make(map[int]struct{}),
		fieldCursor:   0,
		signalOptions: []SignalType{SignalTypeAll, SignalTypeLevenshtein, SignalTypeEntropy},
		signalCursors: make(map[int]struct{}),
		signalCursor:  0,
		textInput:     ti,
		config:        Config{},
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			switch m.step {
			case DBSelectionStep:
				m.config.DBType = m.dbOptions[m.dbCursor]
				m.step = FieldSelectionStep

			case FieldSelectionStep:
				// Convert selected fields to slice
				fields := []string{}
				for i := range m.fieldOptions {
					if _, selected := m.fieldCursors[i]; selected {
						fields = append(fields, m.fieldOptions[i])
					}
				}

				// Make sure at least one field is selected
				if len(fields) == 0 {
					return m, nil
				}

				m.config.SelectedFields = fields
				m.step = SignalSelectionStep

			case SignalSelectionStep:
				// Convert selected signals to slice
				signals := []SignalType{}
				for i, signalType := range m.signalOptions {
					if _, selected := m.signalCursors[i]; selected {
						signals = append(signals, signalType)
					}
				}

				// Make sure at least one signal is selected
				if len(signals) == 0 {
					return m, nil
				}

				m.config.SelectedSignals = signals
				m.step = RowCountStep

			case RowCountStep:
				val, err := strconv.Atoi(m.textInput.Value())
				if err != nil || val <= 0 {
					m.err = fmt.Errorf("please enter a valid positive number")
					return m, nil
				}
				m.err = nil
				m.config.RowCount = val
				m.step = FinishedStep
				return m, tea.Quit
			}

		case "up", "k":
			switch m.step {
			case DBSelectionStep:
				m.dbCursor--
				if m.dbCursor < 0 {
					m.dbCursor = len(m.dbOptions) - 1
				}

			case FieldSelectionStep:
				m.fieldCursor--
				if m.fieldCursor < 0 {
					m.fieldCursor = len(m.fieldOptions) - 1
				}

			case SignalSelectionStep:
				m.signalCursor--
				if m.signalCursor < 0 {
					m.signalCursor = len(m.signalOptions) - 1
				}
			}

		case "down", "j":
			switch m.step {
			case DBSelectionStep:
				m.dbCursor = (m.dbCursor + 1) % len(m.dbOptions)

			case FieldSelectionStep:
				m.fieldCursor = (m.fieldCursor + 1) % len(m.fieldOptions)

			case SignalSelectionStep:
				m.signalCursor = (m.signalCursor + 1) % len(m.signalOptions)
			}

		case " ": // Spacebar
			if m.step == FieldSelectionStep {
				// Toggle selection
				if _, ok := m.fieldCursors[m.fieldCursor]; ok {
					delete(m.fieldCursors, m.fieldCursor)
				} else {
					m.fieldCursors[m.fieldCursor] = struct{}{}
				}
			} else if m.step == SignalSelectionStep {
				// Toggle selection
				if m.signalCursor == 0 { // "All" option
					// If "All" is selected, clear other selections
					if _, ok := m.signalCursors[0]; !ok {
						m.signalCursors = make(map[int]struct{})
						m.signalCursors[0] = struct{}{}
					} else {
						delete(m.signalCursors, 0)
					}
				} else {
					// If selecting a specific signal, remove "All" selection
					delete(m.signalCursors, 0)

					// Toggle this signal
					if _, ok := m.signalCursors[m.signalCursor]; ok {
						delete(m.signalCursors, m.signalCursor)
					} else {
						m.signalCursors[m.signalCursor] = struct{}{}
					}
				}
			}
		}
	}

	// Handle text input for row count
	if m.step == RowCountStep {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	var s string

	switch m.step {
	case DBSelectionStep:
		s = titleStyle.Render("Which database logs do you want to simulate?") + "\n\n"

		for i, option := range m.dbOptions {
			cursor := " "
			if m.dbCursor == i {
				cursor = ">"
				s += activeItemStyle.Render(fmt.Sprintf("%s %s", cursor, option)) + "\n"
			} else {
				s += itemStyle.Render(fmt.Sprintf("%s %s", cursor, option)) + "\n"
			}
		}

		s += "\n" + helpStyle.Render("↑/↓: Navigate • Enter: Select")

	case FieldSelectionStep:
		s = titleStyle.Render("Select fields to simulate (use spacebar to select):") + "\n\n"

		for i, option := range m.fieldOptions {
			cursor := " "
			if m.fieldCursor == i {
				cursor = ">"
			}

			checked := " "
			if _, ok := m.fieldCursors[i]; ok {
				checked = "✓"
			}

			if m.fieldCursor == i {
				s += activeItemStyle.Render(fmt.Sprintf("%s [%s] %s", cursor, checked, option)) + "\n"
			} else {
				s += itemStyle.Render(fmt.Sprintf("%s [%s] %s", cursor, checked, option)) + "\n"
			}
		}

		s += "\n" + helpStyle.Render("↑/↓: Navigate • Space: Toggle • Enter: Confirm")

	case SignalSelectionStep:
		s = titleStyle.Render("Select signal generators to use:") + "\n\n"

		for i, option := range m.signalOptions {
			cursor := " "
			if m.signalCursor == i {
				cursor = ">"
			}

			checked := " "
			if _, ok := m.signalCursors[i]; ok {
				checked = "✓"
			}

			if m.signalCursor == i {
				s += activeItemStyle.Render(fmt.Sprintf("%s [%s] %s", cursor, checked, option)) + "\n"
			} else {
				s += itemStyle.Render(fmt.Sprintf("%s [%s] %s", cursor, checked, option)) + "\n"
			}
		}

		s += "\n" + helpStyle.Render("↑/↓: Navigate • Space: Toggle • Enter: Confirm")

	case RowCountStep:
		s = titleStyle.Render("How many rows do you want to generate?") + "\n\n"
		s += m.textInput.View() + "\n"
		if m.err != nil {
			s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.err.Error())
		}
		s += "\n" + helpStyle.Render("Enter: Confirm")
	}

	return s
}

// GetConfig returns the configuration after the user has made their selections
func GetConfig() (Config, error) {
	p := tea.NewProgram(InitialModel())
	m, err := p.Run()
	if err != nil {
		return Config{}, err
	}

	model, ok := m.(Model)
	if !ok || model.step != FinishedStep {
		return Config{}, fmt.Errorf("configuration cancelled")
	}

	return model.config, nil
}

// DumpConfig returns a string representation of the configuration
func (c Config) String() string {
	return fmt.Sprintf("DB Type: %s\nSelected Fields: %s\nSelected Signals: %s\nRow Count: %d",
		c.DBType,
		strings.Join(c.SelectedFields, ", "),
		formatSignalTypes(c.SelectedSignals),
		c.RowCount)
}

// formatSignalTypes converts a slice of signal types to a readable string
func formatSignalTypes(signals []SignalType) string {
	if len(signals) == 0 {
		return "None"
	}

	strs := make([]string, len(signals))
	for i, s := range signals {
		strs[i] = string(s)
	}
	return strings.Join(strs, ", ")
}
