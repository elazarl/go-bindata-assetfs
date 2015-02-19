package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if !isInPath("go-bindata") && !isInPath("go-bindata.exe") {
		fmt.Println("Cannot find go-bindata executable in path")
		fmt.Println("Maybe you need: go get github.com/elazarl/go-bindata-assetfs/...")
		os.Exit(1)
	}
	cmd := exec.Command("go-bindata", os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
	in, err := os.Open("bindata.go")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot read 'bindata.go'", err)
		return
	}
	out, err := os.Create("bindata_assetfs.go")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot write 'bindata_assetfs.go'", err)
		return
	}
	r := bufio.NewReader(in)
	done := false
	for line, isPrefix, err := r.ReadLine(); err == nil; line, isPrefix, err = r.ReadLine() {
		line = append(line, '\n')
		if _, err := out.Write(line); err != nil {
			fmt.Fprintln(os.Stderr, "Cannot write to 'bindata_assetfs.go'", err)
			return
		}
		if !done && !isPrefix && bytes.HasPrefix(line, []byte("import (")) {
			fmt.Fprintln(out, "\t\"github.com/elazarl/go-bindata-assetfs\"")
			done = true
		}
	}
	/* Note by khaos: Prefix omitted due to strange behaviour. */
	fmt.Fprintln(out, `
func assetFS() *assetfs.AssetFS {
	for _ = range _bintree.Children {
		return &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: ""}
	}
	panic("unreachable")
}`)
	// Close files BEFORE remove calls (don't use defer).
	in.Close()
	out.Close()
	if err := os.Remove("bindata.go"); err != nil {
		fmt.Fprintln(os.Stderr, "Cannot remove bindata_assetfs.go", err)
	}
}

func isInPath(filename string) bool {
	// First, check the same directory.
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	// Then other dirs.
	for _, path := range filepath.SplitList(os.Getenv("PATH")) {
		if _, err := os.Stat(filepath.Join(path, filename)); err == nil {
			return true
		}
	}
	return false
}
