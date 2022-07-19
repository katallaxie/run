package tmpl

import (
	"runtime"
	"text/template"

	sprig "github.com/go-task/slim-sprig"
	"golang.org/x/exp/maps"
)

var (
	tmplFuncs template.FuncMap
)

func init() {
	tmplFuncs = make(template.FuncMap)

	funcs := template.FuncMap{
		"ARCH": func() string {
			return runtime.GOARCH
		},
		"OS": func() string {
			return runtime.GOOS
		},
	}

	tmplFuncs = sprig.TxtFuncMap()
	maps.Copy(tmplFuncs, funcs)
}
