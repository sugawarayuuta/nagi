package main

import (
	"fmt"
	"io"
	"log"
	"nagi/prog"
	"nagi/stack"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

type npmVer struct {
	Name    string
	Version string
	Dist    struct {
		Tarball   string
		Integrity string
	}
	Engines struct {
		Node string
	}
	Os      []string
	Cpu     []string
	Bin     map[string]string
	Scripts struct {
		Preinstall  string
		Postinstall string
	}
	Dependencies         map[string]string
	OptionalDependencies map[string]string
}

type npm struct {
	Dist_tags map[string]string `json:"dist-tags"`
	Time      map[string]string
	Versions  map[string]npmVer
}

type dep struct {
	Name         string
	Version      string
	Tarball      string
	Integrity    string
	Engine       string
	Os           []string
	Cpu          []string
	Bin          map[string]string
	Pre          string
	Post         string
	Dependencies map[string]dep
}

type pkg struct {
	Dependencies map[string]string
	Scripts      map[string]string
}

type cache struct {
	Info dep
	Use  []string
}

type field struct {
	Map map[string]interface{}
	Key string
	Val interface{}
}

type writer struct{}

type reader struct {
	Bar    *prog.Bar
	Reader io.Reader
}

const (
	ERR      = "\x1b[31m[ERR]\x1b[0m"
	INFO     = "\x1b[32m[INFO]\x1b[0m"
	REGISTRY = "https://registry.npmjs.org/"
)

var (
	home, _      = os.UserHomeDir()
	work, _      = os.Getwd()
	dir          = work
	nagi         = filepath.Join(home, ".nagi")
	node_modules string
	lock         string
	tab          = new(tabwriter.Writer)
	muts         = sync.Map{}
	bars         = prog.New()
)

func main() {
	now := time.Now()

	log.SetFlags(0)
	log.SetOutput(new(writer))

	var cmd string
	var arg string
	var args []string
	var opts []string

	for index := range os.Args[1:] {
		if index == 0 {
			cmd = os.Args[1+index]
			continue
		}
		if !strings.HasPrefix(os.Args[1+index], "-") {
			if arg == "" {
				arg = os.Args[1+index]
			}
			args = append(args, os.Args[1+index])
			continue
		}
		opts = append(opts, os.Args[1+index])
	}

	if has(opts, "--global", "-g") {
		dir = filepath.Join(nagi, "global")
	}
	node_modules = filepath.Join(dir, "node_modules")
	lock = filepath.Join(dir, "nagi.lock")

	switch cmd {
	case "add", "install", "i":
		for _, arg := range args {
			stacktrace := Add(arg, opts)
			if stacktrace != nil {
				handle(stacktrace)
			}
		}
		stacktrace := Install(opts)
		if stacktrace != nil {
			handle(stacktrace)
		}
		fmt.Println()
	case "remove", "rm", "r", "uninstall", "unlink", "un":
		for _, arg := range args {
			stacktrace := Remove(arg, opts)
			if stacktrace != nil {
				handle(stacktrace)
			}
		}
		stacktrace := Install(opts)
		if stacktrace != nil {
			handle(stacktrace)
		}
		fmt.Println()
	case "update", "upgrade":
		stacktrace := Update(args, opts)
		if stacktrace != nil {
			handle(stacktrace)
		}
	case "cache":
		if arg == "clean" {
			stacktrace := Clean(filepath.Join(nagi, "cache"))
			if stacktrace != nil {
				handle(stacktrace)
			}
		}
	case "ping":
		code, stacktrace := Ping()
		if stacktrace != nil {
			handle(stacktrace)
		}
		log.Println("status code:", code)
	case "list", "ls":
		list, stacktrace := List(args)
		if stacktrace != nil {
			handle(stacktrace)
		}
		fmt.Println(strings.Join(list, "\n"))
	case "set-script":
		stacktrace := Set(args)
		if stacktrace != nil {
			handle(stacktrace)
		}
	case "run-script", "run":
		stacktrace := Run(arg)
		if stacktrace != nil {
			handle(stacktrace)
		}
	case "start", "restart", "stop", "test":
		stacktrace := Run(cmd)
		if stacktrace != nil {
			handle(stacktrace)
		}
	case "init", "create":
		stacktrace := Init(arg)
		if stacktrace != nil {
			handle(stacktrace)
		}
	case "exec":
		stacktrace := Exec(arg)
		if stacktrace != nil {
			handle(stacktrace)
		}
	case "version":
		version := Version()
		fmt.Println(version)
	case "help":
		help := Help(arg)
		tab.Init(os.Stdout, 0, 8, 0, '\t', 0)
		fmt.Fprintln(tab, help)
		tab.Flush()
	default:
		return
	}
	log.Println(INFO, "finished in", time.Since(now).Seconds(), "second(s)")
}

func (writer writer) Write(req []byte) (int, error) {
	return fmt.Print(time.Now().Format(time.Kitchen) + " " + string(req))
}

func handle(stacktrace *stack.Stacktrace) {
	log.Println(ERR, stacktrace.Err)
	fmt.Println("below is where the error happened")
	for _, frame := range stacktrace.Frames {
		fmt.Println("at line:", frame.Ln, "in file:", filepath.Base(frame.Fl))
	}
	os.Exit(1)
}
