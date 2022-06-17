// Пакет orm реализует объектно-реляционное отображение.
package orm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/olegshs/go-tools/cache"
	"github.com/olegshs/go-tools/database"
	"github.com/olegshs/go-tools/database/interfaces"
	"github.com/olegshs/go-tools/database/query"
	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/helpers/typeconv"
)

var (
	relationNameRegexp = regexp.MustCompile(`^(\w+)(\s*\[\s*(.*)\s*])?$`)
)

type Query struct {
	tx         interfaces.DB
	columns    helpers.Slice[string]
	relations  []queryRelation
	conditions []interface{}
	order      []interface{}
	limit      []int

	model      interface{}
	modelValue reflect.Value
	modelInfo  *ModelInfo

	belongsToCache map[string]map[string]reflect.Value
}

type queryRelation struct {
	name string
	with []string
}

type fieldsData struct {
	fields  []reflect.Value
	columns []interface{}
	values  []interface{}
	info    []*FieldInfo
}

func (q *Query) Tx(tx interfaces.DB) *Query {
	q.tx = tx
	return q
}

func (q *Query) Select(columns ...string) *Query {
	q.columns = columns
	return q
}

func (q *Query) With(relations ...string) *Query {
	a := make([]queryRelation, len(relations))
	for i, r := range relations {
		a[i] = q.parseRelationName(r)
	}

	q.relations = append(q.relations, a...)
	return q
}

func (q *Query) Where(conditions ...interface{}) *Query {
	q.conditions = append(q.conditions, conditions...)
	return q
}

func (q *Query) Order(order ...interface{}) *Query {
	q.order = append(q.order, order...)
	return q
}

func (q *Query) Limit(limit ...int) *Query {
	q.limit = limit
	return q
}

func (q *Query) Count(model interface{}) (int, error) {
	t := reflect.TypeOf(model)
	for (t.Kind() == reflect.Ptr) || (t.Kind() == reflect.Slice) {
		t = t.Elem()
	}

	err := q.setModelByType(t)
	if err != nil {
		return 0, err
	}

	db, err := q.modelDB()
	if err != nil {
		return 0, err
	}

	row := db.Select(query.Expr(`COUNT(*)`)).
		From(q.modelInfo.Table).
		Where(q.conditions...).
		Row()

	var count int

	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (q *Query) First(model interface{}) error {
	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	err := q.setModel(model)
	if err != nil {
		return err
	}

	if len(q.conditions) == 0 {
		q.conditions = q.primaryKeyConditions()
	}

	pk := q.primaryKeyFromConditions(q.conditions)
	if (len(pk) > 0) && (len(q.columns) == 0) {
		err = q.cacheGet(pk)
		if err == nil {
			return nil
		}
		if err == cache.ErrEmptyObject {
			return ErrNoRows
		}
	}

	db, err := q.modelDB()
	if err != nil {
		return err
	}

	fd := q.parseFields()

	row := db.Select(fd.columns...).
		From(q.modelInfo.Table).
		Where(q.conditions...).
		Order(q.order...).
		Limit(1).
		Row()

	q.clearModel()

	err = q.scan(row, fd)
	if err != nil {
		if err == ErrNoRows {
			q.cacheSetEmpty(pk)
		}
		return err
	}

	if len(q.columns) == 0 {
		q.cacheSet(pk)
	}

	err = q.loadRelated()
	if err != nil {
		return err
	}

	return nil
}

func (q *Query) Find(models interface{}) error {
	t := reflect.TypeOf(models)
	if t.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	t = t.Elem()
	if t.Kind() != reflect.Slice {
		return ErrNotSlice
	}

	err := q.setModelByType(t.Elem())
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(models).Elem()
	rv.Set(
		reflect.MakeSlice(t, 0, 0),
	)

	db, err := q.modelDB()
	if err != nil {
		return err
	}

	fd := q.parseFields()

	rows, err := db.Select(fd.columns...).
		From(q.modelInfo.Table).
		Where(q.conditions...).
		Order(q.order...).
		Limit(q.limit...).
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		q.clearModel()

		err := q.scan(rows, fd)
		if err != nil {
			return err
		}

		if len(q.columns) == 0 {
			q.cacheSet(q.modelPrimaryKey())
		}

		err = q.loadRelated()
		if err != nil {
			return err
		}

		newItem := q.newValue(q.modelInfo.Type)
		q.copyStruct(newItem.Elem(), q.modelValue)

		if t.Elem().Kind() != reflect.Ptr {
			newItem = newItem.Elem()
		}
		rv.Set(
			reflect.Append(rv, newItem),
		)
	}

	return nil
}

