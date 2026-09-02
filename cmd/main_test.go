package main

import "testing"

func TestIsMigrationCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "starts the HTTP server without a command", args: nil, want: false},
		{name: "runs migrations with the dedicated command", args: []string{"migrate"}, want: true},
		{name: "does not treat extra arguments as a migration command", args: []string{"migrate", "extra"}, want: false},
		{name: "does not treat another command as a migration command", args: []string{"serve"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMigrationCommand(tt.args); got != tt.want {
				t.Errorf("isMigrationCommand(%q) = %t, want %t", tt.args, got, tt.want)
			}
		})
	}
}
