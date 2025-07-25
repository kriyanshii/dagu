package dagpicker

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dagu-org/dagu/internal/digraph"
	"github.com/dagu-org/dagu/internal/logger"
	"github.com/dagu-org/dagu/internal/models"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// State represents the current state of the picker
type State int

const (
	StateSelectingDAG State = iota
	StateEnteringParams
	StateEnteringRunId
	StateConfirming
	StateDone
)

// DAGItem represents a DAG in the list
type DAGItem struct {
	Name   string
	Path   string // Path is stored but not displayed
	Desc   string
	Tags   []string
	Params string // Parameters that the DAG accepts
}

func (i DAGItem) Title() string {
	title := i.Name
	if len(i.Tags) > 0 {
		title += " [" + strings.Join(i.Tags, ", ") + "]"
	}
	return title
}
func (i DAGItem) Description() string {
	desc := i.Desc
	if i.Params != "" {
		if desc != "" {
			desc += " | params: " + i.Params
		} else {
			desc = "params: " + i.Params
		}
	}
	return desc
}
func (i DAGItem) FilterValue() string { return i.Name }

// Model represents the state of the DAG picker
type Model struct {
	// State management
	state State

	// DAG selection
	list   list.Model
	choice *DAGItem

	// Parameter input
	paramInput textinput.Model
	dag        *digraph.DAG
	params     string

	// Run ID input
	runIdInput textinput.Model
	runId      string

	// Confirmation
	confirmed bool

	// General
	quitting        bool
	width           int
	height          int
	allowEditParams bool // enforce runConfig in picker
	allowEditRunId  bool // enforce runConfig for run ID
}

// Result contains the final selections
type Result struct {
	DAGName   string
	DAGPath   string
	Params    string
	Cancelled bool
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle global keys first
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case tea.KeyMsg:
		// Ctrl+C quits from any state
		if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))) {
			m.quitting = true
			m.state = StateDone
			return m, tea.Quit
		}
	}

	// Handle state-specific updates
	switch m.state {
	case StateSelectingDAG:
		return m.updateDAGSelection(msg)
	case StateEnteringParams:
		return m.updateParamInput(msg)
	case StateEnteringRunId:
		return m.updateRunIdInput(msg)
	case StateConfirming:
		return m.updateConfirmation(msg)
	case StateDone:
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m Model) updateDAGSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.quitting = true
			m.state = StateDone
			return m, tea.Quit

		case "enter":
			if item, ok := m.list.SelectedItem().(DAGItem); ok {
				m.choice = &item

				// Check if the selected DAG has parameters
				hasParams := item.Params != ""

				// --- Refactored: Always check RunConfig, default to true if missing ---
				allowEditParams := true
				allowEditRunId := true
				if m.dag != nil && m.dag.RunConfig != nil {
					log.Print("m.dag.RunConfig.AllowEditParams: ", m.dag.RunConfig.AllowEditParams)
					log.Print("m.dag.RunConfig.AllowEditRunId: ", m.dag.RunConfig.AllowEditRunId)
					allowEditParams = m.dag.RunConfig.AllowEditParams
					allowEditRunId = m.dag.RunConfig.AllowEditRunId
				}
				m.allowEditParams = allowEditParams
				m.allowEditRunId = allowEditRunId
				log.Print("m.allowEditRunId: ", m.allowEditRunId)

				if hasParams {
					m.paramInput.SetValue(item.Params)
					if allowEditParams {
						m.state = StateEnteringParams
						return m, textinput.Blink
					} else {
						// Skip editing, go straight to confirmation
						m.params = item.Params
						m.state = StateConfirming
						return m, nil
					}
				} else {
					m.runIdInput.SetValue("")
					if allowEditRunId {
						m.state = StateEnteringRunId
						return m, textinput.Blink
					} else {
						m.runId = "(auto-generated)"
						m.state = StateConfirming
						return m, nil
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateParamInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.allowEditParams {
		// Editing not allowed, go to run ID input or confirmation
		if m.allowEditRunId {
			m.state = StateEnteringRunId
			return m, textinput.Blink
		} else {
			m.runId = "(auto-generated)"
			m.state = StateConfirming
			return m, nil
		}
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Go back to DAG selection
			m.state = StateSelectingDAG
			return m, nil

		case "enter":
			m.params = m.paramInput.Value()
			if m.allowEditRunId {
				m.state = StateEnteringRunId
				return m, textinput.Blink
			} else {
				m.runId = "(auto-generated)"
				m.state = StateConfirming
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.paramInput, cmd = m.paramInput.Update(msg)
	return m, cmd
}

func (m Model) updateRunIdInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.allowEditRunId {
		// Editing not allowed, skip to confirmation
		m.runId = "(auto-generated)"
		m.state = StateConfirming
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Go back to param input
			if m.allowEditParams {
				m.state = StateEnteringParams
				return m, nil
			} else {
				m.state = StateSelectingDAG
				return m, nil
			}
		case "enter":
			m.runId = m.runIdInput.Value()
			m.state = StateConfirming
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.runIdInput, cmd = m.runIdInput.Update(msg)
	return m, cmd
}

func (m Model) updateConfirmation(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case "esc":
			// Go back to parameter input or DAG selection
			hasParams := m.choice != nil && m.choice.Params != ""
			if hasParams {
				m.state = StateEnteringParams
			} else {
				m.state = StateSelectingDAG
			}
			return m, nil

		case "y", "enter":
			m.confirmed = true
			m.state = StateDone
			return m, tea.Quit

		case "n":
			// Cancel and exit
			m.quitting = true
			m.state = StateDone
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.state == StateDone {
		return ""
	}

	switch m.state {
	case StateSelectingDAG:
		return m.viewDAGSelection()
	case StateEnteringParams:
		return m.viewParamInput()
	case StateEnteringRunId:
		return m.viewRunIdInput()
	case StateConfirming:
		return m.viewConfirmation()
	case StateDone:
		return ""
	default:
		return ""
	}
}

func (m Model) viewDAGSelection() string {
	return docStyle.Render(m.list.View())
}

func (m Model) viewParamInput() string {
	var ctx context.Context
	logger.Info(ctx, "viewPAramInput")
	if !m.allowEditParams {
		// Show params as read-only with clear explanation and visual distinction
		lockStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).MarginBottom(1)
		selectedDAGStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).MarginBottom(2)
		paramStyle := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("243")).PaddingLeft(2)
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(2)
		explanationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Italic(true).MarginTop(1)
		content := lipgloss.JoinVertical(lipgloss.Left,
			lockStyle.Render("🔒 Parameter editing is disabled by this DAG's runConfig."),
			titleStyle.Render("Parameters (read-only)"),
			selectedDAGStyle.Render(fmt.Sprintf("DAG: %s", m.choice.Name)),
			"",
			paramStyle.Render(m.paramInput.Value()),
			"",
			explanationStyle.Render("The workflow author has set 'runConfig.allowEditParams: false' in the DAG definition. You cannot modify parameters for this run."),
			helpStyle.Render("ESC: Back • Ctrl+C: Cancel"),
		)
		return docStyle.Render(content)
	}
	// Define styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)

	selectedDAGStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		MarginBottom(2)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Enter Parameters"),
		selectedDAGStyle.Render(fmt.Sprintf("DAG: %s", m.choice.Name)),
		"",
		m.paramInput.View(),
		"",
		helpStyle.Render("Enter: Confirm • ESC: Back • Ctrl+C: Cancel"),
	)

	return docStyle.Render(content)
}

