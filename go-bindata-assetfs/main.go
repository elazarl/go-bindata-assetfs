package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func isDebug(args []string) bool {
	flagset := flag.NewFlagSet("", flag.ContinueOnError)
	debug := flagset.Bool("debug", false, "")
	debugArgs := make([]string, 0)
	for _, arg := range args {
		if strings.HasPrefix(arg, "-debug") {
			debugArgs = append(debugArgs, arg)
		}
	}
	flagset.Parse(debugArgs)
	if debug == nil {
		return false
	}
	return *debug
}

func getBinDataFile() (*os.File, *os.File, []string, error) {
	bindataArgs := make([]string, 0)
	outputLoc := "bindata.go"

	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-o" {
			outputLoc = os.Args[i+1]
			i++
		} else {
			bindataArgs = append(bindataArgs, os.Args[i])
		}
	}

	tempFile, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		return &os.File{}, &os.File{}, nil, err
	}

	stat, err := os.Lstat(outputLoc)
	if err != nil {
		if !os.IsNotExist(err) {
			return &os.File{}, &os.File{}, nil, err
		}

		// File does not exist. This is fine, just make
		// sure the directory it is to be in exists.
		dir, _ := filepath.Split(outputLoc)
		if dir != "" {
			if err = os.MkdirAll(dir, 0744); err != nil {
				return &os.File{}, &os.File{}, nil, fmt.Errorf("create output directory: %v", err)
			}
		}
	}
	if stat != nil && stat.IsDir() {
		return &os.File{}, &os.File{}, nil, fmt.Errorf("output loc is a directory")
	}

	outputFile, err := os.Create(outputLoc)
	if err != nil {
		return &os.File{}, &os.File{}, nil, err
	}

	bindataArgs = append([]string{"-o", tempFile.Name()}, bindataArgs...)
	return outputFile, tempFile, bindataArgs, nil
}

func main() {
	path, err := exec.LookPath("go-bindata")
	if err != nil {
		fmt.Println("Cannot find go-bindata executable in path")
		fmt.Println("Maybe you need: go get github.com/elazarl/go-bindata-assetfs/...")
		os.Exit(1)
	}
	out, in, args, err := getBinDataFile()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: cannot create temporary file", err)
		os.Exit(1)
	}
	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: go-bindata: ", err)
		os.Exit(1)
	}
	debug := isDebug(os.Args[1:])
	r := bufio.NewReader(in)
	done := false
	for line, isPrefix, err := r.ReadLine(); err == nil; line, isPrefix, err = r.ReadLine() {
		if !isPrefix {
			line = append(line, '\n')
		}
		if _, err := out.Write(line); err != nil {
			fmt.Fprintln(os.Stderr, "Cannot write to ", out.Name(), err)
			return
		}
		if !done && !isPrefix && bytes.HasPrefix(line, []byte("import (")) {
			if debug {
				fmt.Fprintln(out, "\t\"net/http\"")
			} else {
				fmt.Fprintln(out, "\t\"github.com/elazarl/go-bindata-assetfs\"")
			}
			done = true
		}
	}
	if debug {
		fmt.Fprintln(out, `
func assetFS() http.FileSystem {
	for k := range _bintree.Children {
		return http.Dir(k)
	}
	panic("unreachable")
}`)
		fmt.Fprintln(out, `
func AssetFS() http.FileSystem {
	return assetFS()
}`)
	} else {
		fmt.Fprintln(out, `
func assetFS() *assetfs.AssetFS {
	for k := range _bintree.Children {
		return &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: k}
	}
	panic("unreachable")
}`)
		fmt.Fprintln(out, `
func AssetFS() *assetfs.AssetFS {
	return assetFS()
}`)
	}
	// Close files BEFORE remove calls (don't use defer).
	in.Close()
	out.Close()
	if err := os.Remove(in.Name()); err != nil {
		fmt.Fprintln(os.Stderr, "Cannot remove", in.Name(), err)
	}
}
