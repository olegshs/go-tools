package orm

import (
	"regexp"
	"strings"

	"github.com/olegshs/go-tools/database/query"
	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/helpers/typeconv"
)

var (
	fieldTagOrderRegexp = regexp.MustCompile(`(?i)^(.+?)(\s+(ASC|DESC))?$`)
)

type FieldInfo struct {
	FieldIndex    [][]int
	Database      string
	Table         string
	Column        string
	Primary       bool
	AutoIncrement bool
	ForeignKey    []string
	Filter        string
	Order         []interface{}
	Limit         int
	Base64        bool
	Skip          bool
}

func ParseFieldTag(s string) *FieldInfo {
	s = helpers.Trim(s)
	if s == "-" {
		return nil
	}

	fi := new(FieldInfo)

	props := fi.splitProperties(s, ";")
	for _, prop := range props {
		key, value := fi.splitKeyValue(prop)
		fi.setProperty(key, value)
	}

	return fi
}

func (fi *FieldInfo) setProperty(key, value string) {
	switch strings.ToLower(key) {
	case "database":
		fi.Database = value
	case "table":
		fi.Table = value
	case "column":
		fi.Column = value
	case "primary":
		fi.Primary = true
	case "auto_increment":
		fi.AutoIncrement = true
	case "foreign_key":
		fi.ForeignKey = fi.splitProperties(value, ",")
	case "filter":
		fi.Filter = value
	case "order":
		fi.Order = fi.parseOrder(fi.splitProperties(value, ","))
	case "limit":
		fi.Limit = typeconv.Int(value)
	case "base64":
		fi.Base64 = true
	case "skip":
		fi.Skip = true
	}
}

func (fi *FieldInfo) splitProperties(s string, sep string) []string {
	a := strings.Split(s, sep)
	props := make([]string, 0, len(a))

	for _, v := range a {
		v = strings.Trim(v, " \t")
		if v == "" {
			continue
		}

		props = append(props, v)
	}

	return props
}

func (fi *FieldInfo) splitKeyValue(s string) (string, string) {
	a := strings.Split(s, "=")

	key := strings.Trim(a[0], " \t")

	value := ""
	if len(a) > 1 {
		value = strings.Join(a[1:], "=")
		value = strings.Trim(value, " \t")
	}

	return key, value
}

func (fi *FieldInfo) parseOrder(a []string) []interface{} {
	order := make([]interface{}, 0, len(a))

	for _, s := range a {
		m := fieldTagOrderRegexp.FindStringSubmatch(s)
		if m == nil {
			continue
		}

		v := query.Order{m[1], strings.ToUpper(m[3])}
		order = append(order, v)
	}

	return order
}
