package lib

import "testing"

func TestDetectMime(t *testing.T) {
	cases := map[string]string{
		"csv":  "text/csv",
		"json": "application/json",
		"avro": "avro/binary",
		"gz":   "application/gzip",
		"html": "text/html",
		"jpg":  "image/jpeg",
		"":     "application/octet-stream",
	}

	for k, expected := range cases {
		actual := detectMime(k)
		if actual != expected {
			t.Errorf("actual: %v expected: %v", actual, expected)
		}
	}
}
