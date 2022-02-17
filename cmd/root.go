package cmd

import (
	"bytes"
	"fmt"
	"github.com/Sterrenhemel/ddl2struct/pkg/tpl"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/Sterrenhemel/ddl2struct/pkg/parser"
	_ "github.com/Sterrenhemel/ddl2struct/pkg/parser_driver"

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
		fmt.Println(err)
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
	sql, err := ioutil.ReadFile(inputPath)
	if err != nil {
		panic(err)
	}

	var isDir bool
	s, err := os.Stat(outputPath)
	if err != nil {
		isDir = false
	} else {
		if s.IsDir() {
			isDir = true
		} else {
			isDir = false
		}
	}

	parser := parser.New(inputPath, outputPath, isDir, packageName)
	if err := parser.Parse(string(sql)); err != nil {
		panic(err)
	}

	structFiles, err := parser.ToStructs(true)
	if err != nil {
		panic(err)
	}

	for fileName, fileBytes := range structFiles {
		t := template.Must(template.New(fileName).Funcs(map[string]interface{}{
			"mapExists": mapExists,
		}).Parse(tpl.TableTemplate))
		buf := &bytes.Buffer{}
		err := t.Execute(buf, TemplateVar{
			InputFile:   inputPath,
			PackageName: packageName,
			Imports:     parser.FileImports[fileName],
			FileContent: string(fileBytes),
		})
		if err != nil {
			panic(err)
		}
		content := buf.Bytes()

		if fileName != "" {
			if err := ioutil.WriteFile(fileName, content, 0644); err != nil {
				panic(err)
			}
		}
		fmt.Printf("%s", content)
	}
}

type TemplateVar struct {
	InputFile   string
	PackageName string
	Imports     map[string]string
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
