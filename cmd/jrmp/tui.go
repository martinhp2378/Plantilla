package main

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tuiScreen int

const (
	screenMenu tuiScreen = iota
	screenWizard
	screenRunning
	screenResult
)

type tuiActionID string

const (
	actionCreate       tuiActionID = "create"
	actionInit         tuiActionID = "init"
	actionNewModule    tuiActionID = "new-module"
	actionRemoveModule tuiActionID = "remove-module"
	actionUp           tuiActionID = "up"
	actionDown         tuiActionID = "down"
	actionCheck        tuiActionID = "check"
	actionSecurity     tuiActionID = "security"
	actionExit         tuiActionID = "exit"
)

type tuiMenuItem struct {
	id          tuiActionID
	title       string
	commandHint string
	description string
	needsWizard bool
}

type tuiFieldKind int

const (
	fieldText tuiFieldKind = iota
	fieldChoice
)

type tuiFormField struct {
	key         string
	label       string
	placeholder string
	value       string
	kind        tuiFieldKind
	options     []string
	optionIndex int
	required    bool
}

type tuiActionResultMsg struct {
	action        tuiActionID
	title         string
	summary       string
	updatedModule string
	err           error
}

type tuiRunnerFunc func(action tuiActionID, values map[string]string) (string, error)

type tuiModel struct {
	workspaceModule string
	screen          tuiScreen
	cursor          int
	items           []tuiMenuItem
	activeAction    tuiActionID
	wizardTitle     string
	fields          []tuiFormField
	fieldIndex      int
	validationErr   string
	resultTitle     string
	resultSummary   string
	resultErr       error
	quitting        bool
	runner          tuiRunnerFunc
}

var (
	headerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#7A1C32", Dark: "#D4AF37"}).
			Padding(0, 2).
			MarginBottom(1)

	headerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "#7A1C32", Dark: "#F4D06F"})

	headerSubStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#4B5563", Dark: "#9CA3AF"})

	badgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#111827"}).
			Background(lipgloss.AdaptiveColor{Light: "#7A1C32", Dark: "#D4AF37"}).
			Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "#7A1C32", Dark: "#F4D06F"})

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#E5E7EB"})

	cmdHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"})

	descBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#7A1C32", Dark: "#D4AF37"}).
			PaddingLeft(1).
			MarginTop(1).
			Foreground(lipgloss.AdaptiveColor{Light: "#374151", Dark: "#D1D5DB"})

	activeInputBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#7A1C32", Dark: "#F4D06F"}).
				Padding(0, 1)

	inactiveInputBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"}).
				Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"})

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#34D399"})

	footerHelpStyle = lipgloss.NewStyle().
			MarginTop(1).
			Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"})
)

func defaultMenuItems() []tuiMenuItem {
	return []tuiMenuItem{
		{
			id:          actionCreate,
			title:       "🌱 Crear proyecto desde cero",
			commandHint: "jrmp create",
			description: "Clona la Plantilla Institucional, desvincula Git, renombra el módulo Go y genera el .env.",
			needsWizard: true,
		},
		{
			id:          actionInit,
			title:       "🏷️  Inicializar proyecto clonado",
			commandHint: "jrmp init",
			description: "Renombra el módulo Go en un repositorio ya clonado, limpia origin y configura el .env local.",
			needsWizard: true,
		},
		{
			id:          actionNewModule,
			title:       "⚡ Crear nuevo módulo CRUD",
			commandHint: "jrmp new-module",
			description: "Genera handler, service, repository, vistas Templ, exportador PDF/Excel, tests y migraciones SQL.",
			needsWizard: true,
		},
		{
			id:          actionRemoveModule,
			title:       "🧹 Eliminar módulo existente",
			commandHint: "jrmp remove-module",
			description: "Elimina carpeta del módulo, migraciones, registro en cmd/main.go y revierte su tabla en PostgreSQL.",
			needsWizard: true,
		},
		{
			id:          actionUp,
			title:       "🐳 Levantar entorno Docker",
			commandHint: "jrmp up",
			description: "Construye y enciende el servidor y PostgreSQL con recarga en vivo (docker-compose.local.yml).",
			needsWizard: false,
		},
		{
			id:          actionDown,
			title:       "🛑 Detener entorno Docker",
			commandHint: "jrmp down",
			description: "Detiene los contenedores locales del proyecto.",
			needsWizard: false,
		},
		{
			id:          actionCheck,
			title:       "✅ Ejecutar validación completa",
			commandHint: "jrmp check",
			description: "Verifica formato .templ, templ generate, golangci-lint y pruebas unitarias.",
			needsWizard: false,
		},
		{
			id:          actionSecurity,
			title:       "🛡️  Auditar seguridad (govulncheck & Trivy)",
			commandHint: "jrmp security",
			description: "Escanea vulnerabilidades en código Go (govulncheck) y contenedores/dependencias (Trivy).",
			needsWizard: false,
		},
		{
			id:          actionExit,
			title:       "👋 Salir",
			commandHint: "q / esc",
			description: "Cierra el asistente interactivo y regresa a la terminal.",
			needsWizard: false,
		},
	}
}

