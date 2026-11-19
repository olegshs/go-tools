package view

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/olegshs/go-tools/config"
)

func TestView_RenderConcurrent(t *testing.T) {
	baseDir := t.TempDir()

	templates := map[string]string{
		"layout.html": `<main>{{.Content}}</main>`,
		"page.html":   `{{layout "layout.html"}}<p>{{.Title}} {{printf "%s" (file "data.txt")}}</p>`,
		"data.txt":    `data`,
	}
	for name, content := range templates {
		err := os.WriteFile(filepath.Join(baseDir, name), []byte(content), 0600)
		if err != nil {
			t.Fatal(err)
		}
	}

	oldBaseDir := config.BaseDir
	config.BaseDir = baseDir
	defer func() { config.BaseDir = oldBaseDir }()

	// Views are rendered concurrently, as in an HTTP server, and share the
	// package-level template and file caches.
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			title := fmt.Sprint(i)
			html, err := New().Render("page.html", map[string]interface{}{"Title": title})
			if err != nil {
				t.Error(err)
				return
			}

			expected := "<main><p>" + title + " data</p></main>"
			if html != expected {
				t.Errorf("Render() = %q, expected %q", html, expected)
			}
		}(i)
	}
	wg.Wait()
}
