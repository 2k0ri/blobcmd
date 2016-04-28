package command

import (
	"github.com/codegangsta/cli"
	"log"
	"github.com/Azure/azure-sdk-for-go/storage"
	"fmt"
	"os"
	"github.com/2k0ri/blobcmd/lib"
)

func CmdCp(c *cli.Context) {
	// parse flags
	accountName := c.String("account-name")
	accountKey := c.String("account-key")
	container := c.String("container")
	entrypoint := c.String("azure-storage-entrypoint")
	useHTTPS := !c.Bool("disable-https")
	// recursive := c.Bool("recursive")

	fmt.Println(useHTTPS)

	// get blob client
	client, err := storage.NewClient(accountName, accountKey, entrypoint, "2015-02-21", useHTTPS)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	blobService := client.GetBlobService()

	// parse arguments
	if c.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "too few arguments")
		os.Exit(1)
	} else if c.NArg() >= 3 {
		fmt.Fprintln(os.Stderr, "too much arguments")
		os.Exit(1)
	}

	var outfile *os.File
	outfileName := c.Args().Get(c.NArg() - 1)
	if outfileName == "-" {
		outfile := os.Stdout
	} else if lib.IsBlobURI(outfile) {
		
	} else {
		
	}

	infileName := c.Args().First()
	if infileName == "-" {
		infile := os.Stdin
	} else if lib.IsBlobURI(infileName) {
		iAccountName, iContainer, iPrefix, iEntrypoint, _ := ParseBlobURI(infileName)
		iBlobService := GetBlobService(iAccountName, accountKey, iEntrypoint, useHTTPS)
	} else {

	}

	fmt.Println(blobService.BlobExists(container, "test"))
}
