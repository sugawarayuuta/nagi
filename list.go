package main

import (
	"encoding/gob"
	"encoding/json"
	"nagi/stack"
	"os"
	"path/filepath"
)

func List(args []string) ([]string, *stack.Stacktrace) {
	pkg := new(pkg)
	opn, _ := os.Open(filepath.Join(dir, "package.json"))
	if opn != nil {
		err := json.NewDecoder(opn).Decode(pkg)
		if err != nil {
			return nil, stack.Stack(err, 0)
		}
	}
	defer opn.Close()

	dep := map[string]dep{}
	opn, _ = os.Open(lock)
	if opn != nil {
		err := gob.NewDecoder(opn).Decode(&dep)
		if err != nil {
			return nil, stack.Stack(err, 0)
		}
	}
	defer opn.Close()

	res := make([]string, 0)
	for dependency, version := range pkg.Dependencies {
		red, err := os.Readlink(filepath.Join(node_modules, dependency))
		if err != nil {
			return nil, stack.Stack(err, 0)
		}
		if len(args) > 0 {
			for _, arg := range args {
				if arg != dependency {
					continue
				}
				res = append(res, dependency+"@"+version+" -> "+red)
				for name := range dep[dependency].Dependencies {
					res = append(res, " -"+name)
				}
			}
			continue
		}
		
		res = append(res, dependency+"@"+version+" -> "+red)
		for name := range dep[dependency].Dependencies {
			res = append(res, " -"+name)
		}
	}
	return res, nil
}
