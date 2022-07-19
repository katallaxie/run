package runner

import (
	"context"
	"fmt"

	"github.com/katallaxie/run/pkg/spec"
)

// Ctx ...
type Ctx struct {
	funcs      []RunFunc
	idx        int
	cmd        Cmd
	env        Env
	runner     *Runner
	vars       Vars
	workingDir WorkingDir
	task       spec.Task
}

// Vars ...
type Vars map[string]string

// Env ...
type Env map[string]string

// WorkingDir ..
type WorkingDir string

// WorkingDir ..
func (c *Ctx) WorkingDir() WorkingDir {
	return c.workingDir
}

// Cmd ...
type Cmd string

// Cmd ...
func (c *Ctx) Cmd() Cmd {
	return c.cmd
}

// Runner ...
func (c *Ctx) Runner() *Runner {
	return c.runner
}

// Task ...
func (c *Ctx) Task() spec.Task {
	return c.task
}

// Reset ...
func (c *Ctx) Reset() {
	c.env = make(Env)
	c.vars = make(Vars)
	c.cmd = ""
	c.workingDir = ""
}

// Next ...
func (c *Ctx) Next() error {
	c.idx++
	if c.idx < len(c.funcs) {
		if err := c.funcs[c.idx](c); err != nil {
			return err
		}
	}

	return nil
}

// Context ...
func (c *Ctx) Context() context.Context {
	return c.runner.Context()
}

// Env ...
func (c *Ctx) Env() []string {
	env := make([]string, len(c.env))
	for k, v := range c.env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}
