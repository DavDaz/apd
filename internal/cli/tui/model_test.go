package tui

import (
	"strings"
	"testing"
	"time"

	"apd/internal/app"
	"apd/internal/document"
	"apd/internal/templates"
	tea "github.com/charmbracelet/bubbletea"
)

type fakeStore struct{ docs []document.Document }

func (s *fakeStore) Save(doc document.Document) (string, error) {
	s.docs = append(s.docs, doc)
	return "session.yaml", nil
}

type fakeExporter struct{ docs []document.Document }

func (e *fakeExporter) Write(doc document.Document, _ templates.Template) (string, error) {
	e.docs = append(e.docs, doc)
	return "document.md", nil
}

type modelFixture struct {
	model    Model
	store    *fakeStore
	exporter *fakeExporter
}

func TestModelSelectionTransitionsAndView(t *testing.T) {
	fx := newModelFixture(t)
	model := fx.model
	view := model.View()
	if !strings.Contains(view, "Product — desc (2 sections)") || !strings.Contains(view, "Next: press enter to start a guided session.") {
		t.Fatalf("selection view missing guidance:\n%s", view)
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	view = model.View()
	if !strings.Contains(view, "[current] [pending] 1. One") || !strings.Contains(view, "[pending] 2. Two") {
		t.Fatalf("author view missing section orientation:\n%s", view)
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlG}).(Model)
	if !strings.Contains(model.View(), "Section guidance") || !strings.Contains(model.View(), "Help: help text") {
		t.Fatalf("help toggle did not render help:\n%s", model.View())
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlS}).(Model)
	if !strings.Contains(model.View(), "[skipped] 1. One") || !strings.Contains(model.View(), "[current] [pending] 2. Two") {
		t.Fatalf("skip did not update section states:\n%s", model.View())
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlP}).(Model)
	if !strings.Contains(model.View(), "[current] [skipped] 1. One") {
		t.Fatalf("back did not return to previous section:\n%s", model.View())
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlF}).(Model)
	if model.phase != phaseConfirm {
		t.Fatalf("phase = %s, want confirm before finish", model.phase)
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	if model.phase != phaseDone || model.finalResult.MarkdownPath != "document.md" {
		t.Fatalf("finish did not finalize document: %+v", model.finalResult)
	}
}

func TestModelSpaceKeyInsertsSpace(t *testing.T) {
	model := startAuthoring(t)
	model = typeText(model, "hello")
	model = press(model, tea.KeyMsg{Type: tea.KeySpace}).(Model)
	model = typeText(model, "world")
	if got := model.textarea.Value(); got != "hello world" {
		t.Fatalf("textarea value = %q, want %q", got, "hello world")
	}
}

func TestModelEnterInsertsNewlineWithoutSaving(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = typeText(model, "first line")
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = typeText(model, "second line")
	if got := model.textarea.Value(); got != "first line\nsecond line" {
		t.Fatalf("textarea value = %q, want multiline editor content", got)
	}
	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author", model.phase)
	}
	if got := model.workflow.Document().CurrentSectionIndex; got != 0 {
		t.Fatalf("current section = %d, want 0 before explicit submit", got)
	}
	if len(fx.store.docs) != 0 {
		t.Fatalf("Save calls = %d, want 0 before explicit submit", len(fx.store.docs))
	}
}

func TestModelCommandLikeCharactersRemainTextInEditMode(t *testing.T) {
	model := startAuthoring(t)
	for _, msg := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
		{Type: tea.KeySpace},
		{Type: tea.KeyRunes, Runes: []rune{'s'}},
		{Type: tea.KeySpace},
		{Type: tea.KeyRunes, Runes: []rune{'b'}},
		{Type: tea.KeySpace},
		{Type: tea.KeyRunes, Runes: []rune{'f'}},
		{Type: tea.KeySpace},
		{Type: tea.KeyRunes, Runes: []rune{'?'}},
	} {
		model = press(model, msg).(Model)
	}
	if got := model.textarea.Value(); got != "q s b f ?" {
		t.Fatalf("textarea value = %q, want %q", got, "q s b f ?")
	}
	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author", model.phase)
	}
	if strings.Contains(model.View(), "Help: help text") {
		t.Fatalf("text entry unexpectedly toggled help:\n%s", model.View())
	}
}

