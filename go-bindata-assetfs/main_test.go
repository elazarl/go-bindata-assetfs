package main

import "testing"
import "regexp"
import "os"

var temp_pattern *regexp.Regexp = regexp.MustCompile(`^/tmp/`)
var exec_pattern *regexp.Regexp = regexp.MustCompile(`go-bindata`)

func TestConfigParseEmpty(t *testing.T) {
	c, err := parseConfig([]string{})
	if err != nil {
		t.Fatalf("parseConfig failed: %v", err)
	}

	if !temp_pattern.MatchString(c.TempPath) {
		t.Fatalf("TempPath should live in system temp directory, got %s", c.TempPath)
	}
	defer os.Remove(c.TempPath)

	if c.Debug != false {
		t.Fatalf("Debug should not default to true")
	}
	if !exec_pattern.MatchString(c.ExecPath) {
		t.Fatalf("ExecPath should point to go-bindata binary, got %s", c.ExecPath)
	}
	if c.OutPath != "bindata.go" {
		t.Fatalf("OutPath should default to bindata.go, got %s", c.OutPath)
	}
	if len(c.Args) != 0 {
		t.Fatalf("Args should be empty, got %v", c.Args)
	}
}

func TestConfigParseDebug(t *testing.T) {
	c, err := parseConfig([]string{"-debug"})
	if err != nil {
		t.Fatalf("parseConfig failed: %v", err)
	}

	if !temp_pattern.MatchString(c.TempPath) {
		t.Fatalf("TempPath should live in system temp directory, got %s", c.TempPath)
	}
	defer os.Remove(c.TempPath)

	if c.Debug != true {
		t.Fatalf("Debug was requested")
	}
	if !exec_pattern.MatchString(c.ExecPath) {
		t.Fatalf("ExecPath should point to go-bindata binary, got %s", c.ExecPath)
	}
	if c.OutPath != "bindata.go" {
		t.Fatalf("OutPath should default to bindata.go, got %s", c.OutPath)
	}
	if len(c.Args) != 0 {
		t.Fatalf("Args should be empty, got %v", c.Args)
	}
}

func TestConfigParseArgs(t *testing.T) {
	c, err := parseConfig([]string{"x", "y", "-debug", "z"})
	if err != nil {
		t.Fatalf("parseConfig failed: %v", err)
	}

	if !temp_pattern.MatchString(c.TempPath) {
		t.Fatalf("TempPath should live in system temp directory, got %s", c.TempPath)
	}
	defer os.Remove(c.TempPath)

	if c.Debug != true {
		t.Fatalf("Debug was requested")
	}
	if !exec_pattern.MatchString(c.ExecPath) {
		t.Fatalf("ExecPath should point to go-bindata binary, got %s", c.ExecPath)
	}
	if c.OutPath != "bindata.go" {
		t.Fatalf("OutPath should default to bindata.go, got %s", c.OutPath)
	}
	if len(c.Args) != 3 || c.Args[0] != "x" || c.Args[1] != "y" || c.Args[2] != "z" {
		t.Fatalf("Args should be [x,y,z], got %v", c.Args)
	}
}

func TestConfigParsePaths(t *testing.T) {
	c, err := parseConfig([]string{"-t", "tempfile.go", "-o", "outfile.go"})
	if err != nil {
		t.Fatalf("parseConfig failed: %v", err)
	}

	if c.TempPath != "tempfile.go" {
		t.Fatalf("TempPath should be tempfile.go, got %s", c.TempPath)
		if temp_pattern.MatchString(c.TempPath) {
			os.Remove(c.TempPath)
		}
	}

	if c.Debug != false {
		t.Fatalf("Debug was not requested")
	}
	if !exec_pattern.MatchString(c.ExecPath) {
		t.Fatalf("ExecPath should point to go-bindata binary, got %s", c.ExecPath)
	}
	if c.OutPath != "outfile.go" {
		t.Fatalf("OutPath should be outfile.go, got %s", c.OutPath)
	}
	if len(c.Args) != 0 {
		t.Fatalf("Args should be empty, got %v", c.Args)
	}
}
