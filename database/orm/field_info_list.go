package orm

type FieldInfoList []*FieldInfo

func (list FieldInfoList) ByColumn(column string) *FieldInfo {
	for _, fi := range list {
		if fi.Column == column {
			return fi
		}
	}
	return nil
}
