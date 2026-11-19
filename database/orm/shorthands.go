package orm

import (
	"github.com/olegshs/go-tools/database/interfaces"
)

// Tx starts a query that runs on tx. See Query.Tx.
func Tx(tx interfaces.DB) *Query {
	return new(Query).Tx(tx)
}

// Select starts a query that reads only the given columns.
func Select(columns ...string) *Query {
	return new(Query).Select(columns...)
}

// With starts a query that loads the named relations. See Query.With.
func With(relations ...string) *Query {
	return new(Query).With(relations...)
}

// Where starts a query with conditions.
func Where(conditions ...interface{}) *Query {
	return new(Query).Where(conditions...)
}

// Order starts a query with a sort order.
func Order(order ...interface{}) *Query {
	return new(Query).Order(order...)
}

// Limit starts a query with a row limit.
func Limit(limit ...int) *Query {
	return new(Query).Limit(limit...)
}

// Count returns the number of rows in the table of model.
func Count(model interface{}) (int, error) {
	return new(Query).Count(model)
}

// First reads model by its primary key. See Query.First.
func First(model interface{}) error {
	return new(Query).First(model)
}

// Find reads all rows of the model type into models. See Query.Find.
func Find(models interface{}) error {
	return new(Query).Find(models)
}

// LoadRelated loads relations of a model. See Query.LoadRelated.
func LoadRelated(model interface{}, relations ...string) error {
	return new(Query).LoadRelated(model, relations...)
}

// Create inserts model. See Query.Create.
func Create(model interface{}) error {
	return new(Query).Create(model)
}

// Save updates or inserts model. See Query.Save.
func Save(model interface{}, columns ...string) error {
	return new(Query).Save(model, columns...)
}

// Delete deletes model by its primary key. See Query.Delete.
func Delete(model interface{}) error {
	return new(Query).Delete(model)
}

// DeleteAll deletes every row in the table of model.
func DeleteAll(model interface{}) error {
	return new(Query).DeleteAll(model)
}

// DeleteFromCache removes model from the cache. See Query.DeleteFromCache.
func DeleteFromCache(model interface{}) error {
	return new(Query).DeleteFromCache(model)
}
