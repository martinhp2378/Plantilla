package shared

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"

	"PlantillaInstitucional/shared/tenantcontext"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
)

const (
	MaxReportRows = 10_000
	MaxReportSize = 50 << 20
)

var (
	ErrReportRowLimit  = errors.New("report exceeds the 10000 row limit")
	ErrReportSizeLimit = errors.New("report exceeds the 50 MiB file limit")
)

type limitedBuffer struct {
	bytes.Buffer
	max int
}

func newLimitedBuffer(max int) *limitedBuffer { return &limitedBuffer{max: max} }

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.max-b.Len() {
		return 0, ErrReportSizeLimit
	}
	return b.Buffer.Write(p)
}

func validateReportRows(rows int) error {
	if rows > MaxReportRows {
		return ErrReportRowLimit
	}
	return nil
}

func reportStorageKey(ctx context.Context, name, extension string) (string, error) {
	tenant, err := tenantcontext.TenantFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("report tenant unavailable: %w", err)
	}
	var clean strings.Builder
	for _, r := range strings.ToLower(name) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			clean.WriteRune(r)
		} else if clean.Len() > 0 && !strings.HasSuffix(clean.String(), "-") {
			clean.WriteByte('-')
		}
		if clean.Len() >= 64 {
			break
		}
	}
	base := strings.Trim(clean.String(), "-")
	if base == "" {
		base = "report"
	}
	id, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("generate report identifier: %w", err)
	}
	extension = "." + strings.TrimLeft(extension, ".")
	return path.Join("reportes", tenant.Schema, base+"-"+id.String()+extension), nil
}

func storePrivateReport(ctx context.Context, manager *archivos.ManejadorArchivos, data []byte, name, extension string) (string, error) {
	if len(data) > MaxReportSize {
		return "", ErrReportSizeLimit
	}
	key, err := reportStorageKey(ctx, name, extension)
	if err != nil {
		return "", err
	}
	if err := manager.ArchivoPrivado(key).Guardar(ctx, bytes.NewReader(data)); err != nil {
		return "", err
	}
	return "/private-reports/" + key, nil
}

// PrivateDownloadHandler proxies private objects through authenticated application routes.
func PrivateDownloadHandler(manager *archivos.ManejadorArchivos) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenant, err := tenantcontext.TenantFromContext(r.Context())
		key := chi.URLParam(r, "*")
		parts := strings.Split(key, "/")
		if err != nil || len(parts) != 3 || parts[0] != "reportes" || parts[1] != tenant.Schema || path.Clean(key) != key || strings.Contains(key, `\`) {
			http.NotFound(w, r)
			return
		}
		reader, err := manager.ArchivoPrivado(key).Get(r.Context())
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer reader.Close()
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, path.Base(key)))
		w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(key)))
		_, _ = io.Copy(w, reader)
	}
}
