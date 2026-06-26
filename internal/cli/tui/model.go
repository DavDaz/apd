package tui

import (
	"fmt"
	"io"
	"strings"

	"apd/internal/app"
	apdcli "apd/internal/cli"
	"apd/internal/document"
	"apd/internal/templates"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type phase string

const (
	phaseSelect      phase = "select"
	phaseAuthor      phase = "author"
	phaseJump        phase = "jump"
	phaseReview      phase = "review"
	phaseFinalReview phase = "final-review"
	phaseConfirm     phase = "confirm"
	phaseDone        phase = "done"
)

type confirmAction string

const (
	confirmSkip   confirmAction = "skip"
	confirmBack   confirmAction = "back"
	confirmFinish confirmAction = "finish"
	confirmJump   confirmAction = "jump"
	confirmQuit   confirmAction = "quit"
)

type selectKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Choose key.Binding
	Quit   key.Binding
}

func (k selectKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Choose, k.Quit}
}

func (k selectKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Choose, k.Quit}}
}

type authorKeyMap struct {
	NewLine key.Binding
	Submit  key.Binding
	Next    key.Binding
	Prev    key.Binding
	Jump    key.Binding
	Help    key.Binding
	Skip    key.Binding
	Finish  key.Binding
	Quit    key.Binding
}

func (k authorKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NewLine, k.Submit, k.Help, k.Jump, k.Quit}
}

func (k authorKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.NewLine, k.Submit, k.Help, k.Jump}, {k.Next, k.Prev, k.Skip, k.Finish, k.Quit}}
}

type reviewKeyMap struct {
	Confirm key.Binding
	Edit    key.Binding
	Quit    key.Binding
}

func (k reviewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Edit, k.Quit}
}

func (k reviewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Confirm, k.Edit, k.Quit}}
}

type finalReviewKeyMap struct {
	Generate key.Binding
	Edit     key.Binding
	Quit     key.Binding
}

func (k finalReviewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Generate, k.Edit, k.Quit}
}

func (k finalReviewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Generate, k.Edit, k.Quit}}
}

type confirmKeyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
}

func (k confirmKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Cancel}
}

func (k confirmKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Confirm, k.Cancel}}
}

// Model is the first guided Bubble Tea slice for APD authoring.
type Model struct {
	templates       []templates.Template
	selected        int
	workflow        *app.GuidedWorkflow
	newWorkflow     func(templates.Template) *app.GuidedWorkflow
	phase           phase
	textarea        textarea.Model
	help            help.Model
	selectKeys      selectKeyMap
	authorKeys      authorKeyMap
	jumpKeys        jumpKeyMap
	reviewKeys      reviewKeyMap
	finalReviewKeys finalReviewKeyMap
	confirmKeys     confirmKeyMap
	showHelp        bool
	status          string
	sessionPath     string
	err             error
	finalResult     app.WorkflowResult
	pending         confirmAction
	pendingJump     int
	jumpSelection   int
	width           int
}

// NewModel constructs the guided TUI model.
func NewModel(req app.TUIRequest) (Model, error) {
	items := make([]templates.Template, 0, len(req.Registry.SupportedTypes()))
	for _, id := range req.Registry.SupportedTypes() {
		tmpl, _ := req.Registry.Resolve(id)
		items = append(items, tmpl)
	}
	m := Model{
		templates:       items,
		newWorkflow:     req.NewWorkflow,
		phase:           phaseSelect,
		textarea:        newTextarea(),
		help:            help.New(),
		selectKeys:      defaultSelectKeyMap(),
		authorKeys:      defaultAuthorKeyMap(),
		jumpKeys:        defaultJumpKeyMap(),
		reviewKeys:      defaultReviewKeyMap(),
		finalReviewKeys: defaultFinalReviewKeyMap(),
		confirmKeys:     defaultConfirmKeyMap(),
		pendingJump:     -1,
		width:           72,
	}
	if req.InitialType == "" {
		return m, nil
	}
	for _, tmpl := range items {
		if strings.EqualFold(tmpl.ID, req.InitialType) {
			m.start(tmpl)
			return m, nil
		}
	}
	return Model{}, fmt.Errorf("unsupported document type %q", req.InitialType)
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.textarea.SetWidth(max(20, typed.Width-6))
		return m, nil
	case tea.KeyMsg:
		switch m.phase {
		case phaseSelect:
			return m.updateSelect(typed)
		case phaseAuthor:
			return m.updateAuthor(typed)
		case phaseJump:
			return m.updateJump(typed)
		case phaseReview:
			return m.updateReview(typed)
		case phaseFinalReview:
			return m.updateFinalReview(typed)
		case phaseConfirm:
			return m.updateConfirm(typed)
		default:
			return m, nil
		}
	default:
		if m.phase != phaseAuthor {
			return m, nil
		}
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
}