func (m Model) viewRunIdInput() string {
	if !m.allowEditRunId {
		lockStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).MarginBottom(1)
		selectedDAGStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).MarginBottom(2)
		runIdStyle := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("243")).PaddingLeft(2)
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(2)
		explanationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Italic(true).MarginTop(1)
		content := lipgloss.JoinVertical(lipgloss.Left,
			lockStyle.Render("🔒 Run ID editing is disabled by this DAG's runConfig."),
			titleStyle.Render("Run ID (read-only)"),
			selectedDAGStyle.Render(fmt.Sprintf("DAG: %s", m.choice.Name)),
			"",
			runIdStyle.Render("(auto-generated)"),
			"",
			explanationStyle.Render("The workflow author has set 'runConfig.allowEditRunId: false' in the DAG definition. You cannot specify a custom run ID for this run."),
			helpStyle.Render("ESC: Back • Ctrl+C: Cancel"),
		)
		return docStyle.Render(content)
	}
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).MarginBottom(1)
	selectedDAGStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).MarginBottom(2)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(2)
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Enter Run ID (optional)"),
		selectedDAGStyle.Render(fmt.Sprintf("DAG: %s", m.choice.Name)),
		"",
		m.runIdInput.View(),
		"",
		helpStyle.Render("Enter: Confirm • ESC: Back • Ctrl+C: Cancel"),
	)
	return docStyle.Render(content)
}

func (m Model) viewConfirmation() string {
	// Define styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(2)

	dagNameStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))

	paramStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("243"))

	promptStyle := lipgloss.NewStyle().
		Bold(true).
		MarginTop(2)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var content []string
	content = append(content, titleStyle.Render("Confirm Execution"))
	content = append(content, fmt.Sprintf("Ready to run DAG: %s", dagNameStyle.Render(m.choice.Name)))

	if m.params != "" {
		content = append(content, fmt.Sprintf("With parameters: %s", paramStyle.Render(m.params)))
	}

	content = append(content, "")
	content = append(content, promptStyle.Render("Run this DAG? [Y/n]"))
	content = append(content, helpStyle.Render("Y/Enter: Run • N: Cancel • ESC: Back"))

	return docStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

// pickerModel holds context data for the picker
type pickerModel struct {
	ctx      context.Context
	dagStore models.DAGStore
	dagMap   map[string]*digraph.DAG
}

