package prompt

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// selectModel represents the Bubble Tea model for target selection.
type selectModel struct {
	choices           []TargetChoice
	filteredChoices   []TargetChoice
	filteredIndices   []int // maps filtered index to original index
	cursor            int
	filter            string
	showWorktreeLabel bool
	showTypePrefix    bool
	selected          *TargetChoice
	quitting          bool
}

// initialSelectModel creates a new select model.
func initialSelectModel(choices []TargetChoice, showWorktreeLabel bool) selectModel {
	// Determine if we should show type prefixes based on whether we have mixed types
	showTypePrefix := false
	if len(choices) > 0 {
		firstType := choices[0].Type
		for _, choice := range choices {
			if choice.Type != firstType {
				showTypePrefix = true
				break
			}
		}
	}

	return selectModel{
		choices:           choices,
		filteredChoices:   choices,
		filteredIndices:   makeRange(len(choices)),
		cursor:            0,
		filter:            "",
		showWorktreeLabel: showWorktreeLabel,
		showTypePrefix:    showTypePrefix,
		selected:          nil,
		quitting:          false,
	}
}

// makeRange creates a slice of integers from 0 to n-1.
func makeRange(n int) []int {
	result := make([]int, n)
	for i := range result {
		result[i] = i
	}
	return result
}

// Init initializes the model.
func (m selectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		return m.handleKeyInput(msg)
	}

	return m, nil
}

// handleKeyInput processes key input and returns the updated model and command.
func (m *selectModel) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	
	// Handle special keys
	if m.handleSpecialKeys(key) {
		return m, tea.Quit
	}
	
	// Handle navigation keys
	m.handleNavigationKeys(key)
	
	// Handle filter keys
	m.handleFilterKeys(key)
	
	return m, nil
}

// handleSpecialKeys handles special keys that cause the program to quit.
func (m *selectModel) handleSpecialKeys(key string) bool {
	switch key {
	case "ctrl+c", "q":
		m.quitting = true
		return true
	case "enter":
		if len(m.filteredChoices) > 0 && m.cursor < len(m.filteredChoices) {
			selected := m.filteredChoices[m.cursor]
			m.selected = &selected
			return true
		}
	}
	return false
}

// handleNavigationKeys handles navigation keys (up/down).
func (m *selectModel) handleNavigationKeys(key string) {
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.filteredChoices)-1 {
			m.cursor++
		}
	}
}

// handleFilterKeys handles filter-related keys.
func (m *selectModel) handleFilterKeys(key string) {
	switch key {
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.updateFilteredChoices()
		}
	case "esc":
		m.filter = ""
		m.updateFilteredChoices()
	default:
		// Handle regular character input for filtering
		if len(key) == 1 {
			m.filter += key
			m.updateFilteredChoices()
		}
	}
}

// updateFilteredChoices updates the filtered choices based on the current filter.
func (m *selectModel) updateFilteredChoices() {
	if m.filter == "" {
		m.filteredChoices = m.choices
		m.filteredIndices = makeRange(len(m.choices))
	} else {
		m.filteredChoices = []TargetChoice{}
		m.filteredIndices = []int{}

		filterLower := strings.ToLower(m.filter)
		for i, choice := range m.choices {
			if strings.Contains(strings.ToLower(choice.Name), filterLower) {
				m.filteredChoices = append(m.filteredChoices, choice)
				m.filteredIndices = append(m.filteredIndices, i)
			}
		}
	}

	// Reset cursor if it's out of bounds
	if m.cursor >= len(m.filteredChoices) {
		m.cursor = 0
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// View renders the UI.
func (m selectModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Header
	s.WriteString("? Choose repository or workspace:  [Use arrows to move, type to filter]\n\n")

	// Show filter if active
	if m.filter != "" {
		s.WriteString(fmt.Sprintf("Filter: %s\n\n", m.filter))
	}

	// Show choices
	for i, choice := range m.filteredChoices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		choiceText := formatChoice(choice, m.showWorktreeLabel, m.showTypePrefix)
		s.WriteString(fmt.Sprintf("%s %s\n", cursor, choiceText))
	}

	// Footer
	s.WriteString("\nPress Enter to select, Ctrl+C or q to quit")
	if m.filter != "" {
		s.WriteString(", Esc to clear filter")
	}

	return s.String()
}

// formatChoice formats a choice for display.
func formatChoice(choice TargetChoice, showWorktreeLabel bool, showTypePrefix bool) string {
	var result string

	if showTypePrefix {
		var prefix string
		switch choice.Type {
		case TargetRepository:
			prefix = "[repository]"
		case TargetWorkspace:
			prefix = "[workspace]"
		default:
			prefix = "[unknown]"
		}
		result = fmt.Sprintf("%s %s", prefix, choice.Name)
	} else {
		result = choice.Name
	}

	if showWorktreeLabel && choice.Worktree != "" {
		result += fmt.Sprintf(" : %s", choice.Worktree)
	}

	return result
}

// promptSelectTargetBubbleTea runs the Bubble Tea program for target selection.
func promptSelectTargetBubbleTea(choices []TargetChoice, showWorktreeLabel bool) (TargetChoice, error) {
	// Create and run the program
	p := tea.NewProgram(initialSelectModel(choices, showWorktreeLabel))

	// Run the program
	finalModel, err := p.Run()
	if err != nil {
		return TargetChoice{}, fmt.Errorf("failed to run selection program: %w", err)
	}

	// Cast to our model type
	model, ok := finalModel.(selectModel)
	if !ok {
		return TargetChoice{}, fmt.Errorf("unexpected model type")
	}

	// Check if user quit without selecting
	if model.selected == nil {
		return TargetChoice{}, fmt.Errorf("no selection made")
	}

	return *model.selected, nil
}
