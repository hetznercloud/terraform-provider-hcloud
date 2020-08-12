package testtemplate

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"text/template"
	"time"
)

var rng *rand.Rand

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

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
			ts.RandInt = rng.Int()
		}
		if ts.RandName == "" {
			ts.RandName = "r_" + strconv.Itoa(ts.RandInt)
		}

		ts.tmpl = template.New("testdata")
		ts.tmpl.Funcs(template.FuncMap{
			"StringsJoin":      strings.Join,
			"StringsTrimSpace": strings.TrimSpace,
			"DQuoteS": func(ss []string) []string {
				res := make([]string, len(ss))
				for i, s := range ss {
					res[i] = fmt.Sprintf("\"%s\"", s)
				}
				return res
			},
		})

		// We can't use template/Template.ParseGlob here since we want to add a
		// prefix to the name.
		rGlob := filepath.Join(ResourceTemplateDir(t), "*.tf.tmpl")
		parseTmplGlob(t, ts.tmpl, "testdata/r", rGlob)

		dGlob := filepath.Join(DataSourceTemplateDir(t), "*.tf.tmpl")
		parseTmplGlob(t, ts.tmpl, "testdata/d", dGlob)
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
		data, ok := args[i+1].(Data)
		if !ok {
			t.Fatalf("args[%d]: data required: %T", i+1, args[i+1])
		}
		data.SetRInt(ts.RandInt)
		if data.RName() == "" {
			data.SetRName(ts.RandName)
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

	hcl := strings.TrimSpace(buf.String())
	t.Logf("\n\nHCL:\n%s\n", addLineNumbers(hcl))
	return hcl
}

func parseTmplGlob(t *testing.T, root *template.Template, prefix, glob string) {
	fileNames, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("list files in %s: %v", glob, err)
	}
	for _, fileName := range fileNames {
		str, err := ioutil.ReadFile(fileName)
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
		sb.WriteString(fmt.Sprintf("%5d: %s\n", i+1, l))
	}

	return sb.String()
}