func (m Model) updateSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.selectKeys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.selectKeys.Up):
		if m.selected > 0 {
			m.selected--
		}
	case key.Matches(msg, m.selectKeys.Down):
		if m.selected < len(m.templates)-1 {
			m.selected++
		}
	case key.Matches(msg, m.selectKeys.Choose):
		m.start(m.templates[m.selected])
	}
	return m, nil
}

func (m Model) updateAuthor(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.authorKeys.Help):
		return m.apply(apdcli.Intent{Kind: apdcli.IntentHelp})
	case key.Matches(msg, m.authorKeys.Jump):
		m.phase = phaseJump
		m.jumpSelection = m.currentJumpSelection()
		m.status = "Jump mode: revisit a current, answered, or skipped section."
		m.textarea.Blur()
		return m, nil
	case key.Matches(msg, m.authorKeys.Next):
		return m.continueSequentially()
	case key.Matches(msg, m.authorKeys.Skip):
		return m.confirmOrApply(confirmSkip, apdcli.Intent{Kind: apdcli.IntentSkip})
	case key.Matches(msg, m.authorKeys.Prev):
		return m.moveToPreviousSection()
	case key.Matches(msg, m.authorKeys.Finish):
		return m.confirmOrApply(confirmFinish, apdcli.Intent{Kind: apdcli.IntentDone})
	case key.Matches(msg, m.authorKeys.Quit):
		return m.confirmQuit()
	case key.Matches(msg, m.authorKeys.Submit):
		answer := m.textarea.Value()
		if strings.TrimSpace(answer) == "" {
			m.status = "Answer is empty. Use ctrl+s to skip or keep editing."
			return m, nil
		}
		m.phase = phaseReview
		m.status = "Review the answer before saving."
		m.textarea.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	if msg.Type == tea.KeySpace {
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	}
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m Model) updateJump(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.jumpKeys.Cancel):
		m.phase = phaseAuthor
		m.status = "Continue editing."
		m.textarea.Focus()
		return m, nil
	case key.Matches(msg, m.jumpKeys.Up):
		m.jumpSelection = m.moveJumpSelection(-1)
		return m, nil
	case key.Matches(msg, m.jumpKeys.Down):
		m.jumpSelection = m.moveJumpSelection(1)
		return m, nil
	case key.Matches(msg, m.jumpKeys.Choose):
		if !m.canJumpTo(m.jumpSelection) {
			m.status = "Pending future sections unlock once you reach them."
			return m, nil
		}
		if m.jumpSelection == m.currentSectionIndex() {
			m.phase = phaseAuthor
			m.status = "Continue editing the current section."
			m.textarea.Focus()
			return m, nil
		}
		if m.hasUnsavedChanges() {
			m.pending = confirmJump
			m.pendingJump = m.jumpSelection
			m.phase = phaseConfirm
			m.status = ""
			return m, nil
		}
		return m.jumpToSection(m.jumpSelection), nil
	}
	return m, nil
}

func (m Model) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.reviewKeys.Edit):
		m.phase = phaseAuthor
		m.status = "Continue editing before saving."
		m.textarea.Focus()
		return m, nil
	case key.Matches(msg, m.reviewKeys.Quit):
		return m.confirmQuit()
	case key.Matches(msg, m.reviewKeys.Confirm):
		m.textarea.Focus()
		return m.apply(apdcli.Intent{Kind: apdcli.IntentAnswer, Answer: m.textarea.Value()})
	}
	return m, nil
}

func (m Model) updateFinalReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.finalReviewKeys.Edit):
		return m.returnFromFinalReviewToEditing()
	case key.Matches(msg, m.finalReviewKeys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.finalReviewKeys.Generate):
		return m.apply(apdcli.Intent{Kind: apdcli.IntentDone})
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.confirmKeys.Cancel):
		m.pending = ""
		m.pendingJump = -1
		m.phase = phaseAuthor
		m.status = "Continue editing."
		m.textarea.Focus()
		return m, nil
	case key.Matches(msg, m.confirmKeys.Confirm):
		action := m.pending
		jumpIndex := m.pendingJump
		m.pending = ""
		m.pendingJump = -1
		switch action {
		case confirmQuit:
			return m, tea.Quit
		case confirmJump:
			return m.jumpToSection(jumpIndex), nil
		case confirmSkip:
			m.textarea.Focus()
			return m.apply(apdcli.Intent{Kind: apdcli.IntentSkip})
		case confirmBack:
			m.textarea.Focus()
			return m.apply(apdcli.Intent{Kind: apdcli.IntentBack})
		case confirmFinish:
			m.textarea.Focus()
			return m.apply(apdcli.Intent{Kind: apdcli.IntentDone})
		}
	}
	return m, nil
}

