package lib

import (
	"github.com/ToQoz/gopwt/assert"
	"testing"
)

func TestParseBlobURIHTTP(t *testing.T) {
	b, err := ParseBlobURI("https://myaccount.blob.core.windows.net/mycontainer/myblob/a/b?c=d")
	if err != nil {
		t.Fail()
	}
	assert.OK(t, b.AccountName == "myaccount", "AccountName parsed")
	assert.OK(t, b.Container == "mycontainer", "AccountContainer parsed")
	assert.OK(t, b.EntryPoint == "core.windows.net", "Entrypoint parsed")
	assert.OK(t, b.UseHTTPS == true, "UseHTTPS parsed")
}

func TestParseBlobURIWASB(t *testing.T) {
	b, err := ParseBlobURI("wasbs://<containername>@<accountname>.blob.core.windows.net/<path>/a/b?c=d")
	if err != nil {
		t.Fail()
	}
	assert.OK(t, b.AccountName == "<accountname>", "AccountName parsed")
	assert.OK(t, b.Container == "<containername>", "AccountContainer parsed")
	assert.OK(t, b.EntryPoint == "core.windows.net", "Entrypoint parsed")
	assert.OK(t, b.UseHTTPS == true, "UseHTTPS parsed")
}

func TestParseBlobURIUseHTTPS(t *testing.T) {
	b, err := ParseBlobURI("http://myaccount.blob.core.windows.net/mycontainer/myblob/a/b?c=d")
	if err != nil {
		t.Fail()
	}
	assert.OK(t, b.UseHTTPS == false, "UseHTTPS parsed from http://")
	b, err = ParseBlobURI("wasb://<containername>@<accountname>.blob.core.windows.net/<path>")
	if err != nil {
		t.Fail()
	}
	assert.OK(t, b.UseHTTPS == false, "UseHTTPS parsed from wasb://")
}

func TestParseBlobName(t *testing.T) {
	n, err := ParseBlobName("https://myaccount.blob.core.windows.net/mycontainer/myblob/a/b?c=d")
	if err != nil {
		t.Fail()
	}
	assert.OK(t, n == "myblob/a/b")
	n, err = ParseBlobName("wasbs://<containername>@<accountname>.blob.core.windows.net/<path>/a/b?c=d")
	if err != nil {
		t.Fail()
	}
	assert.OK(t, n == "<path>/a/b")
}

// func TestList(t *testing.T) {
// 	a := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
// 	k := os.Getenv("AZURE_STORAGE_ACCECSS_KEY")
// 	b := BlobContext{AccountName:a, AccountKey:k, Container:"zzzz-test", EntryPoint:"core.windows.net", UseHTTPS:true}
// }
