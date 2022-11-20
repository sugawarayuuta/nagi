package main

import "strings"

func Help(arg string) string {

	help := []string{
		"command\tdescription\tusage\tnpm",
		"\t\t\t",
		"help\tshows this menu\thelp <command>\tnpm help",
		"version\tshows versions\tversion\tnpm version",
		"install\tinstalls packages\tinstall <package>\tnpm install",
		"uninstall\tremoves packages\tuninstall <package>\tnpm uninstall",
		"cache\tcleans cache\tcache clean\tnpm cache clean",
		"ping\tping the registry\tping\tnpm ping",
		"ls\tlist dependencies\tls <package>\tnpm ls",
		"set-script\tset a task\tset-script <task> <command>\tnpm set-script",
		"run\trun a task\trun <task>\tnpm run",
		"start/stop/restart/test\trun the specific task\tstart <task>\tnpm start",
		"init\tinitialize a project\tinit <package>\tnpm init",
		"exec\trun a task inside a package\texec <package>\tnpm exec/npx",
	}

	if arg == "" {
		return strings.Join(help, "\n")
	}

	switch arg {
	case "help":
		return strings.Join(append(help[:2], help[0+2]), "\n")
	case "version":
		return strings.Join(append(help[:2], help[1+2]), "\n")
	case "add", "install", "i":
		return strings.Join(append(help[:2], help[2+2]), "\n")
	case "remove", "rm", "r", "uninstall", "unlink", "un":
		return strings.Join(append(help[:2], help[3+2]), "\n")
	case "cache":
		return strings.Join(append(help[:2], help[4+2]), "\n")
	case "ping":
		return strings.Join(append(help[:2], help[5+2]), "\n")
	case "list", "ls":
		return strings.Join(append(help[:2], help[6+2]), "\n")
	case "set-script":
		return strings.Join(append(help[:2], help[7+2]), "\n")
	case "run-script", "run":
		return strings.Join(append(help[:2], help[8+2]), "\n")
	case "start", "stop", "restart", "test":
		return strings.Join(append(help[:2], help[9+2]), "\n")
	case "init", "create":
		return strings.Join(append(help[:2], help[10+2]), "\n")
	case "exec":
		return strings.Join(append(help[:2], help[11+2]), "\n")
	}
	return "help for " + arg + " was not found"
}
