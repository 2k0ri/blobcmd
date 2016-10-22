// port/forked from github.com/taskgraph/taskgraph/filesystem/azure.go

package lib

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/storage"
)

type AzureClient struct {
	client     *storage.Client
	blobClient storage.BlobStorageClient
}

// AzureFile implements io.WriteCloser
type AzureFile struct {
	path   string
	client storage.BlobStorageClient
}

// convertToAzurePath function
// convertToAzurePath splits the given name into two parts
// The first part represents the container's name, and the length of it shoulb be 32 due to Azure restriction
// The second part represents the blob's name
// It will return any error while converting
func convertToAzurePath(name string) (string, string, error) {
	afterSplit := strings.Split(name, "/")
	if len(afterSplit[0]) > 32 {
		return "", "", fmt.Errorf("azureClient : the length of container must be shorter than equal 32")
	}
	blobName := ""
	if len(afterSplit) > 1 {
		blobName = name[len(afterSplit[0])+1:]
	}
	return afterSplit[0], blobName, nil
}

// Remove function
// Delete specific Blob for input path
func (c *AzureClient) Remove(name string) error {
	afterSplit := strings.Split(name, "/")
	if len(afterSplit) == 1 && len(afterSplit[0]) == 32 {
		_, err := c.blobClient.DeleteContainerIfExists(name)
		if err != nil {
			return err
		}
		return nil
	}
	containerName, blobName, err := convertToAzurePath(name)
	if err != nil {
		return err
	}
	_, err = c.blobClient.DeleteBlobIfExists(containerName, blobName, nil)
	return err

}

// Exists function
// support check the contianer or blob if exist or not
func (c *AzureClient) Exists(name string) (bool, error) {
	containerName, blobName, err := convertToAzurePath(name)
	if err != nil {
		return false, err
	}
	if blobName != "" {
		return c.blobClient.BlobExists(containerName, blobName)
	} else {
		return c.blobClient.ContainerExists(containerName)
	}
}

