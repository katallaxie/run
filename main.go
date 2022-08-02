package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/debug"
	"time"

	"github.com/andersnormal/pkg/utils/files"
	"github.com/katallaxie/run/pkg/config"
	"github.com/katallaxie/run/pkg/plugin"
	"github.com/katallaxie/run/pkg/runner"
	"github.com/katallaxie/run/pkg/spec"
	"mvdan.cc/sh/syntax"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	version = ""
)

const usage = `Usage: run [-cflvsdpw] [--config] [--force] [--list] [--verbose] [--silent] [--dry] [--plugin] [--watch] [--validate] [--var] [--init] [--version] [--dir] [task...] 

'''
spec: 	 1
tasks:
  test:
    steps:
      - cmd: go test -v ./...
'''

Options:
`

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	cfg := config.New()

	err := cfg.InitDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}

	pflag.Usage = func() {
		log.Print(usage)
		pflag.PrintDefaults()
	}

	pflag.BoolVarP(&cfg.Flags.Verbose, "verbose", "v", cfg.Flags.Verbose, "verbose output")
	pflag.BoolVarP(&cfg.Flags.Help, "help", "h", cfg.Flags.Help, "show help")
	pflag.BoolVar(&cfg.Flags.Init, "init", cfg.Flags.Init, "init config")
	pflag.BoolVarP(&cfg.Flags.Force, "force", "f", cfg.Flags.Force, "force init")
	pflag.BoolVarP(&cfg.Flags.Dry, "dry", "d", cfg.Flags.Dry, "dry run")
	pflag.BoolVarP(&cfg.Flags.Silent, "silent", "s", cfg.Flags.Silent, "silent mode")
	pflag.StringVarP(&cfg.File, "config", "c", cfg.File, "config file")
	pflag.StringSliceVarP(&cfg.Flags.Env, "env", "e", cfg.Flags.Env, "environment variables")
	pflag.StringVarP(&cfg.Flags.Plugin, "plugin", "p", cfg.Flags.Plugin, "plugin")
	pflag.BoolVarP(&cfg.Flags.Validate, "validate", "V", cfg.Flags.Validate, "validate config")
	pflag.BoolVarP(&cfg.Flags.List, "list", "l", cfg.Flags.List, "list tasks")
	pflag.DurationVarP(&cfg.Flags.Timeout, "timeout", "t", time.Second*300, "timeout")
	pflag.BoolVar(&cfg.Flags.Version, "version", cfg.Flags.Version, "version")
	pflag.StringSliceVar(&cfg.Flags.Vars, "var", cfg.Flags.Vars, "variables")
	pflag.BoolVarP(&cfg.Flags.Watch, "watch", "w", cfg.Flags.Watch, "watch")
	pflag.StringVar(&cfg.Flags.Dir, "dir", "", "working directory")
	pflag.Parse()

	cwd, err := cfg.Cwd()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Flags.Dir != "" {
		cwd = cfg.Flags.Dir
	}

	if cfg.Flags.Verbose {
		start := time.Now()
		defer func() { log.Printf("time: %s", time.Since(start)) }()
	}

	if cfg.Flags.Version {
		fmt.Printf("%s\n", getVersion())
		return
	}

	if cfg.Flags.Help {
		pflag.Usage()
		os.Exit(0)
	}

	s, err := cfg.LoadSpec()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Flags.Validate {
		err = s.Validate()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if cfg.Flags.List {
		for k, t := range s.Tasks {
			log.Printf("%s (%s)", k, t.Name)
		}
		os.Exit(0)
	}

	if cfg.Flags.Init {
		s := &spec.Spec{
			Spec:    1,
			Version: "0.0.1",
			Tasks:   map[string]spec.Task{},
		}

		b, err := yaml.Marshal(&s)
		if err != nil {
			log.Fatal(err)
		}

		ok, err := files.FileExists(cfg.File)
		if err != nil {
			log.Fatal(err)
		}

		if ok && !cfg.Flags.Force {
			log.Fatalf("%s already exists, use --force to overwrite", cfg.File)
		}

		f, err := os.Create(cfg.File)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		_, err = f.Write(b)
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	args, cliArgs, err := parseArgs()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := runner.WithContext(ctx, runner.WithSpec(s), runner.WithWorkingDir(cwd))

	r.Lock()
	defer r.Unlock()

	if cfg.Flags.Plugin != "" {
		m := &plugin.Meta{Path: cfg.Flags.Plugin}
		f := m.Factory(ctx)

		p, err := f()
		if err != nil {
			log.Fatal(err)
		}
		defer p.Close()

		pp := s.Vars

		if _, err := p.Execute(plugin.ExecuteRequest{
			Vars:      pp,
			Arguments: cliArgs,
		}); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	tasks, err := s.Find(args...)
	if err != nil {
		log.Fatal(err)
	}

	defaultTasks := s.Default()

	if len(tasks) == 0 && len(defaultTasks) == 0 {
		log.Fatal("no default task")
	}

	if len(tasks) == 0 {
		tasks = defaultTasks
	}

	if err := r.RunTasks(tasks...); err != nil {
		log.Fatal(err)
	}
}

func parseArgs() ([]string, []string, error) {
	args := pflag.Args()
	dashPos := pflag.CommandLine.ArgsLenAtDash()

	if dashPos == -1 {
		return args, []string{}, nil
	}

	cliArgs := make([]string, 0)
	for _, arg := range args[dashPos:] {
		arg = syntax.QuotePattern(arg)
		cliArgs = append(cliArgs, arg)
	}

	return args[:dashPos], cliArgs, nil
}

func getVersion() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" {
		return "unknown"
	}

	version = info.Main.Version
	if info.Main.Sum != "" {
		version += fmt.Sprintf(" (%s)", info.Main.Sum)
	}

	return version
}
