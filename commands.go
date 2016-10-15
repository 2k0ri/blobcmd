package main

import (
	"fmt"
	"os"

	"github.com/2k0ri/blobcmd/command"
	"github.com/codegangsta/cli"
	"runtime"
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
		Name:   "container, C",
		Value:  "",
		Usage:  "container",
	},
	cli.StringFlag{
		EnvVar: "AZURE_STORAGE_CONNECTION_STRING",
		Name:   "connection-string, c",
		Value:  "",
		Usage:  "connection string(under construction)",
	},
	cli.StringFlag{
		EnvVar: "AZURE_STORAGE_SAS",
		Name:   "sas, s",
		Value:  "",
		Usage:  "shared access signature(under construction)",
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
	cli.IntFlag{
		EnvVar: "AZURE_PARALLELISM",
		Name:   "parallelism, P",
		Value:  runtime.NumCPU(),
		Usage:  "parallelism for goroutines",
	},
}

var Commands = []cli.Command{
	{
		Name:    "ls",
		Aliases: []string{"list"},
		Usage:   "list",
		Action:  command.CmdLs,
		Flags: append(GlobalFlags,
			cli.BoolFlag{
				Name:  "recursive, r",
				Usage: "list recursively",
			},
		),
	},
	{
		Name:    "cp",
		Aliases: []string{"copy"},
		Usage:   "copy INFILE... OUTFILE",
		Action:  command.CmdCp,
		Flags: append(GlobalFlags,
			cli.BoolFlag{
				Name:  "recursive, r",
				Usage: "copy recursively(src and dest must be directory)",
			},
			cli.BoolFlag{
				Name:  "input-as-blob",
				Usage: "",
			},
			cli.BoolFlag{
				Name:  "stdin-as-list",
				Usage: "",
			},
			cli.StringFlag{
				Name:  "out-account-key",
				Usage: "account key for output blob",
			},
			cli.BoolFlag{
				Name:  "concat",
				Usage: "concat INFILEs and upload as one OUTFILE",
			},
			cli.BoolFlag{
				Name:  "dryrun, dry-run, n",
				Usage: "dry-run",
			},
		),
	},
	{
		Name:    "rm",
		Aliases: []string{"remove"},
		Usage:   "remove",
		Action:  command.CmdRm,
		Flags: append(GlobalFlags,
			cli.BoolFlag{
				Name:  "recursive, r",
				Usage: "remove recursively",
			},
			cli.BoolFlag{
				Name:  "dryrun, dry-run, n",
				Usage: "dry-run",
			},
		),
	},
	{
		Name:   "sync",
		Usage:  "sync",
		Action: command.CmdSync,
		Flags:  GlobalFlags,
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