func (m Model) apply(intent apdcli.Intent) (tea.Model, tea.Cmd) {
	before := m.workflow.Document()
	result, err := m.workflow.Apply(intent)
	if err != nil {
		m.err = err
		m.phase = phaseDone
		return m, tea.Quit
	}
	if result.ShowHelp {
		m.showHelp = !m.showHelp
		if m.showHelp {
			m.status = "Section help opened. Continue editing."
		} else {
			m.status = "Section help hidden. Continue editing."
		}
		m.phase = phaseAuthor
		m.textarea.Focus()
		return m, nil
	}
	if result.SessionPath != "" {
		m.sessionPath = result.SessionPath
	}
	after := m.workflow.Document()
	if result.ReadyForFinalize {
		m.phase = phaseFinalReview
		m.status = completionStatus(intent, before)
		m.textarea.Blur()
		return m, nil
	}
	if result.Done {
		m.finalResult = result
		m.phase = phaseDone
		m.status = "Generated: " + result.MarkdownPath
		return m, tea.Quit
	}
	if after.Complete() && !after.AllSectionsResolved() {
		return m.resumeFirstPendingSection(), nil
	}
	m.phase = phaseAuthor
	m.syncTextarea()
	if status := transitionStatus(intent, before, after); status != "" {
		m.status = status
	} else if result.Message != "" {
		m.status = result.Message
	}
	return m, nil
}

func (m *Model) start(tmpl templates.Template) {
	m.workflow = m.newWorkflow(tmpl)
	m.phase = phaseAuthor
	m.syncTextarea()
	m.showHelp = false
	m.status = ""
	m.sessionPath = ""
	m.finalResult = app.WorkflowResult{}
	m.pending = ""
	m.pendingJump = -1
	m.jumpSelection = 0
}

// View implements tea.Model.
func (m Model) View() string {
	if m.phase == phaseSelect {
		var b strings.Builder
		b.WriteString("APD Guided Authoring\n")
		b.WriteString("Choose a document type\n\n")
		for i, tmpl := range m.templates {
			cursor := " "
			if i == m.selected {
				cursor = ">"
			}
			fmt.Fprintf(&b, "%s %s — %s (%d sections)\n", cursor, tmpl.Name, tmpl.Description, len(tmpl.Sections))
		}
		b.WriteString("\nNext: press enter to start a guided session.\n")
		b.WriteString("\n" + m.help.View(m.selectKeys) + "\n")
		return b.String()
	}
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n"
	}
	doc := m.workflow.Document()
	tmpl := m.workflow.Template()
	if m.phase == phaseDone {
		return fmt.Sprintf("Document generated.\nMarkdown: %s\nSession: %s\n", m.finalResult.MarkdownPath, m.finalResult.SessionPath)
	}
	answered, skipped, pending := sectionCounts(doc)
	var b strings.Builder
	fmt.Fprintf(&b, "APD Guided Authoring — %s\n", tmpl.Name)
	fmt.Fprintf(&b, "Progress: %d answered • %d skipped • %d pending\n", answered, skipped, pending)
	if m.sessionPath != "" {
		fmt.Fprintf(&b, "Session: %s\n", m.sessionPath)
	}
	b.WriteString("\nSections\n")
	for i, section := range doc.Sections {
		fmt.Fprintf(&b, "%s\n", m.sectionRailLine(doc, i, section))
	}
	if m.status != "" {
		fmt.Fprintf(&b, "\nStatus\n%s\n", m.status)
	}
	switch m.phase {
	case phaseAuthor:
		current := tmpl.Sections[doc.CurrentSectionIndex]
		fmt.Fprintf(&b, "\nCurrent section\n%s\n%s\n", current.Title, current.Description)
		if m.showHelp {
			b.WriteString("\nSection guidance\n")
			b.WriteString(renderSectionGuidance(current))
		}
		b.WriteString("\nEditor\n")
		b.WriteString(m.textarea.View())
		b.WriteString("\nNext: write the answer, press ctrl+d to review it, ctrl+h for guidance, ctrl+n to continue, or ctrl+j to jump/revisit.\n")
	case phaseJump:
		b.WriteString("\nJump mode\n")
		b.WriteString("Select a current, answered, or skipped section to revisit or edit. Pending future sections stay blocked until you reach them sequentially.\n")
		if target := sectionTitleAt(doc, m.jumpSelection); target != "" {
			fmt.Fprintf(&b, "Selected: %s\n", target)
		}
		b.WriteString("\nNext: press enter to revisit the selected section, or esc to keep the current draft.\n")
	case phaseReview:
		b.WriteString("\nReview answer\n")
		b.WriteString(indentBlock(m.textarea.Value()))
		b.WriteString("\n\nNext: press enter to save and continue, or esc to keep editing.\n")
	case phaseFinalReview:
		b.WriteString("\nFinal review\n")
		fmt.Fprintf(&b, "Answered: %d • Skipped: %d • Pending: %d\n", answered, skipped, pending)
		b.WriteString("All sections are complete. Review the summary, then choose what to do next.\n")
		b.WriteString("\nNext: press enter to generate, esc to edit the last section, or ctrl+c to quit.\n")
	case phaseConfirm:
		b.WriteString("\nConfirm action\n")
		b.WriteString(confirmPrompt(m.pending, m.workflow.Document(), m.hasUnsavedChanges(), sectionTitleAt(m.workflow.Document(), m.pendingJump)))
		b.WriteString("\n")
	}
	b.WriteString("\n" + m.help.View(m.currentKeyMap()) + "\n")
	return b.String()
}

