# go-bindata-assetfs

Serve embedded files from [go-bindata](https://github.com/go-bindata/go-bindata) with `net/http`.

[GoDoc](http://godoc.org/github.com/elazarl/go-bindata-assetfs)

### Installation

Install with

    $ go get github.com/go-bindata/go-bindata/...
    $ go get github.com/elazarl/go-bindata-assetfs/...

### Creating embedded data

Usage is identical to [go-bindata](https://github.com/go-bindata/go-bindata) usage,
instead of running `go-bindata` run `go-bindata-assetfs`.

The tool will create a `bindata_assetfs.go` file, which contains the embedded data.

A typical use case is

    $ go-bindata-assetfs data/...

### Using assetFS in your code

The generated file provides an `assetFS()` function that returns a `http.Filesystem`
wrapping the embedded files. What you usually want to do is:

    http.Handle("/", http.FileServer(assetFS()))

This would run an HTTP server serving the embedded files.

## Without running binary tool

You can always just run the `go-bindata` tool, and then

use

```go
import "github.com/elazarl/go-bindata-assetfs"
...
http.Handle("/",
http.FileServer(
&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "data"}))
```

to serve files embedded from the `data` directory.

## SPA applications

For single page applications you can use `Fallback: "index.html"` in AssetFS context, so if route doesn't match the pattern it will fallback to file specified.

example

```go
import "github.com/elazarl/go-bindata-assetfs"
...
http.Handle("/",
http.FileServer(
&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "data", Fallback: "index.html"}))
```

## Specifying multiple directories

To generate binary data from calling the `go-bindata-assetfs` cmd when declaring multiple directories, then you can use the `AssetDIR(dir string)` function. For example, running the following command inside a directory contaning a `dist/..` & `public/...` directories:

```
$ go-bindata-assetfs dist/... public/...
```

Our server code might look like this using the `AssetDIR(dir string)` function

```go
func main() {
	r := mux.NewRouter() // Example uses "github.com/gorilla/mux"
    // Here we want to serve static content from 2 directories
	r.PathPrefix("/dist").Handler(http.StripPrefix("/dist", http.FileServer(AssetDIR("dist"))))
	r.PathPrefix("/public").Handler(http.StripPrefix("/public", http.FileServer(AssetDIR("public"))))
    // rest of server code...
	server := &http.Server{ Addr:    ":7777", Handler: r, }
	server.ListenAndServe()
}
```