func TestModelCtrlHShowsSectionGuidanceWithoutDeletingDraft(t *testing.T) {
	model := startAuthoring(t)
	model = typeText(model, "draft help text")
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlH}).(Model)
	if got := model.textarea.Value(); got != "draft help text" {
		t.Fatalf("textarea value = %q, want draft preserved after ctrl+h", got)
	}
	view := model.View()
	if !strings.Contains(view, "Section guidance") {
		t.Fatalf("ctrl+h did not open section guidance:\n%s", view)
	}
	if !strings.Contains(view, "Description: first") || !strings.Contains(view, "Help: help text") || !strings.Contains(view, "Example: example") || !strings.Contains(view, "Questions:") {
		t.Fatalf("ctrl+h guidance missing template metadata:\n%s", view)
	}
	if !strings.Contains(view, "ctrl+h") || !strings.Contains(view, "section help") {
		t.Fatalf("help UI does not advertise ctrl+h:\n%s", view)
	}

	model = press(model, tea.KeyMsg{Type: tea.KeyBackspace}).(Model)
	if got := model.textarea.Value(); got != "draft help tex" {
		t.Fatalf("backspace value = %q, want standard delete behavior after ctrl+h help alias", got)
	}
}

func TestModelExplicitSubmitReviewConfirmsAndAdvances(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = typeText(model, "ships before finish?")
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = typeText(model, "second line")
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlD}).(Model)
	if model.phase != phaseReview {
		t.Fatalf("phase = %s, want review", model.phase)
	}
	if len(fx.store.docs) != 0 {
		t.Fatalf("Save calls = %d, want 0 before confirm", len(fx.store.docs))
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	doc := model.workflow.Document()
	if got := doc.Sections[0].Answer; got != "ships before finish?\nsecond line" {
		t.Fatalf("answer = %q, want multiline preservation", got)
	}
	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author to continue on next section", model.phase)
	}
	if got := doc.CurrentSectionIndex; got != 1 {
		t.Fatalf("current section = %d, want 1 after confirm", got)
	}
	if len(fx.store.docs) != 1 {
		t.Fatalf("Save calls = %d, want 1 after confirm", len(fx.store.docs))
	}
	view := model.View()
	if !strings.Contains(view, "Saved One. Now editing Two.") {
		t.Fatalf("submit status missing next-step message:\n%s", view)
	}
	if !strings.Contains(view, "Session: session.yaml") {
		t.Fatalf("submit view missing session path:\n%s", view)
	}
}

func TestModelHappyPathReviewConfirmShowsFinalReviewOnlyAfterAllSectionsResolved(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = typeText(model, "first answer")
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlD}).(Model)
	if model.phase != phaseReview {
		t.Fatalf("phase = %s, want review after ctrl+d", model.phase)
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author after first confirm", model.phase)
	}
	model = typeText(model, "second answer")
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlD}).(Model)
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	if model.phase != phaseFinalReview {
		t.Fatalf("phase = %s, want final review after all sections resolve", model.phase)
	}
	if len(fx.exporter.docs) != 0 {
		t.Fatalf("Export calls = %d, want 0 before explicit generate", len(fx.exporter.docs))
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	if model.phase != phaseDone {
		t.Fatalf("phase = %s, want done after explicit generate", model.phase)
	}
	if len(fx.exporter.docs) != 1 {
		t.Fatalf("Export calls = %d, want 1 after explicit generate", len(fx.exporter.docs))
	}
}

func TestModelNextAdvancesResolvedSectionAndProtectsUnsavedDraft(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = submitAnswer(model, "first answer")
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlP}).(Model)
	if got := model.workflow.Document().CurrentSectionIndex; got != 0 {
		t.Fatalf("current section = %d, want 0 after previous", got)
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlN}).(Model)
	if got := model.workflow.Document().CurrentSectionIndex; got != 1 {
		t.Fatalf("current section = %d, want 1 after next on saved section", got)
	}
	model = typeText(model, "draft second")
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlN}).(Model)
	if model.phase != phaseReview {
		t.Fatalf("phase = %s, want review before advancing unsaved draft", model.phase)
	}
	if len(fx.store.docs) != 1 {
		t.Fatalf("Save calls = %d, want 1 before confirming second section", len(fx.store.docs))
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyEsc}).(Model)
	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author after canceling review", model.phase)
	}
	if got := model.textarea.Value(); got != "draft second" {
		t.Fatalf("textarea value = %q, want draft preserved", got)
	}
}

