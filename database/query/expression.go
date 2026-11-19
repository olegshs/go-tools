package query

import (
	"strconv"
)

// Expression is a piece of raw SQL with arguments. Placeholders in Sql are
// written as $1, $2, ... and are replaced with the driver's placeholders.
// Use \$1 for a literal $1.
type Expression struct {
	Sql  string
	Args []interface{}
}

// Expr creates an Expression, such as Expr("COUNT(*)") or
// Expr("price * $1", rate).
func Expr(sql string, args ...interface{}) Expression {
	return Expression{
		Sql:  sql,
		Args: args,
	}
}

// build returns the SQL with each $N replaced by the placeholder returned
// by appendArg for the N-th argument, from left to right. \$N becomes a
// literal $N. Placeholders without a matching argument are left as they
// are.
func (e Expression) build(appendArg func(interface{}) string) string {
	n := len(e.Args)

	return exprPlaceholderRegexp.ReplaceAllStringFunc(e.Sql, func(m string) string {
		if m[0] == '\\' {
			return m[1:]
		}

		i, err := strconv.Atoi(m[1:])
		if err != nil || i < 1 || i > n {
			return m
		}

		return appendArg(e.Args[i-1])
	})
}
