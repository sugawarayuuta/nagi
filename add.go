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
	"sync"
)

func Add(req string, opts []string) *stack.Stacktrace {
	name, tag := semver.Split(req)

	link := filepath.Join(node_modules, name)
	err := os.Remove(link)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return stack.Stack(err, 0)
	}

	os.MkdirAll(filepath.Join(node_modules, ".bin"), 0755)
	os.MkdirAll(filepath.Join(nagi, "cache"), 0755)
	os.Mkdir(filepath.Dir(filepath.Join(node_modules, name)), 0755)

	dep := map[string]dep{}
	opn, _ := os.Open(lock)
	if opn != nil {
		err = gob.NewDecoder(opn).Decode(&dep)
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
			register(cmd, bin, link)
		}

		pkg := map[string]interface{}{}
		opn, _ := os.Open(filepath.Join(dir, "package.json"))
		if opn != nil {
			err = json.NewDecoder(opn).Decode(&pkg)
			if err != nil {
				return stack.Stack(err, 0)
			}
		}
		defer opn.Close()

		which, what := "dependencies", "~"+dep.Version
		switch {
		case has(opts, "--no-save"):
			return nil
		case has(opts, "--save-exact", "-E"):
			what = dep.Version
			fallthrough
		case has(opts, "--save-dev", "-D"):
			which = "devDependencies"
		case has(opts, "--save-optional", "-O"):
			which = "optionalDependencies"
		}
		if pkg[which] == nil {
			pkg[which] = map[string]interface{}{}
		}
		pkg[which].(map[string]interface{})[name] = what

		crt, err := os.Create(filepath.Join(dir, "package.json"))
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
		return stack.Stack(err, 0)
	}

	deps, stacktrace := mnf.Versions[found].bin(link)
	if stacktrace != nil {
		return stacktrace
	}

	for cmd, bin := range mnf.Versions[found].Bin {
		register(cmd, bin, link)
	}

	pkg := map[string]interface{}{}
	opn, _ = os.Open(filepath.Join(dir, "package.json"))
	if opn != nil {
		err = json.NewDecoder(opn).Decode(&pkg)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	defer opn.Close()

	which, what := "dependencies", "~"+found
	switch {
	case has(opts, "--no-save"):
		return nil
	case has(opts, "--save-exact", "-E"):
		what = found
		fallthrough
	case has(opts, "--save-dev", "-D"):
		which = "devDependencies"
	case has(opts, "--save-optional", "-O"):
		which = "optionalDependencies"
	}
	if pkg[which] == nil {
		pkg[which] = map[string]interface{}{}
	}
	pkg[which].(map[string]interface{})[name] = what

	crt, err := os.Create(filepath.Join(dir, "package.json"))
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

	err = fix(deps.Info)
	if err != nil {
		return stack.Stack(err, 0)
	}

	return nil
}

