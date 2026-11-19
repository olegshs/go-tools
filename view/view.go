// Package view renders html/template files with layouts, parent templates
// and a set of helper functions.
//
// Template names are file paths relative to config.BaseDir. A template can
// choose its own layout or parent from inside, and render other templates:
//
//	{{layout "views/layout.html"}}
//	<h1>{{.Title}}</h1>
//	{{render "views/partials/list.html" .Items}}
//
// The layout or parent receives the rendered content as {{.Content}}.
// Besides the data passed to Render and the view's own values (Set), every
// template gets .time (the current time), .template (its name), .level (the
// nesting depth) and .args (the arguments of render).
//
// Template functions:
//
//	get, set, unset   read and change the view's values
//	layout NAME       render the result inside the layout NAME
//	parent NAME       render this template inside NAME (like layout, but
//	                  for nested templates)
//	render NAME ARGS  insert another template rendered with the same data
//	url NAME PARAMS   URL of a named route of the default router
//	raw, json         insert a string or a value as JSON without escaping
//	base64            encode bytes as base64
//	file NAME         contents of a file
//
// Template sources and files read with "file" are cached for the life of
// the process, so changes on disk need a restart.
package view

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/router"
)

// View renders templates. It keeps state while rendering, so it is not safe
// for concurrent use; create one per request with New.
type View struct {
	// layout is the template that wraps the result of Render.
	layout string
	// parent is the template that wraps the template being rendered. It is
	// reset after use, so each template sets its own.
	parent string
	// level is the current nesting depth, limited by MaxRecursion.
	level     int
	templates map[string]*template.Template
	funcs     template.FuncMap
	// data holds values shared by all templates of this view.
	data map[string]interface{}
	// currentData is the data of the template being rendered, passed on to
	// templates included with render.
	currentData map[string]interface{}
}

var (
	// MaxRecursion limits how deeply layouts, parents and render calls can
	// nest. ErrMaxRecursion is returned when it is exceeded.
	MaxRecursion    = 10
	ErrMaxRecursion = errors.New("maximum recursion depth exceeded")

	// templates and files cache template sources and file contents by name.
	// They are shared by all views and guarded by cacheMutex.
	templates  = map[string]string{}
	files      = map[string][]byte{}
	cacheMutex sync.RWMutex
)

// New creates a view without a layout.
func New() *View {
	view := new(View)
	view.layout = ""
	view.parent = ""
	view.level = 0
	view.templates = map[string]*template.Template{}
	view.funcs = view.defaultFuncs()
	view.data = map[string]interface{}{}
	view.currentData = map[string]interface{}{}

	return view
}

// Get returns a value shared by all templates of the view.
func (view *View) Get(key string) interface{} {
	return view.data[key]
}

// Set sets a value shared by all templates of the view. Data passed to
// Render takes precedence over it.
func (view *View) Set(key string, value interface{}) {
	view.data[key] = value
}

// Delete removes a shared value.
func (view *View) Delete(key string) {
	delete(view.data, key)
}

// Layout returns the name of the layout template.
func (view *View) Layout() string {
	return view.layout
}

// SetLayout sets the layout template that wraps the result of Render.
// Templates can also set it with the "layout" function.
func (view *View) SetLayout(layout string) {
	view.layout = layout
}

// SetFunc adds a template function. It must be called before the templates
// that use it are first rendered by this view.
func (view *View) SetFunc(name string, f interface{}) {
	view.funcs[name] = f
}

// DeleteFunc removes a template function.
func (view *View) DeleteFunc(name string) {
	delete(view.funcs, name)
}

// Render renders the template name with data, then its parents, and then
// the layout if one is set. Rendered content is trimmed and has its line
// endings converted to "\n". data may be nil; it is modified by adding the
// "Content" key.
func (view *View) Render(name string, data map[string]interface{}) (string, error) {
	if data == nil {
		data = map[string]interface{}{}
	}

	content, err := view.renderTemplate(name, data)
	if err != nil {
		return content, err
	}

	if len(view.layout) > 0 && view.layout != name {
		view.level++
		if view.level > MaxRecursion {
			return content, ErrMaxRecursion
		}

		data["Content"] = template.HTML(content)
		content, err = view.renderTemplate(view.layout, data)

		view.level--

		if err != nil {
			return content, err
		}
	}

	return content, nil
}

