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
	// modelInfoCache holds the model info of each type, by full type name.
	modelInfoCache = map[string]*ModelInfo{}
	// modelInfoStack lists the types whose info is being built, to stop
	// infinite recursion between models that refer to each other.
	modelInfoStack []string
	modelInfoMutex = sync.Mutex{}

	// dumper produces a stable text dump of a ModelInfo for its checksum.
	dumper = spew.ConfigState{
		Indent:                  "\t",
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
	}
)

// ModelInfo describes how a model type maps to a table: its columns,
// primary key and relations. Checksum identifies the mapping in cache keys.
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

// Relation describes a belongs-to or has-many relation. Key holds the
// columns of the related table that are matched with the Reference
// columns of the model. Filter, Order and Limit apply to has-many
// relations.
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

// GetModelInfo returns the cached model info for the type of model, which
// may be a struct or a pointer to one. It returns nil for other types.
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

// NewModelInfo builds the model info for the type of model without using
// the cache.
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

// getModelInfoByType returns the model info of r, building it if needed.
// It returns nil if r is already being built further up the stack, which
// leaves out relations that point back to it. Info built while another
// type is being built is not cached, because it may lack such relations.
// The caller must hold modelInfoMutex.
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

// newModelInfoByType builds the model info of a struct type.
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

// addFields adds the fields of modelType. Fields of embedded structs are
// added as if they belonged to the model, and the tag of the embedded
// field itself can set the table and database.
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

// addField adds a column or relation field. indexes is the path of
// embedded structs that contain the field.
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

// addFieldUpdate applies the model-level settings of a field tag: database,
// table and primary key.
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

// addBelongsTo adds a relation where this model's foreign key references
// the primary key of the field's type.
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

// addHasMany adds a relation where the foreign key of the slice element
// type references this model's primary key.
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

// fieldInfo parses the tag of field and fills in the defaults. It returns
// nil for fields tagged "-".
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

// foreignKey returns the fields of mi that reference the primary key of
// fmi. Without explicit columns, the names are derived from fmi's table and
// primary key: "<singular table>_<column>".
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

// filter calls the filter method named in fi on a new model value and
// returns the condition it returns, or nil.
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

// checksum returns a hash of the model info.
func (mi *ModelInfo) checksum() uint64 {
	b := new(bytes.Buffer)
	dumper.Fdump(b, mi)
	return xxhash.Sum64(b.Bytes())
}

// isStruct reports whether fieldType is a struct type that forms a
// relation. time.Time is a plain column.
func isStruct(fieldType reflect.Type) bool {
	if fieldType.Kind() != reflect.Struct {
		return false
	}
	if fieldType.String() == "time.Time" {
		return false
	}
	return true
}

// isSliceOfStruct reports whether fieldType is a slice of structs or
// pointers to structs.
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
