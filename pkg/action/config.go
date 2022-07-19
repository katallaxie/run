package action

import (
	"strings"

	"github.com/sethvargo/go-githubactions"
)

// Config ...
type Config struct {
	File    string
	Tasks   []string
	Verbose bool
}

// NewFromInputs ...
func NewFromInputs(action *githubactions.Action) (*Config, error) {
	c := &Config{}
	c.File = action.GetInput("config")

	tasks := strings.Split(action.GetInput("tasks"), "")
	for i, t := range tasks {
		tasks[i] = strings.TrimSpace(t)
	}
	c.Tasks = tasks

	return c, nil
}

// InitActionConfig ...
func (c *Config) InitActionConfig(action *githubactions.Action) error {
	c.File = action.GetInput("config")

	return nil
}