func newTUIModel(workspaceModule string) tuiModel {
	if strings.TrimSpace(workspaceModule) == "" {
		workspaceModule = "Fuera de proyecto (modo creación)"
	}
	return tuiModel{
		workspaceModule: workspaceModule,
		screen:          screenMenu,
		cursor:          0,
		items:           defaultMenuItems(),
		runner:          defaultTUIRunner,
	}
}

func defaultTUIRunner(action tuiActionID, values map[string]string) (string, error) {
	switch action {
	case actionCreate:
		projectName := strings.TrimSpace(values["project_name"])
		serviceName := strings.TrimSpace(values["service_name"])
		dbPass := strings.TrimSpace(values["db_password"])
		remoteURL := strings.TrimSpace(values["remote_url"])
		startDocker := values["start_docker"] == "sí"
		if err := createProjectCore(defaultTemplateRepo, projectName, serviceName, dbPass, remoteURL, startDocker); err != nil {
			return "", err
		}
		_ = os.Chdir(projectName)
		return fmt.Sprintf("✨ Proyecto '%s' creado y activado en ./%s (ya podés crearle módulos desde este menú)", projectName, projectName), nil

	case actionInit:
		newMod := strings.TrimSpace(values["module_name"])
		svc := strings.TrimSpace(values["service_name"])
		dbPass := strings.TrimSpace(values["db_password"])
		if err := initProjectCore(newMod, svc, dbPass); err != nil {
			return "", err
		}
		return fmt.Sprintf("✨ Proyecto inicializado y renombrado a '%s'", newMod), nil

	case actionNewModule:
		return scaffoldModuleCore(
			values["name"],
			values["schema"],
			values["title"],
			"",
			false,
			false,
		)

	case actionRemoveModule:
		return scaffoldModuleCore(
			values["name"],
			"sistema",
			"",
			"",
			false,
			true,
		)

	case actionUp:
		if err := requireProjectRoot(); err != nil {
			return "", err
		}
		if err := runCmdInDir(".", "docker", "compose", "-f", "docker-compose.local.yml", "up", "--build", "-d"); err != nil {
			return "", err
		}
		return "🐳 Contenedores Docker iniciados en http://localhost:8080/login", nil

	case actionDown:
		if err := requireProjectRoot(); err != nil {
			return "", err
		}
		if err := runCmdInDir(".", "docker", "compose", "-f", "docker-compose.local.yml", "down"); err != nil {
			return "", err
		}
		return "🛑 Contenedores Docker detenidos exitosamente.", nil

	case actionCheck:
		if err := checkProjectCore(); err != nil {
			return "", err
		}
		return "✅ Validación completa exitosa (templ fmt + templ generate + golangci-lint + go test).", nil

	case actionSecurity:
		if err := requireProjectRoot(); err != nil {
			return "", err
		}
		if err := runSecurityCore(); err != nil {
			return "", err
		}
		return "🛡️  Auditoría de seguridad completada con éxito (govulncheck + Trivy).", nil

	default:
		return "", fmt.Errorf("acción no soportada: %s", action)
	}
}

