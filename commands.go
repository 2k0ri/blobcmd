package main

import (
	"fmt"
	"os"

	"github.com/2k0ri/blobcmd/command"
	"github.com/codegangsta/cli"
)

var GlobalFlags = []cli.Flag{
	cli.StringFlag{
		EnvVar: "AZURE_STORAGE_PREFIX",
		Name:   "prefix, p",
		Value:  "",
		Usage:  "prefix",
	},
	cli.StringFlag{
		EnvVar: "AZURE_STORAGE_ACCOUNT",
		Name:   "account-name, a",
		Value:  "",
		Usage:  "account-name",
	},
	cli.StringFlag{
		EnvVar: "AZURE_STORAGE_CONTAINER",
		Name:   "container, c",
		Value:  "",
		Usage:  "container",
	},
	cli.StringFlag{
		EnvVar: "AZURE_STORAGE_CONNECTION_STRING",
		Name:   "connection-string, C",
		Value:  "",
		Usage:  "connection string",
	},
	cli.StringFlag{
		EnvVar: "AZURE_STORAGE_ACCESS_KEY",
		Name:   "account-key, k",
		Value:  "",
		Usage:  "account key",
	},
	cli.StringFlag{
		EnvVar: "AZURE_STORAGE_ENTRYPOINT",
		Name:   "azure-storage-entrypoint",
		Value:  "core.windows.net",
		Usage:  "azure storage entrypoint",
	},
	cli.BoolFlag{
		EnvVar: "AZURE_DISABLE_HTTPS",
		Name:   "disable-https",
		Usage:  "disable HTTPS",
	},
}

var Commands = []cli.Command{
	{
		Name:    "ls",
		Aliases: []string{"list"},
		Usage:   "list",
		Action:  command.CmdLs,
		Flags:   GlobalFlags,
	},
	{
		Name:    "cp",
		Aliases: []string{"copy"},
		Usage:   "copy INFILE... OUTFILE",
		Action:  command.CmdCp,
		Flags:   append(GlobalFlags,
			cli.BoolFlag{
				Name:   "recursive, r",
				Usage:  "copy recursively(src and dest must be directory)",
			},
			cli.BoolFlag{
				Name:   "input-as-blob",
				Usage:  "",
			},
			cli.BoolFlag{
				Name:   "output-as-blob",
				Usage:  "",
			},
			cli.BoolFlag{
				Name:   "concat",
				Usage:  "concat INFILEs and upload as one OUTFILE",
			},
		),
	},
	{
		Name:    "rm",
		Aliases: []string{"remove"},
		Usage:   "remove",
		Action:  command.CmdRm,
		Flags:   append(GlobalFlags,
			cli.BoolFlag{
				Name:   "recursive, r",
				Usage:  "remove recursively",
			},
		),
	},
	{
		Name:    "sync",
		Usage:   "sync",
		Action:  command.CmdSync,
		Flags:   GlobalFlags,
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
