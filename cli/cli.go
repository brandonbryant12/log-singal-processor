package cli

import (
	"fmt"
	"log-signal-processor/logsimulator"
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
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	infoStyle         = lipgloss.NewStyle().MarginLeft(4).Foreground(lipgloss.Color("39"))
	navigationStyle   = lipgloss.NewStyle().MarginTop(1).MarginLeft(4).Foreground(lipgloss.Color("241"))
)

// Step represents the current step in the configuration process
type Step int

const (
	DBSelectionStep Step = iota
	FieldSelectionStep
	SignalSelectionStep
	EncryptionSelectionStep
	EncryptionModeStep // New step for encryption mode (AES, ChaCha20)
	AESKeyBitSizeStep  // New step for AES key bit size
	AESModeStep        // New step for AES mode of operation
	EncryptionPercentageStep
	RowCountStep
	ConfigSummaryStep // New step to show summary before finishing
	FinishedStep
)

// SignalType represents a type of signal generator
type SignalType string

const (
	SignalTypeAll         SignalType = "All"
	SignalTypeLevenshtein SignalType = "Levenshtein"
	SignalTypeEntropy     SignalType = "Entropy"
)

// AESMode represents AES mode of operation
type AESMode string

const (
	AESModeCBC AESMode = "CBC"
	AESModeCTR AESMode = "CTR"
	AESModeGCM AESMode = "GCM"
)

// AESKeyBitSize represents AES key bit sizes
type AESKeyBitSize int

const (
	AESKeyBitSize128 AESKeyBitSize = 128
	AESKeyBitSize192 AESKeyBitSize = 192
	AESKeyBitSize256 AESKeyBitSize = 256
)

// OutputFormat represents output format options
type OutputFormat string

const (
	OutputFormatJSON OutputFormat = "JSON"
)

// Config holds the user's configuration choices
type Config struct {
	DBType               string
	SelectedFields       []string
	SelectedSignals      []SignalType
	EncryptionType       logsimulator.EncryptionType
	AESMode              AESMode       // New field for AES mode
	AESKeyBitSize        AESKeyBitSize // New field for AES key bit size
	EncryptionPercentage int
	RowCount             int
	OutputFormat         OutputFormat // New field for output format
}

// Model represents the application state
type Model struct {
	step                 Step
	dbOptions            []string
	dbCursor             int
	fieldOptions         []string
	fieldCursors         map[int]struct{} // Selected fields
	fieldCursor          int              // Current cursor position
	signalOptions        []SignalType
	signalCursors        map[int]struct{} // Selected signals
	signalCursor         int              // Current signal cursor position
	encryptionOptions    []logsimulator.EncryptionType
	encryptionCursor     int
	aesModeOptions       []AESMode // New field for AES modes
	aesModeCursor        int
	aesKeyBitSizeOptions []AESKeyBitSize // New field for AES key bit sizes
	aesKeyBitSizeCursor  int
	encryptionPercentage textinput.Model
	rowCountInput        textinput.Model

	config Config
	err    error
	// Navigation tracking
	previousSteps []Step // For back button functionality
}

