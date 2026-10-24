// Package navigation projects explicit consultar grants for presentation only.
// Backend route authorization remains authoritative.
package navigation

import (
	"context"
	"strings"
	"sync"
)

// ItemMenu representa un elemento de menú dinámico para la navegación lateral.
// Puede ser un ítem simple con enlace directo o un menú desplegable con SubItems.
type ItemMenu struct {
	Label    string
	URL      string
	Icon     string
	Codigo   string
	SubItems []ItemMenu
}

// TieneSubItems indica si el ítem contiene elementos subordinados (desplegable).
func (m ItemMenu) TieneSubItems() bool {
	return len(m.SubItems) > 0
}

// EsActivo comprueba si el ítem coincide exactamente con el código o la ruta activa.
func (m ItemMenu) EsActivo(codigo, url string) bool {
	if m.Codigo != "" && codigo != "" && m.Codigo == codigo {
		return true
	}
	if m.URL != "" && m.URL != "#" && url != "" && (m.URL == url || landingPath(m.URL) == landingPath(url)) {
		return true
	}
	return false
}

// TieneHijoActivo comprueba recursivamente si alguno de los subítems está activo.
func (m ItemMenu) TieneHijoActivo(codigo, url string) bool {
	for _, sub := range m.SubItems {
		if sub.EsActivo(codigo, url) || sub.TieneHijoActivo(codigo, url) {
			return true
		}
	}
	return false
}

// TieneAcceso evalúa si el usuario tiene permiso para ver este ítem o alguno de sus subítems.
func (m ItemMenu) TieneAcceso(ctx context.Context, esSuperAdmin bool) bool {
	if esSuperAdmin {
		return true
	}
	if !m.TieneSubItems() {
		if m.URL == "" || m.URL == "#" {
			return false
		}
		return Allowed(ctx, m.URL)
	}
	for _, sub := range m.SubItems {
		if sub.TieneAcceso(ctx, esSuperAdmin) {
			return true
		}
	}
	return false
}

var (
	itemsMenuMu sync.RWMutex
	itemsMenu   []ItemMenu
)

// itemCoincide verifica si dos definiciones representan el mismo nodo de menú.
func itemCoincide(a, b ItemMenu) bool {
	if a.Codigo != "" && b.Codigo != "" && a.Codigo == b.Codigo {
		return true
	}
	if a.URL != "" && a.URL != "#" && b.URL != "" && b.URL != "#" && a.URL == b.URL {
		return true
	}
	if (a.URL == "" || a.URL == "#") && (b.URL == "" || b.URL == "#") && a.Codigo == "" && b.Codigo == "" {
		return a.Label == b.Label
	}
	return false
}

func agregarSubItemIdempotente(padre *ItemMenu, nuevoSub ItemMenu) {
	for i, sub := range padre.SubItems {
		if itemCoincide(sub, nuevoSub) {
			if len(nuevoSub.SubItems) > 0 {
				for _, subSub := range nuevoSub.SubItems {
					agregarSubItemIdempotente(&padre.SubItems[i], subSub)
				}
			}
			return
		}
	}
	padre.SubItems = append(padre.SubItems, nuevoSub)
}

// RegistrarItemMenu agrega un módulo o grupo de forma idempotente a la navegación lateral.
// Si ya existe y se proporcionan SubItems, los fusiona de manera idempotente.
func RegistrarItemMenu(item ItemMenu) {
	itemsMenuMu.Lock()
	defer itemsMenuMu.Unlock()
	for i, existente := range itemsMenu {
		if itemCoincide(existente, item) {
			if len(item.SubItems) > 0 {
				for _, nuevoSub := range item.SubItems {
					agregarSubItemIdempotente(&itemsMenu[i], nuevoSub)
				}
			}
			return
		}
	}
	itemsMenu = append(itemsMenu, item)
}

// RegistrarSubItemMenu agrega un subítem a un menú padre identificado por su código o label.
// Si el menú padre no existe aún, lo crea automáticamente utilizando el defaultPadre si se proporciona.
func RegistrarSubItemMenu(codigoPadre string, subItem ItemMenu, defaultPadre ...ItemMenu) {
	itemsMenuMu.Lock()
	defer itemsMenuMu.Unlock()

	for i := range itemsMenu {
		if itemsMenu[i].Codigo == codigoPadre || itemsMenu[i].Label == codigoPadre {
			agregarSubItemIdempotente(&itemsMenu[i], subItem)
			return
		}
	}

	var padre ItemMenu
	if len(defaultPadre) > 0 {
		padre = defaultPadre[0]
	} else {
		padre = ItemMenu{
			Label:  codigoPadre,
			Codigo: codigoPadre,
			Icon:   "folder",
		}
	}
	padre.SubItems = append(padre.SubItems, subItem)
	itemsMenu = append(itemsMenu, padre)
}

// ObtenerItemsMenu retorna una copia profunda de todos los elementos registrados dinámicamente.
func ObtenerItemsMenu() []ItemMenu {
	itemsMenuMu.RLock()
	defer itemsMenuMu.RUnlock()
	copia := make([]ItemMenu, len(itemsMenu))
	for i, item := range itemsMenu {
		copia[i] = clonarItem(item)
	}
	return copia
}

func clonarItem(item ItemMenu) ItemMenu {
	c := item
	if len(item.SubItems) > 0 {
		c.SubItems = make([]ItemMenu, len(item.SubItems))
		for j, sub := range item.SubItems {
			c.SubItems[j] = clonarItem(sub)
		}
	}
	return c
}

// ResetItemsMenu reinicia el estado de navegación (útil para tests).
func ResetItemsMenu() {
	itemsMenuMu.Lock()
	defer itemsMenuMu.Unlock()
	itemsMenu = nil
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
