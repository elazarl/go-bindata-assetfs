package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	// Do not embed assets, just structure
	Debug bool
	// Path to temporary file produced by go-bindata (default: random file in OS tempdir)
	TempPath string
	// Final output path (default: bindata.go)
	OutPath string
	// Remaining misc args to pass through to go-bindata
	Args []string
}

// If this function succeeds, the caller is responsible for deleting the
// temporary file at c.TempPath. This usually makes sense to do with a:
//
//	defer os.Remove(c.TempPath)
//
// right after checking for errors.
func parseConfig(args []string) (Config, error) {
	c := Config{
		Debug:    false,
		TempPath: "",
		OutPath:  "bindata.go",
		Args:     []string{},
	}

	// Do this dumb manually-tracked for loop so we can do skips.
	i := 0
	for {
		if i >= len(args) {
			break
		}
		arg := args[i]

		if arg == "-debug" {
			c.Debug = true
		} else if arg == "-t" && i+1 < len(args) {
			c.TempPath = args[i+1]
			i = i + 1
		} else if arg == "-o" && i+1 < len(args) {
			c.OutPath = args[i+1]
			i = i + 1
		} else {
			c.Args = append(c.Args, arg)
		}
		i = i + 1
	}

	// We don't hold onto the original file handle, as we can expect
	// go-bindata to replace it. We do establish the existence of the
	// file on-disk, though, to avoid collisions.
	if c.TempPath == "" {
		f, err := os.CreateTemp("", "go-bindata-assetfs")
		if err != nil {
			return c, err
		}
		c.TempPath = f.Name()
		f.Close()
	}
	return c, nil
}

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
