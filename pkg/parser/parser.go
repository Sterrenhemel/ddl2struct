package parser

import (
	"bytes"
	"os"
	"path"
	"regexp"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/types"
)

var (
	goFileRegex = regexp.MustCompile("([a-zA-Z0-9_]*\\.go)")
)

type DDLParser struct {
	FileTables  map[string]map[string]*Table // fileName -> TableName -> Table
	FileImports map[string]map[string]string // fileName -> alias -> importName
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
	parser.FileTables = make(map[string]map[string]*Table)
	parser.FileImports = make(map[string]map[string]string)
	parser.Index = make(map[string]Indexes)

	for _, node := range nodes {
		node.Accept(parser)
		if parser.err != nil {
			return errors.Wrap(err, "sql parsing error")
		}
	}

	return nil
}

//func (parser DDLParser) ToStructs(withTag bool) (fileContentMap map[string][]byte, err error) {
//	fileContentMap = make(map[string][]byte)
//	var builder strings.Builder
//	for fileName, tables := range parser.FileTables {
//		for tableName, columns := range tables {
//			builder.WriteString(fmt.Sprintf("type %s struct { %s }\n\n", strcase.ToCamel(tableName), columns.ToStructFields(withTag)))
//		}
//		fileContentMap[fileName], err = format.Source([]byte(builder.String()))
//		if err != nil {
//			return
//		}
//		builder.Reset()
//	}
//	return
//}

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
	fileName, tableComment := parser.parseOutput(stmt)
	tableName := stmt.Table.Name.String()
	if parser.FileImports[fileName] == nil {
		parser.FileImports[fileName] = make(map[string]string)
	}
	if parser.FileTables[fileName] == nil {
		parser.FileTables[fileName] = make(map[string]*Table)
	}

	if _, ok := parser.FileTables[fileName][tableName]; ok {
		return errors.Errorf("duplicate table name :%s", tableName)
	} else {
		table := &Table{
			TableName:    tableName,
			TableComment: tableComment,
			Columns:      []Column{},
		}
		parser.FileTables[fileName][tableName] = table
		for _, col := range stmt.Cols {
			var colComment string
			for _, option := range col.Options {
				if option.Tp == ast.ColumnOptionComment {
					if option.StrValue != "" {
						colComment = option.StrValue
					} else if option.Text() != "" {
						colComment = option.Text()
					} else {
						var buf bytes.Buffer
						option.Expr.Format(&buf)
						colComment = buf.String()
					}

					break
				}
			}
			tableColumn := Column{
				Name:    col.Name.Name.String(),
				Type:    parser.getColumnType(col.Tp),
				Comment: colComment,
			}
			parser.addImport(fileName, tableColumn)
			table.Columns = append(table.Columns, tableColumn)
		}
	}
	return nil
}

func (parser *DDLParser) parseOutput(stmt *ast.CreateTableStmt) (fileName string, tableComment string) {
	s, err := os.Stat(parser.OutputFile)
	if err != nil {
		fileName = parser.OutputFile
	} else {
		if s.IsDir() {
			for _, option := range stmt.Options {
				if option.Tp == ast.TableOptionComment {
					fileName = goFileRegex.FindString(option.StrValue)
					//strings.ReplaceAll(option.StrValue, "*.go", "")
					tableComment = option.StrValue
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
		} else {
			fileName = parser.OutputFile
		}
	}
	return
}

func (parser *DDLParser) addImport(fileName string, column Column) {
	switch column.Type {
	case "time.Time":
		parser.FileImports[fileName]["time"] = "time"
	}
}

func (parser *DDLParser) parseCreateIndexStmt(stmt *ast.CreateIndexStmt) error {
	return nil
}

func (parser *DDLParser) getColumnType(fieldType *types.FieldType) string {
	fieldType.EvalType()
	switch fieldType.Tp {
	case mysql.TypeTiny:
		return "int8"
	case mysql.TypeShort, mysql.TypeInt24, mysql.TypeYear:
		return "int"
	case mysql.TypeLong, mysql.TypeLonglong, mysql.TypeBit:
		return "int64"
	case mysql.TypeFloat:
		return "float32"
	case mysql.TypeDouble, mysql.TypeNewDecimal:
		return "float64"
	case mysql.TypeDate, mysql.TypeDatetime:
		return "time.Time"
	case mysql.TypeTimestamp:
		return "time.Time"
	default:
		return "string"
	}
}

func New(input string, output string, packageName string) *DDLParser {
	return &DDLParser{
		p:           parser.New(),
		InputFile:   input,
		OutputFile:  output,
		packageName: packageName,
	}
}