func (mnfVer npmVer) bin(pack string) (*cache, *stack.Stacktrace) {

	if mnfVer.Engines.Node != "" {
		ok := false
		version := semver.Version(sys.NODE)
		for _, sem := range semver.Build(mnfVer.Engines.Node) {
			if sem.Match(version) {
				ok = true
				break
			}
		}
		if !ok {
			return nil, stack.Stack(fmt.Errorf("unsupported node engine"), 0)
		}
	}

	if len(mnfVer.Os) != 0 {
		ok := false
		for index := range mnfVer.Os {
			not := strings.HasPrefix(mnfVer.Os[index], "!")
			if not && strings.TrimPrefix(mnfVer.Os[index], "!") != sys.OS {
				ok = true
				break
			} else if !not && mnfVer.Os[index] == sys.OS {
				ok = true
				break
			}
		}
		if !ok {
			return nil, stack.Stack(fmt.Errorf("unsupported operating system"), 0)
		}
	}

	if len(mnfVer.Cpu) != 0 {
		ok := false
		for index := range mnfVer.Cpu {
			not := strings.HasPrefix(mnfVer.Cpu[index], "!")
			if not && strings.TrimPrefix(mnfVer.Cpu[index], "!") != sys.ARCH {
				ok = true
				break
			} else if !not && mnfVer.Cpu[index] == sys.ARCH {
				ok = true
				break
			}
		}
		if !ok {
			return nil, stack.Stack(fmt.Errorf("unsupported cpu architecture"), 0)
		}
	}

	module := filepath.Join(nagi, "cache", mnfVer.Name+"@"+mnfVer.Version)
	deps := &cache{
		Info: dep{
			Name:         mnfVer.Name,
			Version:      mnfVer.Version,
			Tarball:      mnfVer.Dist.Tarball,
			Integrity:    mnfVer.Dist.Integrity,
			Engine:       mnfVer.Engines.Node,
			Os:           mnfVer.Os,
			Cpu:          mnfVer.Cpu,
			Pre:          mnfVer.Scripts.Preinstall,
			Post:         mnfVer.Scripts.Postinstall,
			Bin:          mnfVer.Bin,
			Dependencies: map[string]dep{},
		},
		Use: []string{pack},
	}

	opn, mut, _ := open(filepath.Join(module, "nagi.cache"))
	defer mut.Unlock()
	if opn != nil {
		defer opn.Close()

		cache := new(cache)
		err := gob.NewDecoder(opn).Decode(cache)
		if err != nil {
			return nil, stack.Stack(err, 0)
		}

		cache.Use = append(cache.Use, pack)
		crt, err := os.Create(filepath.Join(module, "nagi.cache"))
		if err != nil {
			return nil, stack.Stack(err, 0)
		}
		defer crt.Close()

		err = gob.NewEncoder(crt).Encode(cache)
		if err != nil {
			return nil, stack.Stack(err, 0)
		}

		err = os.Symlink(module, pack)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return nil, stack.Stack(err, 0)
		}

		return cache, nil
	}

	os.Mkdir(module, 0755)

	get, err := http.Get(mnfVer.Dist.Tarball)
	if err != nil {
		return nil, stack.Stack(err, 0)
	}
	if get.StatusCode != http.StatusOK {
		return nil, stack.Stack(fmt.Errorf("bad status code was returned: %d", get.StatusCode), 0)
	}
	defer get.Body.Close()

	if mnfVer.Scripts.Preinstall != "" {
		shell, flag := "sh", "-c"
		if os.PathListSeparator == ';' {
			shell, flag = "cmd.exe", "/c"
		}
		cmd := exec.Command(shell, flag, mnfVer.Scripts.Preinstall)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.Dir = module

		err := cmd.Run()
		if err != nil {
			return nil, stack.Stack(err, 0)
		}
	}

	buf := new(bytes.Buffer)
	sha := sha512.New()
	bar := &prog.Bar{
		Msg:  mnfVer.Name,
		Max:  int(get.ContentLength),
		Curr: 0,
		Pre: func(bar *prog.Bar) string {
			return fmt.Sprintf("%dKB/%dKB", bar.Curr/1000, bar.Max/1000)
		},
	}

	bars.Add(bar)
	_, err = io.Copy(sha, io.TeeReader(&reader{Bar: bar, Reader: get.Body}, buf))
	if err != nil {
		return nil, stack.Stack(err, 0)
	}
	if base64.StdEncoding.EncodeToString(sha.Sum(nil)) != strings.TrimPrefix(mnfVer.Dist.Integrity, "sha512-") {
		return nil, stack.Stack(fmt.Errorf("failed to verify %s", mnfVer.Name), 0)
	}

	err = tgz.Save(buf, module)
	if err != nil {
		return nil, stack.Stack(err, 0)
	}

	err = os.Symlink(module, pack)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, stack.Stack(err, 0)
	}

	err = os.Mkdir(filepath.Join(module, "node_modules"), 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, stack.Stack(err, 0)
	}

	optionalStacks := make(chan interface{}, len(mnfVer.OptionalDependencies))
	for name, tag := range mnfVer.OptionalDependencies {
		delete(mnfVer.Dependencies, name)
		go mnfVer.follow(name, tag, optionalStacks)
	}
	for index := 0; index < len(mnfVer.OptionalDependencies); index++ {
		optionalStacks := <-optionalStacks
		if dep, ok := optionalStacks.(*cache); ok {
			deps.Info.Dependencies[dep.Info.Name] = dep.Info
		}
	}
	close(optionalStacks)

	stacks := make(chan interface{}, len(mnfVer.Dependencies))
	for name, tag := range mnfVer.Dependencies {
		go mnfVer.follow(name, tag, stacks)
	}
	for index := 0; index < len(mnfVer.Dependencies); index++ {
		stacks := <-stacks
		if stacktrace, ok := stacks.(*stack.Stacktrace); ok {
			return nil, stacktrace
		}
		dep := stacks.(*cache)
		deps.Info.Dependencies[dep.Info.Name] = dep.Info
	}
	close(stacks)

	crt, err := os.Create(filepath.Join(module, "nagi.cache"))
	if err != nil {
		return nil, stack.Stack(err, 0)
	}
	defer crt.Close()

	err = gob.NewEncoder(crt).Encode(&deps)
	if err != nil {
		return nil, stack.Stack(err, 0)
	}

	if mnfVer.Scripts.Postinstall != "" {
		shell, flag := "sh", "-c"
		if os.PathListSeparator == ';' {
			shell, flag = "cmd.exe", "/c"
		}
		cmd := exec.Command(shell, flag, mnfVer.Scripts.Postinstall)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.Dir = module

		err := cmd.Run()
		if err != nil {
			return nil, stack.Stack(err, 0)
		}
	}

	return deps, nil
}

func open(file string) (*os.File, *sync.Mutex, error) {
	var opn *os.File
	var mut *sync.Mutex
	var err error
	if any, ok := muts.Load(file); ok {
		mut = any.(*sync.Mutex)
		opn, err = os.Open(file)
	} else {
		mut = new(sync.Mutex)
		muts.Store(file, mut)
	}
	mut.Lock()
	return opn, mut, err
}