func TestModelPreviousReturnsToSavedAnswer(t *testing.T) {
	model := startAuthoring(t)
	model = submitAnswer(model, "first answer")
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlP}).(Model)
	if got := model.workflow.Document().CurrentSectionIndex; got != 0 {
		t.Fatalf("current section = %d, want 0 after previous", got)
	}
	if got := model.textarea.Value(); got != "first answer" {
		t.Fatalf("textarea value = %q, want saved answer", got)
	}
}

func TestModelJumpModeOpensWithoutTurningPrintableKeysIntoNavigation(t *testing.T) {
	model := startAuthoring(t)
	model = press(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}).(Model)
	if got := model.textarea.Value(); got != "j" {
		t.Fatalf("textarea value = %q, want %q", got, "j")
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlJ}).(Model)
	if model.phase != phaseJump {
		t.Fatalf("phase = %s, want jump", model.phase)
	}
	view := model.View()
	if !strings.Contains(view, "jump/revisit") && !strings.Contains(view, "revisit or edit") {
		t.Fatalf("jump copy should explain revisit behavior:\n%s", view)
	}
}

func TestModelSectionRailShowsCurrentAndStatusTogether(t *testing.T) {
	model := startAuthoring(t)
	model = submitAnswer(model, "first answer")
	view := model.View()
	if !strings.Contains(view, "[answered] 1. One") {
		t.Fatalf("section rail missing answered state:\n%s", view)
	}
	if !strings.Contains(view, "[current] [pending] 2. Two") {
		t.Fatalf("section rail missing current and pending state together:\n%s", view)
	}
}

func TestModelJumpLoadsPreviousAnswerIntoTextarea(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = submitAnswer(model, "first answer")

	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlJ}).(Model)
	if model.phase != phaseJump {
		t.Fatalf("phase = %s, want jump", model.phase)
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyUp}).(Model)
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author after jump", model.phase)
	}
	if got := model.workflow.Document().CurrentSectionIndex; got != 0 {
		t.Fatalf("current section = %d, want 0 after jump", got)
	}
	if got := model.textarea.Value(); got != "first answer" {
		t.Fatalf("textarea value = %q, want saved answer", got)
	}
}

func TestModelJumpWithUnsavedTextRequiresConfirmationAndCancelPreservesDraft(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = submitAnswer(model, "first answer")
	model = typeText(model, "draft second")

	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlJ}).(Model)
	model = press(model, tea.KeyMsg{Type: tea.KeyUp}).(Model)
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	if model.phase != phaseConfirm {
		t.Fatalf("phase = %s, want confirm before jump", model.phase)
	}
	if !strings.Contains(model.View(), "Jump to One and discard the current draft?") {
		t.Fatalf("jump confirmation missing warning:\n%s", model.View())
	}

	model = press(model, tea.KeyMsg{Type: tea.KeyEsc}).(Model)
	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author after cancel", model.phase)
	}
	if got := model.workflow.Document().CurrentSectionIndex; got != 1 {
		t.Fatalf("current section = %d, want 1 after cancel", got)
	}
	if got := model.textarea.Value(); got != "draft second" {
		t.Fatalf("textarea value = %q, want draft preserved", got)
	}
}

func TestModelConfirmedJumpDoesNotExportOrQuit(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = submitAnswer(model, "first answer")
	model = typeText(model, "draft second")

	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlJ}).(Model)
	model = press(model, tea.KeyMsg{Type: tea.KeyUp}).(Model)
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)

	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author after confirmed jump", model.phase)
	}
	if got := model.workflow.Document().CurrentSectionIndex; got != 0 {
		t.Fatalf("current section = %d, want 0 after confirmed jump", got)
	}
	if len(fx.exporter.docs) != 0 {
		t.Fatalf("Export calls = %d, want 0 after jump", len(fx.exporter.docs))
	}
	if model.finalResult.MarkdownPath != "" {
		t.Fatalf("markdown path = %q, want empty before explicit generate", model.finalResult.MarkdownPath)
	}
}

