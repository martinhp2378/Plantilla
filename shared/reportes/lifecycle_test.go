package shared

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"PlantillaInstitucional/shared/tenantcontext"

	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
)

type fakeRetentionRepo struct {
	mu          sync.Mutex
	items       []retainedReport
	completed   int
	retried     int
	removed     int
	activateErr error
	removeErr   error
	claimed     chan struct{}
}

func (r *fakeRetentionRepo) prepare(context.Context, string, string) (uuid.UUID, error) {
	return uuid.Must(uuid.NewV4()), nil
}
func (r *fakeRetentionRepo) activate(context.Context, string, uuid.UUID) error { return r.activateErr }
func (r *fakeRetentionRepo) remove(context.Context, string, uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removed++
	return r.removeErr
}
func (r *fakeRetentionRepo) tenants(context.Context) ([]string, error) {
	return []string{"tenant-id"}, nil
}
func (r *fakeRetentionRepo) claim(context.Context, string, int) ([]retainedReport, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := r.items
	r.items = nil
	if r.claimed != nil {
		select {
		case r.claimed <- struct{}{}:
		default:
		}
	}
	return items, nil
}
func (r *fakeRetentionRepo) complete(context.Context, retainedReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.completed++
	return nil
}
func (r *fakeRetentionRepo) retry(context.Context, retainedReport, error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.retried++
	return nil
}

type fakeObjects struct {
	existsValue bool
	existsErr   error
	deleteErr   error
	deleteCalls *int
}

func (o fakeObjects) exists(context.Context, string) (bool, error) { return o.existsValue, o.existsErr }
func (o fakeObjects) delete(context.Context, string) error {
	if o.deleteCalls != nil {
		*o.deleteCalls = *o.deleteCalls + 1
	}
	return o.deleteErr
}

func TestRetentionSweepHandlesSuccessMissingFailureAndCancellation(t *testing.T) {
	tests := []struct {
		name      string
		objects   fakeObjects
		cancel    bool
		completed int
		retried   int
	}{
		{name: "deletes existing object", objects: fakeObjects{existsValue: true}, completed: 1},
		{name: "removes metadata for missing object", completed: 1},
		{name: "preserves retry after transient failure", objects: fakeObjects{existsValue: true, deleteErr: errors.New("temporary")}, retried: 1},
		{name: "releases claim after cancellation", cancel: true, retried: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRetentionRepo{items: []retainedReport{{ID: uuid.Must(uuid.NewV4()), TenantID: "tenant-id", Path: "reportes/tenant/report.pdf"}}}
			lifecycle := newLifecycle(repo, tt.objects, time.Hour)
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancel {
				cancel()
			} else {
				defer cancel()
			}
			_ = lifecycle.Sweep(ctx)
			if repo.completed != tt.completed || repo.retried != tt.retried {
				t.Fatalf("completed=%d retried=%d, want %d/%d", repo.completed, repo.retried, tt.completed, tt.retried)
			}
		})
	}
}

func TestRetentionStartRunsImmediateSweepAndStops(t *testing.T) {
	repo := &fakeRetentionRepo{claimed: make(chan struct{}, 1)}
	lifecycle := newLifecycle(repo, fakeObjects{}, time.Hour)
	lifecycle.Start(context.Background())
	select {
	case <-repo.claimed:
	case <-time.After(time.Second):
		t.Fatal("startup sweep did not run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := lifecycle.Stop(ctx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestStoreCompensatesWriteFailure(t *testing.T) {
	root := t.TempDir()
	manager, err := archivos.NuevoManejador(&archivos.Config{Ambiente: archivos.AmbienteLocal, DirArchivosLocal: root})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "private"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A file where Guardar must create its report directory deterministically fails the write.
	if err := os.WriteFile(filepath.Join(root, "private", "reportes"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	deletionErr := errors.New("object deletion unavailable")
	removalErr := errors.New("metadata removal unavailable")
	deleteCalls := 0
	repo := &fakeRetentionRepo{removeErr: removalErr}
	lifecycle := newLifecycle(repo, fakeObjects{deleteErr: deletionErr, deleteCalls: &deleteCalls}, time.Hour)
	ctx := tenantcontext.WithTenant(context.Background(), tenantcontext.TenantContext{ID: "tenant-id", Schema: "tenant", CanonicalHost: "tenant.example"})

	_, err = StorePrivateReport(lifecycle.Context(ctx), manager, []byte("report"), "users", ".pdf")
	if !errors.Is(err, syscall.ENOTDIR) || !errors.Is(err, deletionErr) || !errors.Is(err, removalErr) {
		t.Fatalf("StorePrivateReport() error = %v, want write, deletion, and metadata removal errors", err)
	}
	if deleteCalls != 1 {
		t.Fatalf("object deletion attempts = %d, want 1", deleteCalls)
	}
	if repo.removed != 1 {
		t.Fatalf("removed metadata rows = %d, want 1", repo.removed)
	}
}

func TestStoreCompensatesActivationFailure(t *testing.T) {
	manager, err := archivos.NuevoManejador(&archivos.Config{Ambiente: archivos.AmbienteLocal, DirArchivosLocal: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	activationErr := errors.New("database unavailable")
	deletionErr := errors.New("object deletion unavailable")
	deleteCalls := 0
	repo := &fakeRetentionRepo{activateErr: activationErr}
	lifecycle := newLifecycle(repo, fakeObjects{deleteErr: deletionErr, deleteCalls: &deleteCalls}, time.Hour)
	ctx := tenantcontext.WithTenant(context.Background(), tenantcontext.TenantContext{ID: "tenant-id", Schema: "tenant", CanonicalHost: "tenant.example"})
	_, err = StorePrivateReport(lifecycle.Context(ctx), manager, []byte("report"), "users", ".pdf")
	if !errors.Is(err, activationErr) || !errors.Is(err, deletionErr) {
		t.Fatalf("storePrivateReport() error = %v, want activation and deletion errors", err)
	}
	if deleteCalls != 1 {
		t.Fatalf("object deletion attempts = %d, want 1", deleteCalls)
	}
	if repo.removed != 1 {
		t.Fatalf("removed metadata rows = %d, want 1", repo.removed)
	}
}
