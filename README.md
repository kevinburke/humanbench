# github.com/kevinburke/humanbench

Humanbench prints human-readable Go benchmarks. If your benchmark takes
milliseconds or minutes, humanbench will print `ms/op` or `min/op` instead of
`ns/op`.

### Usage

Put `humanbench` in front of the normal Go benchmark invocation (or any Go test
invocation). Rows printed to stdout that begin with "Benchmark" will be
reprinted using human-readable benchmarks.

```
humanbench go test -bench=.
```

I may add a version that you can pipe to, e.g. `go test | humanbench`, though
I worry. If `go test` exits 1, I don't have a reliable way of determining that
in the piped script, besides parsing the exit status text. This may lead people
to believe that their tests passed, when they didn't.

(You are supposed to set `bash -o pipefail` though few people do that).

### Installation

```
go get -u github.com/kevinburke/humanbench`
```
