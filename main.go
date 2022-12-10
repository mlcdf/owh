package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/command"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/view"
	"go.mlcdf.fr/sally/cache"

	"github.com/mattn/go-isatty"
	"github.com/mitchellh/cli"
)

var Version = "(devel)"

func HelpFunc(app, description string) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf(
			"Usage: %s [--version] [--help] <command> [<args>]\n\n",
			app))
		buf.WriteString(fmt.Sprintf("%s\n\n", description))
		buf.WriteString("Available commands are:\n")

		// Get the list of keys so we can sort them, and also get the maximum
		// key length so they can be aligned properly.
		keys := make([]string, 0, len(commands))
		maxKeyLen := 0
		for key := range commands {
			if len(key) > maxKeyLen {
				maxKeyLen = len(key)
			}

			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			commandFunc, ok := commands[key]
			if !ok {
				// This should never happen since we JUST built the list of
				// keys.
				panic("command not found: " + key)
			}

			command, err := commandFunc()
			if err != nil {
				log.Printf("[ERR] cli: Command '%s' failed to load: %s",
					key, err)
				continue
			}

			key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLen-len(key)))
			buf.WriteString(fmt.Sprintf("    %s    %s\n", key, command.Synopsis()))
		}

		return buf.String()
	}
}

func main() {
	app := &command.App{
		IsInteractive:    isatty.IsTerminal(os.Stdout.Fd()),
		LinkFunc:         config.EnsureLink,
		APIClientFactory: api.NewClient,
		HTTPClient:       http.DefaultClient,
		View:             &view.View{Writer: os.Stdout},
	}

	var err error

	app.Config, err = config.New(app.IsInteractive)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cli := cli.CLI{
		Args:         os.Args[1:],
		Name:         "owh",
		Version:      Version,
		Autocomplete: true,
		HelpWriter:   os.Stdout,
		HelpFunc:     HelpFunc("owh", "Deploy websites to OVHcloud Web Hosting."),
		Commands: map[string]cli.CommandFactory{
			"deploy": func() (cli.Command, error) {
				return &command.DeployCommand{App: *app}, nil
			},
			"domains": func() (cli.Command, error) {
				return &command.DomainsCommand{App: *app}, nil
			},
			"domains attach": func() (cli.Command, error) {
				return &command.AttachCommand{App: *app}, nil
			},
			"domains detach": func() (cli.Command, error) {
				return &command.DetachCommand{App: *app}, nil
			},
			"hostings": func() (cli.Command, error) {
				return &command.HostingsCommand{App: *app}, nil
			},
			"info": func() (cli.Command, error) {
				return &command.InfoCommand{App: *app}, err
			},
			"link": func() (cli.Command, error) {
				return &command.LinkCommand{App: *app}, nil
			},
			"login": func() (cli.Command, error) {
				return &command.LoginCommand{App: *app}, nil
			},
			"logs": func() (cli.Command, error) {
				return &command.LogsCommand{
					App: *app,
					CacheFactory: func() (cache.Cache, error) {
						return cache.New("owh", "app.cache", nil)
					},
				}, nil
			},
			"open": func() (cli.Command, error) {
				return &command.OpenCommand{App: *app}, nil
			},
			"remove": func() (cli.Command, error) {
				return &command.RemoveCommand{App: *app}, nil
			},
			"tasks": func() (cli.Command, error) {
				return &command.TasksCommand{App: *app}, nil
			},
			"tool": func() (cli.Command, error) {
				return &command.ToolCommand{App: *app}, nil
			},
			"tool ci": func() (cli.Command, error) {
				return &command.CICommand{App: *app}, nil
			},
			"tool ssh": func() (cli.Command, error) {
				return &command.SSHCommand{App: *app}, nil
			},
			"users": func() (cli.Command, error) {
				return &command.UsersCommand{App: *app}, nil
			},
			"users changepass": func() (cli.Command, error) {
				return &command.UsersChangePassCommand{App: *app}, nil
			},
			"users remove": func() (cli.Command, error) {
				return &command.UsersRemoveCommand{App: *app}, nil
			},
			"whoami": func() (cli.Command, error) {
				return &command.WhoamiCommand{App: *app}, nil
			},
		},
	}

	code, err := cli.Run()

	if err != nil {
		fmt.Fprintf(os.Stdout, "Error executing CLI: %s\n", err)
		os.Exit(1)
	}

	os.Exit(code)
}
