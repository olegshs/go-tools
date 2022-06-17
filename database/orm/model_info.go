package orm

import (
	"bytes"
	"reflect"
	"sync"

	"github.com/cespare/xxhash"
	"github.com/davecgh/go-spew/spew"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"

	"github.com/olegshs/go-tools/database"
)

var (
	modelInfoCache = map[string]*ModelInfo{}
	modelInfoStack []string
	modelInfoMutex = sync.Mutex{}

	dumper = spew.ConfigState{
		Indent:                  "\t",
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
	}
)

type ModelInfo struct {
	Type      reflect.Type
	Database  string
	Table     string
	Fields    FieldInfoList
	Primary   FieldInfoList
	BelongsTo map[string]*Relation
	HasMany   map[string]*Relation
	Checksum  uint64
}

type Relation struct {
	FieldIndex []int
	Database   string
	Table      string
	Key        FieldInfoList
	Reference  FieldInfoList
	Filter     interface{}
	Order      []interface{}
	Limit      int
}

func GetModelInfo(model interface{}) *ModelInfo {
	modelType := reflect.TypeOf(model)
	for modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return nil
	}

	modelInfoMutex.Lock()
	defer modelInfoMutex.Unlock()

	mi := getModelInfoByType(modelType)
	return mi
}

func NewModelInfo(model interface{}) *ModelInfo {
	modelType := reflect.TypeOf(model)
	for modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return nil
	}

	mi := newModelInfoByType(modelType)
	return mi
}

func getModelInfoByType(r reflect.Type) *ModelInfo {
	modelName := r.PkgPath() + "." + r.Name()

	mi, ok := modelInfoCache[modelName]
	if ok {
		return mi
	}

	for _, v := range modelInfoStack {
		if modelName == v {
			return nil
		}
	}
	stackLen := len(modelInfoStack)
	modelInfoStack = append(modelInfoStack, modelName)

	mi = newModelInfoByType(r)

	modelInfoStack = modelInfoStack[:stackLen]
	if stackLen == 0 {
		modelInfoCache[modelName] = mi
	}

	return mi
}

func newModelInfoByType(modelType reflect.Type) *ModelInfo {
	for modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	mi := new(ModelInfo)
	mi.Type = modelType
	mi.Database = database.DefaultDB
	mi.Table = strcase.ToSnake(inflection.Plural(modelType.Name()))
	mi.BelongsTo = make(map[string]*Relation)
	mi.HasMany = make(map[string]*Relation)

	mi.addFields(modelType)

	mi.Checksum = mi.checksum()

	return mi
}

func (mi *ModelInfo) addFields(modelType reflect.Type, indexes ...[]int) {
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if field.PkgPath != "" { // skip unexported fields
			continue
		}

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			fi := mi.fieldInfo(&field)
			mi.addFieldUpdate(fi)

			mi.addFields(field.Type, field.Index)
			continue
		}

		mi.addField(&field, indexes...)
	}
}

func (mi *ModelInfo) addField(field *reflect.StructField, indexes ...[]int) {
	fi := mi.fieldInfo(field)
	if fi == nil {
		return
	}

	if len(indexes) > 0 {
		fi.FieldIndex = append(indexes, fi.FieldIndex...)
	}

	t := field.Type
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if isStruct(t) {
		mi.addBelongsTo(field, fi)
		return
	}

	if isSliceOfStruct(t) {
		mi.addHasMany(field, fi)
		return
	}

	mi.Fields = append(mi.Fields, fi)

	mi.addFieldUpdate(fi)
}

func (mi *ModelInfo) addFieldUpdate(fi *FieldInfo) {
	if fi.Database != "" {
		mi.Database = fi.Database
	}

	if fi.Table != "" {
		mi.Table = fi.Table
	}

	if fi.Primary {
		mi.Primary = append(mi.Primary, fi)
	}
}

func (mi *ModelInfo) addBelongsTo(field *reflect.StructField, fi *FieldInfo) {
	t := field.Type
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fmi := getModelInfoByType(t)
	if fmi == nil {
		return
	}

	rel := &Relation{
		FieldIndex: field.Index,
		Database:   fmi.Database,
		Table:      fmi.Table,
		Key:        fmi.Primary,
		Reference:  mi.foreignKey(fmi, fi.ForeignKey),
	}

	mi.BelongsTo[field.Name] = rel
}

func (mi *ModelInfo) addHasMany(field *reflect.StructField, fi *FieldInfo) {
	t := field.Type
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	e := t.Elem()
	for e.Kind() == reflect.Ptr {
		e = e.Elem()
	}

	fmi := getModelInfoByType(e)
	if fmi == nil {
		return
	}

	rel := &Relation{
		FieldIndex: field.Index,
		Database:   fmi.Database,
		Table:      fmi.Table,
		Key:        fmi.foreignKey(mi, fi.ForeignKey),
		Reference:  mi.Primary,
		Filter:     mi.filter(fi),
		Order:      fi.Order,
		Limit:      fi.Limit,
	}

	mi.HasMany[field.Name] = rel
}

func (mi *ModelInfo) fieldInfo(field *reflect.StructField) *FieldInfo {
	var fi *FieldInfo

	if tag, ok := field.Tag.Lookup("orm"); ok {
		fi = ParseFieldTag(tag)
		if fi == nil {
			return nil
		}
	} else {
		fi = new(FieldInfo)
	}

	fi.FieldIndex = [][]int{field.Index}

	if fi.Database == "" {
		fi.Database = mi.Database
	}

	if fi.Table == "" {
		fi.Table = mi.Table
	}

	if fi.Column == "" {
		fi.Column = strcase.ToSnake(field.Name)
	}

	return fi
}

func (mi *ModelInfo) foreignKey(fmi *ModelInfo, columns []string) FieldInfoList {
	if len(columns) == 0 {
		for _, fi := range fmi.Primary {
			c := strcase.ToSnake(inflection.Singular(fmi.Table)) + "_" + fi.Column
			columns = append(columns, c)
		}
	}

	fk := make(FieldInfoList, len(columns))
	for i, col := range columns {
		fi := mi.Fields.ByColumn(col)
		if fi == nil {
			fk[i] = &FieldInfo{Column: col}
		} else {
			fk[i] = fi
		}
	}

	return fk
}

func (mi *ModelInfo) filter(fi *FieldInfo) interface{} {
	var filter interface{}

	filterMethod := reflect.New(mi.Type).MethodByName(fi.Filter)
	if filterMethod.IsValid() {
		a := filterMethod.Call(nil)
		if len(a) > 0 {
			filter = a[0].Interface()
		}
	}

	return filter
}

func (mi *ModelInfo) checksum() uint64 {
	b := new(bytes.Buffer)
	dumper.Fdump(b, mi)
	return xxhash.Sum64(b.Bytes())
}

func isStruct(fieldType reflect.Type) bool {
	if fieldType.Kind() != reflect.Struct {
		return false
	}
	if fieldType.String() == "time.Time" {
		return false
	}
	return true
}

func isSliceOfStruct(fieldType reflect.Type) bool {
	if fieldType.Kind() != reflect.Slice {
		return false
	}

	elemType := fieldType.Elem()
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return false
	}

	return true
}
