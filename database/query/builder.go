package query

import (
	"strings"

	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/helpers/typeconv"
)

type builder struct {
	query      *Query
	args       []interface{}
	argsOffset int
}

func (b *builder) build() (string, []interface{}) {
	var s string
	b.args = make([]interface{}, 0)

	switch b.query.statement {
	case "SELECT":
		s = b.buildSelect()
	case "INSERT":
		s = b.buildInsert()
	case "UPDATE":
		s = b.buildUpdate()
	case "DELETE":
		s = b.buildDelete()
	}

	s = helpers.Trim(s)
	return s, b.args
}

func (b *builder) buildSelect() string {
	s := "SELECT "

	if len(b.query.columns) > 0 {
		s += b.buildColumns(b.query.columns...)
	} else {
		s += "*"
	}

	if len(b.query.tables) > 0 {
		s += "\nFROM " + b.buildFrom(b.query.tables...)
	}

	if len(b.query.joins) > 0 {
		s += "\n" + b.buildJoins(b.query.joins...)
	}

	if len(b.query.where) > 0 {
		s += "\nWHERE " + b.buildConditions(b.query.where...)
	}

	if len(b.query.group) > 0 {
		s += "\nGROUP BY " + b.buildColumns(b.query.group...)
	}

	if len(b.query.having) > 0 {
		s += "\nHAVING " + b.buildConditions(b.query.having...)
	}

	if len(b.query.order) > 0 {
		s += "\nORDER BY " + b.buildOrder(b.query.order...)
	}

	if b.query.limit > 0 {
		s += "\nLIMIT " + b.appendArg(b.query.limit)
		if b.query.offset > 0 {
			s += " OFFSET " + b.appendArg(b.query.offset)
		}
	}

	return s
}

func (b *builder) buildInsert() string {
	switch t := b.query.data.(type) {
	default:
		return b.buildInsertData(t.(Data))
	case *Query:
		return b.buildInsertSubQuery(t)
	}
}

func (b *builder) buildInsertData(data Data) string {
	keys := data.SortedKeys()
	values := make([]string, len(keys))

	for i, k := range keys {
		keys[i] = b.escapeName(k)
		v := data[k]

		switch t := v.(type) {
		case Expression:
			values[i] = b.buildExpr(t)
		default:
			values[i] = b.appendArg(t)
		}
	}

	s := "INSERT INTO " + b.escapeName(b.query.table) +
		"\n(" + strings.Join(keys, ", ") + ")" +
		"\nVALUES (" + strings.Join(values, ", ") + ")"

	if len(b.query.returning) > 0 {
		s += "\nRETURNING " + b.buildColumns(b.query.returning...)
	}

	return s
}

func (b *builder) buildInsertSubQuery(q *Query) string {
	s := "INSERT INTO " + b.escapeName(b.query.table) +
		"\n" + b.buildSubQuery(q)

	if len(b.query.returning) > 0 {
		s += "\nRETURNING " + b.buildColumns(b.query.returning...)
	}

	return s
}

func (b *builder) buildUpdate() string {
	data := b.query.data.(Data)
	keys := data.SortedKeys()
	a := make([]string, len(keys))

	for i, k := range keys {
		v := data[k]
		s := b.escapeName(k) + " = "

		switch t := v.(type) {
		case Expression:
			s += b.buildExpr(t)
		default:
			s += b.appendArg(t)
		}

		a[i] = s
	}

	s := "UPDATE " + b.escapeName(b.query.table) +
		"\nSET " + strings.Join(a, ", ")

	if len(b.query.where) > 0 {
		s += "\nWHERE " + b.buildConditions(b.query.where...)
	}

	if len(b.query.order) > 0 {
		s += "\nORDER BY " + b.buildOrder(b.query.order...)
	}

	if b.query.limit > 0 {
		s += "\nLIMIT " + b.appendArg(b.query.limit)
	}

	return s
}

func (b *builder) buildDelete() string {
	s := "DELETE FROM " + b.escapeName(b.query.table)

	if len(b.query.where) > 0 {
		s += "\nWHERE " + b.buildConditions(b.query.where...)
	}

	if len(b.query.order) > 0 {
		s += "\nORDER BY " + b.buildOrder(b.query.order...)
	}

	if b.query.limit > 0 {
		s += "\nLIMIT " + b.appendArg(b.query.limit)
	}

	return s
}

func (b *builder) buildColumns(columns ...interface{}) string {
	a := make([]string, len(columns))

	for i, v := range columns {
		var s string

		switch t := v.(type) {
		case Column:
			s = b.escapeName(string(t))
		case Expression:
			s = b.buildExpr(t)
		case *Query:
			s = b.buildSubQuery(t)
		case string:
			s = b.escapeName(t)
		default:
			s = b.escapeName(typeconv.String(t))
		}

		a[i] = s
	}

	s := strings.Join(a, ", ")
	return s
}

func (b *builder) buildFrom(tables ...interface{}) string {
	a := make([]string, len(tables))

	for i, table := range tables {
		var s string

		switch t := table.(type) {
		case Expression:
			s = b.buildExpr(t)
		case *Query:
			s = b.buildSubQuery(t)
		case string:
			s = b.escapeName(t)
		default:
			s = b.escapeName(typeconv.String(t))
		}

		a[i] = s
	}

	s := strings.Join(a, ", ")
	return s
}

func (b *builder) buildJoins(joins ...join) string {
	a := make([]string, len(joins))

	for i, j := range joins {
		a[i] = j.Type + " JOIN " + b.escapeName(j.Table) + " ON " + b.buildConditions(j.Conditions...)
	}

	s := strings.Join(a, "\n")
	return s
}