func (q *Query) LoadRelated(model interface{}, relations ...string) error {
	err := q.setModel(model)
	if err != nil {
		return err
	}

	q.With(relations...)

	err = q.loadRelated()
	if err != nil {
		return err
	}

	return nil
}

func (q *Query) Create(model interface{}) error {
	err := q.setModel(model)
	if err != nil {
		return err
	}

	data := query.Data{}
	var autoIncrement *FieldInfo

	for _, fi := range q.modelInfo.Fields {
		if fi.AutoIncrement {
			autoIncrement = fi
			continue
		}

		data[fi.Column] = q.value(fi)
	}

	db, err := q.modelDB()
	if err != nil {
		return err
	}

	if autoIncrement != nil {
		var id int64

		if db.Driver() == database.DriverPostgres {
			row := db.Insert(q.modelInfo.Table, data).
				Returning(autoIncrement.Column).
				Row()

			err := row.Scan(&id)
			if db.Helper().IsDuplicateKey(err) {
				return ErrDuplicateKey{err}
			}
			if err != nil {
				return err
			}
		} else {
			res, err := db.Insert(q.modelInfo.Table, data).
				Exec()
			if db.Helper().IsDuplicateKey(err) {
				return ErrDuplicateKey{err}
			}
			if err != nil {
				return err
			}

			id, err = res.LastInsertId()
			if err != nil {
				return err
			}
		}

		field := fieldByIndex(q.modelValue, autoIncrement.FieldIndex...)
		value := reflect.ValueOf(id).Convert(field.Type())
		field.Set(value)
	} else {
		_, err := db.Insert(q.modelInfo.Table, data).
			Exec()
		if db.Helper().IsDuplicateKey(err) {
			return ErrDuplicateKey{err}
		}
		if err != nil {
			return err
		}
	}

	q.cacheClear()

	return nil
}

func (q *Query) Save(model interface{}, columns ...string) error {
	err := q.setModel(model)
	if err != nil {
		return err
	}

	conditions := q.primaryKeyConditions()
	if len(conditions) == 0 {
		return q.Create(model)
	}

	q.columns = columns

	data := make(query.Data)
	for _, fi := range q.modelInfo.Fields {
		if fi.Primary {
			continue
		}

		column := fi.Column
		if len(q.columns) > 0 {
			if q.columns.IndexOf(column) < 0 {
				continue
			}
		} else if fi.Skip {
			continue
		}

		data[column] = q.value(fi)
	}
	if len(data) == 0 {
		return nil
	}

	db, err := q.modelDB()
	if err != nil {
		return err
	}

	_, err = db.Update(q.modelInfo.Table, data).
		Where(conditions).
		Exec()
	if db.Helper().IsDuplicateKey(err) {
		return ErrDuplicateKey{err}
	}
	if err != nil {
		return err
	}

	q.cacheClear()

	return nil
}

func (q *Query) Delete(model interface{}) error {
	err := q.setModel(model)
	if err != nil {
		return err
	}

	conditions := q.primaryKeyConditions()
	if len(conditions) == 0 {
		return ErrNoPrimaryKey
	}

	db, err := q.modelDB()
	if err != nil {
		return err
	}

	_, err = db.Delete(q.modelInfo.Table).
		Where(conditions...).
		Exec()
	if err != nil {
		return err
	}

	q.hasManyCacheClear()
	q.cacheClear()

	return nil
}