func TestModelSkipWithUnsavedTextRequiresConfirmation(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = typeText(model, "draft")

	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlS}).(Model)
	if model.phase != phaseConfirm {
		t.Fatalf("phase = %s, want confirm", model.phase)
	}
	if got := model.workflow.Document().Sections[0].Status; got != document.StatusPending {
		t.Fatalf("section status = %s, want pending before confirm", got)
	}
	if len(fx.store.docs) != 0 {
		t.Fatalf("Save calls = %d, want 0 before confirm", len(fx.store.docs))
	}
	view := model.View()
	if !strings.Contains(view, "Skip this section and discard the current draft?") {
		t.Fatalf("skip confirmation missing warning:\n%s", view)
	}
	if !strings.Contains(view, "You will move to Two.") {
		t.Fatalf("skip confirmation missing next-step copy:\n%s", view)
	}
	if strings.Contains(strings.ToLower(view), "not saved") {
		t.Fatalf("skip confirmation uses misleading save language:\n%s", view)
	}

	model = press(model, tea.KeyMsg{Type: tea.KeyEsc}).(Model)
	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author after cancel", model.phase)
	}
	if got := model.textarea.Value(); got != "draft" {
		t.Fatalf("textarea value = %q, want draft preserved", got)
	}
	if len(fx.store.docs) != 0 {
		t.Fatalf("Save calls = %d, want 0 after cancel", len(fx.store.docs))
	}
}

func TestModelSkipWithUnsavedTextOnNonFinalSectionAdvancesWithStatus(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = typeText(model, "draft")

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	model = next.(Model)
	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = next.(Model)

	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author after skip confirm", model.phase)
	}
	if got := model.workflow.Document().CurrentSectionIndex; got != 1 {
		t.Fatalf("current section = %d, want 1 after skip", got)
	}
	if len(fx.exporter.docs) != 0 {
		t.Fatalf("Export calls = %d, want 0 after non-final skip", len(fx.exporter.docs))
	}
	view := model.View()
	if !strings.Contains(view, "Skipped One. Now editing Two.") {
		t.Fatalf("skip status missing next-step message:\n%s", view)
	}
}

func TestModelBackWithUnsavedTextRequiresConfirmation(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = typeText(model, "first")
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlD}).(Model)
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = typeText(model, "draft second")

	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlB}).(Model)
	if model.phase != phaseConfirm {
		t.Fatalf("phase = %s, want confirm", model.phase)
	}
	if got := model.workflow.Document().CurrentSectionIndex; got != 1 {
		t.Fatalf("current section = %d, want 1 before confirm", got)
	}
	if len(fx.store.docs) != 1 {
		t.Fatalf("Save calls = %d, want 1 before confirm", len(fx.store.docs))
	}
	if !strings.Contains(model.View(), "Return to One and discard the current draft?") {
		t.Fatalf("back confirmation missing warning:\n%s", model.View())
	}

	model = press(model, tea.KeyMsg{Type: tea.KeyEsc}).(Model)
	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author after cancel", model.phase)
	}
	if got := model.textarea.Value(); got != "draft second" {
		t.Fatalf("textarea value = %q, want draft preserved", got)
	}
	if got := model.workflow.Document().CurrentSectionIndex; got != 1 {
		t.Fatalf("current section = %d, want 1 after cancel", got)
	}
}

func TestModelFinishRequiresConfirmationAndShowsSummary(t *testing.T) {
	model := startAuthoring(t)
	model = typeText(model, "draft")

	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlF}).(Model)
	if model.phase != phaseConfirm {
		t.Fatalf("phase = %s, want confirm", model.phase)
	}
	view := model.View()
	if !strings.Contains(view, "Generate partial document?") {
		t.Fatalf("finish confirmation missing heading:\n%s", view)
	}
	if !strings.Contains(view, "2 pending section(s) will stay unfinished.") {
		t.Fatalf("finish confirmation missing partial warning:\n%s", view)
	}
	if !strings.Contains(view, "The current draft will be discarded.") {
		t.Fatalf("finish confirmation missing unsaved warning:\n%s", view)
	}
	if !strings.Contains(view, "Answered: 0 • Skipped: 0 • Pending: 2") {
		t.Fatalf("finish confirmation missing section summary:\n%s", view)
	}
}

