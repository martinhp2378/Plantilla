package utils

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

// RenderIgnore executes a templ component and ignores the error.
// Use this when you intentionally don't need to handle render errors
// (e.g., in HTTP handlers where the response is already being sent).
func RenderIgnore(ctx context.Context, w io.Writer, c templ.Component) {
	_ = c.Render(ctx, w)
}
