package tui

import (
	"fmt"
	"io"
	"strings"

	"apd/internal/wiki"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// WikiRequest contains the state and explicit initialization action for the wiki dashboard.
type WikiRequest struct {
	Target        string
	Workspace     wiki.Workspace
	Initialized   bool
	CanInitialize bool
	Initialize    func() (wiki.Workspace, error)
	Boundary      string
	Register      func(source, notes string) (wiki.Workspace, error)
	Prepare       func(wiki.Workspace, string) (wiki.Workspace, error)
	Emit          func(wiki.Workspace) (wiki.Workspace, string, error)
}

type wikiPhase int

const (
	wikiDashboard wikiPhase = iota
	wikiSourcePath
	wikiSourceNotes
	wikiTargetPath
)

// WikiModel presents the current safe wiki action without semantic integration behavior.
type WikiModel struct {
	request     WikiRequest
	width       int
	err         error
	phase       wikiPhase
	sourceInput textinput.Model
	notesInput  textinput.Model
	targetInput textinput.Model
	status      string
}

// NewWikiModel constructs a dashboard model from a read-only workspace snapshot.
func NewWikiModel(request WikiRequest) WikiModel {
	sourceInput := textinput.New()
	sourceInput.Prompt = "Source path: "
	sourceInput.Placeholder = "./notes.txt"
	sourceInput.CharLimit = 4096
	notesInput := textinput.New()
	notesInput.Prompt = "Notes or emphasis (optional): "
	notesInput.Placeholder = "What should a later integrator pay attention to?"
	notesInput.CharLimit = 4096
	targetInput := textinput.New()
	targetInput.Prompt = "Wiki target path: "
	targetInput.Placeholder = "wiki/topic.md"
	targetInput.CharLimit = 4096
	return WikiModel{request: request, width: 80, sourceInput: sourceInput, notesInput: notesInput, targetInput: targetInput}
}

// Init implements tea.Model.
func (m WikiModel) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m WikiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.updateInputWidths()
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.phase {
		case wikiSourcePath, wikiSourceNotes:
			return m.updateRegistration(msg)
		case wikiTargetPath:
			return m.updatePreparation(msg)
		}
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "i":
			if m.request.CanInitialize && m.request.Initialize != nil {
				workspace, err := m.request.Initialize()
				if err != nil {
					m.err = err
					return m, nil
				}
				m.request.Workspace = workspace
				m.request.Initialized = true
				m.request.CanInitialize = false
			}
		case "r":
			if m.canRegister() {
				m.phase = wikiSourcePath
				m.status = ""
				return m, m.sourceInput.Focus()
			}
		case "p":
			if m.canPrepare() {
				m.phase = wikiTargetPath
				m.status = ""
				return m, m.targetInput.Focus()
			}
		case "e":
			if m.canEmit() {
				workspace, path, err := m.request.Emit(m.request.Workspace)
				if err != nil {
					m.err = err
					return m, nil
				}
				m.request.Workspace = workspace
				m.status = "Integration request emitted: " + path
				m.err = nil
			}
		}
	}
	return m, nil
}

func (m WikiModel) updatePreparation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.phase = wikiDashboard
		m.targetInput.Blur()
		m.status = "Request preparation cancelled."
		return m, nil
	}
	if msg.String() == "enter" {
		target := strings.TrimSpace(m.targetInput.Value())
		if target == "" {
			m.err = fmt.Errorf("wiki target path is required")
			return m, nil
		}
		workspace, err := m.request.Prepare(m.request.Workspace, target)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.request.Workspace = workspace
		m.phase = wikiDashboard
		m.targetInput.Blur()
		m.status = "Request prepared. Next: " + workspace.NextAction
		m.err = nil
		return m, nil
	}
	var cmd tea.Cmd
	m.targetInput, cmd = m.targetInput.Update(msg)
	return m, cmd
}

func (m WikiModel) updateRegistration(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.phase = wikiDashboard
		m.sourceInput.Blur()
		m.notesInput.Blur()
		m.status = "Registration cancelled."
		return m, nil
	}
	if msg.String() == "enter" {
		if m.phase == wikiSourcePath {
			if strings.TrimSpace(m.sourceInput.Value()) == "" {
				m.err = fmt.Errorf("source path is required")
				return m, nil
			}
			m.phase = wikiSourceNotes
			m.err = nil
			m.sourceInput.Blur()
			return m, m.notesInput.Focus()
		}
		workspace, err := m.request.Register(strings.TrimSpace(m.sourceInput.Value()), strings.TrimSpace(m.notesInput.Value()))
		if err != nil {
			m.err = err
			return m, nil
		}
		m.request.Workspace = workspace
		m.phase = wikiDashboard
		m.notesInput.Blur()
		m.status = "Source registered. Next: " + workspace.NextAction
		m.err = nil
		return m, nil
	}
	var cmd tea.Cmd
	if m.phase == wikiSourcePath {
		m.sourceInput, cmd = m.sourceInput.Update(msg)
	} else {
		m.notesInput, cmd = m.notesInput.Update(msg)
	}
	return m, cmd
}

