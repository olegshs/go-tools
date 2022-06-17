// Пакет query реализует конструктор SQL запросов.
package query

import (
	"sync"

	"github.com/olegshs/go-tools/database/interfaces"
)

type Query struct {
	db     interfaces.DB
	helper interfaces.Helper

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

	data interface{}

	query string
	args  []interface{}
	alias string

	changed bool
	mutex   sync.Mutex
}

func New(db interfaces.DB, helper interfaces.Helper) interfaces.Query {
	q := new(Query)
	q.db = db
	q.helper = helper
	return q
}

func (q *Query) Select(columns ...interface{}) interfaces.Query {
	q.statement = "SELECT"
	q.columns = columns
	q.changed = true
	return q
}

func (q *Query) From(tables ...interface{}) interfaces.Query {
	q.tables = tables
	q.changed = true
	return q
}

func (q *Query) Insert(table string, data interface{}) interfaces.Query {
	q.statement = "INSERT"
	q.table = table
	q.data = data
	q.changed = true
	return q
}

func (q *Query) Update(table string, data interface{}) interfaces.Query {
	q.statement = "UPDATE"
	q.table = table
	q.data = data
	q.changed = true
	return q
}

func (q *Query) Delete(table string) interfaces.Query {
	q.statement = "DELETE"
	q.table = table
	q.changed = true
	return q
}

func (q *Query) Join(joinType string, table string, conditions ...interface{}) interfaces.Query {
	q.joins = append(q.joins, join{
		joinType,
		table,
		conditions,
	})
	q.changed = true
	return q
}

func (q *Query) InnerJoin(table string, conditions ...interface{}) interfaces.Query {
	q.Join("INNER", table, conditions...)
	return q
}

func (q *Query) LeftJoin(table string, conditions ...interface{}) interfaces.Query {
	q.Join("LEFT", table, conditions...)
	return q
}

func (q *Query) RightJoin(table string, conditions ...interface{}) interfaces.Query {
	q.Join("RIGHT", table, conditions...)
	return q
}

func (q *Query) Where(conditions ...interface{}) interfaces.Query {
	if len(conditions) > 0 {
		q.where = append(q.where, conditions...)
		q.changed = true
	}
	return q
}

func (q *Query) Group(columns ...interface{}) interfaces.Query {
	if len(columns) > 0 {
		q.group = append(q.group, columns...)
		q.changed = true
	}
	return q
}

func (q *Query) Having(conditions ...interface{}) interfaces.Query {
	if len(conditions) > 0 {
		q.having = append(q.having, conditions...)
		q.changed = true
	}
	return q
}

func (q *Query) Order(order ...interface{}) interfaces.Query {
	if len(order) > 0 {
		q.order = append(q.order, order...)
		q.changed = true
	}
	return q
}

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

func (q *Query) Returning(columns ...interface{}) interfaces.Query {
	if len(columns) > 0 {
		q.returning = columns
		q.changed = true
	}
	return q
}

func (q *Query) As(alias string) interfaces.Query {
	q.alias = alias
	return q
}

func (q *Query) String() string {
	q.build()
	return q.query
}

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

func (q *Query) Exec() (interfaces.Result, error) {
	return q.db.Exec(q.String(), q.Args()...)
}

func (q *Query) Rows() (interfaces.Rows, error) {
	return q.db.Query(q.String(), q.Args()...)
}

func (q *Query) Row() interfaces.Row {
	return q.db.QueryRow(q.String(), q.Args()...)
}
