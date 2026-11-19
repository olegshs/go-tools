// Package config holds the application configuration in a global,
// concurrency-safe key/value store.
//
// Configuration is loaded from JSON or YAML files. Each file becomes a
// section, and nested values can be read with dotted keys:
//
//	config.LoadAll("config/*.yaml") // config/database.yaml -> "database"
//	driver := config.GetString("database.default.driver", "sqlite3")
//
// Other packages read their settings into structs with GetStruct, starting
// from their DefaultConfig, so only the values present in the files
// override the defaults. See helpers/structmap for how keys map to struct
// fields.
package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/olegshs/go-tools/helpers/structmap"
	"github.com/olegshs/go-tools/helpers/typeconv"
)

var (
	// BaseDir is the directory that AbsPath resolves relative paths against.
	// It is set to the working directory at startup.
	BaseDir = "."

	// data stores every value under its full dotted key. A map value is also
	// stored under each of its children's keys, so both "a" and "a.b" can be
	// read directly.
	data  = map[string]interface{}{}
	mutex = sync.RWMutex{}
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	BaseDir = dir
}

// AbsPath returns filename as an absolute path. Relative paths are resolved
// against BaseDir.
func AbsPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filepath.Clean(filename)
	}

	fullPath := filepath.Join(BaseDir, filename)

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fullPath
	}

	return absPath
}

// Load reads a JSON or YAML file and merges it into the section key. Files
// ending in ".json" are parsed as JSON, and all others as YAML.
//
// If a file with the same name exists in a "local" subdirectory next to the
// file, it is loaded afterwards and overrides it. This is meant for
// machine-specific settings that are not committed.
func Load(key string, filename string) error {
	m, err := read(filename)
	if err != nil {
		return err
	}

	Set(key, m)

	localFilename := localFilename(filename)
	if _, err := os.Stat(localFilename); err == nil {
		m, err := read(localFilename)
		if err != nil {
			return err
		}

		Set(key, m)
	}

	return nil
}

// LoadAll calls Load for every file in a directory whose name matches the
// pattern, such as "config/*.yaml". The file name without its extension is
// used as the section name. Subdirectories are not searched.
func LoadAll(pattern string) error {
	dirname := path.Dir(pattern)
	pattern = path.Base(pattern)

	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()

		m, err := filepath.Match(pattern, filename)
		if err != nil {
			return err
		}
		if !m {
			continue
		}

		ext := path.Ext(filename)
		name := filename[:len(filename)-len(ext)]

		err = Load(name, dirname+"/"+filename)
		if err != nil {
			return err
		}
	}

	return nil
}

// Set stores value under key. Map values are deep-merged with the existing
// value. A dotted key whose parent sections do not exist creates them.
//
// If the parent section already exists, it is not updated: after
// Set("a.b", 1), Get("a.b") returns 1 but the map returned by Get("a") does
// not change. Set whole sections when other code reads them with GetStruct.
func Set(key string, value interface{}) {
	mutex.Lock()
	defer mutex.Unlock()

	set(key, value)
}

// set implements Set. The caller must hold the write lock.
func set(key string, value interface{}) {
	a := strings.Split(key, ".")
	n := len(a)
	if n > 1 {
		parentKey := strings.Join(a[:n-1], ".")
		_, parentExists := data[parentKey]
		// Store the value through its parent, which in turn stores the
		// value under this key.
		if !parentExists {
			set(parentKey, map[string]interface{}{
				a[n-1]: value,
			})
			return
		}
	}

	data[key] = merge(data[key], value)

	// Also store each child of a map under its own dotted key.
	m, ok := value.(map[string]interface{})
	if ok {
		for k, v := range m {
			set(key+"."+k, v)
		}
	}
}

// Exists reports whether key has a value.
func Exists(key string) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	_, ok := data[key]
	return ok
}

// Get returns the value of key, or defaultValue if it is not set.
// Maps are returned as stored, so callers must not modify them.
func Get(key string, defaultValue interface{}) interface{} {
	mutex.RLock()
	defer mutex.RUnlock()

	value, ok := data[key]
	if !ok {
		return defaultValue
	}

	return value
}

// GetBool returns the value of key converted with typeconv.Bool.
func GetBool(key string, defaultValue bool) bool {
	value := Get(key, defaultValue)
	return typeconv.Bool(value)
}

// GetInt returns the value of key converted with typeconv.Int.
func GetInt(key string, defaultValue int) int {
	value := Get(key, defaultValue)
	return typeconv.Int(value)
}

// GetString returns the value of key converted with typeconv.String.
func GetString(key string, defaultValue string) string {
	value := Get(key, defaultValue)
	return typeconv.String(value)
}

// GetStruct copies the section key into the struct that dst points to,
// using structmap.ToStruct. Fields not present in the section are left
// unchanged. It does nothing if key is not set or is not a map.
func GetStruct(key string, dst interface{}) {
	m, ok := Get(key, nil).(map[string]interface{})
	if !ok {
		return
	}

	structmap.ToStruct(m, dst)
}

// GetSlice copies the list at key into the slice that dst points to,
// using structmap.ToSlice. It does nothing if key is not set or is not a
// list.
func GetSlice(key string, dst interface{}) {
	a, ok := Get(key, nil).([]interface{})
	if !ok {
		return
	}

	structmap.ToSlice(a, dst)
}

// localFilename returns the path of the local override for filename:
// dir/local/name.
func localFilename(filename string) string {
	dir := path.Dir(filename)
	base := path.Base(filename)
	local := filepath.Join(dir, "local", base)
	return local
}

// read parses a config file into a map.
func read(filename string) (map[string]interface{}, error) {
	filename = AbsPath(filename)
	ext := path.Ext(filename)

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var m map[string]interface{}

	switch ext {
	case ".json":
		m, err = decodeJson(f)
	default:
		m, err = decodeYaml(f)
	}

	if err != nil {
		return nil, err
	}

	return m, nil
}

// decodeJson decodes JSON keeping numbers as json.Number, so large integers
// are not rounded through float64.
func decodeJson(f io.Reader) (map[string]interface{}, error) {
	var m map[string]interface{}

	d := json.NewDecoder(f)
	d.UseNumber()

	err := d.Decode(&m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// decodeYaml decodes YAML and converts its map[interface{}]interface{}
// values to map[string]interface{}, as returned by the JSON decoder.
func decodeYaml(f io.Reader) (map[string]interface{}, error) {
	var m map[string]interface{}

	d := yaml.NewDecoder(f)

	err := d.Decode(&m)
	if err != nil {
		return nil, err
	}

	m = convertMap(m).(map[string]interface{})
	return m, nil
}

// convertMap recursively converts YAML maps to map[string]interface{}.
func convertMap(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		m := map[string]interface{}{}
		for k, v := range t {
			m[k] = convertMap(v)
		}
		return m
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v := range t {
			m[typeconv.String(k)] = convertMap(v)
		}
		return m
	case []interface{}:
		a := make([]interface{}, len(t))
		for k, v := range t {
			a[k] = convertMap(v)
		}
		return a
	default:
		return t
	}
}

// merge returns b deep-merged over a. If b is not a map, b replaces a.
// Neither a nor b is modified.
func merge(a, b interface{}) interface{} {
	bm, ok := b.(map[string]interface{})
	if !ok {
		return b
	}

	c := map[string]interface{}{}
	am, ok := a.(map[string]interface{})
	if ok {
		for k, v := range am {
			c[k] = v
		}
	}

	for k, v := range bm {
		vm, ok := v.(map[string]interface{})
		if ok {
			c[k] = merge(c[k], vm)
		} else {
			c[k] = v
		}
	}

	return c
}
