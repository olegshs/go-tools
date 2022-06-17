package orm

import (
	"github.com/olegshs/go-tools/database/interfaces"
)

func Tx(tx interfaces.DB) *Query {
	return new(Query).Tx(tx)
}

func Select(columns ...string) *Query {
	return new(Query).Select(columns...)
}

func With(relations ...string) *Query {
	return new(Query).With(relations...)
}

func Where(conditions ...interface{}) *Query {
	return new(Query).Where(conditions...)
}

func Order(order ...interface{}) *Query {
	return new(Query).Order(order...)
}

func Limit(limit ...int) *Query {
	return new(Query).Limit(limit...)
}

func Count(model interface{}) (int, error) {
	return new(Query).Count(model)
}

func First(model interface{}) error {
	return new(Query).First(model)
}

func Find(models interface{}) error {
	return new(Query).Find(models)
}

func LoadRelated(model interface{}, relations ...string) error {
	return new(Query).LoadRelated(model, relations...)
}

func Create(model interface{}) error {
	return new(Query).Create(model)
}

func Save(model interface{}, columns ...string) error {
	return new(Query).Save(model, columns...)
}

func Delete(model interface{}) error {
	return new(Query).Delete(model)
}

func DeleteAll(model interface{}) error {
	return new(Query).DeleteAll(model)
}

func DeleteFromCache(model interface{}) error {
	return new(Query).DeleteFromCache(model)
}
