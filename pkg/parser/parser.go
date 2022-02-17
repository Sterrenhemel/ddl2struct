package parser

import (
	"fmt"
	"go/format"
	"path"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/types"
)

var (
	goFileRegex = regexp.MustCompile("([a-zA-Z0-9_]*\\.go)")
)

type DDLParser struct {
	Files       map[string]map[string]Columns // fileName -> TableName -> Columns
	Index       map[string]Indexes
	InputFile   string
	OutputFile  string
	IsDir       bool
	packageName string
	err         error
	p           *parser.Parser
}

func (parser *DDLParser) Parse(sql string) error {
	nodes, _, err := parser.p.Parse(sql, "", "")
	if err != nil {
		return errors.Wrap(err, "sql parsing error")
	}
	parser.Files = make(map[string]map[string]Columns)
	parser.Index = make(map[string]Indexes)

	for _, node := range nodes {
		node.Accept(parser)
		if parser.err != nil {
			return errors.Wrap(err, "sql parsing error")
		}
	}

	return nil
}

func (parser DDLParser) ToStructs(withTag bool) (fileContentMap map[string][]byte, err error) {
	fileContentMap = make(map[string][]byte)
	var builder strings.Builder
	for fileName, tables := range parser.Files {
		for tableName, columns := range tables {
			builder.WriteString(fmt.Sprintf("type %s struct { %s }\n\n", strcase.ToCamel(tableName), columns.ToStructFields(withTag)))
		}
		fileContentMap[fileName], err = format.Source([]byte(builder.String()))
		if err != nil {
			return
		}
	}
	return
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
	fileName := ""
	tableName := stmt.Table.Name.String()
	if !parser.IsDir {
		fileName = parser.OutputFile
	} else {
		for _, option := range stmt.Options {
			if option.Tp == ast.TableOptionComment {
				fileName = goFileRegex.FindString(option.StrValue)
				break
			}
		}
		if fileName == "" {
			fileName = path.Base(parser.InputFile)
			fileSuffix := path.Ext(parser.InputFile)
			filePrefix := fileName[0 : len(fileName)-len(fileSuffix)]
			if fileName != "" {
				fileName = path.Join(parser.OutputFile, filePrefix+".go")
			} else {
				fileName = path.Join(parser.OutputFile, "tables.go")
			}
		} else {
			fileName = path.Join(parser.OutputFile, fileName)
		}
	}
	if _, ok := parser.Files[fileName]; !ok {
		parser.Files[fileName] = make(map[string]Columns)
	}

	if _, ok := parser.Files[fileName][tableName]; ok {
		return errors.Errorf("duplicate table name :%s", tableName)
	} else {
		for _, col := range stmt.Cols {
			parser.Files[fileName][tableName] = append(parser.Files[fileName][tableName], Column{
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

func New(input string, output string, isDir bool, packageName string) *DDLParser {
	return &DDLParser{
		p:           parser.New(),
		InputFile:   input,
		OutputFile:  output,
		IsDir:       isDir,
		packageName: packageName,
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
