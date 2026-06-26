package tui

import (
	"fmt"
	"strings"

	apdcli "apd/internal/cli"
	"apd/internal/document"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type jumpKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Choose key.Binding
	Cancel key.Binding
}

func (k jumpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Choose, k.Cancel}
}

func (k jumpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Choose, k.Cancel}}
}

func defaultJumpKeyMap() jumpKeyMap {
	return jumpKeyMap{
		Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move")),
		Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move")),
		Choose: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "revisit")),
		Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
}

func confirmPrompt(action confirmAction, doc document.Document, unsaved bool, jumpTarget string) string {
	var b strings.Builder
	switch action {
	case confirmJump:
		if jumpTarget != "" {
			fmt.Fprintf(&b, "Jump to %s and discard the current draft?\n", jumpTarget)
		} else {
			b.WriteString("Jump away and discard the current draft?\n")
		}
	case confirmSkip:
		if isFinalSection(doc) {
			b.WriteString("Skip this final section and discard the current draft?\n")
			b.WriteString("You will go to final review before generating anything.\n")
		} else {
			b.WriteString("Skip this section and discard the current draft?\n")
			if nextTitle := sectionTitleAt(doc, doc.CurrentSectionIndex+1); nextTitle != "" {
				fmt.Fprintf(&b, "You will move to %s.\n", nextTitle)
			}
		}
	case confirmBack:
		if previousTitle := sectionTitleAt(doc, doc.CurrentSectionIndex-1); previousTitle != "" {
			fmt.Fprintf(&b, "Return to %s and discard the current draft?\n", previousTitle)
		} else {
			b.WriteString("Return to the previous section and discard the current draft?\n")
		}
	case confirmFinish:
		answered, skipped, pending := sectionCounts(doc)
		if pending > 0 {
			b.WriteString("Generate partial document?\n")
			fmt.Fprintf(&b, "%d pending section(s) will stay unfinished.\n", pending)
		} else {
			b.WriteString("Generate the document now?\n")
		}
		if unsaved {
			b.WriteString("The current draft will be discarded.\n")
		}
		fmt.Fprintf(&b, "Answered: %d • Skipped: %d • Pending: %d\n", answered, skipped, pending)
	case confirmQuit:
		b.WriteString("Quit and discard the current draft?\n")
	}
	b.WriteString("Press enter to confirm or esc to cancel.")
	return b.String()
}

func (m Model) sectionRailLine(doc document.Document, idx int, section document.SectionState) string {
	prefix := " "
	if m.phase == phaseJump && idx == m.jumpSelection {
		prefix = "›"
	}
	return fmt.Sprintf("%s %s %d. %s", prefix, sectionTags(section.Status, idx == doc.CurrentSectionIndex), idx+1, section.Title)
}

func sectionTags(status document.SectionStatus, current bool) string {
	label := string(status)
	if label == "" {
		label = string(document.StatusPending)
	}
	tags := make([]string, 0, 2)
	if current {
		tags = append(tags, "current")
	}
	tags = append(tags, label)
	parts := make([]string, 0, len(tags))
	for _, tag := range tags {
		parts = append(parts, fmt.Sprintf("[%s]", tag))
	}
	return strings.Join(parts, " ")
}

func (m Model) currentJumpSelection() int {
	current := m.currentSectionIndex()
	if m.canJumpTo(current) {
		return current
	}
	if m.workflow == nil {
		return 0
	}
	for i := len(m.workflow.Document().Sections) - 1; i >= 0; i-- {
		if m.canJumpTo(i) {
			return i
		}
	}
	return 0
}

func (m Model) moveJumpSelection(step int) int {
	if m.workflow == nil || step == 0 {
		return m.jumpSelection
	}
	sections := m.workflow.Document().Sections
	for idx := m.jumpSelection + step; idx >= 0 && idx < len(sections); idx += step {
		if m.canJumpTo(idx) {
			return idx
		}
	}
	return m.jumpSelection
}

func (m Model) canJumpTo(idx int) bool {
	if m.workflow == nil {
		return false
	}
	doc := m.workflow.Document()
	if idx < 0 || idx >= len(doc.Sections) {
		return false
	}
	if idx == doc.CurrentSectionIndex {
		return true
	}
	return doc.Sections[idx].Status != document.StatusPending
}

