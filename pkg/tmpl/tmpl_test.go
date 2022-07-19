package tmpl_test

import (
	"errors"
	"runtime"
	"testing"

	"github.com/katallaxie/run/pkg/tmpl"

	"github.com/stretchr/testify/assert"
)

func TestTemplate_New(t *testing.T) {
	tmpl := tmpl.New()
	assert.NotNil(t, tmpl)
}

func TestTemplate_Apply(t *testing.T) {
	type test struct {
		input    string
		want     string
		hasError bool
		err      error
		opts     []tmpl.Opt
	}

	tests := []test{
		{input: "{{OS}}", want: runtime.GOOS, err: nil},
		{input: "{{ARCH}}", want: runtime.GOARCH, err: nil},
		{input: "{{.FOO}}", want: "", err: nil},
		{input: "{{.FOO}}", want: "<no value>", err: nil, opts: []tmpl.Opt{tmpl.WithDisableReplaceNoValue()}},
		{input: "{{.FOO}}", want: "", hasError: true, err: errors.New("map has no entry for key"), opts: []tmpl.Opt{tmpl.WithFailOnMissing()}},
	}

	for _, tc := range tests {
		tmpl := tmpl.New(tc.opts...)
		got, err := tmpl.Apply(tc.input)

		assert.Equal(t, tc.want, got)
		if tc.hasError {
			assert.ErrorContains(t, err, tc.err.Error())
		} else {
			assert.NoError(t, err)
		}
	}
}
