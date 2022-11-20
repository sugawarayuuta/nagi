package main

import (
	"encoding/json"
	"fmt"
	"nagi/stack"
	"os"
	"os/exec"
	"path/filepath"
)

func Set(args []string) *stack.Stacktrace {
	if len(args) < 2 {
		return stack.Stack(fmt.Errorf("required fields are empty"), 0)
	}

	pkg := map[string]interface{}{}
	opn, _ := os.Open(filepath.Join(work, "package.json"))
	if opn != nil {
		err := json.NewDecoder(opn).Decode(&pkg)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	defer opn.Close()

	if pkg["scripts"] == nil {
		pkg["scripts"] = map[string]interface{}{}
	}
	scripts := pkg["scripts"].(map[string]interface{})
	scripts[args[0]] = args[1]

	crt, err := os.Create(filepath.Join(work, "package.json"))
	if err != nil {
		return stack.Stack(err, 0)
	}
	defer crt.Close()

	enc := json.NewEncoder(crt)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err = enc.Encode(&pkg)
	if err != nil {
		return stack.Stack(err, 0)
	}
	return nil
}

func Run(arg string) *stack.Stacktrace {
	if arg == "" {
		return stack.Stack(fmt.Errorf("required fields are empty"), 0)
	}

	pkg := new(pkg)
	opn, _ := os.Open(filepath.Join(work, "package.json"))
	if opn != nil {
		err := json.NewDecoder(opn).Decode(&pkg)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	defer opn.Close()

	shell, flag := "sh", "-c"
	if os.PathListSeparator == ';' {
		shell, flag = "cmd.exe", "/c"
	}
	if script, ok := pkg.Scripts[arg]; ok {
		if script, ok := pkg.Scripts["pre"+arg]; ok {
			cmd := exec.Command(shell, flag, script)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			err := cmd.Run()
			if err != nil {
				return stack.Stack(err, 0)
			}
		}

		cmd := exec.Command(shell, flag, script)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		err := cmd.Run()
		if err != nil {
			return stack.Stack(err, 0)
		}

		if script, ok := pkg.Scripts["post"+arg]; ok {
			cmd := exec.Command(shell, flag, script)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			err := cmd.Run()
			if err != nil {
				return stack.Stack(err, 0)
			}
		}

		return nil
	}
	return stack.Stack(fmt.Errorf("task not found"), 0)
}
