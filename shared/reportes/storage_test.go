package shared

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"PlantillaInstitucional/shared/tenantcontext"

	"github.com/issemym-repos/archivos"
)

func TestReportLimitsRejectRowsAndOutputSize(t *testing.T) {
	if err := validateReportRows(MaxReportRows + 1); !errors.Is(err, ErrReportRowLimit) {
		t.Fatalf("validateReportRows() error = %v, want ErrReportRowLimit", err)
	}
	buf := newLimitedBuffer(4)
	if _, err := buf.Write([]byte("12345")); !errors.Is(err, ErrReportSizeLimit) {
		t.Fatalf("limited buffer error = %v, want ErrReportSizeLimit", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("rejected write retained %d bytes", buf.Len())
	}
}

func TestReportStorageKeysAreTenantScopedAndUniqueInParallel(t *testing.T) {
	ctx := tenantcontext.WithTenant(context.Background(), tenantcontext.TenantContext{
		ID: "tenant-id", Schema: "tenant_one", CanonicalHost: "tenant-one.example",
	})
	const workers = 100
	keys := make(chan string, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key, err := reportStorageKey(ctx, "../../User Report!?", ".pdf")
			if err != nil {
				t.Errorf("reportStorageKey() error = %v", err)
				return
			}
			keys <- key
		}()
	}
	wg.Wait()
	close(keys)
	seen := make(map[string]bool, workers)
	for key := range keys {
		if !strings.HasPrefix(key, "reportes/tenant_one/user-report-") || !strings.HasSuffix(key, ".pdf") {
			t.Errorf("unsafe or unscoped key %q", key)
		}
		if seen[key] {
			t.Fatalf("duplicate parallel key %q", key)
		}
		seen[key] = true
	}
	if len(seen) != workers {
		t.Fatalf("generated %d keys, want %d", len(seen), workers)
	}
}

func TestStoreReportUsesPrivateStorageAndApplicationURL(t *testing.T) {
	manager, err := archivos.NuevoManejador(&archivos.Config{Ambiente: archivos.AmbienteLocal, DirArchivosLocal: t.TempDir()})
	if err != nil {
		t.Fatalf("NuevoManejador() error = %v", err)
	}
	ctx := tenantcontext.WithTenant(context.Background(), tenantcontext.TenantContext{ID: "tenant-id", Schema: "tenant_one", CanonicalHost: "tenant-one.example"})
	ctx = newLifecycle(&fakeRetentionRepo{}, managedObjects{manager}, reportRetention).Context(ctx)
	url, err := StorePrivateReport(ctx, manager, []byte("report"), "Users", ".pdf")
	if err != nil {
		t.Fatalf("storePrivateReport() error = %v", err)
	}
	key := strings.TrimPrefix(url, "/private-reports/")
	privateExists, _ := manager.ArchivoPrivado(key).Existe(ctx)
	publicExists, _ := manager.ArchivoPublico(key).Existe(ctx)
	if !privateExists || publicExists {
		t.Fatalf("stored report private=%t public=%t", privateExists, publicExists)
	}
}
