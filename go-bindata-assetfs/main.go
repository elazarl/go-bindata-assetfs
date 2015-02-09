package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if !isInPath("go-bindata") {
		fmt.Println("Cannot find go-bindata executable in path")
		fmt.Println("Maybe you need: go get github.com/elazarl/go-bindata-assetfs/...")
		os.Exit(1)
	}
	cmd := exec.Command("go-bindata", os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Error running go-bindata: %v", err)
		os.Exit(2)
	}
	in, err := os.Open("bindata.go")
	if err != nil {
		log.Printf("Cannot read 'bindata.go': %v", err)
		os.Exit(3)
	}
	out, err := os.Create("bindata_assetfs.go")
	if err != nil {
		log.Printf("Cannot write 'bindata_assetfs.go': %v", err)
		os.Exit(4)
	}
	r := bufio.NewReader(in)
	w := bufio.NewWriter(out)
	defer in.Close()
	defer func() {
		w.Flush()
		out.Close()
	}()
	importPrefix := "import ("
LinesLoop:
	for {
		isImport := false
		b, err := r.Peek(len(importPrefix))
		if err == nil {
			isImport = string(b) == importPrefix
		}
		for {
			c, _, err := r.ReadRune()
			if err != nil {
				if err == io.EOF {
					break LinesLoop
				} else {
					log.Printf("Unable to read from bindata file: %v", err)
					os.Exit(5)
				}
			}
			_, err = w.WriteRune(c)
			if err != nil {
				log.Printf("Unable to write to assetfs file: %v", err)
				os.Exit(6)
			}
			if c == '\n' {
				if isImport {
					_, err = fmt.Fprintf(w, "\t\"github.com/elazarl/go-bindata-assetfs\"\n")
					if err != nil {
						log.Printf("Unable to write import to assetfs out: %v", err)
						os.Exit(7)
					}
				}
				break
			}
		}
	}

	_, err = fmt.Fprintln(w, `
func assetFS() *assetfs.AssetFS {
	for k := range _bintree.Children {
		return &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: k}
	}
	panic("unreachable")
}`)
	if err != nil {
		log.Printf("Unable to write assetFS function to assetfs out: %v", err)
		os.Exit(8)
	}
	if err := os.Remove("bindata.go"); err != nil {
		log.Printf("Cannot remove bindata_assetfs.go: %v", err)
		os.Exit(9)
	}
}

func isInPath(filename string) bool {
	for _, path := range filepath.SplitList(os.Getenv("PATH")) {
		if _, err := os.Stat(filepath.Join(path, "go-bindata")); err == nil {
			return true
		}
	}
	return false
}