func TestModelQuitWithUnsavedTextRequiresConfirmation(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = typeText(model, "draft")

	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlC}).(Model)
	if model.phase != phaseConfirm {
		t.Fatalf("phase = %s, want confirm", model.phase)
	}
	if !strings.Contains(model.View(), "Quit and discard the current draft?") {
		t.Fatalf("quit confirmation missing warning:\n%s", model.View())
	}
	if len(fx.store.docs) != 0 {
		t.Fatalf("Save calls = %d, want 0 before quit confirm", len(fx.store.docs))
	}

	model = press(model, tea.KeyMsg{Type: tea.KeyEsc}).(Model)
	if got := model.textarea.Value(); got != "draft" {
		t.Fatalf("textarea value = %q, want draft preserved after cancel", got)
	}

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model = next.(Model)
	if model.phase != phaseConfirm {
		t.Fatalf("phase = %s, want confirm on second quit attempt", model.phase)
	}
	next, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if _, ok := next.(Model); !ok {
		t.Fatalf("next model type = %T, want tui.Model", next)
	}
	if cmd == nil {
		t.Fatal("quit confirm returned nil cmd, want tea.Quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("quit confirm cmd() = %T, want tea.QuitMsg", cmd())
	}
}

func TestModelSkipWithUnsavedTextOnFinalSectionOpensFinalReview(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = submitAnswer(model, "first answer")
	model = typeText(model, "last draft")

	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlS}).(Model)
	view := model.View()
	if !strings.Contains(view, "Skip this final section and discard the current draft?") {
		t.Fatalf("final skip confirmation missing final-review copy:\n%s", view)
	}
	if !strings.Contains(view, "You will go to final review before generating anything.") {
		t.Fatalf("final skip confirmation missing final-review guidance:\n%s", view)
	}
	if strings.Contains(strings.ToLower(view), "not saved") {
		t.Fatalf("final skip confirmation uses misleading save language:\n%s", view)
	}

	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	if model.phase != phaseFinalReview {
		t.Fatalf("phase = %s, want final review", model.phase)
	}
	if len(fx.exporter.docs) != 0 {
		t.Fatalf("Export calls = %d, want 0 before explicit generate", len(fx.exporter.docs))
	}
	view = model.View()
	if !strings.Contains(view, "Final review") || !strings.Contains(view, "Answered: 1 • Skipped: 1 • Pending: 0") {
		t.Fatalf("final review missing summary:\n%s", view)
	}
	if !strings.Contains(view, "Skipped Two. Final review is ready.") {
		t.Fatalf("final review missing status:\n%s", view)
	}
}

func TestModelFinalReviewGenerateExportsAndQuits(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = submitAnswer(model, "first answer")
	model = submitAnswer(model, "second answer")
	if model.phase != phaseFinalReview {
		t.Fatalf("phase = %s, want final review before generate", model.phase)
	}

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = next.(Model)
	if model.phase != phaseDone {
		t.Fatalf("phase = %s, want done after generate", model.phase)
	}
	if model.finalResult.MarkdownPath != "document.md" {
		t.Fatalf("markdown path = %q, want document.md", model.finalResult.MarkdownPath)
	}
	if len(fx.exporter.docs) != 1 {
		t.Fatalf("Export calls = %d, want 1 after explicit generate", len(fx.exporter.docs))
	}
	if cmd == nil {
		t.Fatal("generate returned nil cmd, want tea.Quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("generate cmd() = %T, want tea.QuitMsg", cmd())
	}
}

func TestModelFinalReviewEditReturnsToEditingSafely(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = submitAnswer(model, "first answer")
	model = submitAnswer(model, "second answer")
	if model.phase != phaseFinalReview {
		t.Fatalf("phase = %s, want final review", model.phase)
	}

	model = press(model, tea.KeyMsg{Type: tea.KeyEsc}).(Model)
	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author after editing from final review", model.phase)
	}
	if got := model.workflow.Document().CurrentSectionIndex; got != 1 {
		t.Fatalf("current section = %d, want last section after final review edit", got)
	}
	if got := model.textarea.Value(); got != "second answer" {
		t.Fatalf("textarea value = %q, want saved last answer", got)
	}
	if len(fx.exporter.docs) != 0 {
		t.Fatalf("Export calls = %d, want 0 while returning to edit", len(fx.exporter.docs))
	}
	if !strings.Contains(model.View(), "Back to Two. Continue editing.") {
		t.Fatalf("edit-from-final-review missing status:\n%s", model.View())
	}
}