// renderTemplate renders one template and the parent it sets, if any.
func (view *View) renderTemplate(name string, data map[string]interface{}, args ...interface{}) (string, error) {
	if view.level > MaxRecursion {
		return "", ErrMaxRecursion
	}

	t, err := view.template(name)
	if err != nil {
		return "", err
	}

	view.currentData = data

	m := map[string]interface{}{
		"time":     time.Now(),
		"template": name,
		"level":    view.level,
		"args":     args,
	}
	for k, v := range view.data {
		m[k] = v
	}
	for k, v := range data {
		m[k] = v
	}

	buf := new(bytes.Buffer)

	err = t.Execute(buf, m)
	if err != nil {
		return "", err
	}

	content := buf.String()
	content = strings.Replace(content, "\r\n", "\n", -1)
	content = helpers.Trim(content)

	if len(view.parent) > 0 && view.parent != name {
		view.level++
		if view.level > MaxRecursion {
			return content, ErrMaxRecursion
		}

		parent := view.parent
		view.parent = ""

		data["Content"] = template.HTML(content)
		content, err = view.renderTemplate(parent, data, args...)

		view.level--

		if err != nil {
			return content, err
		}
	}

	return content, nil
}

// template returns the parsed template, parsing it on first use by this
// view.
func (view *View) template(name string) (*template.Template, error) {
	t, ok := view.templates[name]
	if ok {
		return t, nil
	}

	t, err := view.newTemplate(name)
	if err != nil {
		return nil, err
	}

	view.templates[name] = t

	return t, nil
}

// newTemplate parses the template file with the view's functions.
func (view *View) newTemplate(name string) (*template.Template, error) {
	t := template.New(path.Base(name))
	t.Funcs(view.funcs)

	c, err := templateContent(name)
	if err != nil {
		return nil, err
	}

	_, err = t.Parse(c)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// defaultFuncs returns the built-in template functions. See the package
// documentation.
func (view *View) defaultFuncs() template.FuncMap {
	return template.FuncMap{
		"get": func(key string) interface{} {
			return view.Get(key)
		},
		"set": func(key string, value interface{}) interface{} {
			view.Set(key, value)
			return nil
		},
		"unset": func(key string) interface{} {
			view.Delete(key)
			return nil
		},
		"layout": func(name string) interface{} {
			view.layout = name
			return nil
		},
		"parent": func(name string) interface{} {
			view.parent = name
			return nil
		},
		// raw and json bypass escaping; use them only with trusted content.
		"raw": func(html string) template.HTML {
			return template.HTML(html)
		},
		"json": func(a interface{}) template.HTML {
			j, _ := json.Marshal(a)
			return template.HTML(string(j))
		},
		"url": func(name string, params ...interface{}) string {
			url, _ := router.Url(name, params...)
			return url
		},
		"base64": func(data []byte) string {
			return base64.StdEncoding.EncodeToString(data)
		},
		"file": fileContent,
		// render errors are ignored and produce empty output.
		"render": func(name string, args ...interface{}) template.HTML {
			view.level++
			parent := view.parent
			view.parent = ""

			data := view.currentData
			content, _ := view.renderTemplate(name, data, args...)

			view.level--
			view.parent = parent

			return template.HTML(content)
		},
	}
}

// templateContent returns the source of a template file, reading it on
// first use.
func templateContent(name string) (string, error) {
	cacheMutex.RLock()
	t, ok := templates[name]
	cacheMutex.RUnlock()
	if ok {
		return t, nil
	}

	b, err := ioutil.ReadFile(config.AbsPath(name))
	if err != nil {
		return "", err
	}

	t = string(b)

	cacheMutex.Lock()
	templates[name] = t
	cacheMutex.Unlock()

	return t, nil
}

// fileContent returns the contents of a file, reading it on first use. It
// returns nil if the file cannot be read, and caches that result too.
func fileContent(filename string) []byte {
	cacheMutex.RLock()
	b, ok := files[filename]
	cacheMutex.RUnlock()
	if ok {
		return b
	}

	b, _ = ioutil.ReadFile(config.AbsPath(filename))

	cacheMutex.Lock()
	files[filename] = b
	cacheMutex.Unlock()

	return b
}
