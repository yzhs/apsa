package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"

	"github.com/yzhs/apsa"
)

func main() {
	apsa.InitConfig()

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	libraryDir := fmt.Sprintf("%s/.apsa/library", home)
	entries, err := os.ReadDir(libraryDir)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		file := entry.Name()
		if !strings.HasSuffix(file, ".md") {
			continue
		}

		fileContent, err := ioutil.ReadFile(libraryDir + "/" + file)
		if err != nil {
			panic(err)
		}

		recipe := apsa.Parse(file[:len(file)-3], string(fileContent))
		if len(recipe.Ingredients) == 0 {
			fmt.Printf("Could not parse ingredients for recipe '%s' (%s)\n", recipe.Title, recipe.Id)
			continue
		}

		tmpl := template.Must(template.New("template.yaml").Funcs(sprig.TxtFuncMap()).Parse(Template))
		tmpl.Execute(os.Stdout, recipe)
		break
	}
}

const Template = `title: {{ .Title }}
portions: {{ .Portions }}
source: {{ .Source }}
tags:
  {{- range .Tags }}
  - {{ . }}
  {{- end }}

steps:
- ingredients:
  {{- range .Ingredients }}
  - {{ . }}
  {{- end }}
  instructions: |
    {{- .Content | indent 4 }}
`
