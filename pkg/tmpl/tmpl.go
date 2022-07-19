package tmpl

import (
	"bytes"
	"io/ioutil"
	"os"
	"text/template"
)

// Template ...
type Template struct {
	opts Opts
}

// Fields ...
type Fields map[string]interface{}

// New ...
func New(opts ...Opt) *Template {
	options := NewOpts()
	options.Configure(opts...)

	return &Template{
		opts: options,
	}
}

// ApplyFile ...
func (t *Template) ApplyFile(in, out string) error {
	i, err := ioutil.ReadFile(in)
	if err != nil {
		return err
	}

	o, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer o.Close()

	s, err := t.Apply(string(i))
	if err != nil {
		return err
	}

	_, err = o.WriteString(s)
	if err != nil {
		return err
	}

	return nil
}

// Apply ...
func (t *Template) Apply(s string) (string, error) {
	tmpl, err := template.New("").Funcs(t.opts.Funcs).Parse(s)
	if err != nil {
		return "", err
	}

	if t.opts.FailOnMissing {
		tmpl.Option("missingkey=error")
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, t.opts.Fields); err != nil {
		return "", err
	}

	if t.opts.DisableReplaceNoValue {
		return out.String(), nil
	}

	b := bytes.ReplaceAll(out.Bytes(), []byte("<no value>"), []byte(""))

	return string(b), err
}
