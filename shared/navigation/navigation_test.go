package navigation

import (
	"context"
	"testing"
)

func TestTrailingSlashLandingPermissions(t *testing.T) {
	for _, path := range []string{"/dashboard", "/institucion/personas", "/configuracion/usuarios", "/crud"} {
		t.Run(path, func(t *testing.T) {
			for _, grant := range []string{path, path + "/"} {
				ctx := WithPaths(context.Background(), []string{grant})
				for _, landing := range []string{path, path + "/"} {
					if !Allowed(ctx, landing) {
						t.Errorf("grant %q hides landing %q", grant, landing)
					}
				}
				for _, denied := range []string{path + "/child", path + "-other", "/configuracion/politicas"} {
					if Allowed(ctx, denied) {
						t.Errorf("grant %q unexpectedly allows %q", grant, denied)
					}
				}
			}
		})
	}
}

func TestRootPermissionRemainsDistinctFromEmptyPath(t *testing.T) {
	ctx := WithPaths(context.Background(), []string{"/"})
	if !Allowed(ctx, "/") || Allowed(ctx, "") || Allowed(ctx, "/dashboard") {
		t.Fatal("root permission no longer remains literal")
	}
}

func TestImmutableRequestPermissions(t *testing.T) {
	paths := []string{"/dashboard"}
	first := WithPaths(context.Background(), paths)
	paths[0] = "/configuracion/usuarios"
	second := WithPaths(first, nil)
	if !Allowed(first, "/dashboard") || Allowed(first, paths[0]) || Allowed(second, "/dashboard") || Allowed(context.Background(), "/dashboard") {
		t.Fatal("permissions leaked or mutated")
	}
	if Allowed(first, "/dashboard/child") {
		t.Fatal("literal grant became prefix")
	}
}

func TestItemMenuHierarchyAndActiveState(t *testing.T) {
	menu := ItemMenu{
		Label:  "Administración",
		Codigo: "admin",
		Icon:   "admin_panel_settings",
		SubItems: []ItemMenu{
			{Label: "Usuarios", URL: "/configuracion/usuarios", Codigo: "usuarios"},
			{Label: "Perfiles", URL: "/configuracion/perfiles", Codigo: "perfiles"},
		},
	}

	if !menu.TieneSubItems() {
		t.Fatal("expected menu to have subitems")
	}

	simple := ItemMenu{Label: "Inicio", URL: "/dashboard", Codigo: "dashboard"}
	if simple.TieneSubItems() {
		t.Fatal("expected simple menu to not have subitems")
	}

	// Active check by code
	if !menu.TieneHijoActivo("usuarios", "") {
		t.Fatal("expected menu to have active child for code 'usuarios'")
	}
	// Active check by URL
	if !menu.TieneHijoActivo("", "/configuracion/perfiles/") {
		t.Fatal("expected menu to have active child for url '/configuracion/perfiles/'")
	}
	// Inactive check
	if menu.TieneHijoActivo("dashboard", "/dashboard") {
		t.Fatal("menu should not have active child for 'dashboard'")
	}
}

func TestItemMenuAccessControl(t *testing.T) {
	ctx := WithPaths(context.Background(), []string{"/configuracion/usuarios"})

	simpleAllowed := ItemMenu{URL: "/configuracion/usuarios"}
	simpleDenied := ItemMenu{URL: "/configuracion/politicas"}

	if !simpleAllowed.TieneAcceso(ctx, false) {
		t.Fatal("expected simpleAllowed to have access")
	}
	if simpleDenied.TieneAcceso(ctx, false) {
		t.Fatal("expected simpleDenied to not have access")
	}
	if !simpleDenied.TieneAcceso(ctx, true) {
		t.Fatal("expected superadmin to have access to simpleDenied")
	}

	dropdown := ItemMenu{
		Label:  "Configuración",
		Codigo: "config",
		SubItems: []ItemMenu{
			{Label: "Políticas", URL: "/configuracion/politicas"},
			{Label: "Usuarios", URL: "/configuracion/usuarios"},
		},
	}
	// Should have access because at least one child is allowed
	if !dropdown.TieneAcceso(ctx, false) {
		t.Fatal("expected dropdown to have access when one child is allowed")
	}

	dropdownDenied := ItemMenu{
		Label:  "Parámetros",
		Codigo: "params",
		SubItems: []ItemMenu{
			{Label: "Globales", URL: "/configuracion/parametros"},
		},
	}
	if dropdownDenied.TieneAcceso(ctx, false) {
		t.Fatal("expected dropdownDenied to be hidden when no children are allowed")
	}
	if !dropdownDenied.TieneAcceso(ctx, true) {
		t.Fatal("superadmin should have access to dropdownDenied")
	}
}

func TestRegistrarItemMenuAndSubItems(t *testing.T) {
	ResetItemsMenu()
	defer ResetItemsMenu()

	// Register simple item
	RegistrarItemMenu(ItemMenu{Label: "Dashboard", URL: "/dashboard", Codigo: "dash"})
	// Register duplicate
	RegistrarItemMenu(ItemMenu{Label: "Dashboard", URL: "/dashboard", Codigo: "dash"})

	items := ObtenerItemsMenu()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	// Register dropdown group with SubItems
	RegistrarItemMenu(ItemMenu{
		Label:  "Gestión",
		Codigo: "gestion",
		Icon:   "folder",
		SubItems: []ItemMenu{
			{Label: "Personas", URL: "/institucion/personas", Codigo: "personas"},
		},
	})

	items = ObtenerItemsMenu()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if len(items[1].SubItems) != 1 {
		t.Fatalf("expected 1 subitem in Gestión, got %d", len(items[1].SubItems))
	}

	// Add another subitem dynamically via RegistrarSubItemMenu
	RegistrarSubItemMenu("gestion", ItemMenu{
		Label:  "Contratos",
		URL:    "/institucion/contratos",
		Codigo: "contratos",
	})

	items = ObtenerItemsMenu()
	if len(items[1].SubItems) != 2 {
		t.Fatalf("expected 2 subitems in Gestión, got %d", len(items[1].SubItems))
	}

	// Adding duplicate subitem must be idempotent
	RegistrarSubItemMenu("gestion", ItemMenu{
		Label:  "Contratos",
		URL:    "/institucion/contratos",
		Codigo: "contratos",
	})
	items = ObtenerItemsMenu()
	if len(items[1].SubItems) != 2 {
		t.Fatalf("expected 2 subitems after duplicate register, got %d", len(items[1].SubItems))
	}

	// Mutating the returned slice should not mutate internal items
	items[1].SubItems[0].Label = "MUTATED"
	fresh := ObtenerItemsMenu()
	if fresh[1].SubItems[0].Label == "MUTATED" {
		t.Fatal("internal menu state was mutated via external reference")
	}
}
