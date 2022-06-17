package interfaces

type Query interface {
	Select(columns ...interface{}) Query
	From(tables ...interface{}) Query
	Insert(table string, data interface{}) Query
	Update(table string, data interface{}) Query
	Delete(table string) Query
	Join(joinType string, table string, conditions ...interface{}) Query
	InnerJoin(table string, conditions ...interface{}) Query
	LeftJoin(table string, conditions ...interface{}) Query
	RightJoin(table string, conditions ...interface{}) Query
	Where(conditions ...interface{}) Query
	Group(columns ...interface{}) Query
	Having(conditions ...interface{}) Query
	Order(order ...interface{}) Query
	Limit(limit ...int) Query
	Returning(columns ...interface{}) Query
	As(alias string) Query
	String() string
	Args() []interface{}
	Exec() (Result, error)
	Rows() (Rows, error)
	Row() Row
}
