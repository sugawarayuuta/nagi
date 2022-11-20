package main

import (
	"encoding/gob"
	"errors"
	"nagi/stack"
	"os"
	"path/filepath"
	"strings"
)

func Clean(dir string) *stack.Stacktrace {
	fls, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return stack.Stack(err, 0)
	}
	for _, fl := range fls {
		if fl.IsDir() && strings.HasPrefix(fl.Name(), "@") {
			stacktrace := Clean(filepath.Join(dir, fl.Name()))
			if stacktrace != nil {
				return stacktrace
			}
			continue
		}

		if !fl.IsDir() {
			continue
		}

		opn, err := os.Open(filepath.Join(dir, fl.Name(), "nagi.cache"))
		if err != nil {
			return stack.Stack(err, 0)
		}
		defer opn.Close()
		cache := new(cache)
		err = gob.NewDecoder(opn).Decode(cache)
		if err != nil {
			return stack.Stack(err, 0)
		}

		purged := purge(cache.Use)
		for index := 0; index < len(purged); index++ {
			_, err := os.Stat(purged[index])
			if errors.Is(err, os.ErrNotExist) {
				purged = append(purged[:index], purged[index+1:]...)
				index--
			}
		}

		if len(purged) == 0 {
			os.RemoveAll(filepath.Join(dir, fl.Name()))
		}
	}
	return nil
}

func purge(reqs []string) []string {
	keys := make(map[string]bool, 0)
	res := make([]string, len(reqs))

	for index, req := range reqs {
		if _, key := keys[req]; !key {
			keys[req] = true
			res[index] = req
		}
	}
	return res
}
