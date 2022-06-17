package query

import (
	"strconv"
)

type Expression struct {
	Sql  string
	Args []interface{}
}

func Expr(sql string, args ...interface{}) Expression {
	return Expression{
		Sql:  sql,
		Args: args,
	}
}

func (e Expression) build(appendArg func(interface{}) string) string {
	s := e.Sql
	n := len(e.Args)

	indexes := exprPlaceholderRegexp.FindAllStringSubmatchIndex(e.Sql, -1)
	for _, index := range indexes {
		prefix := s[index[2]:index[3]]
		if prefix == "\\" {
			s = s[0:index[2]] + s[index[3]:]
			continue
		}

		i, _ := strconv.Atoi(s[index[6]:index[7]])
		i--
		if i >= n {
			continue
		}

		s = s[0:index[4]] + appendArg(e.Args[i]) + s[index[5]:]
	}

	return s
}