func InitialModel() Model {
	// Set up row count input
	rowCount := textinput.New()
	rowCount.Placeholder = "Enter a number"
	rowCount.Focus()
	rowCount.CharLimit = 5
	rowCount.Width = 20

	// Set up encryption percentage input
	encPercent := textinput.New()
	encPercent.Placeholder = "Enter percentage (0-100)"
	encPercent.Focus()
	encPercent.CharLimit = 3
	encPercent.Width = 20

	return Model{
		step:                 DBSelectionStep,
		dbOptions:            []string{"oracle", "postgres"},
		dbCursor:             0,
		fieldOptions:         []string{"bio", "email", "phone", "address"},
		fieldCursors:         make(map[int]struct{}),
		fieldCursor:          0,
		signalOptions:        []SignalType{SignalTypeAll, SignalTypeLevenshtein, SignalTypeEntropy},
		signalCursors:        make(map[int]struct{}),
		signalCursor:         0,
		encryptionOptions:    []logsimulator.EncryptionType{logsimulator.EncryptionTypeNone, logsimulator.EncryptionTypeAES, logsimulator.EncryptionTypeChaCha20},
		encryptionCursor:     0,
		aesModeOptions:       []AESMode{AESModeCBC, AESModeCTR, AESModeGCM},
		aesModeCursor:        0,
		aesKeyBitSizeOptions: []AESKeyBitSize{AESKeyBitSize128, AESKeyBitSize192, AESKeyBitSize256},
		aesKeyBitSizeCursor:  2, // Default to 256-bit
		encryptionPercentage: encPercent,
		rowCountInput:        rowCount,
		config:               Config{OutputFormat: OutputFormatJSON}, // Set default output format to JSON
		previousSteps:        []Step{},
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// goToStep handles transitions between steps, tracking history for back button
func (m *Model) goToStep(newStep Step) {
	m.previousSteps = append(m.previousSteps, m.step)
	m.step = newStep
}

// goBack returns to the previous step
func (m *Model) goBack() {
	if len(m.previousSteps) > 0 {
		m.step = m.previousSteps[len(m.previousSteps)-1]
		m.previousSteps = m.previousSteps[:len(m.previousSteps)-1]
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "backspace", "esc":
			// Back button functionality
			if m.step > DBSelectionStep {
				m.goBack()
			}
			return m, nil

		case "enter":
			switch m.step {
			case DBSelectionStep:
				m.config.DBType = m.dbOptions[m.dbCursor]
				m.goToStep(FieldSelectionStep)

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
					m.err = fmt.Errorf("please select at least one field")
					return m, nil
				}
				m.err = nil
				m.config.SelectedFields = fields
				m.goToStep(SignalSelectionStep)

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
					m.err = fmt.Errorf("please select at least one signal type")
					return m, nil
				}
				m.err = nil
				m.config.SelectedSignals = signals
				m.goToStep(EncryptionSelectionStep)

			case EncryptionSelectionStep:
				m.config.EncryptionType = m.encryptionOptions[m.encryptionCursor]

				// If AES selected, go to AES-specific options
				if m.config.EncryptionType == logsimulator.EncryptionTypeAES {
					m.goToStep(AESModeStep)
				} else if m.config.EncryptionType == logsimulator.EncryptionTypeNone {
					// Skip percentage step if "None" is selected
					m.config.EncryptionPercentage = 0
					m.goToStep(RowCountStep)
					m.rowCountInput.Focus()
				} else {
					// For other encryption types go to percentage
					m.goToStep(EncryptionPercentageStep)
					m.encryptionPercentage.Focus()
				}

			case AESModeStep:
				m.config.AESMode = m.aesModeOptions[m.aesModeCursor]
				m.goToStep(AESKeyBitSizeStep)

			case AESKeyBitSizeStep:
				m.config.AESKeyBitSize = m.aesKeyBitSizeOptions[m.aesKeyBitSizeCursor]
				m.goToStep(EncryptionPercentageStep)
				m.encryptionPercentage.Focus()

			case EncryptionPercentageStep:
				val, err := strconv.Atoi(m.encryptionPercentage.Value())
				if err != nil || val < 0 || val > 100 {
					m.err = fmt.Errorf("please enter a valid percentage (0-100)")
					return m, nil
				}
				m.err = nil
				m.config.EncryptionPercentage = val
				m.goToStep(RowCountStep)
				m.rowCountInput.Focus()

			case RowCountStep:
				val, err := strconv.Atoi(m.rowCountInput.Value())
				if err != nil || val <= 0 {
					m.err = fmt.Errorf("please enter a valid positive number")
					return m, nil
				}
				m.err = nil
				m.config.RowCount = val
				m.goToStep(ConfigSummaryStep)

			case ConfigSummaryStep:
				m.goToStep(FinishedStep)
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

			case EncryptionSelectionStep:
				m.encryptionCursor--
				if m.encryptionCursor < 0 {
					m.encryptionCursor = len(m.encryptionOptions) - 1
				}

			case AESModeStep:
				m.aesModeCursor--
				if m.aesModeCursor < 0 {
					m.aesModeCursor = len(m.aesModeOptions) - 1
				}

			case AESKeyBitSizeStep:
				m.aesKeyBitSizeCursor--
				if m.aesKeyBitSizeCursor < 0 {
					m.aesKeyBitSizeCursor = len(m.aesKeyBitSizeOptions) - 1
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

			case EncryptionSelectionStep:
				m.encryptionCursor = (m.encryptionCursor + 1) % len(m.encryptionOptions)

			case AESModeStep:
				m.aesModeCursor = (m.aesModeCursor + 1) % len(m.aesModeOptions)

			case AESKeyBitSizeStep:
				m.aesKeyBitSizeCursor = (m.aesKeyBitSizeCursor + 1) % len(m.aesKeyBitSizeOptions)

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

	// Handle text input for encryption percentage
	if m.step == EncryptionPercentageStep {
		m.encryptionPercentage, cmd = m.encryptionPercentage.Update(msg)
		return m, cmd
	}

	// Handle text input for row count
	if m.step == RowCountStep {
		m.rowCountInput, cmd = m.rowCountInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	var s string

	// Common header for all screens
	s = titleStyle.Render("Log Signal Processor Configuration") + "\n\n"

	switch m.step {
	case DBSelectionStep:
		s += titleStyle.Render("Which database logs do you want to simulate?") + "\n\n"

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
		s += titleStyle.Render("Select fields to simulate (use spacebar to select):") + "\n\n"

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

		if m.err != nil {
			s += "\n" + errorStyle.Render(m.err.Error())
		}

		s += "\n" + helpStyle.Render("↑/↓: Navigate • Space: Toggle • Enter: Confirm • Esc: Back")

	case SignalSelectionStep:
		s += titleStyle.Render("Select signal generators to use:") + "\n\n"

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

		if m.err != nil {
			s += "\n" + errorStyle.Render(m.err.Error())
		}

		s += "\n" + helpStyle.Render("↑/↓: Navigate • Space: Toggle • Enter: Confirm • Esc: Back")

	case EncryptionSelectionStep:
		s += titleStyle.Render("Select encryption type for simulated attacks:") + "\n\n"

		for i, option := range m.encryptionOptions {
			cursor := " "
			if m.encryptionCursor == i {
				cursor = ">"
				s += activeItemStyle.Render(fmt.Sprintf("%s %s", cursor, option)) + "\n"
			} else {
				s += itemStyle.Render(fmt.Sprintf("%s %s", cursor, option)) + "\n"
			}
		}

		s += "\n" + helpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Back")

	case AESModeStep:
		s += titleStyle.Render("Select AES mode of operation:") + "\n\n"
		s += infoStyle.Render("Each mode offers different security properties") + "\n\n"

		for i, option := range m.aesModeOptions {
			cursor := " "
			if m.aesModeCursor == i {
				cursor = ">"
			}

			description := ""
			switch option {
			case AESModeCBC:
				description = "- Cipher Block Chaining (legacy)"
			case AESModeCTR:
				description = "- Counter Mode (stream cipher)"
			case AESModeGCM:
				description = "- Galois/Counter Mode (authenticated)"
			}

			if m.aesModeCursor == i {
				s += activeItemStyle.Render(fmt.Sprintf("%s %s %s", cursor, option, description)) + "\n"
			} else {
				s += itemStyle.Render(fmt.Sprintf("%s %s %s", cursor, option, description)) + "\n"
			}
		}

		s += "\n" + helpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Back")

	case AESKeyBitSizeStep:
		s += titleStyle.Render("Select AES key bit size:") + "\n\n"
		s += infoStyle.Render("Larger keys provide more security but may be slightly slower") + "\n\n"

		for i, option := range m.aesKeyBitSizeOptions {
			cursor := " "
			if m.aesKeyBitSizeCursor == i {
				cursor = ">"
			}

			description := ""
			switch option {
			case AESKeyBitSize128:
				description = "- Fast, recommended minimum"
			case AESKeyBitSize192:
				description = "- Medium strength"
			case AESKeyBitSize256:
				description = "- Maximum security (recommended)"
			}

			if m.aesKeyBitSizeCursor == i {
				s += activeItemStyle.Render(fmt.Sprintf("%s %d-bit %s", cursor, option, description)) + "\n"
			} else {
				s += itemStyle.Render(fmt.Sprintf("%s %d-bit %s", cursor, option, description)) + "\n"
			}
		}

		s += "\n" + helpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Back")

	case EncryptionPercentageStep:
		s += titleStyle.Render("What percentage of rows should be encrypted? (0-100)") + "\n\n"
		s += m.encryptionPercentage.View() + "\n"
		if m.err != nil {
			s += "\n" + errorStyle.Render(m.err.Error())
		}
		s += "\n" + helpStyle.Render("Enter: Confirm • Esc: Back")

	case RowCountStep:
		s += titleStyle.Render("How many rows do you want to generate?") + "\n\n"
		s += m.rowCountInput.View() + "\n"
		if m.err != nil {
			s += "\n" + errorStyle.Render(m.err.Error())
		}
		s += "\n" + helpStyle.Render("Enter: Confirm • Esc: Back")

	case ConfigSummaryStep:
		s += titleStyle.Render("Configuration Summary:") + "\n\n"
		s += m.config.String() + "\n\n"
		s += helpStyle.Render("Enter: Start Processing • Esc: Go Back")
	}

	// Add navigation help if not on first screen
	if m.step > DBSelectionStep && m.step != FinishedStep {
		s += "\n" + navigationStyle.Render("Press Esc to go back to previous step")
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

// GetEncryptionConfig converts the encryption settings to the simulator's config format
func (c *Config) GetEncryptionConfig() logsimulator.EncryptionConfig {
	// For AES, include mode and key size
	if c.EncryptionType == logsimulator.EncryptionTypeAES {
		// Convert AESKeyBitSize to bytes
		keySize := int(c.AESKeyBitSize) / 8

		return logsimulator.EncryptionConfig{
			Type:       c.EncryptionType,
			Percentage: c.EncryptionPercentage,
			AESMode:    string(c.AESMode),
			KeySize:    keySize,
		}
	}

	// For other encryption types
	return logsimulator.EncryptionConfig{
		Type:       c.EncryptionType,
		Percentage: c.EncryptionPercentage,
	}
}

// DumpConfig returns a string representation of the configuration
func (c Config) String() string {
	encryptionDetails := "None"
	if c.EncryptionType != logsimulator.EncryptionTypeNone {
		if c.EncryptionType == logsimulator.EncryptionTypeAES {
			encryptionDetails = fmt.Sprintf("%s-%d-%s (%d%%)",
				c.EncryptionType,
				c.AESKeyBitSize,
				c.AESMode,
				c.EncryptionPercentage)
		} else {
			encryptionDetails = fmt.Sprintf("%s (%d%%)", c.EncryptionType, c.EncryptionPercentage)
		}
	}

	return fmt.Sprintf("DB Type: %s\nSelected Fields: %s\nSelected Signals: %s\nEncryption: %s\nRow Count: %d\nOutput Format: %s",
		c.DBType,
		strings.Join(c.SelectedFields, ", "),
		formatSignalTypes(c.SelectedSignals),
		encryptionDetails,
		c.RowCount,
		c.OutputFormat)
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
