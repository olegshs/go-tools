// Пакет config предоставляет функции для работы с конфигурационной информацией.
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

	"github.com/olegshs/go-tools/helpers/structmap"
	"github.com/olegshs/go-tools/helpers/typeconv"
)

var (
	// BaseDir хранит базовую директорию, которая используется функцией AbsPath.
	BaseDir = "."

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

// AbsPath возвращает абсолютный путь к файлу,
// принимая путь относительно директории BaseDir.
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

// Load загружает в указанный раздел конфигурацию из файла JSON или YAML.
// Если в директории с файлом есть поддиректория "local",
// то из неё будет загружен файл с тем же именем, после загрузки указанного файла.
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

// LoadAll загружает конфигурации из всех файлов JSON или YAML,
// которые соответствуют указанной маске.
// Имена файлов считаются именами разделов.
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

// Set устанавливает значение параметра.
func Set(key string, value interface{}) {
	mutex.Lock()
	defer mutex.Unlock()

	set(key, value)
}

func set(key string, value interface{}) {
	a := strings.Split(key, ".")
	n := len(a)
	if n > 1 {
		parentKey := strings.Join(a[:n-1], ".")
		_, parentExists := data[parentKey]
		if !parentExists {
			set(parentKey, map[string]interface{}{
				a[n-1]: value,
			})
			return
		}
	}

	data[key] = merge(data[key], value)

	m, ok := value.(map[string]interface{})
	if ok {
		for k, v := range m {
			set(key+"."+k, v)
		}
	}
}

// Exists проверяет существование параметра.
func Exists(key string) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	_, ok := data[key]
	return ok
}

// Get возвращает значение параметра.
func Get(key string, defaultValue interface{}) interface{} {
	mutex.RLock()
	defer mutex.RUnlock()

	value, ok := data[key]
	if !ok {
		return defaultValue
	}

	return value
}

// GetBool возвращает значение параметра с преобразованием в булев тип.
func GetBool(key string, defaultValue bool) bool {
	value := Get(key, defaultValue)
	return typeconv.Bool(value)
}

// GetInt возвращает значение параметра с преобразованием в целое число.
func GetInt(key string, defaultValue int) int {
	value := Get(key, defaultValue)
	return typeconv.Int(value)
}

// GetString возвращает значение параметра с преобразованием в строку.
func GetString(key string, defaultValue string) string {
	value := Get(key, defaultValue)
	return typeconv.String(value)
}

// GetStruct копирует значения раздела в структуру.
func GetStruct(key string, dst interface{}) {
	m, ok := Get(key, nil).(map[string]interface{})
	if !ok {
		return
	}

	structmap.ToStruct(m, dst)
}

// GetSlice копирует значения раздела в массив.
func GetSlice(key string, dst interface{}) {
	a, ok := Get(key, nil).([]interface{})
	if !ok {
		return
	}

	structmap.ToSlice(a, dst)
}

func localFilename(filename string) string {
	dir := path.Dir(filename)
	base := path.Base(filename)
	local := filepath.Join(dir, "local", base)
	return local
}

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
