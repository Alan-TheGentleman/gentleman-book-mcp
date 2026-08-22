package book

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// newLocaleFixture creates a temp book path with the given locale dirs,
// each containing one .mdx file.
func newLocaleFixture(t *testing.T, dirs ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, d := range dirs {
		dir := filepath.Join(root, d)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "index.mdx"), []byte("# test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestGetAvailableLocales(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		want    []string
		wantErr bool
	}{
		{
			name: "Discovery of all locale dirs",
			setup: func(t *testing.T) string {
				return newLocaleFixture(t, "en", "es", "ai-agent", "harness", "secret-knowledge")
			},
			want: []string{"ai-agent", "en", "es", "harness", "secret-knowledge"},
		},
		{
			name: "Hidden directory excluded",
			setup: func(t *testing.T) string {
				return newLocaleFixture(t, "en", ".git")
			},
			want: []string{"en"},
		},
		{
			name: "Directory without .mdx files",
			setup: func(t *testing.T) string {
				root := t.TempDir()
				if err := os.MkdirAll(filepath.Join(root, "images"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "images", "cover.png"), []byte("png"), 0o644); err != nil {
					t.Fatal(err)
				}
				return root
			},
			want: nil,
		},
		{
			name: "Empty book path",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "does-not-exist")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.setup(t))
			got, err := p.GetAvailableLocales()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAvailableLocales() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !sort.StringsAreSorted(got) {
				t.Errorf("GetAvailableLocales() not sorted: %v", got)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("GetAvailableLocales() = %v, want %v", got, tt.want)
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("GetAvailableLocales()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestGetAvailableLocalesSorted(t *testing.T) {
	root := newLocaleFixture(t, "zeta", "alpha", "mid")
	p := NewParser(root)
	got, err := p.GetAvailableLocales()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"alpha", "mid", "zeta"}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("got[%d]=%q want %q (full: %v)", i, v, want[i], got)
		}
	}
}
