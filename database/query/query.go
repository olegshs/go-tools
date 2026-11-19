// Package query builds SQL statements with a fluent API:
//
//	q := db.Select("p.id", "p.title", query.Expr("COUNT(c.id)")).
//		From("posts p").
//		LeftJoin("comments c", "c.post_id = p.id").
//		Where(query.Eq{"p.status": 1}, query.Gt{"p.created": since}).
//		Group("p.id").
//		Order(query.Desc("p.created")).
//		Limit(20)
//	rows, err := q.Rows()
//
// Values are always passed as arguments, never inlined. Names given as
// strings are quoted with the driver's EscapeName, so "posts p" becomes a
// table with an alias and "p.id" a qualified column.
//
// Conditions are combined with AND. A condition can be:
//
//   - a map type such as Eq, Ne, Lt, In or Like, with column names as keys;
//     several keys are combined with AND. A nil value becomes IS NULL, a
//     bool becomes IS TRUE or IS FALSE, a []interface{} becomes a list,
//     a Column compares with another column, and a *Query a subquery;
//   - And, Or and Not, which group other conditions;
//   - Between, Condition, CompositeCondition or Expression;
//   - a raw SQL string, inserted as is.
//
// Raw strings and Expression SQL are not escaped, so never build them from
// user input.
package query

import (
	"sync"

	"github.com/olegshs/go-tools/database/interfaces"
)

// Query is a query under construction. It implements interfaces.Query.
// Methods change the query in place and return it for chaining. The SQL is
// built lazily, and rebuilt only after a change.
type Query struct {
	db     interfaces.DB
	helper interfaces.Helper

	// statement is SELECT, INSERT, UPDATE or DELETE.
	statement string
	table     string
	tables    []interface{}
	columns   []interface{}
	joins     []join
	where     []interface{}
	group     []interface{}
	having    []interface{}
	order     []interface{}
	offset    int
	limit     int
	returning []interface{}

	// data is the INSERT or UPDATE data: a Data map, or a *Query for
	// INSERT ... SELECT.
	data interface{}

	// query and args cache the built SQL until the next change.
	query string
	args  []interface{}
	// alias is used when the query is a subquery in FROM or a column list.
	alias string

	changed bool
	mutex   sync.Mutex
}

// New creates an empty query that runs on db and uses helper for the SQL
// dialect. db may be nil if the query is only built and not run.
func New(db interfaces.DB, helper interfaces.Helper) interfaces.Query {
	q := new(Query)
	q.db = db
	q.helper = helper
	return q
}

// Select makes q a SELECT of the given columns, or of * if there are none.
// A column can be a string or Column (a quoted name), an Expression or a
// *Query (a subquery).
func (q *Query) Select(columns ...interface{}) interfaces.Query {
	q.statement = "SELECT"
	q.columns = columns
	q.changed = true
	return q
}

// From sets the tables of a SELECT. A table can be a string (a quoted name
// with an optional alias), an Expression or a *Query with an alias set
// by As.
func (q *Query) From(tables ...interface{}) interfaces.Query {
	q.tables = tables
	q.changed = true
	return q
}

// Insert makes q an INSERT into table. data is a Data map of column values,
// which may be Expressions, or a *Query for INSERT ... SELECT.
func (q *Query) Insert(table string, data interface{}) interfaces.Query {
	q.statement = "INSERT"
	q.table = table
	q.data = data
	q.changed = true
	return q
}

// Update makes q an UPDATE of table. data is a Data map of column values,
// which may be Expressions.
func (q *Query) Update(table string, data interface{}) interfaces.Query {
	q.statement = "UPDATE"
	q.table = table
	q.data = data
	q.changed = true
	return q
}

// Delete makes q a DELETE from table.
func (q *Query) Delete(table string) interfaces.Query {
	q.statement = "DELETE"
	q.table = table
	q.changed = true
	return q
}

// Join adds a join of the given type (e.g. "INNER") with the join
// conditions. Each call adds one more join.
func (q *Query) Join(joinType string, table string, conditions ...interface{}) interfaces.Query {
	q.joins = append(q.joins, join{
		joinType,
		table,
		conditions,
	})
	q.changed = true
	return q
}

// InnerJoin adds an INNER JOIN.
func (q *Query) InnerJoin(table string, conditions ...interface{}) interfaces.Query {
	q.Join("INNER", table, conditions...)
	return q
}

// LeftJoin adds a LEFT JOIN.
func (q *Query) LeftJoin(table string, conditions ...interface{}) interfaces.Query {
	q.Join("LEFT", table, conditions...)
	return q
}

// RightJoin adds a RIGHT JOIN.
func (q *Query) RightJoin(table string, conditions ...interface{}) interfaces.Query {
	q.Join("RIGHT", table, conditions...)
	return q
}

// Where adds WHERE conditions. Conditions from all calls are combined with
// AND.
func (q *Query) Where(conditions ...interface{}) interfaces.Query {
	if len(conditions) > 0 {
		q.where = append(q.where, conditions...)
		q.changed = true
	}
	return q
}

// Group adds GROUP BY columns.
func (q *Query) Group(columns ...interface{}) interfaces.Query {
	if len(columns) > 0 {
		q.group = append(q.group, columns...)
		q.changed = true
	}
	return q
}

// Having adds HAVING conditions, combined with AND.
func (q *Query) Having(conditions ...interface{}) interfaces.Query {
	if len(conditions) > 0 {
		q.having = append(q.having, conditions...)
		q.changed = true
	}
	return q
}

// Order adds ORDER BY columns. An item can be a column name, Desc,
// Order{"column", "asc|desc"} or an Expression.
func (q *Query) Order(order ...interface{}) interfaces.Query {
	if len(order) > 0 {
		q.order = append(q.order, order...)
		q.changed = true
	}
	return q
}

// Limit sets the row limit: Limit(n) or Limit(offset, n). Zero means no
// limit. OFFSET is only used with SELECT.
func (q *Query) Limit(limit ...int) interfaces.Query {
	n := len(limit)
	if n > 1 {
		q.offset = limit[0]
		q.limit = limit[1]
		q.changed = true
	} else if n > 0 {
		q.offset = 0
		q.limit = limit[0]
		q.changed = true
	}
	return q
}

// Returning adds a RETURNING clause to an INSERT, for databases that
// support it.
func (q *Query) Returning(columns ...interface{}) interfaces.Query {
	if len(columns) > 0 {
		q.returning = columns
		q.changed = true
	}
	return q
}

// As sets the alias used when q is a subquery.
func (q *Query) As(alias string) interfaces.Query {
	q.alias = alias
	return q
}

// String returns the SQL of the query.
func (q *Query) String() string {
	q.build()
	return q.query
}

// Args returns the argument values for the placeholders in String.
func (q *Query) Args() []interface{} {
	q.build()
	return q.args
}

func (q *Query) build() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.changed {
		b := new(builder)
		b.query = q

		q.query, q.args = b.build()
		q.changed = false
	}
}

// Exec runs the query as a statement.
func (q *Query) Exec() (interfaces.Result, error) {
	return q.db.Exec(q.String(), q.Args()...)
}

// Rows runs the query and returns its rows, which the caller must close.
func (q *Query) Rows() (interfaces.Rows, error) {
	return q.db.Query(q.String(), q.Args()...)
}

// Row runs the query and returns its first row. Scan must be called on the
// result.
func (q *Query) Row() interfaces.Row {
	return q.db.QueryRow(q.String(), q.Args()...)
}
