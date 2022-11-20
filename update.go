package main

import (
	"encoding/json"
	"nagi/stack"
	"os"
	"path/filepath"
)

func Update(args []string, opts []string) *stack.Stacktrace {

	if !has(opts, "--save") {
		opts = append(opts, "--no-save")
	}

	if len(args) > 0 {
		for _, arg := range args {
			stacktrace := Add(arg, opts)
			if stacktrace != nil {
				return stacktrace
			}
		}
		return nil
	}

	pkg := new(pkg)
	opn, _ := os.Open(filepath.Join(dir, "package.json"))
	if opn != nil {
		err := json.NewDecoder(opn).Decode(pkg)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	defer opn.Close()

	for name, tag := range pkg.Dependencies {
		stacktrace := Add(name + "@" + tag, opts)
		if stacktrace != nil {
			return stacktrace
		}
	}
	return nil
}
