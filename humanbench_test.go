package main

import (
	"go/build"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/tools/godoc"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/gatefs"
)

var sample = `
BenchmarkIndexRune-4             	50000000	         22.9ns/op	  43.69MB/s 	    0.00B/op	          0.00/op
BenchmarkIndexRuneLongString-4   	50000000	         26.2ns/op	    0.00B/op	          0.00/op
BenchmarkNewDirectory-4          	       5	          319ms/op	    227MB/op	          513k/op
`

func TestBenchmarkFmt(t *testing.T) {
	text := "BenchmarkIndexRune-4             	50000000	        23.2 ns/op	  44.75 MB/s	       0 B/op	       0 allocs/op"
	want := "BenchmarkIndexRune-4             \t50000000\t         23.2ns/op\t   44.75MB/s \t    0.00B/op\t       0 allocs/op"
	fields := strings.Fields(text)
	firstTabPos := strings.IndexByte(text, '\t')
	result := toString(firstTabPos, fields)
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

// Taken from the strings package.

const benchmarkString = "some_text=some☺value"

var sink = 0

func BenchmarkIndexRune(b *testing.B) {
	if got := strings.IndexRune(benchmarkString, '☺'); got != 14 {
		b.Fatalf("wrong index: expected 14, got=%d", got)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = strings.IndexRune(benchmarkString, '☺')
		b.SetBytes(1)
	}
}

var benchmarkLongString = strings.Repeat(" ", 100) + benchmarkString
var longSink = 0

func BenchmarkIndexRuneLongString(b *testing.B) {
	if got := strings.IndexRune(benchmarkLongString, '☺'); got != 114 {
		b.Fatalf("wrong index: expected 114, got=%d", got)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		longSink = strings.IndexRune(benchmarkLongString, '☺')
	}
}

func BenchmarkNewDirectory(b *testing.B) {
	if testing.Short() {
		b.Skip("not running tests requiring large file scan in short mode")
	}

	fsGate := make(chan bool, 20)

	goroot := runtime.GOROOT()
	rootfs := gatefs.New(vfs.OS(goroot), fsGate)
	var fs = vfs.NameSpace{}
	fs.Bind("/", rootfs, "/", vfs.BindReplace)
	for _, p := range filepath.SplitList(build.Default.GOPATH) {
		fs.Bind("/src/golang.org", gatefs.New(vfs.OS(p), fsGate), "/src/golang.org", vfs.BindAfter)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for tries := 0; tries < b.N; tries++ {
		corpus := godoc.NewCorpus(fs)
		corpus.Init()
	}
}
