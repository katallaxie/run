package spec

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/katallaxie/run/pkg/plugin"
	"github.com/katallaxie/run/pkg/tmpl"
	"github.com/katallaxie/run/pkg/utils"

	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"
)

var (
	ErrTaskNotFound = fmt.Errorf("task not found")
)

// DefaultFilename ...
const (
	DefaultFilename = ".run.yml"
)

// Spec ...
type Spec struct {
	// Spec ...
	Spec int `validate:"required" yaml:"spec"`
	// Version ...
	Version string `validate:"required" yaml:"version,omitempty"`
	// Description ...
	Description string `yaml:"description,omitempty"`
	// Authors ...
	Authors Authors `validate:"required" yaml:"authors,omitempty"`
	// Homepage ...
	Homepage string `yaml:"homepage,omitempty"`
	// License ...
	License string `yaml:"license,omitempty"`
	// Repository ...
	Repository string `yaml:"repository,omitempty"`
	// Plugins ...
	Plugins Plugins `yaml:"plugins,omitempty"`
	// Tasks ...
	Tasks Tasks `yaml:"tasks"`
	// Vars ...
	Vars Vars `yaml:"vars"`
	// Env ...
	Env Env `yaml:"env"`
}

// Fields ...
func (s *Spec) Fields() tmpl.Fields {
	fields := tmpl.Fields{
		"Spec":        s.Spec,
		"Version":     s.Version,
		"Description": s.Description,
		"Authors":     s.Authors,
		"License":     s.License,
		"Repository":  s.Repository,
	}

	return fields
}

// Validate ..
func (s *Spec) Validate() error {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	err := v.Struct(s)
	if err != nil {
		return err
	}

	return v.Struct(s)
}

// Environ ...
func (s *Spec) Environ() []string {
	if s.Env == nil {
		return nil
	}

	environ := os.Environ()

	for k, v := range s.Env {
		environ = append(environ, fmt.Sprintf("%s=%s", k, v))
	}

	return environ
}

// Load ...
func Load(file string) (*Spec, error) {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var spec Spec
	err = yaml.Unmarshal(f, &spec)
	if err != nil {
		return nil, err
	}

	return &spec, nil
}

// Default ...
func (s *Spec) Default() []string {
	tt := make([]string, 0)

	for k, t := range s.Tasks {
		if t.Default {
			tt = append(tt, k)
		}
	}

	return tt
}

// Find ...
func (s *Spec) Find(names ...string) ([]string, error) {
	all := make(map[string]bool)
	tt := make([]string, 0)

	for _, name := range names {
		if _, exists := all[name]; exists {
			continue
		}

		t, ok := s.Tasks[name]
		if !ok {
			return nil, ErrTaskNotFound
		}

		for _, dep := range t.DependsOn {
			if _, exists := all[dep]; exists {
				continue
			}

			d, ok := s.Tasks[dep]
			if !ok {
				return nil, ErrTaskNotFound
			}
			tt = append(tt, dep)
			all[d.Name] = true
		}

		all[name] = true
		tt = append(tt, name)
	}

	return tt, nil
}

// Authors ...
type Authors []string

// Plugins ...
type Plugins []Plugin

// Tasks ...
type Tasks map[string]Task

// Task ...
type Task struct {
	If        string    `yaml:"if"`
	Default   bool      `yaml:"default"`
	DependsOn DependsOn `yaml:"depends-on"`
	Name      string    `yaml:"name"`
	Disabled  bool      `yaml:"disabled"`
	Env       Env       `yaml:"env"`
	Vars      Vars      `yaml:"vars"`
	Templates Templates `yaml:"template,omitempty"`

	Watch      Watch      `yaml:"watch"`
	WorkingDir WorkingDir `yaml:"working-dir"`
	Steps      Steps      `yaml:"steps"`
}

// RunOpt ...
type RunOpt func(*RunOpts)

// RunOpts ...
type RunOpts struct {
	WorkingDir WorkingDir
	Vars       Vars
	Env        Env
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
}

// Configure ...
func (o *RunOpts) Configure(opts ...RunOpt) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithExtraVars ...
func WithExtraVars(vars Vars) RunOpt {
	return func(o *RunOpts) {
		o.Vars = vars
	}
}

// WithStdin ...
func WithStdin(r io.Reader) RunOpt {
	return func(o *RunOpts) {
		o.Stdin = r
	}
}

// WithStdout ...
func WithStdout(w io.Writer) RunOpt {
	return func(o *RunOpts) {
		o.Stdout = w
	}
}

// WithStderr ...
func WithStderr(w io.Writer) RunOpt {
	return func(o *RunOpts) {
		o.Stderr = w
	}
}

// WithExtraEnv ...
func WithExtraEnv(env Env) RunOpt {
	return func(o *RunOpts) {
		o.Env = env
	}
}

