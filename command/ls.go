package command

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/2k0ri/blobcmd/lib"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/codegangsta/cli"
	"code.google.com/p/go.crypto/ssh/terminal"
)

func CmdLs(c *cli.Context) {
	// parse flags
	a := c.String("account-name")
	k := c.String("account-key")
	C := c.String("container")
	ep := c.String("azure-storage-entrypoint")
	p := c.String("prefix")
	u := !c.Bool("disable-https")
	r := c.Bool("recursive")

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
		b.Container = C
	}
	b.AccountKey = k
	b.UseHTTPS = u
	err = b.Validate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if b.Container != "" {
		// print header if tty
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			fmt.Println("Name\tBlobType\tLength\tContent-Type\tLast-Modified")
		}
		// @TODO separate list and asynchronous print
		_, err = lib.List(&b, p, r, true)
	} else {
		// list containers instead of blobs

		// print header if tty
		// if terminal.IsTerminal(int(os.Stdout.Fd())) {
		// 	fmt.Println("ContainerName")
		// }
		_, err = ListContainers(&b)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	// fmt.Println(strings.Join(list, "\n"))
	os.Exit(0)
}

func ListContainers(b *lib.BlobContext, ) ([]string, error) {
	var (
		m sync.RWMutex
		w sync.WaitGroup
		names []string
	)
	c, err := b.GetBlobClient()
	if err != nil {
		return names, err
	}

	p := storage.ListContainersParameters{}
	for {
		res, err := c.ListContainers(p)
		if err != nil {
			return names, err
		}

		// parse names
		w.Add(1)
		go func(blobs []storage.Container) {
			defer w.Done()
			m.Lock()
			n := make([]string, len(blobs))
			for i, blob := range blobs {
				n[i] = blob.Name
			}
			// names = append(names, n...)
			fmt.Println(strings.Join(n, "\n"))
			m.Unlock()
		}(res.Containers)

		// recursive list request
		if res.NextMarker == "" {
			break
		}
		p.Marker = res.NextMarker
	}
	w.Wait()
	return names, nil
}
