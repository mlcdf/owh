package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/urfave/cli/v2"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/commands"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/sally/logging"
)

var Version = "(devel)"

func main() {
	log.SetFlags(0)

	app := cli.NewApp()

	app.Name = "owh"
	app.Usage = "Deploy to OVHcloud Web Hosting"
	app.EnableBashCompletion = true
	app.Version = Version
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			Usage:   "enable verbose output",
			Aliases: []string{"d"},
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:  "login",
			Usage: "Login to your OVHcloud account",
			Action: func(cCtx *cli.Context) error {
				return commands.Login()
			},
		},
		{
			Name:    "hosting:list",
			Aliases: []string{"hl"},
			Usage:   "List all the hostings",
			Action: func(cCtx *cli.Context) error {
				var hosting string

				if cCtx.Args().Len() == 1 {
					hosting = cCtx.Args().First()
				}

				err := config.GlobalOpts.Validate()
				if err != nil {
					return err
				}

				client, err := api.NewClient(config.GlobalOpts.Region)
				if err != nil {
					return err
				}

				return commands.Hosting(client, hosting)
			},
		},
		{
			Name:    "domains:list",
			Aliases: []string{"dl"},
			Usage:   "List attached domains",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "hosting",
					Usage:       "hosting",
					Destination: new(string),
				},
			},
			Action: func(cCtx *cli.Context) error {
				err := config.GlobalOpts.Validate()
				if err != nil {
					return err
				}

				client, err := api.NewClient(config.GlobalOpts.Region)
				if err != nil {
					return err
				}

				return commands.ListDomains(client, cCtx.String("hosting"))
			},
		},
		{
			Name:  "domains:attach",
			Usage: "Attach a domain",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "hosting",
					Usage:       "hosting",
					Destination: new(string),
				},
			},
			Action: func(cCtx *cli.Context) error {
				err := config.GlobalOpts.Validate()
				if err != nil {
					return err
				}

				client, err := api.NewClient(config.GlobalOpts.Region)
				if err != nil {
					return err
				}

				return commands.AttachDomain(client, cCtx.String("hosting"), cCtx.Args().First())
			},
		},
		{
			Name:  "domains:detach",
			Usage: "Detach a domain",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "hosting",
					Usage:       "hosting",
					Destination: new(string),
				},
			},
			Action: func(cCtx *cli.Context) error {
				err := config.GlobalOpts.Validate()
				if err != nil {
					return err
				}

				client, err := api.NewClient(config.GlobalOpts.Region)
				if err != nil {
					return err
				}

				return commands.DetachDomain(client, cCtx.String("hosting"), cCtx.Args().First())
			},
		},
		{
			Name:  "deploy",
			Usage: "Deploy the content of a folder to a website",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "dir",
					Usage:       "dir",
					Destination: new(string),
					Value:       ".",
				},
				&cli.BoolFlag{
					Name:        "www",
					Usage:       "www",
					Destination: new(bool),
					Value:       false,
				},
			},
			Action: func(cCtx *cli.Context) error {
				options := &commands.DeployOptions{
					Directory: cCtx.String("dir"),
					WWW:       cCtx.Bool("www"),
				}

				err := config.GlobalOpts.Validate()
				if err != nil {
					return err
				}

				client, err := api.NewClient(config.GlobalOpts.Region)
				if err != nil {
					return err
				}

				return commands.Deploy(client, options)
			},
		},
		{
			Name:  "open",
			Usage: "Open browser to current deployed website",
			Action: func(cCtx *cli.Context) error {
				return commands.Open()
			},
		},
		{
			Name:  "users:list",
			Usage: "List ssh/ftp users",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "hosting",
					Usage:       "hosting",
					Destination: new(string),
				},
			},
			Action: func(cCtx *cli.Context) error {
				err := config.GlobalOpts.Validate()
				if err != nil {
					return err
				}

				client, err := api.NewClient(config.GlobalOpts.Region)
				if err != nil {
					return err
				}

				return commands.Users(client, cCtx.String("hosting"))
			},
		},
		{
			Name:  "users:changepw",
			Usage: "Change ssh/ftp user password",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "hosting",
					Usage:       "hosting",
					Destination: new(string),
				},
			},
			Action: func(cCtx *cli.Context) error {
				err := config.GlobalOpts.Validate()
				if err != nil {
					return err
				}

				client, err := api.NewClient(config.GlobalOpts.Region)
				if err != nil {
					return err
				}

				return commands.ResetPassword(client, cCtx.String("hosting"), cCtx.String("user"))
			},
		},
		{
			Name:  "users:delete",
			Usage: "Delete ssh/ftp users",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "hosting",
					Usage:       "hosting",
					Destination: new(string),
				},
			},
			Action: func(cCtx *cli.Context) error {
				err := config.GlobalOpts.Validate()
				if err != nil {
					return err
				}

				client, err := api.NewClient(config.GlobalOpts.Region)
				if err != nil {
					return err
				}

				return commands.DeleteUser(client, cCtx.String("hosting"), cCtx.Args().First())
			},
		},
		{
			Name:    "remove",
			Aliases: []string{"rm"},
			Usage:   "Remove a website (files & attached domains)",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "hosting",
					Usage:       "hosting",
					Destination: new(string),
				},
			},
			Action: func(cCtx *cli.Context) error {
				err := config.GlobalOpts.Validate()
				if err != nil {
					return err
				}

				client, err := api.NewClient(config.GlobalOpts.Region)
				if err != nil {
					return err
				}

				return commands.Remove(client, cCtx.String("hosting"), cCtx.Args().First())
			},
		},
		{
			Name:    "tasks:list",
			Aliases: []string{"tasks"},
			Usage:   "List attached tasks",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "hosting",
					Usage:       "hosting",
					Destination: new(string),
				},
			},
			Action: func(cCtx *cli.Context) error {
				err := config.GlobalOpts.Validate()
				if err != nil {
					return err
				}

				client, err := api.NewClient(config.GlobalOpts.Region)
				if err != nil {
					return err
				}

				return commands.Tasks(client, cCtx.String("hosting"))
			},
		},
		{
			Name:  "whoami",
			Usage: "Shows info about the user currently logged in",
			Action: func(cCtx *cli.Context) error {
				err := config.GlobalOpts.Validate()
				if err != nil {
					return err
				}

				client, err := api.NewClient(config.GlobalOpts.Region)
				if err != nil {
					return err
				}

				return commands.Whoami(client)
			},
		},
	}

	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("debug") {
			logging.SetLevel(logging.ERROR)
		}

		return config.New()
	}

	app.ExitErrHandler = func(cCtx *cli.Context, err error) {
		if err == nil {
			os.Exit(0)
		} else if err == cmdutil.ErrSilent {
			os.Exit(1)
		} else if err == cmdutil.ErrCancel {
			os.Exit(2)
		} else if errors.Is(err, terminal.InterruptErr) {
			fmt.Fprint(os.Stderr, "\n")
			os.Exit(2)
		}

		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
