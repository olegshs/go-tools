package query

import (
	"regexp"
	"sort"

	"github.com/olegshs/go-tools/helpers"
)

var (
	// exprPlaceholderRegexp matches $N placeholders in Expression SQL,
	// together with an optional preceding backslash that marks a literal $N.
	exprPlaceholderRegexp = regexp.MustCompile(`\\?\$\d+`)
)

type (
	// Data maps column names to values for INSERT and UPDATE.
	Data map[string]interface{}

	// Column is a column name. As a condition value, it compares with that
	// column instead of a literal value.
	Column string

	// And, Or and Not group conditions. Not negates the AND of its
	// conditions: NOT (a AND b).
	And []interface{}
	Or  []interface{}
	Not []interface{}

	// Comparison conditions, mapping column names to values. In and NotIn
	// take a []interface{} or a *Query.
	Eq      map[string]interface{}
	Ne      map[string]interface{}
	Like    map[string]interface{}
	NotLike map[string]interface{}
	Lt      map[string]interface{}
	Lte     map[string]interface{}
	Gt      map[string]interface{}
	Gte     map[string]interface{}
	In      map[string]interface{}
	NotIn   map[string]interface{}

	// Between maps column names to [low, high] pairs.
	Between map[string][2]interface{}

	// Condition compares a column with a value using any operator, such as
	// Condition{"age", ">=", 18}.
	Condition struct {
		Column   string
		Operator string
		Value    interface{}
	}

	// CompositeCondition compares several columns as a tuple, such as
	// (a, b) IN ((1, 2), (3, 4)).
	CompositeCondition struct {
		Columns  []string
		Operator string
		Values   []interface{}
	}

	// Order is an ORDER BY item: {column, "ASC" or "DESC"}.
	Order [2]string
	// Desc is a column sorted in descending order.
	Desc string

	// join is a JOIN clause.
	join struct {
		Type       string
		Table      string
		Conditions []interface{}
	}
)

// SortedKeys returns the column names in sorted order, so the generated SQL
// is deterministic.
func (m Data) SortedKeys() []string {
	return helpers.Map[string, interface{}](m).SortedKeys()
}

// SortedKeys returns the column names in sorted order.
func (m Between) SortedKeys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}
