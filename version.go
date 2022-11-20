package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Version() string {
	version := []string{"go:" + runtime.Version(), "os:" + runtime.GOOS, "arch:" + runtime.GOARCH}

	exec, err := os.Executable()
	if err != nil {
		return strings.Join(append(version, err.Error()), "\n")
	}

	opn, err := os.Open(filepath.Join(filepath.Dir(exec), "package.json"))
	if err != nil {
		return strings.Join(append(version, err.Error()), "\n")
	}

	pkg := map[string]interface{}{}
	err = json.NewDecoder(opn).Decode(&pkg)
	if err != nil {
		return strings.Join(append(version, err.Error()), "\n")
	}

	return strings.Join(append(version, pkg["version"].(string)), "\n")
}