package command

import (
	"testing"
)

func TestCmdLs(t *testing.T) {
	// Write your code here
}

func BenchmarkListR(b *testing.B) {
	list := make([]string, 300)
	for i, _ := range list {
		list[i] = "a"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		listR(list)
	}
}

func BenchmarkListJ(b *testing.B) {
	list := make([]string, 300)
	for i, _ := range list {
		list[i] = "a"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		listJ(list)
	}
}