func fieldsForAction(action tuiActionID) []tuiFormField {
	switch action {
	case actionCreate:
		return []tuiFormField{
			{
				key:         "project_name",
				label:       "📝 Nombre del nuevo proyecto (sin espacios)",
				placeholder: "SistemaInventario",
				kind:        fieldText,
				required:    true,
			},
			{
				key:         "service_name",
				label:       "🗄️  Nombre corto para servicio y base de datos (opcional)",
				placeholder: "Autogenerado desde el nombre del proyecto",
				kind:        fieldText,
				required:    false,
			},
			{
				key:         "db_password",
				label:       "🔑 Contraseña PostgreSQL local",
				placeholder: "postgres",
				value:       "postgres",
				kind:        fieldText,
				required:    false,
			},
			{
				key:         "remote_url",
				label:       "🔗 URL de tu repositorio Git (opcional)",
				placeholder: "https://github.com/org/repo.git",
				kind:        fieldText,
				required:    false,
			},
			{
				key:         "start_docker",
				label:       "🐳 ¿Construir y encender Docker al terminar?",
				kind:        fieldChoice,
				options:     []string{"no", "sí"},
				optionIndex: 0,
			},
		}
	case actionInit:
		return []tuiFormField{
			{
				key:         "module_name",
				label:       "📝 Nuevo nombre del proyecto / módulo Go",
				placeholder: "SistemaInventario",
				kind:        fieldText,
				required:    true,
			},
			{
				key:         "service_name",
				label:       "🗄️  Nombre corto para servicio y base de datos (opcional)",
				placeholder: "Autogenerado en snake_case",
				kind:        fieldText,
				required:    false,
			},
			{
				key:         "db_password",
				label:       "🔑 Contraseña PostgreSQL local",
				placeholder: "postgres",
				value:       "postgres",
				kind:        fieldText,
				required:    false,
			},
		}
	case actionNewModule:
		return []tuiFormField{
			{
				key:         "name",
				label:       "📦 Nombre del módulo en singular",
				placeholder: "proveedor, orden-compra, articulo...",
				kind:        fieldText,
				required:    true,
			},
			{
				key:         "schema",
				label:       "🗂️  Esquema de PostgreSQL",
				kind:        fieldChoice,
				options:     []string{"sistema", "institucion"},
				optionIndex: 0,
			},
			{
				key:         "title",
				label:       "🏷️  Título visible en menú y pantalla (opcional)",
				placeholder: "Enter para autogenerar desde el plural",
				kind:        fieldText,
				required:    false,
			},
		}
	case actionRemoveModule:
		return []tuiFormField{
			{
				key:         "name",
				label:       "🧹 Nombre del módulo a eliminar",
				placeholder: "proveedor",
				kind:        fieldText,
				required:    true,
			},
		}
	default:
		return nil
	}
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case screenMenu:
			return m.updateMenu(typed)
		case screenWizard:
			return m.updateWizard(typed)
		case screenResult:
			return m.updateResult(typed)
		}
	case tuiActionResultMsg:
		m.screen = screenResult
		m.resultTitle = typed.title
		m.resultSummary = typed.summary
		m.resultErr = typed.err
		if typed.err == nil && typed.updatedModule != "" {
			m.workspaceModule = typed.updatedModule
		}
		return m, nil
	}
	return m, nil
}

func (m tuiModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case "enter":
		selected := m.items[m.cursor]
		if selected.id == actionExit {
			m.quitting = true
			return m, tea.Quit
		}
		m.activeAction = selected.id
		if selected.needsWizard {
			m.screen = screenWizard
			m.wizardTitle = wizardTitleForAction(selected.id)
			m.fields = fieldsForAction(selected.id)
			m.fieldIndex = 0
			m.validationErr = ""
			return m, nil
		}
		m.screen = screenRunning
		return m, m.dispatchAction(selected.id, map[string]string{})
	}
	return m, nil
}