// WithWorkingDir ...
func WithWorkingDir(dir WorkingDir) RunOpt {
	return func(o *RunOpts) {
		o.WorkingDir = dir
	}
}

// Run ...
func (t *Task) Run(ctx context.Context, opts ...RunOpt) error {
	options := new(RunOpts)
	options.Configure(opts...)

	for _, template := range t.Templates {
		ff := make(tmpl.TmplFields)
		for k, v := range template.Vars {
			ff[k] = v
		}

		fmt.Println(template.Vars)

		gen := tmpl.New(tmpl.WithExtraFields(ff))
		err := gen.ApplyFile(template.File, template.Out)
		if err != nil {
			return err
		}
	}

	for _, s := range t.Steps {
		if err := s.Run(ctx, append(opts, WithExtraEnv(t.Env), WithExtraVars(t.Vars))...); err != nil {
			return err
		}
	}

	return nil
}

// Step ...
type Step struct {
	Cmd              string            `yaml:"cmd"`
	ContinueOnError  bool              `yaml:"continue-on-error"`
	Env              Env               `yaml:"env"`
	Id               string            `yaml:"id"`
	If               string            `yaml:"if"`
	TimeoutInSeconds int64             `yaml:"timeout-in-seconds"`
	Uses             string            `yaml:"uses"`
	Vars             Vars              `yaml:"vars"`
	With             map[string]string `yaml:"with"`
	WorkingDir       WorkingDir        `yaml:"working-dir"`
}

// Run ...
func (s *Step) Run(ctx context.Context, opts ...RunOpt) error {
	options := new(RunOpts)
	options.Configure(opts...)

	maps.Copy(options.Env, s.Env)
	maps.Copy(options.Vars, s.Vars)

	if s.WorkingDir != "" {
		options.WorkingDir = s.WorkingDir
	}

	cmds := strings.Split(s.Cmd, "\n")
	timeout := time.Duration(time.Nanosecond * math.MaxInt)
	if s.TimeoutInSeconds > 0 {
		timeout = time.Duration(time.Second * time.Duration(s.TimeoutInSeconds))
	}

	if s.Uses != "" {
		err := s.runRemote(ctx, s.Uses, timeout)
		if err != nil && !s.ContinueOnError {
			return err
		}

		return nil
	}

	for _, cmd := range cmds {

		err := s.runCmd(ctx, cmd, timeout, options)
		if err != nil && !s.ContinueOnError {
			return err
		}
	}

	return nil
}

func (s *Step) runRemote(ctx context.Context, path string, timeout time.Duration) error {
	m := &plugin.Meta{Path: path}
	f := m.Factory(ctx)

	p, err := f()
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	_, err = p.Execute(plugin.ExecuteRequest{})
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (s *Step) runCmd(ctx context.Context, cmd string, timeout time.Duration, opts *RunOpts) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	p, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	if err != nil {
		return err
	}

	r, err := interp.New(
		interp.Dir(string(opts.WorkingDir)),
		interp.Env(expand.ListEnviron(append(os.Environ(), utils.Strings(opts.Env)...)...)),

		interp.Module(interp.DefaultExec),
		interp.Module(interp.OpenDevImpls(interp.DefaultOpen)),

		interp.StdIO(opts.Stdin, opts.Stdout, opts.Stderr),
	)
	if err != nil {
		return err
	}

	err = r.Run(ctx, p)
	if err != nil {
		return err
	}

	return nil
}

// Steps ...
type Steps []Step

// Task ...
func (t *Task) Environ() []string {
	if t.Env == nil {
		return nil
	}

	environ := os.Environ()

	for k, v := range t.Env {
		environ = append(environ, fmt.Sprintf("%s=%s", k, v))
	}

	return environ
}

// WorkingDir ...
type WorkingDir string

// String ...
func (w WorkingDir) String() string {
	return string(w)
}

// Templates ...
type Templates []Template

// Template ...
type Template struct {
	File string `yaml:"file"`
	Out  string `yaml:"out"`
	Vars Vars   `yaml:"var"`
}

// Watch ...
type Watch struct {
	Paths   Paths   `yaml:"paths,omitempty"`
	Ignores Ignores `yaml:"ignores,omitempty"`
}

// Paths ...
type Paths []string

// Ignore ...
type Ignores []string

// Vars ...
type Vars map[string]string

// Merge ...
func (v Vars) Merge(vars Vars) {
	maps.Copy(v, vars)
}

// Env ...
type Env map[string]string

// DependsOn ...
type DependsOn []string

// Commands ...
type Commands []Command

// Command ...
type Command string

// Inputs ...
type Inputs []Input

// Input ...
type Input struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Prompt string `yaml:"prompt"`
	Regex  string `yaml:"regex"`
}

// Includes ...
type Includes []string

// Excludes ...
type Excludes []string

// Plugin ...
type Plugin struct {
	Id          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Path        string `yaml:"path"`
}
