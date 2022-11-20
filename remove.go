package main

import (
	"encoding/gob"
	"encoding/json"
	"nagi/stack"
	"os"
	"path/filepath"
	"sort"
)

func Remove(name string, opts []string) *stack.Stacktrace {
	dep := map[string]dep{}
	opn, _ := os.Open(lock)
	if opn != nil {
		err := gob.NewDecoder(opn).Decode(&dep)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	defer opn.Close()
	err := dep[name].follow(filepath.Join(node_modules, name))
	if err != nil {
		return stack.Stack(err, 0)
	}
	delete(dep, name)
	crt, err := os.Create(lock)
	if err != nil {
		return stack.Stack(err, 0)
	}
	defer crt.Close()
	err = gob.NewEncoder(crt).Encode(&dep)
	if err != nil {
		return stack.Stack(err, 0)
	}

	pkg := map[string]interface{}{}
	opn, _ = os.Open(filepath.Join(dir, "package.json"))
	if opn != nil {
		err := json.NewDecoder(opn).Decode(&pkg)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	defer opn.Close()
	if dependencies, ok := pkg["dependencies"].(map[string]interface{}); ok {
		if has(opts, "--no-save") {
			return nil
		}
		delete(dependencies, name)
		crt, err = os.Create(filepath.Join(dir, "package.json"))
		if err != nil {
			return stack.Stack(err, 0)
		}
		defer crt.Close()
		enc := json.NewEncoder(crt)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "    ")
		err = enc.Encode(pkg)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	os.Remove(filepath.Join(node_modules, name))
	return nil
}

func (dep dep) follow(pack string) error {
	module := filepath.Join(nagi, "cache", dep.Name+"@"+dep.Version)

	cache := new(cache)
	opn, _ := os.Open(filepath.Join(module, "nagi.cache"))
	if opn != nil {
		err := gob.NewDecoder(opn).Decode(cache)
		if err != nil {
			return err
		}
	}
	defer opn.Close()

	if sch := sort.SearchStrings(cache.Use, pack); sch != len(cache.Use) {
		cache.Use = append(cache.Use[:sch], cache.Use[sch+1:]...)
		crt, err := os.Create(filepath.Join(module, "nagi.cache"))
		if err != nil {
			return err
		}
		defer crt.Close()
		err = gob.NewEncoder(crt).Encode(cache)
		if err != nil {
			return err
		}
	}
	for _, dependency := range dep.Dependencies {
		err := dependency.follow(filepath.Join(module, "node_modules", dependency.Name))
		if err != nil {
			return err
		}
	}
	return nil
}
