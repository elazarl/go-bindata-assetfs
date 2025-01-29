package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type Config struct {
	// Do not embed assets, just structure
	Debug bool
	// Path to go-bindata executable
	ExecPath string
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
		ExecPath: "",
		TempPath: "",
		OutPath:  "bindata.go",
		Args:     []string{},
	}
	extra_args := []string{}

	path, err := exec.LookPath("go-bindata")
	if err != nil {
		fmt.Println("Cannot find go-bindata executable in path")
		fmt.Println("Maybe you need: go get github.com/elazarl/go-bindata-assetfs/...")
		return c, err
	}
	c.ExecPath = path

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
			extra_args = append(extra_args, arg)
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

	// Polish up Args with stuff we pulled out/deduped earlier
	c.Args = []string{"-o", c.TempPath}
	if (c.Debug) {
		c.Args = append(c.Args, "-debug")
	}
	c.Args = append(c.Args, extra_args...)
	return c, nil
}

func main() {
	c, err := parseConfig(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		os.Exit(1)
	}
	defer os.Remove(c.TempPath)

	in, err := os.Create(c.TempPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: cannot create temporary file", err)
		os.Exit(1)
	}

	out, err := os.Create(c.OutPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: cannot create output file", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Args:", c.Args)
	cmd := exec.Command(c.ExecPath, c.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: go-bindata: ", err)
		os.Exit(1)
	}
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
			if c.Debug {
				fmt.Fprintln(out, "\t\"net/http\"")
			} else {
				fmt.Fprintln(out, "\t\"github.com/elazarl/go-bindata-assetfs\"")
			}
			done = true
		}
	}
	if c.Debug {
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
