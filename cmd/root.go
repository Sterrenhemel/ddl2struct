package cmd

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/iancoleman/strcase"

	"github.com/Sterrenhemel/ddl2struct/pkg/parser"
	_ "github.com/Sterrenhemel/ddl2struct/pkg/parser_driver"
	"github.com/Sterrenhemel/ddl2struct/pkg/tpl"
	"github.com/Sterrenhemel/ddl2struct/pkg/util/logutil"

	"github.com/spf13/cobra"
)

var cfgFile string
var (
	inputPath   string
	outputPath  string
	packageName string
)

var rootCmd = &cobra.Command{
	Use:   "ddl2struct",
	Short: "create golang struct from ddl",
	Long:  ``,
	Run:   runCommand,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logutil.BgSLogger().Fatal(err)
		os.Exit(1)
	}
}

func init() {
	flag := rootCmd.PersistentFlags()
	flag.StringVarP(&inputPath, "input", "i", "", `sql file path`)
	flag.StringVarP(&outputPath, "output", "o", "", `output file path`)
	flag.StringVarP(&packageName, "package", "p", "", "go file package")
}

func runCommand(cmd *cobra.Command, args []string) {
	s, err := os.Stat(inputPath)
	if err != nil {
		logutil.BgSLogger().Fatal(err)
	} else {
		if s.IsDir() {
			files, err := ioutil.ReadDir(inputPath)
			if err != nil {
				logutil.BgSLogger().Fatal(err)
			}
			for _, file := range files {
				if filepath.Ext(file.Name()) == ".sql" {
					name := filepath.Join(inputPath, file.Name())
					parseFile(name)
				}
			}
		}
	}
}

func parseFile(filepath string) {
	sql, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	ddlParser := parser.New(filepath, outputPath, packageName)
	if err := ddlParser.Parse(string(sql)); err != nil {
		panic(err)
	}

	//structFiles, err := ddlParser.ToStructs(true)
	//if err != nil {
	//	panic(err)
	//}

	for fileName := range ddlParser.FileTables {
		t := template.Must(template.New(fileName).Funcs(map[string]interface{}{
			"mapExists": mapExists,
			"ToCamel":   strcase.ToCamel,
			"ToSnake":   strcase.ToSnake,
		}).Parse(tpl.TableTemplate))
		buf := &bytes.Buffer{}
		err := t.Execute(buf, TemplateVar{
			InputFile:   filepath,
			PackageName: packageName,
			Imports:     ddlParser.FileImports[fileName],
			Structs:     ddlParser.FileTables[fileName],
			WithTag:     true,
			//FileContent: string(fileBytes),
		})
		if err != nil {
			panic(err)
		}
		content := buf.Bytes()
		source, err := format.Source(content)
		if err != nil {
			return
		}
		if fileName != "" {
			if err := ioutil.WriteFile(fileName, source, 0644); err != nil {
				panic(err)
			}
		}
		fmt.Printf("%s", source)
	}
}

type TemplateVar struct {
	InputFile   string
	PackageName string
	Imports     map[string]string
	Structs     map[string]*parser.Table
	WithTag     bool
	TagString   string
	FileContent string
}

func mapExists(v TemplateVar) bool {
	if v.Imports == nil || len(v.Imports) == 0 {
		return false
	}
	return true
}

func generateFileFromBytes(structBytes []byte) {
	if err := ioutil.WriteFile(outputPath, structBytes, 0644); err != nil {
		panic(err)
	}
}
