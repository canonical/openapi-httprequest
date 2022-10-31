package templates

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	errgo "gopkg.in/errgo.v1"
)

type TemplateArg struct {
	Pkg            string
	Types          DefinitionList
	Handlers       HandlerList
	ParamsFile     string
	GenerateServer bool
}

// WriteAll writes the generated packed to the provided outputDir
func WriteAll(outputDir string, args TemplateArg) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, os.ModePerm)
	}
	err := Write(Params, args, filepath.Join(outputDir, args.ParamsFile))
	if err != nil {
		return errgo.Notef(err, "cannot write api-params.go template")
	}
	if args.GenerateServer {
		err := Write(Main, args, filepath.Join(outputDir, "main.go"))
		if err != nil {
			return errgo.Notef(err, "cannot write main.go template")
		}
		err = Write(GoMod, args, filepath.Join(outputDir, "go.mod"))
		if err != nil {
			return errgo.Notef(err, "cannot write go.mod template")
		}
	}
	return nil
}

func Write(template *template.Template, data TemplateArg, filepath string) error {
	var args = struct {
		TemplateArg
		Imports []string
	}{data, []string{}}

	packagesToImport := make(map[string]bool)
	typeToPackageMap := map[string]string{
		"time.Time":       "time",
		"*time.Time":      "time",
		"json.RawMessage": "encoding/json",
	}

	for _, def := range data.Types {
		name, ok := typeToPackageMap[def.TypeStr]
		if ok {
			packagesToImport[name] = true
		}

		for _, prop := range def.Properties {
			name, ok := typeToPackageMap[prop.TypeStr]
			if ok {
				packagesToImport[name] = true
			}
		}
	}

	for k := range packagesToImport {
		args.Imports = append(args.Imports, k)
	}

	var buf bytes.Buffer
	if err := template.Execute(&buf, args); err != nil {
		return errgo.Mask(err)
	}

	outputData := buf.Bytes()
	if strings.HasSuffix(filepath, ".go") {
		data, err := format.Source(outputData)
		if err != nil {
			return errgo.Notef(err, "invalid Go source output")
		}
		outputData = data
	}

	err := ioutil.WriteFile(filepath, outputData, 0666)
	return errgo.Mask(err)
}

// Comment takes a string and turns it into a golang comment using "//", while preserving new lines.
func Comment(str string) string {
	str = strings.TrimSpace(str)
	if str == "" {
		return ""
	}
	return "// " + strings.Replace(str, "\n", "\n// ", -1)
}
