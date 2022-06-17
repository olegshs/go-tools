package query

import (
	"regexp"
	"sort"

	"github.com/olegshs/go-tools/helpers"
)

var (
	exprPlaceholderRegexp = regexp.MustCompile(`(^|.)(\$(\d+))`)
)

type (
	Data map[string]interface{}

	Column string

	And []interface{}
	Or  []interface{}
	Not []interface{}

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

	Between map[string][2]interface{}

	Condition struct {
		Column   string
		Operator string
		Value    interface{}
	}

	CompositeCondition struct {
		Columns  []string
		Operator string
		Values   []interface{}
	}

	Order [2]string
	Desc  string

	join struct {
		Type       string
		Table      string
		Conditions []interface{}
	}
)

func (m Data) SortedKeys() []string {
	return helpers.Map[string, interface{}](m).SortedKeys()
}

func (m Between) SortedKeys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}
