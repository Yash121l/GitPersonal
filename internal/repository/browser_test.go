package repository

import (
	"errors"
	"testing"

	"github.com/yashlunawat/forge/internal/store"
)

func TestNormalizeBrowserPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "empty path", input: "", want: ""},
		{name: "rooted path", input: "/src/main.go", want: "src/main.go"},
		{name: "nested path", input: "internal/server/ui.go", want: "internal/server/ui.go"},
		{name: "dot segment rejected", input: "./README.md", wantErr: store.ErrInvalidArgument},
		{name: "parent traversal rejected", input: "../etc/passwd", wantErr: store.ErrInvalidArgument},
		{name: "embedded traversal rejected", input: "src/../main.go", wantErr: store.ErrInvalidArgument},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeBrowserPath(testCase.input)
			if testCase.wantErr != nil {
				if !errors.Is(err, testCase.wantErr) {
					t.Fatalf("expected error %v, got %v", testCase.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalize browser path: %v", err)
			}
			if got != testCase.want {
				t.Fatalf("normalize browser path = %q, want %q", got, testCase.want)
			}
		})
	}
}