func (q *Query) DeleteAll(model interface{}) error {
	err := q.setModel(model)
	if err != nil {
		return err
	}

	db, err := q.modelDB()
	if err != nil {
		return err
	}

	_, err = db.Delete(q.modelInfo.Table).
		Where(q.conditions...).
		Order(q.order...).
		Limit(q.limit...).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func (q *Query) DeleteFromCache(model interface{}) error {
	err := q.setModel(model)
	if err != nil {
		return err
	}

	q.hasManyCacheClear()
	return q.cacheClear()
}

func (q *Query) setModel(model interface{}) error {
	q.model = model

	q.modelValue = reflect.ValueOf(q.model)
	for q.modelValue.Kind() == reflect.Ptr {
		q.modelValue = q.modelValue.Elem()
	}

	q.modelInfo = GetModelInfo(q.model)
	if q.modelInfo == nil {
		return ErrInvalidModel
	}

	return nil
}

func (q *Query) setModelByType(modelType reflect.Type) error {
	q.modelValue = q.newValue(modelType).Elem()

	q.model = q.modelValue.Interface()

	q.modelInfo = GetModelInfo(q.model)
	if q.modelInfo == nil {
		return ErrInvalidModel
	}

	return nil
}

func (q *Query) clearModel() {
	for i := 0; i < q.modelValue.NumField(); i++ {
		f := q.modelValue.Field(i)
		if f.IsZero() {
			continue
		}

		f.Set(reflect.Zero(f.Type()))
	}
}

func (q *Query) modelDB() (interfaces.DB, error) {
	if q.tx != nil {
		return q.tx, nil
	}

	db, err := database.Get(q.modelInfo.Database)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (q *Query) modelName() string {
	return q.modelNameByType(q.modelInfo.Type)
}

func (q *Query) modelNameByType(t reflect.Type) string {
	return t.PkgPath() + "." + t.Name()
}

func (q *Query) modelPrimaryKey() []interface{} {
	a := make([]interface{}, len(q.modelInfo.Primary))
	for i, fi := range q.modelInfo.Primary {
		v := fieldByIndex(q.modelValue, fi.FieldIndex...).Interface()
		if fi.AutoIncrement && !typeconv.Bool(v) {
			return nil
		}

		a[i] = v
	}

	return a
}

func (q *Query) parseFields() *fieldsData {
	n := len(q.modelInfo.Fields)

	fields := make([]reflect.Value, 0, n)
	columns := make([]interface{}, 0, n)
	values := make([]interface{}, 0, n)
	info := make([]*FieldInfo, 0, n)

	for _, fi := range q.modelInfo.Fields {
		if len(q.columns) > 0 {
			if q.columns.IndexOf(fi.Column) < 0 {
				continue
			}
		} else if fi.Skip {
			continue
		}

		columns = append(columns, fi.Column)

		f := fieldByIndex(q.modelValue, fi.FieldIndex...)
		fields = append(fields, f)

		v := f.Addr().Interface()
		values = append(values, v)

		info = append(info, fi)
	}

	return &fieldsData{
		fields,
		columns,
		values,
		info,
	}
}

func (q *Query) scan(row interfaces.Row, fd *fieldsData) error {
	err := row.Scan(fd.values...)
	if err != nil {
		return err
	}

	for i, f := range fd.fields {
		rv := reflect.ValueOf(fd.values[i]).Elem()

		if fd.info[i].Base64 {
			q.decodeBase64(rv)
		}

		f.Set(rv)
	}

	return nil
}

func (q *Query) value(fi *FieldInfo) interface{} {
	f := fieldByIndex(q.modelValue, fi.FieldIndex...)
	v := f.Interface()

	if fi.Base64 {
		v = q.encodeBase64(f)
	}

	return v
}

func (q *Query) primaryKeyConditions() query.And {
	pk := q.modelPrimaryKey()
	if len(pk) == 0 {
		return nil
	}

	a := make(query.And, len(q.modelInfo.Primary))
	for i, fi := range q.modelInfo.Primary {
		a[i] = query.Eq{fi.Column: pk[i]}
	}

	return a
}

func (q *Query) primaryKeyFromConditions(conditions []interface{}) []interface{} {
	m := q.conditionsMap(conditions)
	if len(m) < len(q.modelInfo.Primary) {
		return nil
	}

	a := make([]interface{}, len(q.modelInfo.Primary))
	for i, fi := range q.modelInfo.Primary {
		v, ok := m[fi.Column]
		if !ok {
			return nil
		}
		a[i] = v
	}

	return a
}

func (q *Query) relationConditions(relation *Relation) query.And {
	a := make(query.And, len(relation.Key))
	for i, fi := range relation.Key {
		value := fieldByIndex(q.modelValue, relation.Reference[i].FieldIndex...).Interface()
		a[i] = query.Eq{
			fi.Column: value,
		}
	}

	return a
}

func (q *Query) conditionsMap(conditions []interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for _, cond := range conditions {
		switch t := cond.(type) {
		case query.And:
			tm := q.conditionsMap(t)
			for k, v := range tm {
				m[k] = v
			}
		case query.Eq:
			for k, v := range t {
				m[k] = v
			}
		}
	}

	return m
}

func (q *Query) cacheClear() error {
	q.belongsToCacheClear()

	pk := q.modelPrimaryKey()

	err := q.cacheDelete(pk)
	if err != nil {
		return err
	}

	return nil
}

func (q *Query) cacheKey(pk []interface{}) string {
	if len(pk) == 0 {
		return ""
	}

	b, err := json.Marshal(pk)
	if err != nil {
		return ""
	}

	key := fmt.Sprintf("%s<%016x>%s", q.modelName(), q.modelInfo.Checksum, b)
	return key
}

func (q *Query) cacheGet(pk []interface{}) error {
	key := q.cacheKey(pk)
	if key == "" {
		return ErrNoPrimaryKey
	}

	err := cacheStorage.Get(key, q.model)
	if err != nil {
		return err
	}

	err = q.loadRelated()
	if err != nil {
		return err
	}

	return nil
}

func (q *Query) cacheSet(pk []interface{}) error {
	key := q.cacheKey(pk)
	if key == "" {
		return ErrNoPrimaryKey
	}

	err := cacheStorage.Set(key, q.modelValue.Interface(), cache.DefaultTTL)
	return err
}

func (q *Query) cacheSetEmpty(pk []interface{}) error {
	key := q.cacheKey(pk)
	if key == "" {
		return ErrNoPrimaryKey
	}

	err := cacheStorage.Set(key, nil, cache.DefaultTTL)
	return err
}

func (q *Query) cacheDelete(pk []interface{}) error {
	key := q.cacheKey(pk)
	if key == "" {
		return ErrNoPrimaryKey
	}

	err := cacheStorage.Delete(key)
	return err
}

func (q *Query) parseRelationName(r string) queryRelation {
	m := relationNameRegexp.FindStringSubmatch(r)
	if m == nil {
		panic(fmt.Sprintf(`invalid relation name: "%s"`, r))
	}

	var with []string
	if m[3] != "" {
		with = strings.Split(m[3], ",")
		for i, w := range with {
			with[i] = strings.TrimSpace(w)
		}
	}

	return queryRelation{
		name: m[1],
		with: with,
	}
}

func (q *Query) loadRelated() error {
	if len(q.relations) == 0 {
		return nil
	}

	for _, relation := range q.relations {
		var err error

		if rel, ok := q.modelInfo.BelongsTo[relation.name]; ok {
			err = q.loadBelongsTo(rel, relation.with...)
		} else if rel, ok := q.modelInfo.HasMany[relation.name]; ok {
			err = q.loadHasMany(rel, relation.with...)
		} else {
			err = fmt.Errorf("unknown relation: %q", relation)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (q *Query) relationFieldName(relation *Relation) string {
	return q.modelInfo.Type.FieldByIndex(relation.FieldIndex).Name
}

func (q *Query) loadBelongsTo(relation *Relation, with ...string) error {
	field := q.modelValue.FieldByIndex(relation.FieldIndex)
	t := field.Type()
	v := q.newValue(t)
	conditions := q.relationConditions(relation)

	cached := q.belongsToCacheGet(relation, conditions)
	if cached != nil {
		v = *cached
	} else {
		err := Where(conditions...).First(v.Interface())
		if err != nil {
			return err
		}

		if t.Kind() != reflect.Ptr {
			v = v.Elem()
		}

		q.belongsToCacheSet(relation, conditions, v)
	}

	if len(with) > 0 {
		err := LoadRelated(v.Interface(), with...)
		if err != nil {
			return err
		}
	}

	field.Set(v)
	return nil
}

func (q *Query) belongsToCacheClear() {
	for _, relation := range q.modelInfo.BelongsTo {
		q.belongsToCacheClearRelation(relation)
	}
}

func (q *Query) belongsToCacheClearRelation(belongsTo *Relation) {
	f := q.modelValue.FieldByIndex(belongsTo.FieldIndex)

	t := f.Type()
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.New(t)
	mi := GetModelInfo(v.Interface())

	for _, hasMany := range mi.HasMany {
		if hasMany.Database != q.modelInfo.Database {
			continue
		}
		if hasMany.Table != q.modelInfo.Table {
			continue
		}

		for i, fi := range belongsTo.Key {
			k := fieldByIndex(q.modelValue, belongsTo.Reference[i].FieldIndex...)
			fieldByIndex(v.Elem(), fi.FieldIndex...).Set(k)
		}

		q := new(Query)
		err := q.setModel(v.Elem().Interface())
		if err != nil {
			continue
		}

		q.hasManyCacheClearRelation(hasMany)
	}
}

func (q *Query) belongsToCacheKey(conditions query.And) string {
	m := q.conditionsMap(conditions)

	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	key := string(b)
	return key
}

func (q *Query) belongsToCacheGet(relation *Relation, conditions query.And) *reflect.Value {
	cacheKey := q.belongsToCacheKey(conditions)
	if cacheKey == "" {
		return nil
	}

	fieldName := q.relationFieldName(relation)

	value, ok := q.belongsToCache[fieldName][cacheKey]
	if !ok {
		return nil
	}

	return &value
}

func (q *Query) belongsToCacheSet(relation *Relation, conditions query.And, value reflect.Value) {
	cacheKey := q.belongsToCacheKey(conditions)
	if cacheKey == "" {
		return
	}

	fieldName := q.relationFieldName(relation)

	if q.belongsToCache == nil {
		q.belongsToCache = map[string]map[string]reflect.Value{}
	}
	if _, ok := q.belongsToCache[fieldName]; !ok {
		q.belongsToCache[fieldName] = map[string]reflect.Value{}
	}

	q.belongsToCache[fieldName][cacheKey] = value
}

func (q *Query) loadHasMany(relation *Relation, with ...string) error {
	field := q.modelValue.FieldByIndex(relation.FieldIndex)
	t := field.Type()
	a := reflect.New(t)
	a.Elem().Set(
		reflect.MakeSlice(t, 0, 0),
	)
	models := a.Interface()
	conditions := q.relationConditions(relation)

	cached := false
	cacheKey := q.hasManyCacheKey(relation)
	if cacheKey != "" {
		err := cacheStorage.Get(cacheKey, models)
		if err == nil {
			cached = true
		}
	}

	if !cached {
		q := Where(conditions...).Order(relation.Order...).Limit(relation.Limit)
		if relation.Filter != nil {
			q.Where(relation.Filter)
		}

		err := q.Find(models)
		if err != nil {
			return err
		}

		cacheStorage.Set(cacheKey, models, cache.DefaultTTL)
	}

	if len(with) > 0 {
		n := a.Elem().Len()
		for i := 0; i < n; i++ {
			err := LoadRelated(a.Elem().Index(i).Interface(), with...)
			if err != nil {
				return err
			}
		}
	}

	field.Set(a.Elem())
	return nil
}

func (q *Query) hasManyCacheClear() {
	for _, relation := range q.modelInfo.HasMany {
		q.hasManyCacheClearRelation(relation)
	}
}

func (q *Query) hasManyCacheClearRelation(relation *Relation) {
	cacheKey := q.hasManyCacheKey(relation)
	if cacheKey == "" {
		return
	}

	cacheStorage.Delete(cacheKey)
}

func (q *Query) hasManyCacheKey(relation *Relation) string {
	pk := q.modelPrimaryKey()
	key := q.cacheKey(pk) + "." + q.modelInfo.Type.FieldByIndex(relation.FieldIndex).Name
	return key
}

func (q *Query) newValue(t reflect.Type) reflect.Value {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t)
}

func (q *Query) copyStruct(dst reflect.Value, src reflect.Value) {
	n := dst.NumField()
	for i := 0; i < n; i++ {
		f := dst.Field(i)
		if f.CanSet() {
			f.Set(src.Field(i))
		}
	}
}

func (q *Query) encodeBase64(rv reflect.Value) string {
	if !q.isSliceOfUint8(rv.Type()) {
		return ""
	}

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	s := base64.StdEncoding.EncodeToString(rv.Bytes())
	return s
}

func (q *Query) decodeBase64(rv reflect.Value) {
	if !q.isSliceOfUint8(rv.Type()) {
		return
	}

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	b, err := base64.StdEncoding.DecodeString(string(rv.Bytes()))
	if err != nil {
		return
	}

	rv.SetBytes(b)
}

func (q *Query) isSliceOfUint8(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return (t.Kind() == reflect.Slice) && (t.Elem().Kind() == reflect.Uint8)
}
