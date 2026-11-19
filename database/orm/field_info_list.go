package orm

// FieldInfoList is a list of fields.
type FieldInfoList []*FieldInfo

// ByColumn returns the field for column, or nil.
func (list FieldInfoList) ByColumn(column string) *FieldInfo {
	for _, fi := range list {
		if fi.Column == column {
			return fi
		}
	}
	return nil
}