// Run executes the Bubble Tea program.
func Run(in io.Reader, out io.Writer, req app.TUIRequest) error {
	model, err := NewModel(req)
	if err != nil {
		return err
	}
	program := tea.NewProgram(model, tea.WithInput(in), tea.WithOutput(out))
	raw, err := program.Run()
	if err != nil {
		return err
	}
	finalModel := raw.(Model)
	return finalModel.err
}

func (m *Model) syncTextarea() {
	value := ""
	if m.workflow != nil {
		doc := m.workflow.Document()
		if doc.CurrentSectionIndex >= 0 && doc.CurrentSectionIndex < len(doc.Sections) {
			value = doc.Sections[doc.CurrentSectionIndex].Answer
		}
	}
	m.textarea = newTextarea()
	m.textarea.SetWidth(max(20, m.width-6))
	m.textarea.SetValue(value)
	m.textarea.Focus()
}

func (m Model) currentKeyMap() help.KeyMap {
	switch m.phase {
	case phaseJump:
		return m.jumpKeys
	case phaseReview:
		return m.reviewKeys
	case phaseFinalReview:
		return m.finalReviewKeys
	case phaseConfirm:
		return m.confirmKeys
	}
	return m.authorKeys
}

func (m Model) confirmOrApply(action confirmAction, intent apdcli.Intent) (tea.Model, tea.Cmd) {
	if action == confirmFinish || m.hasUnsavedChanges() {
		m.pending = action
		m.pendingJump = -1
		m.phase = phaseConfirm
		m.status = ""
		m.textarea.Blur()
		return m, nil
	}
	return m.apply(intent)
}

func (m Model) confirmQuit() (tea.Model, tea.Cmd) {
	if !m.hasUnsavedChanges() {
		return m, tea.Quit
	}
	m.pending = confirmQuit
	m.pendingJump = -1
	m.phase = phaseConfirm
	m.status = ""
	m.textarea.Blur()
	return m, nil
}

func (m Model) hasUnsavedChanges() bool {
	if m.workflow == nil {
		return strings.TrimSpace(m.textarea.Value()) != ""
	}
	doc := m.workflow.Document()
	if doc.Complete() {
		return false
	}
	if doc.CurrentSectionIndex < 0 || doc.CurrentSectionIndex >= len(doc.Sections) {
		return strings.TrimSpace(m.textarea.Value()) != ""
	}
	return m.textarea.Value() != doc.Sections[doc.CurrentSectionIndex].Answer
}

func defaultSelectKeyMap() selectKeyMap {
	return selectKeyMap{
		Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move")),
		Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move")),
		Choose: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func defaultAuthorKeyMap() authorKeyMap {
	return authorKeyMap{
		NewLine: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "new line")),
		Submit:  key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "review")),
		Next:    key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "next")),
		Prev:    key.NewBinding(key.WithKeys("ctrl+p", "ctrl+b"), key.WithHelp("ctrl+p", "previous")),
		Jump:    key.NewBinding(key.WithKeys("ctrl+j"), key.WithHelp("ctrl+j", "jump/revisit")),
		Help:    key.NewBinding(key.WithKeys("ctrl+h", "ctrl+g"), key.WithHelp("ctrl+h", "section help")),
		Skip:    key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "skip")),
		Finish:  key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "finish")),
		Quit:    key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	}
}

