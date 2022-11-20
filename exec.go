package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"nagi/semver"
	"nagi/stack"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func Exec(arg string) *stack.Stacktrace {
	if arg == "" {
		return stack.Stack(fmt.Errorf("required fields are empty"), 0)
	}
	name, tag := semver.Split(arg)

	link := filepath.Join(node_modules, name)
	opn, _ := os.Open(filepath.Join(link, "nagi.cache"))
	if opn != nil {
		cache := new(cache)
		err := gob.NewDecoder(opn).Decode(cache)
		if err != nil {
			return stack.Stack(err, 0)
		}

		for cmd, bin := range cache.Info.Bin {
			if len(cache.Info.Bin) == 1 || path.Base(name) == cmd {
				return stack.Stack(execute(filepath.Join(link, bin)), 0)
			}
		}

		return stack.Stack(fmt.Errorf("command not found"), 0)
	}

	defer os.Remove(link)
	os.Mkdir(filepath.Join(node_modules), 0755)
	os.MkdirAll(filepath.Join(nagi, "cache"), 0755)
	os.Mkdir(filepath.Dir(filepath.Join(node_modules, name)), 0755)

	dep := map[string]dep{}
	opn, _ = os.Open(lock)
	if opn != nil {
		err := gob.NewDecoder(opn).Decode(&dep)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	defer opn.Close()
	if dep, ok := dep[name]; ok {
		stacktrace := dep.bin(link)
		if stacktrace != nil {
			return stacktrace
		}

		for cmd, bin := range dep.Bin {
			if len(dep.Bin) == 1 || path.Base(name) == cmd {
				return stack.Stack(execute(filepath.Join(link, bin)), 0)
			}
		}

		return stack.Stack(fmt.Errorf("command not found"), 0)
	}

	get, err := http.Get(REGISTRY + name)
	if err != nil {
		return stack.Stack(err, 0)
	}
	if get.StatusCode != http.StatusOK {
		return stack.Stack(fmt.Errorf("bad status code was returned: %d", get.StatusCode), 0)
	}
	defer get.Body.Close()
	mnf := new(npm)
	err = json.NewDecoder(get.Body).Decode(mnf)
	if err != nil {
		return stack.Stack(err, 0)
	}
	found, err := find(mnf, tag)
	if err != nil {
		return stack.Stack(fmt.Errorf("couldn't find %s of %s: %v", tag, name, err), 0)
	}
	_, stacktrace := mnf.Versions[found].bin(link)
	if stacktrace != nil {
		return stacktrace
	}

	for cmd, bin := range mnf.Versions[found].Bin {
		if len(mnf.Versions[found].Bin) == 1 || path.Base(name) == cmd {
			return stack.Stack(execute(filepath.Join(link, bin)), 0)
		}
	}

	return stack.Stack(fmt.Errorf("command not found"), 0)
}

func execute(bin string) error {
	cmd := exec.Command(bin)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := os.Chmod(bin, 0755)
	if err != nil {
		return err
	}
	return cmd.Run()
}
