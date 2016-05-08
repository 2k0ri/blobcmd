package command

import (
	"fmt"
	"os"

	"github.com/2k0ri/blobcmd/lib"
	"github.com/codegangsta/cli"
)

/*
func copy(c *cli.Context, in, out string) error {
	if lib.IsBlobURI(in) && lib.IsBlobURI(out) {
		blobCopy(c, in, out)
		return nil
	} else if !lib.IsBlobURI(in) && !lib.IsBlobURI(out) {
		return errors.New("Invalid argument type")
	}

	var (
		i *os.File
		o *io.ReadWriter
		buf bytes.Buffer
		err error
	)

	k := c.String("account-key")
	s := !c.Bool("disable-https")
	// IB := c.Bool("input-as-blob")

	// input detection
	if in == "-" {
		i = os.Stdin
	} else if lib.IsBlobURI(in) {
		iA, iC, iP, iN, _ := lib.ParseBlobURI(in)
		ibs := lib.GetBlobClient(iA, k, iP, s)
		i, err = ibs.GetBlob(iC, iN)
		if err != nil {
			return err
		}
		defer i.Close()
		i.Read(buf)
	} else {
		i, err = os.Open(in)
		if err != nil {
			return err
		}
	}

	// output detection
	if out == "-" {
		o = os.Stdout
	} else if lib.IsBlobURI(out) {
		oA, oC, oP, oN, _ := lib.ParseBlobURI(out)
		obs := lib.GetBlobClient(oA, k, oP, s)
		err = obs.CreateBlockBlob(oC, oN)
		if err != nil {
			log.Fatal("Error: %v", err.Error())
		}
		size := uint64(len(buf))
		if size <= 64 * 1024 * 1024 {
			// lower than equal 64MB
			obs.CreateBlockBlobFromReader(oC, oN, size, i, map[string]string{})
		} else {
			cnum := size / (4 * 1024 * 1024) + 1 // 4MB
			chunk := bufio.NewReaderSize(i, 4 * 1024 * 1024)
			// @TODO chunk upload http://ibis.mach.me.ynu.ac.jp/flower150913.html
			obs.PutBlock(oC, oN, cnum, chunk.Read(buf))
		}
	} else {
		o, err = os.Open(out)
		if err != nil {
			fmt.Fprintln(os.Stderr, "%s: %s", in, err.Error())
			return
		}
	}
}
*/
func blobCopy(c *cli.Context, in, out string) {
	k := c.String("account-key")
	ok := c.String("out-account-key")
	s := !c.Bool("disable-https")
	ib, err := lib.ParseBlobURI(in)
	if err != nil {
		fmt.Fprintln(os.Stdout, "%s: error parsing uri: %s", in, err.Error())
		return
	}
	ib.AccountKey = k
	ib.UseHTTPS = s
	ibs, err := ib.GetBlobClient()
	if err != nil {
		fmt.Fprintln(os.Stdout, "%s: error getting blob client: %s", in, err.Error())
		return
	}

	ob, err := lib.ParseBlobURI(out)
	if err != nil {
		fmt.Fprintln(os.Stdout, "%s: error parsing uri: %s", out, err.Error())
		return
	}
	if ok != "" {
		ob.AccountKey = ok
	} else {
		ob.AccountKey = k
	}
	ob.UseHTTPS = s
	obs, err := ob.GetBlobClient()
	if err != nil {
		fmt.Fprintln(os.Stdout, "%s: error getting blob client: %s", out, err.Error())
		return
	}
	iN, err := lib.ParseBlobName(in)
	if err != nil {
		fmt.Fprintln(os.Stdout, "%s: error getting blob path: %s", in, err.Error())
		return
	}
	oN, err := lib.ParseBlobName(out)
	if err != nil {
		fmt.Fprintln(os.Stdout, "%s: error getting blob path: %s", out, err.Error())
		return
	}
	obs.CopyBlob(ob.Container, oN, ibs.GetBlobURL(ib.Container, iN))
	fmt.Println("%s -> %s copied", in, out)
}

func CmdCp(c *cli.Context) {
	/*
		// parse flags
		accountName := c.String("account-name")
		accountKey := c.String("account-key")
		entryPoint := c.String("azure-storage-entrypoint")
		useHTTPS := !c.Bool("disable-https")

		recursive := c.Bool("recursive")
		stdinAsList := c.Bool("stdin-as-list")
		inputAsBlobPaths := c.Bool("input-as-blobpaths")
		// recursive := c.Bool("recursive")

		fmt.Println(useHTTPS)

		// get blob client
		client, err := storage.NewClient(accountName, accountKey, entryPoint, "2015-02-21", useHTTPS)
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

		infileNames := c.Args()[:c.NArg() - 1]
		for _, infileName := range infileNames {
			if infileName == "-" {
				if stdinAsList {

				}
			} else if lib.IsBlobURI(infileName) {

			} else if inputAsBlobPaths {

			}
		}

		// execute in parallel
		sem := make(chan bool, runtime.NumCPU())
		var (
			wg sync.WaitGroup
			m sync.RWMutex
		)

		s := bufio.NewScanner(in)
		for s.Scan() {
			l := s.Text()

			wg.Add(1)
			go func(l string) {
				sem <- true
				defer wg.Done()

				blob, err := blobService.GetBlob(container, l)
				defer blob.Close()
				if err != nil {
					fmt.Fprintln(os.Stderr, l + ": " + err.Error())
					return
				}

				buf := new(bytes.Buffer)
				buf.ReadFrom(blob)

				defer m.Unlock()

				// output detection
				var out *os.File
				outfileName := c.Args().Get(c.NArg() - 1)
				if recursive {
					// @TODO recursive filepath
					if ! strings.HasSuffix(outfileName, "/") {
						outfileName += "/"
					}
					inl := strings.Split(infileName, "/")
					outfileName = outfileName + inl[len(inl) - 1]
				}
				if outfileName == "-" {
					out = os.Stdout
				} else if lib.IsBlobURI(outfileName) {

				} else {
					if !recursive {
						out, err = os.Open(outfileName)
						if err != nil {
							log.Fatal("Error: %v", err.Error())
						}
					} else {
						// @TODO recursive filepath
						if ! strings.HasSuffix(outfileName, "/") {
							outfileName += "/"
						}
						inl := strings.Split(infileName, "/")
						outpath := outfileName + inl[len(inl) - 1]
					}
				}
				m.Lock()
				out.Write(buf.Bytes())
				<-sem
			}(l)
		}
		if s.Err() != nil {
			// non-EOF error.
			log.Fatal(s.Err())
		}
		wg.Wait()
		fmt.Println(blobService.BlobExists(container, "test"))
	*/
}
