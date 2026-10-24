package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIMenu_NavigationAndSelection(t *testing.T) {
	tests := []struct {
		name           string
		keys           []tea.KeyMsg
		wantCursor     int
		wantScreen     tuiScreen
		wantAction     tuiActionID
		wantQuitting   bool
		wantViewSubstr string
	}{
		{
			name:           "initial state renders header and first item selected",
			keys:           nil,
			wantCursor:     0,
			wantScreen:     screenMenu,
			wantQuitting:   false,
			wantViewSubstr: "JRMP CLI",
		},
		{
			name: "move down with j and down arrow",
			keys: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune{'j'}},
				{Type: tea.KeyDown},
			},
			wantCursor:     2,
			wantScreen:     screenMenu,
			wantQuitting:   false,
			wantViewSubstr: "new-module",
		},
		{
			name: "upper boundary clamping on k and up arrow at index 0",
			keys: []tea.KeyMsg{
				{Type: tea.KeyUp},
				{Type: tea.KeyRunes, Runes: []rune{'k'}},
			},
			wantCursor:   0,
			wantScreen:   screenMenu,
			wantQuitting: false,
		},
		{
			name: "select new-module opens wizard screen",
			keys: []tea.KeyMsg{
				{Type: tea.KeyDown},
				{Type: tea.KeyDown},
				{Type: tea.KeyEnter},
			},
			wantCursor:     2,
			wantScreen:     screenWizard,
			wantAction:     actionNewModule,
			wantQuitting:   false,
			wantViewSubstr: "Nuevo Módulo CRUD",
		},
		{
			name: "quit from main menu with q",
			keys: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune{'q'}},
			},
			wantCursor:   0,
			wantScreen:   screenMenu,
			wantQuitting: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m tea.Model = newTUIModel("PlantillaInstitucional")
			for _, k := range tt.keys {
				m, _ = m.Update(k)
			}

			got := m.(tuiModel)
			if got.cursor != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", got.cursor, tt.wantCursor)
			}
			if got.screen != tt.wantScreen {
				t.Errorf("screen = %v, want %v", got.screen, tt.wantScreen)
			}
			if tt.wantAction != "" && got.activeAction != tt.wantAction {
				t.Errorf("activeAction = %q, want %q", got.activeAction, tt.wantAction)
			}
			if got.quitting != tt.wantQuitting {
				t.Errorf("quitting = %v, want %v", got.quitting, tt.wantQuitting)
			}
			if tt.wantViewSubstr != "" {
				view := got.View()
				if !strings.Contains(view, tt.wantViewSubstr) {
					t.Errorf("View() missing substring %q\nGot:\n%s", tt.wantViewSubstr, view)
				}
			}
		})
	}
}

func TestTUIWizard_FormInputAndExecution(t *testing.T) {
	var capturedAction tuiActionID
	var capturedValues map[string]string

	m := newTUIModel("PlantillaInstitucional")
	m.runner = func(action tuiActionID, values map[string]string) (string, error) {
		capturedAction = action
		capturedValues = values
		return "Módulo Pago generado en /institucion/pagos", nil
	}

	// 1. Navigate to new-module (index 2) and press Enter
	var model tea.Model = m
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	got := model.(tuiModel)
	if got.screen != screenWizard {
		t.Fatalf("expected screenWizard, got %v", got.screen)
	}
	if len(got.fields) != 3 {
		t.Fatalf("expected 3 fields for new-module wizard, got %d", len(got.fields))
	}

	// 2. Pressing Enter on empty required field ("name") shows validation error
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got = model.(tuiModel)
	if got.validationErr == "" {
		t.Fatalf("expected validation error when submitting empty required field")
	}

	// 3. Type "pago" into field 0 ("name"), press Tab to move to field 1 ("schema")
	for _, r := range "pago" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	// 4. Toggle choice field ("schema") from "sistema" to "institucion" with Right arrow
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRight})

	// 5. Move to field 2 ("title") and type "Pagos"
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "Pagos" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// 6. Press Enter on last field to execute action
	var cmd tea.Cmd
	model, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected tea.Cmd on wizard submission, got nil")
	}
	msg := cmd()
	model, _ = model.Update(msg)

	got = model.(tuiModel)
	if got.screen != screenResult {
		t.Fatalf("expected screenResult after execution, got %v", got.screen)
	}
	if capturedAction != actionNewModule {
		t.Errorf("capturedAction = %q, want %q", capturedAction, actionNewModule)
	}
	if capturedValues["name"] != "pago" {
		t.Errorf("values[name] = %q, want %q", capturedValues["name"], "pago")
	}
	if capturedValues["schema"] != "institucion" {
		t.Errorf("values[schema] = %q, want %q", capturedValues["schema"], "institucion")
	}
	if capturedValues["title"] != "Pagos" {
		t.Errorf("values[title] = %q, want %q", capturedValues["title"], "Pagos")
	}
	if !strings.Contains(got.View(), "Módulo Pago generado") {
		t.Errorf("result View() missing summary, got:\n%s", got.View())
	}

	// 7. Press Enter on result screen to return to main menu
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got = model.(tuiModel)
	if got.screen != screenMenu {
		t.Fatalf("expected screenMenu after pressing Enter on result, got %v", got.screen)
	}
}

func TestTUIWizard_BackspaceCancelAndInitWorkspaceUpdate(t *testing.T) {
	m := newTUIModel("PlantillaInstitucional")
	m.runner = func(action tuiActionID, values map[string]string) (string, error) {
		return "Inicializado como " + values["module_name"], nil
	}

	// Select init (index 1)
	var model tea.Model = m
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Type "MiProyectox", then Backspace, then Enter through the 3 fields
	for _, r := range "MiProyectox" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // advances to service_name
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // advances to db_password

	var cmd tea.Cmd
	model, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // submits wizard
	if cmd == nil {
		t.Fatalf("expected submit command for init wizard")
	}
	model, _ = model.Update(cmd())

	got := model.(tuiModel)
	if got.workspaceModule != "MiProyecto" {
		t.Errorf("workspaceModule = %q, want %q", got.workspaceModule, "MiProyecto")
	}
}

func TestShouldLaunchTUI(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		isTTY bool
		want  bool
	}{
		{name: "no args on interactive terminal launches TUI", args: nil, isTTY: true, want: true},
		{name: "no args on non-interactive pipe falls back to help", args: nil, isTTY: false, want: false},
		{name: "explicit tui subcommand launches TUI", args: []string{"tui"}, isTTY: false, want: true},
		{name: "explicit ui subcommand launches TUI", args: []string{"ui"}, isTTY: false, want: true},
		{name: "direct subcommand new-module does not launch TUI", args: []string{"new-module", "pago"}, isTTY: true, want: false},
		{name: "help flag does not launch TUI", args: []string{"--help"}, isTTY: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldLaunchTUI(tt.args, tt.isTTY); got != tt.want {
				t.Errorf("shouldLaunchTUI(%v, %v) = %v, want %v", tt.args, tt.isTTY, got, tt.want)
			}
		})
	}
}



