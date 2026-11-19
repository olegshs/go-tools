package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/olegshs/go-tools/config"
)

func TestStaticMiddleware_BaseDir(t *testing.T) {
	baseDir := t.TempDir()
	err := os.MkdirAll(filepath.Join(baseDir, "public"), 0700)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(baseDir, "public", "file.txt"), []byte("static"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// The static directory is relative to config.BaseDir, not to the
	// working directory.
	oldBaseDir := config.BaseDir
	config.BaseDir = baseDir
	defer func() { config.BaseDir = oldBaseDir }()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("next"))
	})
	handler := staticMiddleware(next, ConfigStatic{Enabled: true, Dir: "public"})

	tests := map[string]string{
		"/file.txt":    "static",
		"/missing.txt": "next",
		"/.hidden":     "next",
	}

	for path, expected := range tests {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

		if body := w.Body.String(); body != expected {
			t.Errorf("GET %s: %q, expected %q", path, body, expected)
		}
	}
}
