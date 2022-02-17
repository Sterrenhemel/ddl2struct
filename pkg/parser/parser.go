package parser

import (
	"fmt"
	"go/format"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/types"
)

type DDLParser struct {
	Table map[string]Columns
	Index map[string]Indexes
	err   error
	p     *parser.Parser
}

func (parser *DDLParser) Parse(sql string) error {
	nodes, _, err := parser.p.Parse(sql, "", "")
	if err != nil {
		return errors.Wrap(err, "sql parsing error")
	}
	parser.Table = make(map[string]Columns)
	parser.Index = make(map[string]Indexes)

	for _, node := range nodes {
		node.Accept(parser)
		if parser.err != nil {
			return errors.Wrap(err, "sql parsing error")
		}
	}

	return nil
}

func (parser DDLParser) ToStructs(withTag bool) ([]byte, error) {
	var builder strings.Builder
	for tableName, columns := range parser.Table {
		builder.WriteString(fmt.Sprintf("type %s struct { %s }\n\n", strcase.ToCamel(tableName), columns.ToStructFields(withTag)))
	}
	s := builder.String()
	return format.Source([]byte(s))
}

func (parser *DDLParser) Enter(n ast.Node) (node ast.Node, skipChildren bool) {
	switch n := n.(type) {
	case *ast.CreateTableStmt:
		parser.err = parser.parseCreateTableStmt(n)
	case *ast.CreateIndexStmt:
		parser.err = parser.parseCreateIndexStmt(n)
	}
	return n, true
}

func (parser *DDLParser) Leave(n ast.Node) (node ast.Node, ok bool) {
	return n, true
}

func (parser *DDLParser) parseCreateTableStmt(stmt *ast.CreateTableStmt) error {
	tableName := stmt.Table.Name.String()
	if _, ok := parser.Table[tableName]; ok {
		return errors.Errorf("duplicate table name :%s", tableName)
	} else {
		for _, col := range stmt.Cols {
			parser.Table[tableName] = append(parser.Table[tableName], Column{
				Name: col.Name.Name.String(),
				Type: parser.getColumnType(col.Tp.EvalType()),
			})
		}
	}
	return nil
}

func (parser *DDLParser) parseCreateIndexStmt(stmt *ast.CreateIndexStmt) error {
	return nil
}

func (parser *DDLParser) getColumnType(typ types.EvalType) string {
	switch typ {
	case types.ETInt:
		return "int"
	case types.ETReal, types.ETDecimal:
		return "float64"
	case types.ETDatetime, types.ETTimestamp:
		return "time.Time"
	case types.ETString:
		return "string"
	default:
		return "string"
	}
}

func New() *DDLParser {
	return &DDLParser{
		p: parser.New(),
	}
}

type Columns []Column

func (columns Columns) ToStructFields(withTag bool) string {
	fields := make([]string, 0)
	for _, column := range columns {
		fields = append(fields, column.ToStructField(withTag))
	}
	return strings.Join(fields, "\n")
}

type Column struct {
	Name string
	Type string
}

func (column Column) ToStructField(withTag bool) string {
	var tag string
	if withTag {
		tag = fmt.Sprintf("`json:\"%s\" gorm:\"column:%s\"`", strcase.ToSnake(column.Name), column.Name)
	}
	return fmt.Sprintf("%s %s", strcase.ToCamel(column.Name), column.Type) + tag
}

type Indexes []Index

type Index struct {
}
