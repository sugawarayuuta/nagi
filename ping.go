package main

import (
	"nagi/stack"
	"net/http"
)

func Ping() (int, *stack.Stacktrace) {
	get, err := http.Get(REGISTRY)
	if err != nil {
		return 0, stack.Stack(err, 0)
	}
	defer get.Body.Close()
	return get.StatusCode, nil
}