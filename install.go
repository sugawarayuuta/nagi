package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"nagi/prog"
	"nagi/semver"
	"nagi/stack"
	"nagi/sys"
	"nagi/tgz"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Install(opts []string) *stack.Stacktrace {
	pkg := new(pkg)
	opn, _ := os.Open(filepath.Join(dir, "package.json"))
	if opn != nil {
		err := json.NewDecoder(opn).Decode(pkg)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	defer opn.Close()

	dep := map[string]dep{}
	opn, _ = os.Open(lock)
	if opn != nil {
		err := gob.NewDecoder(opn).Decode(&dep)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	defer opn.Close()

	os.MkdirAll(filepath.Join(node_modules, ".bin"), 0755)
	os.MkdirAll(filepath.Join(nagi, "cache"), 0755)

	for name, tag := range pkg.Dependencies {
		link := filepath.Join(node_modules, name)

		_, err := os.Stat(link)
		if !errors.Is(err, os.ErrNotExist) {
			continue
		}

		err = os.Mkdir(filepath.Dir(filepath.Join(node_modules, name)), 0755)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return stack.Stack(err, 0)
		}

		if dep, ok := dep[name]; ok {
			stacktrace := dep.bin(link)
			if stacktrace != nil {
				return stacktrace
			}

			for cmd, bin := range dep.Bin {
				register(cmd, bin, link)
			}

			return nil
		}

		stacktrace := Add(name+"@"+tag, append(opts, "--no-save"))
		if stacktrace != nil {
			return stacktrace
		}
	}
	return nil
}

func (lock dep) bin(pack string) *stack.Stacktrace {

	if lock.Engine != "" {
		ok := false
		version := semver.Version(sys.NODE)
		for _, sem := range semver.Build(lock.Engine) {
			if sem.Match(version) {
				ok = true
				break
			}
		}
		if !ok {
			return stack.Stack(fmt.Errorf("unsupported node engine"), 0)
		}
	}

	if len(lock.Os) != 0 {
		ok := false
		for index := range lock.Os {
			not := strings.HasPrefix(lock.Os[index], "!")
			if not && strings.TrimPrefix(lock.Os[index], "!") != sys.OS {
				ok = true
				break
			} else if !not && lock.Os[index] == sys.OS {
				ok = true
				break
			}
		}
		if !ok {
			return stack.Stack(fmt.Errorf("unsupported operating system"), 0)
		}
	}

	if len(lock.Cpu) != 0 {
		ok := false
		for index := range lock.Cpu {
			not := strings.HasPrefix(lock.Cpu[index], "!")
			if not && strings.TrimPrefix(lock.Cpu[index], "!") != sys.ARCH {
				ok = true
				break
			} else if !not && lock.Cpu[index] == sys.ARCH {
				ok = true
				break
			}
		}
		if !ok {
			return stack.Stack(fmt.Errorf("unsupported cpu architecture"), 0)
		}
	}

	module := filepath.Join(nagi, "cache", lock.Name+"@"+lock.Version)
	deps := &cache{
		Info: lock,
		Use:  []string{pack},
	}

	opn, mut, _ := open(filepath.Join(module, "nagi.cache"))
	defer mut.Unlock()
	if opn != nil {
		defer opn.Close()

		cache := new(cache)
		err := gob.NewDecoder(opn).Decode(cache)
		if err != nil {
			return stack.Stack(err, 0)
		}
		cache.Use = append(cache.Use, pack)

		crt, err := os.Create(filepath.Join(module, "nagi.cache"))
		if err != nil {
			return stack.Stack(err, 0)
		}

		err = gob.NewEncoder(crt).Encode(cache)
		if err != nil {
			return stack.Stack(err, 0)
		}

		err = os.Symlink(module, pack)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return stack.Stack(err, 0)
		}

		return nil
	}

	os.Mkdir(module, 0755)

	get, err := http.Get(lock.Tarball)
	if err != nil {
		return stack.Stack(err, 0)
	}
	if get.StatusCode != http.StatusOK {
		return stack.Stack(fmt.Errorf("bad status code was returned: %d", get.StatusCode), 0)
	}
	defer get.Body.Close()

	if lock.Pre != "" {
		shell, flag := "sh", "-c"
		if os.PathListSeparator == ';' {
			shell, flag = "cmd.exe", "/c"
		}
		cmd := exec.Command(shell, flag, lock.Pre)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.Dir = module

		err := cmd.Run()
		if err != nil {
			return stack.Stack(err, 0)
		}
	}

	buf := new(bytes.Buffer)
	sha := sha512.New()
	bar := &prog.Bar{
		Msg:  lock.Name,
		Max:  int(get.ContentLength),
		Curr: 0,

		Pre: func(bar *prog.Bar) string {
			return fmt.Sprintf("%dKB/%dKB", bar.Curr/1000, bar.Max/1000)
		},
	}
	bars.Add(bar)
	_, err = io.Copy(sha, io.TeeReader(&reader{Bar: bar, Reader: get.Body}, buf))
	if err != nil {
		return stack.Stack(err, 0)
	}
	if base64.StdEncoding.EncodeToString(sha.Sum(nil)) != strings.TrimPrefix(lock.Integrity, "sha512-") {
		return stack.Stack(fmt.Errorf("failed to verify %s", lock.Name), 0)
	}
	err = tgz.Save(buf, module)
	if err != nil {
		return stack.Stack(err, 0)
	}

	err = os.Symlink(module, pack)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return stack.Stack(err, 0)
	}

	err = os.Mkdir(filepath.Join(module, "node_modules"), 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return stack.Stack(err, 0)
	}

	stacks := make(chan interface{}, len(lock.Dependencies))

	for name, dependency := range lock.Dependencies {
		go func(name string, dependency dep) {
			err = os.Mkdir(filepath.Dir(filepath.Join(module, "node_modules", name)), 0755)
			if err != nil && !errors.Is(err, os.ErrExist) {
				stacks <- stack.Stack(err, 0)
				return
			}

			stacktrace := dependency.bin(filepath.Join(module, "node_modules", name))
			if stacktrace != nil {
				stacks <- stacktrace
				return
			}

			stacks <- name
		}(name, dependency)
	}

	for index := 0; index < len(lock.Dependencies); index++ {
		stacks := <-stacks
		if stacktrace, ok := stacks.(*stack.Stacktrace); ok {
			return stacktrace
		}
	}
	close(stacks)

	crt, err := os.Create(filepath.Join(module, "nagi.cache"))
	if err != nil {
		return stack.Stack(err, 0)
	}

	err = gob.NewEncoder(crt).Encode(deps)
	if err != nil {
		return stack.Stack(err, 0)
	}

	if lock.Post != "" {
		shell, flag := "sh", "-c"
		if os.PathListSeparator == ';' {
			shell, flag = "cmd.exe", "/c"
		}
		cmd := exec.Command(shell, flag, lock.Post)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.Dir = module

		err := cmd.Run()
		if err != nil {
			return stack.Stack(err, 0)
		}
	}

	return nil
}
