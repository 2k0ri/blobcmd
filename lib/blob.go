package lib

import (
	"strings"
	"github.com/Azure/azure-sdk-for-go/storage"
	"sync"
	"errors"
	"fmt"
)

type BlobContext struct {
	AccountName string
	AccountKey  string
	Container   string
	EntryPoint  string
	UseHTTPS    bool
}

var clients map[BlobContext]*storage.BlobStorageClient

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

func IsBlobURI(uri string) bool {
	return strings.HasPrefix(uri, "https://") || strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "wasb://") || strings.HasPrefix(uri, "wasbs://")
}

// ParseBlobURI parses blob uri and returns BlobContext
func ParseBlobURI(uri string) (BlobContext, error) {
	var b BlobContext
	if ! IsBlobURI(uri) {
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
		u := strings.Split(uri, "/") // []string{"wasb:", "", "<containername>@<accountname>.blob.core.windows.net", "<path>"}
		ua := strings.SplitN(u[2], "@", 2) // []string{"<containername>", "<accountname>.blob.core.windows.net"}
		b.Container = ua[0]
		ud := strings.SplitN(ua[1], ".", 3) // []string{"<accountname>", "blob", "core.windows.net"}
		b.AccountName = ud[0]
		b.EntryPoint = ud[2]
	}

	return b, nil
}

func ParseBlobName(uri string) (string, error) {
	if ! IsBlobURI(uri) {
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
	u := strings.SplitN(uri, "/", i + 1)
	return strings.SplitN(u[i], "?", 2)[0], nil
}

func List(b *BlobContext, prefix string, r bool) ([]string, error) {
	var (
		m sync.RWMutex
		w sync.WaitGroup
		names []string
	)
	c, err := b.GetBlobClient()
	if err != nil {
		return names, err
	}

	p := storage.ListBlobsParameters{Prefix: prefix}
	if !r {
		p.Delimiter = "/"
	}

	for {
		res, err := c.ListBlobs(b.Container, p)
		if err != nil {
			return names, err
		}

		// If not recursive, list blob prefixes
		if !r {
			w.Add(1)
			go func(prefixes []string) {
				defer w.Done()
				m.Lock()
				n := make([]string, len(prefixes))
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