func defaultReviewKeyMap() reviewKeyMap {
	return reviewKeyMap{
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		Edit:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "edit")),
		Quit:    key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	}
}

func defaultFinalReviewKeyMap() finalReviewKeyMap {
	return finalReviewKeyMap{
		Generate: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "generate")),
		Edit:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "edit last")),
		Quit:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	}
}

func defaultConfirmKeyMap() confirmKeyMap {
	return confirmKeyMap{
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		Cancel:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
}

func newTextarea() textarea.Model {
	ta := textarea.New()
	ta.Prompt = "┃ "
	ta.Placeholder = "Write the answer. Enter adds a new line; ctrl+d opens review; ctrl+h shows guidance."
	ta.ShowLineNumbers = false
	ta.KeyMap.DeleteCharacterBackward = key.NewBinding(key.WithKeys("backspace"), key.WithHelp("backspace", "delete character backward"))
	ta.SetWidth(66)
	ta.SetHeight(8)
	ta.Focus()
	return ta
}

func renderSectionGuidance(section templates.Section) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Description: %s\n", section.Description)

	hasExtraGuidance := false
	if section.Help != "" {
		hasExtraGuidance = true
		fmt.Fprintf(&b, "Help: %s\n", section.Help)
	}
	if section.Example != "" {
		hasExtraGuidance = true
		fmt.Fprintf(&b, "Example: %s\n", section.Example)
	}
	if len(section.Questions) > 0 {
		hasExtraGuidance = true
		b.WriteString("Questions:\n")
		for _, question := range section.Questions {
			fmt.Fprintf(&b, "- %s\n", question)
		}
	}
	if len(section.ContextKeys) > 0 {
		hasExtraGuidance = true
		b.WriteString("Context keys:\n")
		for _, key := range section.ContextKeys {
			fmt.Fprintf(&b, "- %s\n", key)
		}
	}
	if !hasExtraGuidance {
		b.WriteString("No extra section metadata is available. Use the title and description as the main guidance.\n")
	}

	return b.String()
}

func indentBlock(value string) string {
	if value == "" {
		return "  (empty)"
	}
	return "  " + strings.ReplaceAll(value, "\n", "\n  ")
}

func transitionStatus(intent apdcli.Intent, before, after document.Document) string {
	switch intent.Kind {
	case apdcli.IntentAnswer:
		currentTitle := sectionTitleAt(before, before.CurrentSectionIndex)
		nextTitle := sectionTitleAt(after, after.CurrentSectionIndex)
		if currentTitle != "" && nextTitle != "" {
			return fmt.Sprintf("Saved %s. Now editing %s.", currentTitle, nextTitle)
		}
	case apdcli.IntentSkip:
		currentTitle := sectionTitleAt(before, before.CurrentSectionIndex)
		nextTitle := sectionTitleAt(after, after.CurrentSectionIndex)
		if currentTitle != "" && nextTitle != "" {
			return fmt.Sprintf("Skipped %s. Now editing %s.", currentTitle, nextTitle)
		}
	case apdcli.IntentBack:
		currentTitle := sectionTitleAt(after, after.CurrentSectionIndex)
		if currentTitle != "" {
			return fmt.Sprintf("Back to %s. Continue editing.", currentTitle)
		}
	}
	return ""
}

func completionStatus(intent apdcli.Intent, before document.Document) string {
	sectionTitle := sectionTitleAt(before, before.CurrentSectionIndex)
	switch intent.Kind {
	case apdcli.IntentAnswer:
		if sectionTitle != "" {
			return fmt.Sprintf("Saved %s. Final review is ready.", sectionTitle)
		}
	case apdcli.IntentSkip:
		if sectionTitle != "" {
			return fmt.Sprintf("Skipped %s. Final review is ready.", sectionTitle)
		}
	}
	return "Final review is ready."
}

func isFinalSection(doc document.Document) bool {
	return doc.CurrentSectionIndex >= 0 && doc.CurrentSectionIndex == len(doc.Sections)-1
}

func sectionTitleAt(doc document.Document, idx int) string {
	if idx < 0 || idx >= len(doc.Sections) {
		return ""
	}
	return doc.Sections[idx].Title
}

func sectionCounts(doc document.Document) (answered, skipped, pending int) {
	for _, section := range doc.Sections {
		switch section.Status {
		case document.StatusAnswered:
			answered++
		case document.StatusSkipped:
			skipped++
		default:
			pending++
		}
	}
	return answered, skipped, pending
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