// Azure prevent user renaming their blob
// Thus this function firstly copy the source blob,
// when finished, delete the source blob.
// http://stackoverflow.com/questions/3734672/azure-storage-blob-rename
func (c *AzureClient) moveBlob(dstContainerName, dstBlobName, srcContainerName, srcBlobName string, isContainerRename bool) error {
	dstBlobUrl := c.blobClient.GetBlobURL(dstContainerName, dstBlobName)
	srcBlobUrl := c.blobClient.GetBlobURL(srcContainerName, srcBlobName)
	if dstBlobUrl != srcBlobUrl {
		err := c.blobClient.CopyBlob(dstContainerName, dstBlobName, srcBlobUrl)
		if err != nil {
			return err
		}
		if !isContainerRename {
			err = c.blobClient.DeleteBlob(srcContainerName, srcBlobName, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AzureClient -> Rename function
// support user renaming contianer/blob
func (c *AzureClient) Rename(oldpath, newpath string) error {
	exist, err := c.Exists(oldpath)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("azureClient : oldpath does not exist")
	}
	srcContainerName, srcBlobName, err := convertToAzurePath(oldpath)
	if err != nil {
		return err
	}
	dstContainerName, dstBlobName, err := convertToAzurePath(newpath)
	if err != nil {
		return err
	}
	if srcBlobName == "" && dstBlobName == "" {
		resp, err := c.blobClient.ListBlobs(srcContainerName, storage.ListBlobsParameters{Marker: ""})
		if err != nil {
			return err
		}
		_, err = c.blobClient.CreateContainerIfNotExists(dstContainerName, storage.ContainerAccessTypeBlob)
		if err != nil {
			return err
		}

		for _, blob := range resp.Blobs {
			err = c.moveBlob(dstContainerName, blob.Name, srcContainerName, blob.Name, true)
			if err != nil {
				return err
			}
		}
		err = c.blobClient.DeleteContainer(srcContainerName)
		if err != nil {
			return err
		}
	} else if srcBlobName != "" && dstBlobName != "" {
		c.moveBlob(dstContainerName, dstBlobName, srcContainerName, srcBlobName, false)
	} else {
		return fmt.Errorf("Rename path does not match")
	}

	return nil
}

// OpenReadCloser function
// implement by the providing function
func (c *AzureClient) OpenReadCloser(name string) (io.ReadCloser, error) {
	containerName, blobName, err := convertToAzurePath(name)
	if err != nil {
		return nil, err
	}
	return c.blobClient.GetBlob(containerName, blobName)
}

// OpenWriteCloser function
// If not exist, Create corresponding Container and blob.
func (c *AzureClient) OpenWriteCloser(name string) (io.WriteCloser, error) {
	exist, err := c.Exists(name)
	if err != nil {
		return nil, err
	}
	containerName, blobName, err := convertToAzurePath(name)
	if err != nil {
		return nil, err
	}
	if !exist {
		_, err = c.blobClient.CreateContainerIfNotExists(containerName, storage.ContainerAccessTypeBlob)
		if err != nil {
			return nil, err
		}
		ctype := detectMime(blobName)
		err = c.blobClient.CreateBlockBlobFromReader(containerName, blobName, 0, nil, map[string]string{"Content-Type": ctype, "x-ms-blob-content-type": ctype})
		if err != nil {
			return nil, err
		}
		err = c.blobClient.SetBlobProperties(containerName, blobName, storage.BlobHeaders{ContentType: ctype})
		if err != nil {
			return nil, err
		}
	} else {
		log.Printf("%s exists", blobName)
	}
	return &AzureFile{
		path:   name,
		client: c.blobClient,
	}, nil
}

// Follow io.Writer rules
// use BlockList mechanism of Azure Blob to achieve this interface
// BlockList divided block into committed block and uncommitted ones
// For every write op of user,
// it create new block list with old commited block and  user write content,
// and commit this one to modifies the blob commited block list
// If failed in the process, it would write nothing to blob
// return 0, error to user
func (f *AzureFile) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, fmt.Errorf("Need write content")
	}
	cnt, blob, err := convertToAzurePath(f.path)
	if err != nil {
		return 0, err
	}

	blockList, err := f.client.GetBlockList(cnt, blob, storage.BlockListTypeAll)
	if err != nil {
		return 0, err
	}

	// blockLen is a naming rule for block,
	// uses base64.StdEncoding.EncodeToString(fmt.Sprintf("%011d\n", blocksLen)))

	// blockLen initially set to committed block size
	// if exist uncommitted block of same name,
	// it will rewrite the uncommitted block content
	blocksLen := len(blockList.CommittedBlocks)
	amendList := []storage.Block{}
	for _, v := range blockList.CommittedBlocks {
		amendList = append(amendList, storage.Block{v.Name, storage.BlockStatusCommitted})
	}

	var chunkSize int = storage.MaxBlobBlockSize
	inputSourceReader := bytes.NewReader(b)
	chunk := make([]byte, chunkSize)
	for {
		n, err := inputSourceReader.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

		blockId := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%011d\n", blocksLen)))
		data := chunk[:n]
		err = f.client.PutBlock(cnt, blob, blockId, data)
		if err != nil {
			return 0, err
		}
		// Add current uncommitted block to temporary block list
		amendList = append(amendList, storage.Block{blockId, storage.BlockStatusUncommitted})
		blocksLen++
	}
	// update block list ot blob committed block list.
	err = f.client.PutBlockList(cnt, blob, amendList)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (f *AzureFile) Close() error {
	return nil
}

// AzureClient -> Glob function
// only supports '*', '?'
// Syntax:
// cntName?/part.*
func (c *AzureClient) Glob(pattern string) (matches []string, err error) {
	afterSplit := strings.Split(pattern, "/")
	cntPattern, blobPattern := afterSplit[0], afterSplit[1]
	if len(afterSplit) != 2 {
		return nil, fmt.Errorf("Glob pattern should follow the Syntax")
	}
	resp, err := c.blobClient.ListContainers(storage.ListContainersParameters{Prefix: ""})
	if err != nil {
		return nil, err
	}
	for _, cnt := range resp.Containers {
		matched, err := path.Match(cntPattern, cnt.Name)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		resp, err := c.blobClient.ListBlobs(cnt.Name, storage.ListBlobsParameters{Marker: ""})
		if err != nil {
			return nil, err
		}
		for _, v := range resp.Blobs {
			matched, err := path.Match(blobPattern, v.Name)
			if err != nil {
				return nil, err
			}
			if matched {
				matches = append(matches, cnt.Name+"/"+v.Name)
			}
		}
	}
	return matches, nil
}

// list inputs storage.BlobListResponse into provided channel
// it covers recursive traverse within azure storage blob
func (c *AzureClient) list(container, prefix string, recursive bool, ch chan storage.BlobListResponse) {
	param := storage.ListBlobsParameters{Prefix: prefix}
	if !recursive {
		param.Delimiter = "/"
	}
	defer close(ch)
	// loop until NextMarker ends
	for {
		res, err := c.blobClient.ListBlobs(container, param)
		if err != nil {
			// @TODO genuine error handling
			log.Println(err)
			continue
		}
		ch <- res
		if res.NextMarker == "" {
			break
		}
		param.Marker = res.NextMarker
	}
}

// ListAndPrint outputs print blobs to Stdout
// @TODO stop when container not found
// @TODO finish channel
func (c *AzureClient) ListAndPrint(container, prefix string, recursive bool, w *sync.WaitGroup) {
	lr := make(chan storage.BlobListResponse, 1024)
	w.Add(2)
	go func() {
		defer w.Done()
		c.list(container, prefix, recursive, lr)
	}()
	go func() {
		defer w.Done()
		FprintBlobList(lr, recursive, os.Stdout)
	}()
	w.Wait()
}

// Find returns AzureFile channel
func (c *AzureClient) Find(b <-chan storage.Blob) <-chan AzureFile {
	ch := make(chan AzureFile, 5000)
	go func() {
		defer close(ch)
	loop:
		for {
			select {
			case blob, ok := <-b:
				if !ok {
					break loop
				}
				ch <- AzureFile{
					path:   blob.Name,
					client: c.blobClient,
				}
			default:
				// channel blocking
			}
		}
	}()
	return ch
}

// SprintBlobListCh returns string channel from storage.BlobListResponse channel
func SprintBlobListCh(lr <-chan storage.BlobListResponse, recursive bool, out chan string) {
	defer close(out)
loop:
	for {
		select {
		case listRes, ok := <-lr:
			if !ok {
				break loop
			}
			if !recursive {
				for _, p := range listRes.BlobPrefixes {
					out <- p
				}
			}
			for _, b := range listRes.Blobs {
				out <- SprintBlob(b)
			}
		default:
			// channel blocking
		}
	}
}

// FprintBlobList returns string channel from storage.BlobListResponse channel
func FprintBlobList(lr <-chan storage.BlobListResponse, recursive bool, out io.Writer) {
loop:
	for {
		select {
		case listRes, ok := <-lr:
			if !ok {
				break loop
			}
			if !recursive {
				for _, p := range listRes.BlobPrefixes {
					fmt.Fprintln(out, p)
				}
			}
			for _, b := range listRes.Blobs {
				fmt.Fprintln(out, SprintBlob(b))
			}
		default:
			// channel blocking
		}
	}
}
func SprintBlobCh(b <-chan storage.Blob) <-chan string {
	ch := make(chan string, 5000)
	go func() {
		defer close(ch)
	loop:
		for {
			select {
			case b, ok := <-b:
				if !ok {
					break loop
				}
				ch <- SprintBlob(b)
			default:
				// channel blocking
			}
		}
	}()
	return ch
}

// SprintBlob returns string from storage.Blob with template
// @TODO pluggable template
func SprintBlob(b storage.Blob) string {
	return fmt.Sprintf("%v\t%v\t%v\t%v\t%v", b.Name, b.Properties.BlobType, b.Properties.ContentLength, b.Properties.ContentType, b.Properties.LastModified)
}

// NewAzureClient function
// NewClient constructs a StorageClient and blobStorageClient.
// This should be used if the caller wants to specify
// whether to use HTTPS, a specific REST API version or a
// custom storage endpoint than Azure Public Cloud.
func NewAzureClient(accountName, accountKey, baseURL string, useHTTPS bool) (*AzureClient, error) {
	cli, err := storage.NewClient(accountName, accountKey, baseURL, storage.DefaultAPIVersion, useHTTPS)
	if err != nil {
		return nil, err
	}
	return &AzureClient{
		client:     &cli,
		blobClient: cli.GetBlobService(),
	}, nil
}

// detectMime returns Content-Type string from extension
func detectMime(blobName string) string {
	_e := strings.Split(blobName, ".")
	ext := strings.ToLower(_e[len(_e)-1])
	var ctype string
	switch ext {
	case "csv":
		ctype = "text/csv"
	case "json":
		ctype = "application/json"
	case "avro":
		ctype = "avro/binary"
	case "gz":
		ctype = "application/gzip"
	default:
		// guess by "mime" package
		ctype = mime.TypeByExtension("." + ext)
		if ctype == "" {
			ctype = "application/octet-stream"
		}
		ctype = strings.Split(ctype, ";")[0]
	}
	// log.Printf("Content-Type: %v\n", ctype)
	return ctype
}
