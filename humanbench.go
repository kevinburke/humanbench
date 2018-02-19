// Command humanbench prints human-readable benchmark output.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"unicode"
)

func usage() {
	fmt.Fprintf(os.Stderr, `humanbench go test [build/test flags] [packages] [build/test flags & test binary flags]

humanbench runs the resulting command, translating benchmarks into
human-readable number outputs.`)
}

func toString(firstTabPos int, fields []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-*s\t", firstTabPos, fields[0])
	fmt.Fprintf(&b, "%8s", fields[1])
	for i := 2; i+2 <= len(fields); i += 2 {
		if fields[i+1] == "MB/s" {
			fmt.Fprintf(&b, "\t%8sMB/s ", fields[i])
			continue
		}
		val, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			panic(err)
		}
		scaler := NewScaler(val, fields[i+1])
		replacement := scaler(val)
		b.WriteByte('\t')
		switch fields[i+1] {
		case "ns/op":
			fmt.Fprintf(&b, "%18s", replacement+"/op")
		case "allocs/op":
			fmt.Fprintf(&b, "%18s", replacement+" allocs/op")
		default:
			//wid := 7 + 1 + len(fields[i+1])
			//fmt.Fprintf(&b, "%"+strconv.Itoa(wid)+"s", replacement)
			fmt.Fprintf(&b, "%12s", replacement+"/op")
		}
	}
	return b.String()
}

func parseLine(line string) string {
	space := strings.IndexFunc(line, unicode.IsSpace)
	if space < 0 {
		return line
	}
	name := line[:space]
	if !strings.HasPrefix(name, "Benchmark") {
		return line
	}
	f := strings.Fields(line)
	if len(f) < 4 {
		return line
	}
	name = f[0]
	if !strings.HasPrefix(name, "Benchmark") {
		return line
	}
	firstTabPos := strings.IndexByte(line, '\t')
	return toString(firstTabPos, f)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	r, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	// wait for the command to finish
	waitCh := make(chan error, 1)
	txtCh := make(chan string, 1)
	bufErrCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
		close(waitCh)
	}()
	bs := bufio.NewScanner(r)
	go func() {
		for bs.Scan() {
			txtCh <- bs.Text()
		}
		if err := bs.Err(); err != nil {
			bufErrCh <- err
		}
		close(txtCh)
		close(bufErrCh)
	}()
	sigs := make(chan os.Signal)
	signal.Notify(sigs)
	for {
		select {
		case line, more := <-txtCh:
			if !more {
				continue
			}
			fmt.Println(parseLine(line))
		case sig, more := <-sigs:
			if !more {
				continue
			}
			if sig == syscall.SIGSTOP || sig == syscall.SIGCHLD {
				continue
			}
			if err = cmd.Process.Signal(sig); err != nil {
				log.Fatalf("could not send signal %s: %v", sig, err)
			}
		case err, more := <-bufErrCh:
			if !more {
				continue
			}
			log.Fatalf("buf error: %v", err)
		case err := <-waitCh:
			var waitStatus syscall.WaitStatus
			if exitError, ok := err.(*exec.ExitError); ok {
				waitStatus = exitError.Sys().(syscall.WaitStatus)
				os.Exit(waitStatus.ExitStatus())
			}
			if err != nil {
				log.Fatalf("%v", err)
			}
			return
		}
	}
}
