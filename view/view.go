// Пакет view предоставляет вспомогательные функции для работы с шаблонами HTML.
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
	"time"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/router"
)

type View struct {
	layout      string
	parent      string
	level       int
	templates   map[string]*template.Template
	funcs       template.FuncMap
	data        map[string]interface{}
	currentData map[string]interface{}
}

var (
	MaxRecursion    = 10
	ErrMaxRecursion = errors.New("maximum recursion depth exceeded")

	templates = map[string]string{}
	files     = map[string][]byte{}
)

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

func (view *View) Get(key string) interface{} {
	return view.data[key]
}

func (view *View) Set(key string, value interface{}) {
	view.data[key] = value
}

func (view *View) Delete(key string) {
	delete(view.data, key)
}

func (view *View) Layout() string {
	return view.layout
}

func (view *View) SetLayout(layout string) {
	view.layout = layout
}

func (view *View) SetFunc(name string, f interface{}) {
	view.funcs[name] = f
}

func (view *View) DeleteFunc(name string) {
	delete(view.funcs, name)
}

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
		"file": func(filename string) []byte {
			b, ok := files[filename]
			if ok {
				return b
			}

			b, _ = ioutil.ReadFile(config.AbsPath(filename))
			files[filename] = b

			return b
		},
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

func templateContent(name string) (string, error) {
	t, ok := templates[name]
	if ok {
		return t, nil
	}

	b, err := ioutil.ReadFile(config.AbsPath(name))
	if err != nil {
		return "", err
	}

	t = string(b)
	templates[name] = t

	return t, nil
}
