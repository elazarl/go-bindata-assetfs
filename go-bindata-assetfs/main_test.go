package main

import "testing"
import "regexp"
import "os"
import "reflect"

var temp_pattern *regexp.Regexp = regexp.MustCompile(`^/tmp/`)
var exec_pattern *regexp.Regexp = regexp.MustCompile(`go-bindata`)

func helper(t *testing.T, tmp string, args []string) Config {
	c, err := parseConfig(args)
	if err != nil {
		t.Fatalf("parseConfig failed: %v", err)
	}

	if tmp != "" {
		if c.TempPath != tmp {
			t.Fatalf("Expected c.TempFile to be %s, got %s", tmp, c.TempPath)
		}
	} else if temp_pattern.MatchString(c.TempPath) {
		os.Remove(c.TempPath)
	} else {
		t.Fatalf("TempPath should live in system temp directory, got %s", c.TempPath)
	}

	if !exec_pattern.MatchString(c.ExecPath) {
		t.Fatalf("ExecPath should point to go-bindata binary, got %s", c.ExecPath)
	}
	return c
}

func TestConfigParseEmpty(t *testing.T) {
	c := helper(t, "", []string{})
	expected := Config {
		Debug: false,
		ExecPath: c.ExecPath,
		TempPath: c.TempPath,
		OutPath: "bindata.go",
		Args: []string{},
	}
	if !reflect.DeepEqual(c, expected) {
		t.Fatalf("Expected %v, got %v", expected, c)
	}
}

func TestConfigParseDebug(t *testing.T) {
	c := helper(t, "", []string{"-debug", "-debug", "-debug"})
	expected := Config {
		Debug: true,
		ExecPath: c.ExecPath,
		TempPath: c.TempPath,
		OutPath: "bindata.go",
		Args: []string{},
	}
	if !reflect.DeepEqual(c, expected) {
		t.Fatalf("Expected %v, got %v", expected, c)
	}
}

func TestConfigParseArgs(t *testing.T) {
	c := helper(t, "", []string{"x", "y", "-debug", "z"})
	expected := Config {
		Debug: true,
		ExecPath: c.ExecPath,
		TempPath: c.TempPath,
		OutPath: "bindata.go",
		Args: []string{"x", "y", "z"},
	}
	if !reflect.DeepEqual(c, expected) {
		t.Fatalf("Expected %v, got %v", expected, c)
	}
}

func TestConfigParsePaths(t *testing.T) {
	c := helper(t, "tempfile.go", []string{"-t", "tempfile.go", "-o", "outfile.go"})
	expected := Config {
		Debug: false,
		ExecPath: c.ExecPath,
		TempPath: "tempfile.go",
		OutPath: "outfile.go",
		Args: []string{},
	}
	if !reflect.DeepEqual(c, expected) {
		t.Fatalf("Expected %v, got %v", expected, c)
	}
}