// View implements tea.Model.
func (m WikiModel) View() string {
	if m.phase == wikiSourcePath || m.phase == wikiSourceNotes || m.phase == wikiTargetPath {
		if m.phase == wikiTargetPath {
			return truncateWikiLines(m.preparationView(), m.width)
		}
		return truncateWikiLines(m.registrationView(), m.width)
	}
	view := WikiSnapshot(m.request)
	if !m.request.Initialized && m.request.CanInitialize {
		view += "\nPress i to initialize this workspace. Press q to quit.\n"
	} else if m.canRegister() {
		view += "\nNext: press r to register a local source. Press q to quit.\n"
	} else if m.canPrepare() {
		view += "\nNext: press p to prepare an external integration request. Press q to quit.\n"
	} else if m.canEmit() {
		view += "\nNext: press e to emit the external integration request. Press q to quit.\n"
	} else {
		view += "\nPress q to quit.\n"
	}
	if m.status != "" {
		view += "\n" + m.status + "\n"
	}
	if m.err != nil {
		view += "\nError: " + m.err.Error() + "\n"
	}
	return truncateWikiLines(view, m.width)
}

func (m WikiModel) canPrepare() bool {
	return m.request.Initialized && m.request.Workspace.Status == wiki.StatusRegistered && m.request.Prepare != nil
}

func (m WikiModel) canEmit() bool {
	return m.request.Initialized && m.request.Workspace.Status == wiki.StatusRequestReady && len(m.request.Workspace.ExpectedTargets) > 0 && m.request.Emit != nil
}

func (m WikiModel) preparationView() string {
	var b strings.Builder
	b.WriteString("APD Wiki: Prepare integration request\n")
	fmt.Fprintf(&b, "Workspace: %s\n", m.request.Target)
	b.WriteString("Name a path inside wiki/. APD will not edit it; an external agent uses it as a target.\n\n")
	b.WriteString(m.targetInput.View())
	b.WriteString("\n\nEnter prepares the request. Esc cancels without changes. Ctrl+C quits.\n")
	if m.err != nil {
		b.WriteString("\nError: " + m.err.Error() + "\n")
	}
	return b.String()
}

func (m WikiModel) canRegister() bool {
	return m.request.Initialized && m.request.Workspace.Status == wiki.StatusInitialized &&
		m.request.Workspace.NextAction == wiki.NextAction(wiki.StatusInitialized) && m.request.Register != nil
}

func (m WikiModel) registrationView() string {
	var b strings.Builder
	b.WriteString("APD Wiki: Register local source\n")
	fmt.Fprintf(&b, "Workspace: %s\n", m.request.Target)
	fmt.Fprintf(&b, "Source boundary: %s\n", m.request.Boundary)
	b.WriteString("Choose a readable regular file inside this boundary and outside the workspace. APD copies it unchanged.\n\n")
	b.WriteString(m.sourceInput.View())
	b.WriteString("\n")
	b.WriteString(m.notesInput.View())
	b.WriteString("\n\nEnter advances or registers. Esc cancels without changes. Ctrl+C quits.\n")
	if m.err != nil {
		b.WriteString("\nError: " + m.err.Error() + "\n")
	}
	return b.String()
}

func (m *WikiModel) updateInputWidths() {
	width := max(8, m.width-30)
	m.sourceInput.Width = width
	m.notesInput.Width = width
	m.targetInput.Width = width
}

// WikiSnapshot renders a deterministic, read-only status and next-action summary.
func WikiSnapshot(request WikiRequest) string {
	if !request.Initialized {
		next := "Select an absent child workspace and open it interactively to initialize it."
		if request.CanInitialize {
			next = "Open this workspace interactively and press i to initialize it."
		}
		return fmt.Sprintf("APD Wiki\nWorkspace: %s\nStatus: not initialized\nPending work: none\nNext action: %s\n", request.Target, next)
	}
	pending := "No source handoff is registered."
	if request.Workspace.Status == wiki.StatusRequestReady {
		pending = "Integration request is prepared but not emitted."
	}
	if request.Workspace.Status == wiki.StatusAwaitingExternalSemanticIntegration {
		pending = "External semantic integration remains pending."
	}
	requestPath := ""
	if request.Workspace.IntegrationRequestPath != "" {
		requestPath = "Integration request: " + request.Workspace.IntegrationRequestPath + "\nExternal agent: perform semantic integration; APD does not edit wiki content.\n"
	}
	return fmt.Sprintf("APD Wiki\nWorkspace: %s\nStatus: %s\nPending work: %s\n%sNext action: %s\n", request.Target, request.Workspace.Status, pending, requestPath, request.Workspace.NextAction)
}

// RunWiki executes the Bubble Tea wiki dashboard.
func RunWiki(in io.Reader, out io.Writer, request WikiRequest) error {
	program := tea.NewProgram(NewWikiModel(request), tea.WithInput(in), tea.WithOutput(out))
	raw, err := program.Run()
	if err != nil {
		return err
	}
	return raw.(WikiModel).err
}

func truncateWikiLines(value string, width int) string {
	if width < 1 {
		return value
	}
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		if len(line) > width {
			lines[i] = line[:max(0, width-3)] + "..."
		}
	}
	return strings.Join(lines, "\n")
}
