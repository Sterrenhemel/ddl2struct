package parser

type Table struct {
	TableName    string
	TableComment string
	Columns      Columns
}

type Columns []Column

//func (columns Columns) ToStructFields(withTag bool) string {
//	fields := make([]string, 0)
//	for _, column := range columns {
//		fields = append(fields, column.ToStructField(withTag))
//	}
//	return strings.Join(fields, "\n")
//}

type Column struct {
	Name       string
	Type       string
	Comment    string // 注释
	DefaultVal string
}

//func (column Column) ToStructField(withTag bool) string {
//	var tag string
//	if withTag {
//		tag = fmt.Sprintf("`json:\"%s\" gorm:\"column:%s\"`", strcase.ToSnake(column.Name), column.Name)
//	}
//	return fmt.Sprintf("%s %s", strcase.ToCamel(column.Name), column.Type) + tag
//}

type Indexes []Index

type Index struct {
}
