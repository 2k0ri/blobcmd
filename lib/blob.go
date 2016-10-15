package lib

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/storage"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

type BlobContext struct {
	AccountName string
	AccountKey  string
	Container   string
	EntryPoint  string
	UseHTTPS    bool
}

var clients map[BlobContext]*storage.BlobStorageClient

type AzureClient struct {
	client     *storage.Client
	blobClient storage.BlobStorageClient
}

type AzureFile struct {
	path   string
	logger *log.Logger
	client storage.BlobStorageClient
}

func (b *BlobContext) GetBlobClient() (*storage.BlobStorageClient, error) {
	if clients == nil {
		clients = map[BlobContext]*storage.BlobStorageClient{}
	}
	if clients[*b] == nil {
		client, err := GetBlobClient(b.AccountName, b.AccountKey, b.EntryPoint, b.UseHTTPS)
		if err != nil {
			return nil, err
		}
		clients[*b] = &client
	}
	return clients[*b], nil
}

func (b *BlobContext) Validate() error {
	var e string
	if b.AccountName == "" {
		e += "AccountName is required\n"
	}
	if b.AccountKey == "" {
		e += "AccountKey is required\n"
	}
	if b.EntryPoint == "" {
		e += "EntryPoint is required\n"
	}
	if e != "" {
		return errors.New(e)
	}
	return nil
}

func (b *BlobContext) ValidateWithContainer() error {
	var e string
	err := b.Validate()
	if err != nil {
		e += err.Error()
	}
	if b.Container == "" {
		e += "Container is required\n"
	}
	if e != "" {
		return errors.New(e)
	}
	return nil
}

func GetBlobClient(accountName, accountKey, entryPoint string, useHTTPS bool) (storage.BlobStorageClient, error) {
	client, err := storage.NewClient(accountName, accountKey, entryPoint, "2015-02-21", useHTTPS)
	if err != nil {
		return storage.BlobStorageClient{}, err
	}
	return client.GetBlobService(), nil
}

func isBlobURI(uri string) bool {
	return strings.HasPrefix(uri, "https://") || strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "wasb://") || strings.HasPrefix(uri, "wasbs://")
}

// ParseBlobURI parses blob uri and returns BlobContext
func ParseBlobURI(uri string) (BlobContext, error) {
	var b BlobContext
	if !isBlobURI(uri) {
		return b, errors.New("not an blob")
	}

	protocol := strings.SplitN(uri, ":", 2)[0]
	b.UseHTTPS = strings.HasSuffix(protocol, "s")

	if strings.HasPrefix(protocol, "http") {
		// http[s]://myaccount.blob.core.windows.net/mycontainer/myblob?...
		u := strings.Split(uri, "/") // []string{"https:", "", "myaccount.blob.core.windows.net", "mycontainer", "myblob"}
		b.Container = u[3]
		ud := strings.SplitN(u[2], ".", 3) // []string{"myaccount", "blob", "core.windows.net"}
		b.AccountName = ud[0]
		b.EntryPoint = ud[2]
	} else if strings.HasPrefix(protocol, "wasb") {
		// wasb[s]://<containername>@<accountname>.blob.core.windows.net/<path>
		u := strings.Split(uri, "/")       // []string{"wasb:", "", "<containername>@<accountname>.blob.core.windows.net", "<path>"}
		ua := strings.SplitN(u[2], "@", 2) // []string{"<containername>", "<accountname>.blob.core.windows.net"}
		b.Container = ua[0]
		ud := strings.SplitN(ua[1], ".", 3) // []string{"<accountname>", "blob", "core.windows.net"}
		b.AccountName = ud[0]
		b.EntryPoint = ud[2]
	}

	return b, nil
}

func ParseBlobName(uri string) (string, error) {
	if !isBlobURI(uri) {
		return "", errors.New("not an blob")
	}
	var i int
	if strings.HasPrefix(uri, "http") {
		// http[s]://myaccount.blob.core.windows.net/mycontainer/myblob?...
		i = 4
	} else {
		// wasb[s]://<containername>@<accountname>.blob.core.windows.net/<path>
		i = 3
	}
	u := strings.SplitN(uri, "/", i+1)
	return strings.SplitN(u[i], "?", 2)[0], nil
}

