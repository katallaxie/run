package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/katallaxie/run/pkg/spec"
)

// Runner ...
type Runner struct {
	ctx   context.Context
	funcs []RunFunc
	pool  sync.Pool
	opts  *Opts

	sync.Mutex
}

// Opt ...
type Opt func(*Opts)

// Opts ...
type Opts struct {
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	Timeout    time.Duration
	File       *spec.Spec
	Vars       Vars
	Env        Env
	WorkingDir spec.WorkingDir
}

// Configure ...
func (o *Opts) Configure(opts ...Opt) {
	for _, opt := range opts {
		opt(o)
	}

	if o.Stdin == nil {
		o.Stdin = os.Stdin
	}

	if o.Stdout == nil {
		o.Stdout = os.Stdout
	}

	if o.Stderr == nil {
		o.Stderr = os.Stderr
	}
}

// WithSpec ...
func WithSpec(file *spec.Spec) Opt {
	return func(o *Opts) {
		o.File = file
	}
}

// WithVars ...
func WithVars(vars Vars) Opt {
	return func(o *Opts) {
		o.Vars = vars
	}
}

// WithEnv ...
func WithEnv(env Env) Opt {
	return func(o *Opts) {
		o.Env = env
	}
}

// WithWorkingDir ...
func WithWorkingDir(cwd string) Opt {
	return func(o *Opts) {
		o.WorkingDir = spec.WorkingDir(cwd)
	}
}

// Context ...
func (r *Runner) Context() context.Context {
	return r.ctx
}

// AcquireCtx ...
func (r *Runner) AcquireCtx() *Ctx {
	c := r.pool.Get().(*Ctx)
	c.Reset()

	c.runner = r

	return c
}

// Stdin ...
func (r *Runner) Stdin() io.Reader {
	return r.opts.Stdin
}

// Stdout ...
func (r *Runner) Stdout() io.Writer {
	return r.opts.Stdout
}

// Stderr ...
func (r *Runner) Stderr() io.Writer {
	return r.opts.Stderr
}

// ReleaseFunc ...
type ReleaseFunc func()

// ReleseCtx ...
func (r *Runner) ReleaseCtx(c *Ctx) ReleaseFunc {
	return func() { r.pool.Put(c) }
}

// RunFunc ...
type RunFunc func(c *Ctx) error

// RunTask ...
func (r *Runner) RunTasks(tasks ...string) error {
	tt := make([]spec.Task, len(tasks))

	for i, task := range tasks {
		t, ok := r.opts.File.Tasks[task]
		if !ok {
			return fmt.Errorf("task %s not found", task)
		}
		tt[i] = t
	}

	for _, t := range tt {
		c := r.AcquireCtx()
		defer r.ReleaseCtx(c)

		if err := t.Run(
			c.Context(),
			spec.WithWorkingDir(r.opts.WorkingDir),
			spec.WithExtraVars(r.opts.File.Vars),
			spec.WithExtraEnv(r.opts.File.Env),
			spec.WithStderr(c.runner.Stderr()),
			spec.WithStdin(c.runner.Stdin()),
			spec.WithStdout(c.runner.Stdout()),
		); err != nil {
			return err
		}
	}

	return nil
}

// WithContext ...
func WithContext(ctx context.Context, opts ...Opt) *Runner {
	options := new(Opts)
	options.Configure(opts...)

	return &Runner{
		ctx:  ctx,
		opts: options,
		pool: sync.Pool{
			New: func() interface{} {
				return new(Ctx)
			},
		},
	}
}

// Use ...
func (r *Runner) Use(funcs ...RunFunc) {
	r.funcs = append(r.funcs, funcs...)
}
