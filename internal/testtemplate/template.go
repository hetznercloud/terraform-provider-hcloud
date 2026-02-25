package testtemplate

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Data marks a struct as containing data for test templates.
type Data interface {
	SetRInt(int)
	SetRName(string)
	RInt() int
	RName() string
}

// DataCommon contains fields required by all template data types.
type DataCommon struct {
	rInt  int
	rName string
}

// SetRInt sets a random integer.
func (tdc *DataCommon) SetRInt(i int) {
	tdc.rInt = i
}

// RInt returns the random integer
func (tdc *DataCommon) RInt() int {
	return tdc.rInt
}

// SetRName sets a resource name.
func (tdc *DataCommon) SetRName(n string) {
	tdc.rName = n
}

// RInt returns the random name
func (tdc *DataCommon) RName() string {
	return tdc.rName
}

// Manager loads and renders Terraform HCL templates.
type Manager struct {
	RandInt  int
	RandName string
	tmpl     *template.Template
	once     sync.Once // ensures templates in <project-root>/internal/testdata/{r,d} are loaded only once.
}

// init loads the templates in <project-root>/internal/testdata/{r,d} exactly
// once.
func (ts *Manager) init(t *testing.T) {
	t.Helper()

	ts.once.Do(func() {
		if ts.RandInt == 0 {
			ts.RandInt = rand.Int() // nolint: gosec
		}
		if ts.RandName == "" {
			ts.RandName = "r_" + strconv.Itoa(ts.RandInt)
		}

		ts.tmpl = template.New("testdata")

		funMap := sprig.TxtFuncMap()
		funMap["quoteEach"] = func(items []string) []string {
			result := make([]string, 0, len(items))
			for _, item := range items {
				result = append(result, fmt.Sprintf("%q", item))
			}
			return result
		}
		ts.tmpl.Funcs(funMap)

		// We can't use template/Template.ParseGlob here since we want to add a
		// prefix to the name.
		rGlob := filepath.Join(ResourceTemplateDir(t), "*.tf.tmpl")
		parseTmplGlob(t, ts.tmpl, "testdata/r", rGlob)

		dGlob := filepath.Join(DataSourceTemplateDir(t), "*.tf.tmpl")
		parseTmplGlob(t, ts.tmpl, "testdata/d", dGlob)

		aGlob := filepath.Join(ActionTemplateDir(t), "*.tf.tmpl")
		parseTmplGlob(t, ts.tmpl, "testdata/a", aGlob)
	})
}

// Render renders the passed template to a string using the provided data.
func (ts *Manager) Render(t *testing.T, args ...interface{}) string {
	t.Helper()
	ts.init(t)

	if len(args)%2 != 0 {
		t.Fatal("even number of args required")
	}

	var buf bytes.Buffer
	for i := 0; i < len(args); i += 2 {
		tmplName, ok := args[i].(string)
		if !ok {
			t.Fatalf("args[%d]: string required: %T", i, args[i])
		}

		var data any
		switch arg := args[i+1].(type) {
		case Data:
			arg.SetRInt(ts.RandInt)
			if arg.RName() == "" {
				arg.SetRName(ts.RandName)
			}
			data = arg
		case string:
			data = arg
		default:
			t.Fatalf("args[%d]: data or string required: %#v", i+1, args[i+1])
		}

		tmpl := ts.tmpl.Lookup(tmplName)
		if tmpl == nil {
			t.Fatalf("template %s: not found", tmplName)
		}
		if err := tmpl.Execute(&buf, data); err != nil {
			t.Fatalf("execute template: %v", err)
		}
		// Just append a new line. If it is the last line we trim extraneous
		// whitespace anyway.
		buf.WriteString("\n")
	}

	definition := buf.String()

	{
		// Remove empty lines
		lines := strings.Split(definition, "\n")
		lines = slices.DeleteFunc(lines, func(line string) bool { return strings.TrimSpace(line) == "" })
		definition = strings.Join(lines, "\n")
	}

	{
		// Format HCL files
		buf := bytes.NewBuffer(nil)
		file, diags := hclwrite.ParseConfig([]byte(definition), "testing.tf", hcl.InitialPos)
		if diags.HasErrors() {
			t.Logf("\n\nHCL:\n%s\n", addLineNumbers(definition))
			t.Fatal(diags.Error())
		}
		if _, err := file.WriteTo(buf); err != nil {
			t.Fatal(err)
		}
		definition = buf.String()
	}

	t.Logf("\n\nHCL:\n%s\n", addLineNumbers(definition))

	return definition
}

func parseTmplGlob(t *testing.T, root *template.Template, prefix, glob string) {
	fileNames, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("list files in %s: %v", glob, err)
	}
	for _, fileName := range fileNames {
		str, err := os.ReadFile(fileName)
		if err != nil {
			t.Fatalf("read template file %s: %v", fileName, err)
		}

		tmplName := filepath.Base(fileName)
		tmplName = strings.TrimSuffix(tmplName, ".tf.tmpl")
		tmplName = fmt.Sprintf("%s/%s", prefix, tmplName)
		tmpl := root.New(tmplName)

		if _, err := tmpl.Parse(string(str)); err != nil {
			t.Fatalf("parse template %s: %v", fileNames, err)
		}
	}
}

func addLineNumbers(s string) string {
	var sb strings.Builder

	lines := strings.Split(s, "\n")
	for i, l := range lines {
		fmt.Fprintf(&sb, "%5d: %s\n", i+1, l)
	}

	return sb.String()
}
