package sys

import (
	"os/exec"
	"runtime"
)

var (
	NODE = Node()
	OS   = Os()
	ARCH = Arch()
)

func Node() string {
	cmd := exec.Command("node", "-v")
	out, _ := cmd.Output()
	return string(out)
}

func Os() string {
	switch runtime.GOOS {
	case "aix", "darwin", "freebsd", "linux", "openbsd":
		return runtime.GOOS
	case "windows":
		return "win32"
	case "solaris":
		return "sunos"
	}
	return ""
}

func Arch() string {
	switch runtime.GOARCH {
	case "arm", "arm64", "mips", "s390x", "ppc64":
		return runtime.GOARCH
	case "amd64":
		return "x64"
	case "386":
		return "ia32"
	}
	return ""
}
