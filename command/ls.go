package command

import (
	"github.com/codegangsta/cli"
	"github.com/2k0ri/blobcmd/lib"
	"os"
	"fmt"
)

func CmdLs(c *cli.Context) {
	// parse flags
	a := c.String("account-name")
	k := c.String("account-key")
	co := c.String("container")
	ep := c.String("azure-storage-entrypoint")
	p := c.String("prefix")
	u := !c.Bool("disable-https")

	var (
		path string
		b lib.BlobContext
		err error
	)
	if c.NArg() >= 1 {
		path = c.Args()[0]
		b, err = lib.ParseBlobURI(path)
		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
			os.Exit(1)
		}
		p, err = lib.ParseBlobName(path)
		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
			os.Exit(1)
		}
	} else {
		b.AccountName = a
		b.EntryPoint = ep
		b.Container = co
	}
	b.AccountKey = k
	b.UseHTTPS = u
	err = b.Validate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if b.Container == "" {
		_, err = lib.ListContainers(&b)
	}
	// list, err := lib.List(&b, p)
	_, err = lib.List(&b, p)
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
		os.Exit(1)
	}
	// fmt.Println(strings.Join(list, "\n"))
	os.Exit(0)
}
