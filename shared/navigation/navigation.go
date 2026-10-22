// Package navigation projects explicit consultar grants for presentation only.
// Backend route authorization remains authoritative.
package navigation

import (
	"context"
	"strings"
	"sync"
)

// ItemMenu representa un elemento de menú dinámico para la navegación lateral.
type ItemMenu struct {
	Label  string
	URL    string
	Icon   string
	Codigo string
}

var (
	itemsMenuMu sync.RWMutex
	itemsMenu   []ItemMenu
)

// RegistrarItemMenu agrega un módulo de forma idempotente a la lista de navegación lateral.
func RegistrarItemMenu(item ItemMenu) {
	itemsMenuMu.Lock()
	defer itemsMenuMu.Unlock()
	for _, existente := range itemsMenu {
		if existente.URL == item.URL {
			return
		}
	}
	itemsMenu = append(itemsMenu, item)
}

// ObtenerItemsMenu retorna una copia de todos los elementos registrados dinámicamente.
func ObtenerItemsMenu() []ItemMenu {
	itemsMenuMu.RLock()
	defer itemsMenuMu.RUnlock()
	copia := make([]ItemMenu, len(itemsMenu))
	copy(copia, itemsMenu)
	return copia
}

type contextKey struct{}

// Reader returns literal GET landing paths granted by consultar policies.
// Legacy prefix-only and pattern-only policies do not create menu grants.
type Reader interface {
	NavigationPaths(context.Context, string) ([]string, error)
}

// WithPaths copies paths into an immutable, request-local permission set.
// Landing paths may differ by one trailing slash, but never match by prefix.
func WithPaths(ctx context.Context, paths []string) context.Context {
	allowed := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		allowed[landingPath(path)] = struct{}{}
	}
	return context.WithValue(ctx, contextKey{}, allowed)
}

// Allowed checks a literal landing path, ignoring one trailing slash, never a prefix.
func Allowed(ctx context.Context, path string) bool {
	allowed, _ := ctx.Value(contextKey{}).(map[string]struct{})
	_, ok := allowed[landingPath(path)]
	return ok
}

func landingPath(path string) string {
	if path == "/" {
		return path
	}
	return strings.TrimSuffix(path, "/")
}
