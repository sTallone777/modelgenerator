package generate

type Field struct {
	Field    string `gorm:"column:field_name"`
	Type     string `gorm:"column:field_type"`
	Nullable string `gorm:"column:field_isnullable"`
	Key      string `gorm:"column:field_isprimarykey"`
	Extra    string `gorm:"column:field_isincrement"`
	Size     string `gorm:"column:field_size"`
	Decimal  string `gorm:"column:field_decimal"`
	Comment  string `gorm:"column:field_comment"`
}
