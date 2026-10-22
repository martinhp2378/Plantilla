// Package navigation projects explicit consultar grants for presentation only.
// Backend route authorization remains authoritative.
package navigation

import (
	"context"
	"strings"
)

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
