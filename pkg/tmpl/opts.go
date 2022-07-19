package tmpl

import "text/template"

// Opt ...
type Opt func(*Opts)

// Opts ...
type Opts struct {
	Fields                TmplFields
	Funcs                 template.FuncMap
	FailOnMissing         bool
	DisableReplaceNoValue bool
}

// TmplFields ...
type TmplFields map[string]interface{}

// NewOpts ...
func NewOpts() Opts {
	return Opts{
		Fields: make(TmplFields),
		Funcs:  tmplFuncs,
	}
}

// Configure os configuring the options.
func (o *Opts) Configure(opts ...Opt) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithExtraFields ...
func WithExtraFields(fields TmplFields) Opt {
	return func(opts *Opts) {
		for f, v := range fields {
			opts.Fields[f] = v
		}
	}
}

// WithExtraFuncs ...
func WithExtraFuncs(funcs template.FuncMap) Opt {
	return func(opts *Opts) {
		for f, v := range funcs {
			opts.Funcs[f] = v
		}
	}
}

// WithDisableReplaceNoValue ...
func WithDisableReplaceNoValue() Opt {
	return func(opts *Opts) {
		opts.DisableReplaceNoValue = true
	}
}

// WithFailOnMissing ...
func WithFailOnMissing() Opt {
	return func(opts *Opts) {
		opts.FailOnMissing = true
	}
}
