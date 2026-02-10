package flags

import "flag"

type Options struct {
	ConfigPath string
	Language   string
	Framework  string
	Name       string
	Dir        string
	DryRun     bool
	NoTUI      bool
}

func Parse(args []string) (Options, error) {
	fs := flag.NewFlagSet("project-initiator", flag.ContinueOnError)

	var opts Options
	fs.StringVar(&opts.ConfigPath, "config", "", "Path to config file")
	fs.StringVar(&opts.Language, "lang", "", "Language to scaffold")
	fs.StringVar(&opts.Framework, "framework", "", "Framework to scaffold")
	fs.StringVar(&opts.Name, "name", "", "Project name")
	fs.StringVar(&opts.Dir, "dir", "", "Base directory for the new project")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "Print actions without writing files")
	fs.BoolVar(&opts.NoTUI, "no-tui", false, "Disable TUI prompts")

	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	return opts, nil
}