func (m tuiModel) updateWizard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.fields) == 0 {
		m.screen = screenMenu
		return m, nil
	}

	current := &m.fields[m.fieldIndex]

	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		m.screen = screenMenu
		m.validationErr = ""
		return m, nil
	case "up", "shift+tab":
		if m.fieldIndex > 0 {
			m.fieldIndex--
			m.validationErr = ""
		}
		return m, nil
	case "down", "tab":
		if m.fieldIndex < len(m.fields)-1 {
			m.fieldIndex++
			m.validationErr = ""
		}
		return m, nil
	case "left":
		if current.kind == fieldChoice && len(current.options) > 0 {
			current.optionIndex = (current.optionIndex - 1 + len(current.options)) % len(current.options)
		}
		return m, nil
	case "right":
		if current.kind == fieldChoice && len(current.options) > 0 {
			current.optionIndex = (current.optionIndex + 1) % len(current.options)
		}
		return m, nil
	case " ":
		if current.kind == fieldChoice && len(current.options) > 0 {
			current.optionIndex = (current.optionIndex + 1) % len(current.options)
			return m, nil
		}
		current.value += " "
		return m, nil
	case "backspace":
		if current.kind == fieldText && len(current.value) > 0 {
			_, size := utf8.DecodeLastRuneInString(current.value)
			current.value = current.value[:len(current.value)-size]
		}
		return m, nil
	case "enter":
		if current.required && strings.TrimSpace(current.value) == "" {
			m.validationErr = fmt.Sprintf("El campo %q es obligatorio.", current.label)
			return m, nil
		}
		if m.fieldIndex < len(m.fields)-1 {
			m.fieldIndex++
			m.validationErr = ""
			return m, nil
		}
		return m.submitWizard()
	}

	if msg.Type == tea.KeyRunes && current.kind == fieldText {
		current.value += string(msg.Runes)
		m.validationErr = ""
	}

	return m, nil
}

func (m tuiModel) submitWizard() (tea.Model, tea.Cmd) {
	values := make(map[string]string, len(m.fields))
	for i, f := range m.fields {
		if f.required && strings.TrimSpace(f.value) == "" {
			m.fieldIndex = i
			m.validationErr = fmt.Sprintf("El campo %q es obligatorio.", f.label)
			return m, nil
		}
		if f.kind == fieldChoice && len(f.options) > 0 {
			values[f.key] = f.options[f.optionIndex]
		} else {
			values[f.key] = strings.TrimSpace(f.value)
		}
	}

	m.screen = screenRunning
	m.validationErr = ""
	return m, m.dispatchAction(m.activeAction, values)
}

func (m tuiModel) dispatchAction(action tuiActionID, values map[string]string) tea.Cmd {
	runner := m.runner
	if runner == nil {
		runner = defaultTUIRunner
	}
	title := wizardTitleForAction(action)
	return func() tea.Msg {
		summary, err := runner(action, values)
		updatedMod := ""
		if err == nil {
			switch action {
			case actionCreate:
				updatedMod = strings.TrimSpace(values["project_name"])
			case actionInit:
				updatedMod = strings.TrimSpace(values["module_name"])
			}
		}
		return tuiActionResultMsg{
			action:        action,
			title:         title,
			summary:       summary,
			updatedModule: updatedMod,
			err:           err,
		}
	}
}

func (m tuiModel) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	case "enter", "esc":
		m.screen = screenMenu
		return m, nil
	}
	return m, nil
}

func wizardTitleForAction(action tuiActionID) string {
	switch action {
	case actionCreate:
		return "🌱 Crear Nuevo Proyecto desde Cero"
	case actionInit:
		return "🏷️  Inicializar Proyecto Clonado"
	case actionNewModule:
		return "⚡ Nuevo Módulo CRUD Institucional"
	case actionRemoveModule:
		return "🧹 Eliminar Módulo Existente"
	case actionUp:
		return "🐳 Levantar Entorno Docker"
	case actionDown:
		return "🛑 Detener Entorno Docker"
	case actionCheck:
		return "✅ Validación de Calidad Institucional"
	case actionSecurity:
		return "🛡️  Auditoría de Seguridad (govulncheck & Trivy)"
	default:
		return string(action)
	}
}

func (m tuiModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	header := fmt.Sprintf(
		"%s  %s\n%s\n%s %s",
		headerTitleStyle.Render("🏛️  JRMP CLI"),
		badgeStyle.Render(jrmpVersion),
		headerSubStyle.Render("Asistente Interactivo de Arquitectura & Scaffolding Institucional"),
		headerSubStyle.Render("Proyecto activo:"),
		selectedItemStyle.Render(m.workspaceModule),
	)
	b.WriteString(headerBoxStyle.Render(header))
	b.WriteString("\n")

	switch m.screen {
	case screenMenu:
		b.WriteString(m.viewMenu())
	case screenWizard:
		b.WriteString(m.viewWizard())
	case screenRunning:
		b.WriteString("⏳ Ejecutando operación...\n")
	case screenResult:
		b.WriteString(m.viewResult())
	}

	return b.String()
}