// PickDAGInteractive shows a unified fullscreen UI for DAG selection, parameter input, and confirmation
func PickDAGInteractive(ctx context.Context, dagStore models.DAGStore, dag *digraph.DAG) (Result, error) {
	// Create an internal picker model
	pickerModel := &pickerModel{
		ctx:      ctx,
		dagStore: dagStore,
		dagMap:   make(map[string]*digraph.DAG),
	}

	// Get list of DAGs
	result, errs, err := dagStore.List(ctx, models.ListDAGsOptions{})
	if err != nil {
		return Result{}, fmt.Errorf("failed to list DAGs: %w", err)
	}

	if len(errs) > 0 {
		// Log errors but continue
		for _, e := range errs {
			fmt.Printf("Warning: %s\n", e)
		}
	}

	if len(result.Items) == 0 {
		return Result{}, fmt.Errorf("no DAGs found in the configured directory")
	}

	// Convert DAGs to list items
	items := make([]list.Item, 0, len(result.Items))

	for _, d := range result.Items {
		// Format parameters for display
		var params string
		if d.DefaultParams != "" {
			params = d.DefaultParams
		} else if len(d.Params) > 0 {
			params = strings.Join(d.Params, " ")
		}

		item := DAGItem{
			Name:   d.Name,
			Path:   d.Location,
			Desc:   d.Description,
			Tags:   d.Tags,
			Params: params,
		}
		items = append(items, item)
		pickerModel.dagMap[d.Name] = d
	}

	// Create list with custom delegate for better rendering
	l := list.New(items, list.NewDefaultDelegate(), 80, 20)
	l.Title = "Select a DAG to run"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("DAG", "DAGs")
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.SetShowTitle(true)
	l.DisableQuitKeybindings()

	// Style the title
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("#6B46C1")).
		Foreground(lipgloss.Color("#FFFDF5")).
		Padding(0, 1)

	// Initialize parameter input
	ti := textinput.New()
	ti.Placeholder = "Enter parameters..."
	ti.Focus()
	ti.CharLimit = 1024
	ti.Width = 60

	// Initialize run ID input
	runIdTi := textinput.New()
	runIdTi.Placeholder = "Enter run ID (optional)..."
	runIdTi.CharLimit = 100
	runIdTi.Width = 60

	m := Model{
		state:      StateSelectingDAG,
		list:       l,
		paramInput: ti,
		dag:        dag,
		runIdInput: runIdTi,
	}

	// Use the default delegate since we'll handle DAG updates differently
	m.list.SetDelegate(list.NewDefaultDelegate())

	// Run the picker
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return Result{}, fmt.Errorf("failed to run DAG picker: %w", err)
	}

	// Get the selection
	finalM, ok := finalModel.(Model)
	if !ok {
		return Result{}, fmt.Errorf("unexpected model type")
	}

	// Build result
	if finalM.quitting || finalM.choice == nil {
		return Result{Cancelled: true}, nil
	}

	return Result{
		DAGName:   finalM.choice.Name,
		DAGPath:   finalM.choice.Path,
		Params:    finalM.params,
		Cancelled: false,
	}, nil
}

// PickDAG shows an interactive DAG picker and returns the selected DAG path
// Deprecated: Use PickDAGInteractive instead for a better user experience
func PickDAG(ctx context.Context, dagStore models.DAGStore) (string, error) {
	result, err := PickDAGInteractive(ctx, dagStore, nil)
	if err != nil {
		return "", err
	}
	if result.Cancelled {
		fmt.Println("No DAG selected.")
		os.Exit(0)
	}
	return result.DAGPath, nil
}

// PromptForParams prompts the user to enter parameters for a DAG
func PromptForParams(dag *digraph.DAG) (string, error) {
	if dag.DefaultParams == "" && len(dag.Params) == 0 {
		return "", nil
	}

	fmt.Println("\nThis DAG accepts the following parameters:")

	// Display default parameters if available
	if dag.DefaultParams != "" {
		fmt.Printf("  Default: %s\n", dag.DefaultParams)
	}

	// Display current parameters if available
	if len(dag.Params) > 0 {
		fmt.Printf("  Current: %s\n", strings.Join(dag.Params, " "))
	}

	fmt.Print("\nEnter parameters (press Enter to use defaults): ")

	// Read full line of input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	// Trim whitespace and return
	return strings.TrimSpace(input), nil
}

// ConfirmRunDAG prompts the user for Y/n confirmation before running a DAG
func ConfirmRunDAG(dagName string, params string) (bool, error) {
	// Build confirmation message
	fmt.Println()
	fmt.Printf("Ready to run DAG: %s\n", lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).Render(dagName))
	if params != "" {
		fmt.Printf("With parameters: %s\n", lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("214")).Render(params))
	}
	fmt.Print("\nRun this DAG? [Y/n]: ")

	// Read user input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	// Trim and convert to lowercase
	response := strings.ToLower(strings.TrimSpace(input))

	// Accept 'y', 'yes', or empty (default to yes)
	return response == "" || response == "y" || response == "yes", nil
}
