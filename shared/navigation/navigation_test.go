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