func TestModelFinishWithPendingSectionsCancelReturnsToEditing(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = submitAnswer(model, "first answer")
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlF}).(Model)
	if model.phase != phaseConfirm {
		t.Fatalf("phase = %s, want confirm", model.phase)
	}
	if !strings.Contains(model.View(), "Generate partial document?") {
		t.Fatalf("finish confirmation missing partial heading:\n%s", model.View())
	}
	model = press(model, tea.KeyMsg{Type: tea.KeyEsc}).(Model)
	if model.phase != phaseAuthor {
		t.Fatalf("phase = %s, want author after cancel", model.phase)
	}
	if len(fx.exporter.docs) != 0 {
		t.Fatalf("Export calls = %d, want 0 after cancel", len(fx.exporter.docs))
	}
	if got := model.workflow.Document().CurrentSectionIndex; got != 1 {
		t.Fatalf("current section = %d, want 1 after cancel", got)
	}
}

func TestModelFinishWithPendingSectionsConfirmExportsPartialDocument(t *testing.T) {
	fx := newModelFixture(t)
	model := press(fx.model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = submitAnswer(model, "first answer")
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlF}).(Model)
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	if model.phase != phaseDone {
		t.Fatalf("phase = %s, want done after confirmed partial generate", model.phase)
	}
	if len(fx.exporter.docs) != 1 {
		t.Fatalf("Export calls = %d, want 1 after confirmed partial generate", len(fx.exporter.docs))
	}
	if got := fx.exporter.docs[0].Sections[1].Status; got != document.StatusPending {
		t.Fatalf("exported section 2 status = %s, want pending for partial export", got)
	}
}

func TestModelHelpTogglePreservesDraft(t *testing.T) {
	model := startAuthoring(t)
	model = typeText(model, "draft help text")
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlG}).(Model)
	if got := model.textarea.Value(); got != "draft help text" {
		t.Fatalf("textarea value = %q, want draft preserved after help toggle", got)
	}
	if !strings.Contains(model.View(), "Help: help text") {
		t.Fatalf("help toggle did not render help:\n%s", model.View())
	}
}

func TestModelSectionGuidanceFallsBackToDescriptionWhenMetadataIsSparse(t *testing.T) {
	model := startAuthoring(t)
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlS}).(Model)
	model = press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlH}).(Model)

	view := model.View()
	if !strings.Contains(view, "Current section\nTwo\nsecond") {
		t.Fatalf("expected to be on second section:\n%s", view)
	}
	if !strings.Contains(view, "Description: second") {
		t.Fatalf("guidance should surface section description:\n%s", view)
	}
	if !strings.Contains(view, "No extra section metadata is available. Use the title and description as the main guidance.") {
		t.Fatalf("guidance should explain sparse metadata:\n%s", view)
	}
}

func newModelFixture(t *testing.T) modelFixture {
	t.Helper()
	reg, err := templates.NewRegistry([]templates.Template{{
		ID:          "product",
		Name:        "Product",
		Version:     1,
		Description: "desc",
		Sections:    []templates.Section{{ID: "one", Title: "One", Description: "first", Help: "help text", Example: "example", Questions: []string{"question one"}}, {ID: "two", Title: "Two", Description: "second"}},
	}})
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	store, exporter := &fakeStore{}, &fakeExporter{}
	model, err := NewModel(app.TUIRequest{
		Registry: reg,
		NewWorkflow: func(tmpl templates.Template) *app.GuidedWorkflow {
			clock := func() func() time.Time {
				now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
				return func() time.Time {
					now = now.Add(time.Minute)
					return now
				}
			}()
			return app.NewGuidedWorkflowWithOptions(tmpl, store, exporter, clock, app.GuidedWorkflowOptions{AutoFinalizeOnComplete: false})
		},
	})
	if err != nil {
		t.Fatalf("NewModel() error = %v", err)
	}
	return modelFixture{model: model, store: store, exporter: exporter}
}

func press(model Model, msg tea.KeyMsg) tea.Model {
	next, _ := model.Update(msg)
	return next
}

func typeText(model Model, text string) Model {
	for _, r := range text {
		model = press(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}).(Model)
	}
	return model
}

func startAuthoring(t *testing.T) Model {
	t.Helper()
	return press(newModelFixture(t).model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
}

func submitAnswer(model Model, answer string) Model {
	model = typeText(model, answer)
	model = press(model, tea.KeyMsg{Type: tea.KeyCtrlD}).(Model)
	return press(model, tea.KeyMsg{Type: tea.KeyEnter}).(Model)
}