func (m tuiModel) viewMenu() string {
	var b strings.Builder
	for i, item := range m.items {
		cursor := "  "
		title := normalItemStyle.Render(item.title)
		if i == m.cursor {
			cursor = selectedItemStyle.Render("▸ ")
			title = selectedItemStyle.Render(item.title)
		}
		hint := cmdHintStyle.Render(fmt.Sprintf("(%s)", item.commandHint))
		_, _ = fmt.Fprintf(&b, "%s%s %s\n", cursor, title, hint)
	}

	if m.cursor >= 0 && m.cursor < len(m.items) {
		sel := m.items[m.cursor]
		b.WriteString(descBoxStyle.Render(sel.description))
		b.WriteString("\n")
	}

	b.WriteString(footerHelpStyle.Render("↑/k subir • ↓/j bajar • enter abrir • q/esc salir"))
	b.WriteString("\n")
	return b.String()
}

func (m tuiModel) viewWizard() string {
	var b strings.Builder
	b.WriteString(selectedItemStyle.Render(m.wizardTitle))
	b.WriteString("\n\n")

	for i, f := range m.fields {
		active := i == m.fieldIndex
		labelStyle := normalItemStyle
		boxStyle := inactiveInputBoxStyle
		if active {
			labelStyle = selectedItemStyle
			boxStyle = activeInputBoxStyle
		}

		_, _ = fmt.Fprintf(&b, "%s\n", labelStyle.Render(f.label))

		var content string
		if f.kind == fieldChoice {
			var opts []string
			for idx, opt := range f.options {
				if idx == f.optionIndex {
					opts = append(opts, badgeStyle.Render("● "+opt))
				} else {
					opts = append(opts, cmdHintStyle.Render("○ "+opt))
				}
			}
			content = strings.Join(opts, "   ")
		} else {
			val := f.value
			if val == "" {
				content = cmdHintStyle.Render(f.placeholder)
			} else {
				content = normalItemStyle.Render(val)
			}
			if active {
				content += selectedItemStyle.Render("█")
			}
		}

		b.WriteString(boxStyle.Render(content))
		b.WriteString("\n\n")
	}

	if m.validationErr != "" {
		b.WriteString(errorStyle.Render("⚠️  " + m.validationErr))
		b.WriteString("\n\n")
	}

	b.WriteString(footerHelpStyle.Render("tab/↓ siguiente • shift+tab/↑ anterior • ←/→ cambiar opción • enter continuar/ejecutar • esc cancelar"))
	b.WriteString("\n")
	return b.String()
}

func (m tuiModel) viewResult() string {
	var b strings.Builder
	b.WriteString(selectedItemStyle.Render(m.resultTitle))
	b.WriteString("\n\n")
	if m.resultErr != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("❌ Error: %v", m.resultErr)))
		b.WriteString("\n")
	} else {
		b.WriteString(successStyle.Render(m.resultSummary))
		b.WriteString("\n")
	}
	b.WriteString(footerHelpStyle.Render("enter/esc volver al menú principal • q salir"))
	b.WriteString("\n")
	return b.String()
}

func shouldLaunchTUI(args []string, isTTY bool) bool {
	if len(args) == 0 {
		return isTTY
	}
	sub := strings.ToLower(strings.TrimSpace(args[0]))
	return sub == "tui" || sub == "ui"
}

func isTerminalFile(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func detectWorkspaceModule() string {
	ensureProjectRoot()
	if _, err := os.Stat("go.mod"); err != nil {
		return "Fuera de proyecto (modo creación)"
	}
	return getModuleName("go.mod")
}

func runTUI() {
	modName := detectWorkspaceModule()
	p := tea.NewProgram(newTUIModel(modName), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("❌ Error en interfaz interactiva: %v\n", err)
		os.Exit(1)
	}
}