func (m Model) currentSectionIndex() int {
	if m.workflow == nil {
		return -1
	}
	return m.workflow.Document().CurrentSectionIndex
}

func (m Model) jumpToSection(idx int) Model {
	if m.workflow == nil || !m.workflow.JumpToSection(idx) {
		m.status = "Could not jump to that section."
		m.phase = phaseAuthor
		m.textarea.Focus()
		return m
	}
	m.phase = phaseAuthor
	m.jumpSelection = idx
	m.syncTextarea()
	section := m.workflow.Document().Sections[idx]
	switch section.Status {
	case document.StatusAnswered:
		m.status = fmt.Sprintf("Editing saved answer for %s.", section.Title)
	case document.StatusSkipped:
		m.status = fmt.Sprintf("Reopened skipped section %s.", section.Title)
	default:
		m.status = fmt.Sprintf("Continue editing %s.", section.Title)
	}
	return m
}

func (m Model) openReview(status string) Model {
	m.phase = phaseReview
	m.status = status
	m.textarea.Blur()
	return m
}

func (m Model) openFinalReview(status string) Model {
	m.phase = phaseFinalReview
	m.status = status
	m.textarea.Blur()
	return m
}

func (m Model) continueSequentially() (tea.Model, tea.Cmd) {
	if m.workflow == nil {
		return m, nil
	}
	doc := m.workflow.Document()
	if doc.Complete() {
		if doc.AllSectionsResolved() {
			return m.openFinalReview("All sections are complete. Final review is ready."), nil
		}
		return m.resumeFirstPendingSection(), nil
	}
	current := doc.Sections[doc.CurrentSectionIndex]
	if current.Status == document.StatusPending || m.hasUnsavedChanges() {
		if strings.TrimSpace(m.textarea.Value()) == "" {
			m.status = "Answer is empty. Use ctrl+s to skip or keep editing."
			return m, nil
		}
		return m.openReview("Review the answer before saving."), nil
	}
	nextIndex := doc.CurrentSectionIndex + 1
	if nextIndex >= len(doc.Sections) {
		if doc.AllSectionsResolved() {
			return m.openFinalReview("All sections are complete. Final review is ready."), nil
		}
		return m.resumeFirstPendingSection(), nil
	}
	return m.jumpToSection(nextIndex), nil
}

func (m Model) moveToPreviousSection() (tea.Model, tea.Cmd) {
	if m.workflow == nil {
		return m, nil
	}
	doc := m.workflow.Document()
	if doc.Complete() {
		if len(doc.Sections) == 0 {
			return m, nil
		}
		return m.jumpToSection(len(doc.Sections) - 1), nil
	}
	if doc.CurrentSectionIndex <= 0 {
		m.status = "Already at the first section."
		return m, nil
	}
	if m.hasUnsavedChanges() {
		m.pending = confirmBack
		m.pendingJump = -1
		m.phase = phaseConfirm
		m.status = ""
		m.textarea.Blur()
		return m, nil
	}
	return m.jumpToSection(doc.CurrentSectionIndex - 1), nil
}

func (m Model) resumeFirstPendingSection() Model {
	if m.workflow == nil {
		return m
	}
	idx := m.workflow.Document().FirstPendingSectionIndex()
	if idx < 0 {
		return m.openFinalReview("All sections are complete. Final review is ready.")
	}
	m = m.jumpToSection(idx)
	if title := sectionTitleAt(m.workflow.Document(), idx); title != "" {
		m.status = fmt.Sprintf("Continue with %s.", title)
	}
	return m
}

func (m Model) returnFromFinalReviewToEditing() (tea.Model, tea.Cmd) {
	if m.workflow == nil {
		return m, nil
	}
	doc := m.workflow.Document()
	if doc.Complete() {
		m.textarea.Focus()
		return m.apply(apdcli.Intent{Kind: apdcli.IntentBack})
	}
	return m.jumpToSection(doc.CurrentSectionIndex), nil
}