func List(b *BlobContext, prefix string, recursive bool, print bool) ([]string, error) {
	var (
		m     sync.RWMutex
		w     sync.WaitGroup
		names []string
	)
	c, err := b.GetBlobClient()
	if err != nil {
		return names, err
	}

	p := storage.ListBlobsParameters{Prefix: prefix}
	if !recursive {
		p.Delimiter = "/"
	}

	for {
		res, err := c.ListBlobs(b.Container, p)
		if err != nil {
			return names, err
		}

		// If not recursive, list blob prefixes
		if !recursive {
			w.Add(1)
			go func(prefixes []string) {
				defer w.Done()
				m.Lock()
				n := make([]string, len(prefixes))
				// @TODO remove duplicate prefixes
				for i, prefix := range prefixes {
					n[i] = prefix
				}

				// @TODO separate list and asynchronous print
				if len(n) > 0 {
					fmt.Println(strings.Join(n, "\n"))
				}
				m.Unlock()
			}(res.BlobPrefixes)
		}

		// list items
		w.Add(1)
		go func(blobs []storage.Blob) {
			defer w.Done()
			m.Lock()
			n := make([]string, len(blobs))
			for i, blob := range blobs {
				// @FIXME BlobType does not appeared(Casting issue?)
				n[i] = fmt.Sprintf("%s\t%s\t%d\t%s\t%s", blob.Name, blob.Properties.BlobType, blob.Properties.ContentLength, blob.Properties.ContentType, blob.Properties.LastModified)
			}

			// @TODO separate list and asynchronous print
			if len(n) > 0 {
				fmt.Println(strings.Join(n, "\n"))
			}
			m.Unlock()
		}(res.Blobs)

		// recursive list request
		if res.NextMarker == "" {
			break
		}
		p.Marker = res.NextMarker
	}
	w.Wait()
	return names, nil
}

func Copy(bC *BlobContext, src storage.Blob, dest os.File) (int64, error) {
	r, err := GetReader(bC, src.Name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	defer dest.Close()

	return io.Copy(dest, r)
}

func GetReader(bc *BlobContext, src string) (io.ReadCloser, error) {
	c, err := bc.GetBlobClient()
	if err != nil {
		return nil, err
	}
	return c.GetBlob(bc.Container, src)
}

func Upload(bc *BlobContext, src io.Reader, dst string) {
	c, err := bc.GetBlobClient()
	if err != nil {
		return nil, err
	}

	var (
		m sync.RWMutex
		w sync.WaitGroup
	)
	// fi, err := src.Stat()
	// if err != nil {
	// 	return nil, err
	// }

	// size := fi.Size()
	// if size <= 64 * 1024 * 1024 {
	// 	// if file is smaller than equal 64MiB, it can be put with single request
	// 	hash := md5.New()
	// 	md, err := io.Copy(hash, src)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	id := base64.URLEncoding.EncodeToString(md)
	// 	c.PutBlockWithLength(bc.Container, dst, id, size, src, map[string]string{})
	// } else {
	// if file is larger than 64MiB

	s := bufio.NewReaderSize(src, 4*1024*1024)
	buf := make([]byte, 4*1024*1024)

	for {
		chunk, er := s.Read(buf)
		if chunk > 0 {
			w.Add(1)
			go func(chunk []byte) {
				hash := md5.New()
				md, err := io.Copy(hash, src)
				if err != nil {
					return nil, err
				}
				id := base64.URLEncoding.EncodeToString(md)
				err = c.PutBlock(bc.Container, dst, id, chunk)
				if err != nil {
					return nil, err
				}
			}(s.Read(buf))
			nw, ew := dst.Write(buf[0:chunk])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if chunk != nw {
				err = ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return written, err

	// }
}