func find(mnf *npm, tag string) (string, error) {
	for dist_tag := range mnf.Dist_tags {
		if dist_tag == tag {
			return mnf.Dist_tags[tag], nil
		}
	}

	vers := make([][]string, 0, len(mnf.Time))
	for ver, time := range mnf.Time {
		if ver == "modified" || ver == "created" {
			continue
		}
		vers = append(vers, []string{ver, time})
	}
	sort := semver.Sort(vers)
	builds := semver.Build(tag)

	for _, build := range builds {
		if build == nil {
			return "", fmt.Errorf("unknown type")
		}
		for index := range sort {
			if _, ok := mnf.Versions[sort[index][0]]; build.Match(semver.Version(sort[index][0])) && ok {
				return sort[index][0], nil
			}
		}
	}
	return mnf.Dist_tags["latest"], nil
}

func has(elems []string, elem ...string) bool {
	for index := range elems {
		curr := elems[index]
		for index := range elem {
			if elem[index] == curr {
				return true
			}
		}
	}
	return false
}

func fix(deps dep) error {
	dep := map[string]dep{}

	opn, _ := os.Open(lock)
	if opn != nil {
		err := gob.NewDecoder(opn).Decode(&dep)
		if err != nil {
			return err
		}
	}
	defer opn.Close()

	dep[deps.Name] = deps

	crt, err := os.Create(lock)
	if err != nil {
		return err
	}
	defer crt.Close()

	err = gob.NewEncoder(crt).Encode(&dep)
	if err != nil {
		return err
	}

	return nil
}

func (reader *reader) Read(bytes []byte) (int, error) {
	size, err := reader.Reader.Read(bytes)

	if size != 0 {
		reader.Bar.Curr += size
		bars.Print()
	}
	return size, err
}

func (mnfVer *npmVer) UnmarshalJSON(bytes []byte) error {
	type npmVerTmp struct {
		Name    string
		Version string
		Dist    struct {
			Tarball   string
			Integrity string
		}
		Engines interface{}
		Os      []string
		Cpu     []string
		Bin     interface{}
		Scripts struct {
			Preinstall  string
			Postinstall string
		}
		Dependencies         map[string]string
		OptionalDependencies map[string]string
	}

	mnfVerTmp := new(npmVerTmp)
	err := json.Unmarshal(bytes, mnfVerTmp)
	if err != nil {
		return err
	}

	mnfVer.Name = mnfVerTmp.Name
	mnfVer.Version = mnfVerTmp.Version
	mnfVer.Dist = mnfVerTmp.Dist
	mnfVer.Os = mnfVerTmp.Os
	mnfVer.Cpu = mnfVerTmp.Cpu
	mnfVer.Bin = make(map[string]string)
	mnfVer.Scripts = mnfVerTmp.Scripts
	mnfVer.Dependencies = mnfVerTmp.Dependencies
	mnfVer.OptionalDependencies = mnfVerTmp.OptionalDependencies

	if bins, ok := mnfVerTmp.Bin.(map[string]interface{}); ok {
		for cmd, bin := range bins {
			mnfVer.Bin[cmd] = bin.(string)
		}
	} else if bin, ok := mnfVerTmp.Bin.(string); ok {
		mnfVer.Bin = map[string]string{mnfVer.Name: bin}
	}

	if engines, ok := mnfVerTmp.Engines.(struct{ Node interface{} }); ok {
		mnfVer.Engines.Node = engines.Node.(string)
	} else {
		mnfVer.Engines.Node = ""
	}

	return nil
}

func register(cmd string, bin string, link string) error {
	err := os.Remove(filepath.Join(node_modules, ".bin", cmd))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	err = os.Chmod(filepath.Join(link, bin), 0755)
	if err != nil {
		return err
	}
	err = os.Symlink(filepath.Join(link, bin), filepath.Join(node_modules, ".bin", cmd))
	if err != nil {
		return err
	}
	return nil
}

func (mnfVer npmVer) follow(name string, tag string, stacks chan interface{}) {

	module := filepath.Join(nagi, "cache", mnfVer.Name+"@"+mnfVer.Version)
	mnf := new(npm)

	get, err := http.Get(REGISTRY + name)
	if err != nil {
		stacks <- stack.Stack(err, 0)
		return
	}
	if get.StatusCode != http.StatusOK {
		stacks <- stack.Stack(fmt.Errorf("bad status code was returned: %d", get.StatusCode), 0)
		return
	}
	defer get.Body.Close()

	err = json.NewDecoder(get.Body).Decode(mnf)
	if err != nil {
		stacks <- stack.Stack(err, 0)
		return
	}

	found, err := find(mnf, tag)
	if err != nil {
		stacks <- stack.Stack(fmt.Errorf("%s's request %s of %s: %s", mnfVer.Name, tag, name, err.Error()), 0)
		return
	}

	err = os.Mkdir(filepath.Dir(filepath.Join(module, "node_modules", name)), 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		stacks <- stack.Stack(err, 0)
		return
	}

	dep, stacktrace := mnf.Versions[found].bin(filepath.Join(module, "node_modules", name))
	if stacktrace != nil {
		stacks <- stacktrace
		return
	}
	stacks <- dep
}
