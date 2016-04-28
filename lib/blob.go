package lib

import (
	"strings"
	"log"
	"github.com/Azure/azure-sdk-for-go/storage"
	"bytes"
	"fmt"
	"os"
)

type Entity struct {
	uri              string
	isBlob           bool
	accountName      string
	accountKey       string
	container        string
	prefix           string
	entryPoint       string
	useHTTPS         bool
	connectionString string
}

func NewEntity(uri string) *Entity {
	e := Entity{uri: uri}
	e.isBlob = IsBlobURI(e.uri)
	return &Entity{e}
}

func (e *Entity) GetBody() []byte {
	if e.isBlob {
		return e.getBlobBody()
	} else {
		return e.getBody()
	}
}

func (e *Entity) getBlobBody() ([]byte, error) {
	e.ParseBlobURI()
	blobService := e.GetBlobService()
	blob, err := blobService.GetBlob(e.container, e.prefix)
	if err != nil {
		fmt.Fprintln(os.Stderr, e.prefix + ": " + err.Error())
		return nil, err
	}
	buf := new(bytes.Buffer)
	return buf.ReadFrom(blob), nil
}

func (e *Entity) getBody() []byte {

}

func (e *Entity) ParseBlobURI() {
	e.accountName, e.container, e.prefix, e.entryPoint, e.useHTTPS = ParseBlobURI(e.uri)
}

func (e *Entity) GetBlobService() *storage.BlobStorageClient {
	blobService := GetBlobService(e.accountName, e.accountKey, e.entryPoint, e.useHTTPS)
	return &storage.BlobStorageClient{blobService}
}

func ParseConnectionString(uri string) {

}

var client storage.Client

func GetBlobService(accountName, accountKey, entryPoint string, useHTTPS bool) storage.BlobStorageClient {
	var err error
	if client == nil {
		client, err = storage.NewClient(accountName, accountKey, entryPoint, "2015-02-21", useHTTPS)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}
	return client.GetBlobService()
}

func IsBlobURI(uri string) bool {
	return strings.HasPrefix(uri, "https://") || strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "wasb://") || strings.HasPrefix(uri, "wasbs://")
}

// ParseBlobURI parses blob uri and returns (accountName, container, prefix, entryPoint, useHTTPS)
func ParseBlobURI(uri string) (string, string, string, string, bool) {
	if ! IsBlobURI(uri) {
		return nil, nil, nil, nil, nil
	}
	var (
		accountName, container, prefix, entryPoint string
		useHTTPS bool
	)

	protocol := strings.SplitN(uri, ":", 2)[0]
	useHTTPS = strings.HasSuffix(protocol, "s")

	if strings.HasPrefix(protocol, "http") {
		// http[s]://myaccount.blob.core.windows.net/mycontainer/myblob?...
		u := strings.Split(uri, "/") // []string{"https:", "", "myaccount.blob.core.windows.net", "mycontainer", "myblob"} 
		container = u[3]
		if len(u) >= 5 {
			prefix = strings.SplitN(u[4], "?", 2)[0]
		}
		ud := strings.SplitN(u[2], ".", 3) // []string{"myaccount", "blob", "core.windows.net"}
		accountName = ud[0]
		entryPoint = ud[2]
	} else if strings.HasPrefix(protocol, "wasb") {
		// wasb[s]://<containername>@<accountname>.blob.core.windows.net/<path>
		u := strings.Split(uri, "/") // []string{"wasb:", "", "<containername>@<accountname>.blob.core.windows.net", "<path>"}
		if len(u) >= 4 {
			prefix = strings.SplitN(u[3], "?", 2)[0]
		}

		ua := strings.SplitN(u[2], "@", 2) // []string{"<containername>", "<accountname>.blob.core.windows.net"}
		container = ua[0]

		ud := strings.SplitN(ua[1], ".", 3) // []string{"<accountname>", "blob", "core.windows.net"}
		accountName = ud[0]
		entryPoint = ud[2]
	}

	return accountName, container, prefix, entryPoint, useHTTPS
}

func 