func (b *builder) buildConditions(conditions ...interface{}) string {
	a := make([]string, len(conditions))

	for i, v := range conditions {
		var s string

		switch t := v.(type) {
		case And:
			s = b.buildLogic("AND", t)
		case Or:
			s = b.buildLogic("OR", t)
		case Not:
			s = b.buildLogic("AND", t)
			s = "!(" + s + ")"
		case Expression:
			s = b.buildExpr(t)
		case map[string]interface{}:
			s = b.buildCondition("=", t)
		case Eq:
			s = b.buildCondition("=", t)
		case Ne:
			s = b.buildCondition("!=", t)
		case Like:
			s = b.buildCondition("LIKE", t)
		case NotLike:
			s = b.buildCondition("NOT LIKE", t)
		case Lt:
			s = b.buildCondition("<", t)
		case Lte:
			s = b.buildCondition("<=", t)
		case Gt:
			s = b.buildCondition(">", t)
		case Gte:
			s = b.buildCondition(">=", t)
		case In:
			s = b.buildCondition("IN", t)
		case NotIn:
			s = b.buildCondition("NOT IN", t)
		case Between:
			s = b.buildBetween(t)
		case Condition:
			s = b.escapeName(t.Column) + " " + t.Operator + " " + b.appendArgs(t.Value)
		case CompositeCondition:
			s = b.buildCompositeCondition(t)
		case string:
			s = t
		default:
			s = typeconv.String(t)
		}

		a[i] = s
	}

	s := b.joinWithParentheses(a, "AND")
	return s
}

func (b *builder) buildLogic(operator string, conditions []interface{}) string {
	a := make([]string, len(conditions))

	for i, v := range conditions {
		a[i] = b.buildConditions(v)
	}

	s := b.joinWithParentheses(a, operator)
	return s
}

func (b *builder) buildCondition(operator string, m map[string]interface{}) string {
	keys := helpers.Map[string, interface{}](m).SortedKeys()
	a := make([]string, len(keys))

	for i, k := range keys {
		v := m[k]
		op := operator
		var r string

		switch t := v.(type) {
		case Column:
			r = b.escapeName(string(t))
		case Expression:
			r = b.buildExpr(t)
		case *Query:
			r = b.buildSubQuery(t)
		case bool:
			op = b.eqToIs(op)
			if t {
				r = "TRUE"
			} else {
				r = "FALSE"
			}
		case nil:
			op = b.eqToIs(op)
			r = "NULL"
		default:
			r = b.appendArgs(t)
		}

		a[i] = b.escapeName(k) + " " + op + " " + r
	}

	s := b.joinWithParentheses(a, "AND")
	return s
}

func (b *builder) buildCompositeCondition(condition CompositeCondition) string {
	columns := make([]string, len(condition.Columns))
	for i, column := range condition.Columns {
		columns[i] = b.escapeName(column)
	}

	values := b.appendArgs(condition.Values)

	s := "(" + strings.Join(columns, ", ") + ") " + condition.Operator + " " + values
	return s
}

func (b *builder) buildBetween(between Between) string {
	keys := between.SortedKeys()
	a := make([]string, len(keys))

	for i, k := range keys {
		v := between[k]
		a[i] = b.escapeName(k) + " BETWEEN " + b.appendArg(v[0]) + " AND " + b.appendArg(v[1])
	}

	s := strings.Join(a, " AND ")
	return s
}

func (b *builder) buildOrder(order ...interface{}) string {
	a := make([]string, len(order))

	for i, v := range order {
		var s string

		switch t := v.(type) {
		case Column:
			s = b.escapeName(string(t))
		case Desc:
			s = b.escapeName(string(t)) + " DESC"
		case Expression:
			s = b.buildExpr(t)
		case Order:
			s = b.escapeName(t[0])

			switch strings.ToUpper(t[1]) {
			case "ASC":
				s += " ASC"
			case "DESC":
				s += " DESC"
			}
		case string:
			s = b.escapeName(t)
		default:
			s = b.escapeName(typeconv.String(t))
		}

		a[i] = s
	}

	s := strings.Join(a, ", ")
	return s
}

func (b *builder) buildExpr(e Expression) string {
	return e.build(b.appendArgs)
}

func (b *builder) buildSubQuery(q *Query) string {
	sub := new(builder)
	sub.query = q
	sub.argsOffset = len(b.args) + b.argsOffset

	s, a := sub.build()

	s = "(" + s + ")"
	if q.alias != "" {
		s += " AS " + b.escapeName(q.alias)
	}

	b.appendArgs(a)

	return s
}

func (b *builder) appendArgs(v interface{}) string {
	switch t := v.(type) {
	case []interface{}:
		a := make([]string, len(t))
		for i, v := range t {
			a[i] = b.appendArgs(v)
		}
		return "(" + strings.Join(a, ", ") + ")"
	case *Query:
		return b.buildSubQuery(t)
	default:
		return b.appendArg(t)
	}
}

func (b *builder) appendArg(v interface{}) string {
	b.args = append(b.args, v)

	n := len(b.args) + b.argsOffset
	placeholder := b.query.helper.ArgPlaceholder(n)

	return placeholder
}

func (b *builder) escapeName(s string) string {
	s = b.query.helper.EscapeName(s)
	return s
}

func (b *builder) eqToIs(operator string) string {
	if operator == "=" {
		return "IS"
	} else {
		return "IS NOT"
	}
}

func (b *builder) joinWithParentheses(a []string, sep string) string {
	if len(a) < 1 {
		return ""
	}
	if len(a) == 1 {
		return a[0]
	}

	sep = ") " + sep + " ("

	s := strings.Join(a, sep)
	s = "(" + s + ")"

	return s
}
