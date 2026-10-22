package service

import (
	"context"
	"reflect"
	"testing"
)

type fakeNavigationRepo struct {
	PolicyRepository
	paths []string
}

func (f *fakeNavigationRepo) NavigationPaths(ctx context.Context, profile string) ([]string, error) {
	return f.paths, nil
}

func TestNavigationPathsIntegration(t *testing.T) {
	ctx := context.Background()
	svc := NewService(&fakeNavigationRepo{paths: []string{"/dashboard", "/institucion/personas"}})

	got, err := svc.NavigationPaths(ctx, "superadmin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"/dashboard", "/institucion/personas"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NavigationPaths() = %v, want %v", got, want)
	}
}

func TestNavigationPathsLegacyRepoFailsClosed(t *testing.T) {
	ctx := context.Background()
	svc := NewService(nil)

	got, err := svc.NavigationPaths(ctx, "superadmin")
	if err != nil || got != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", got, err)
	}
}
