package context

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type EnvContext struct {
	OS             string
	Distro         string
	Arch           string
	PackageManager string
	Shell          string
	CWD            string
}

func Detect() (*EnvContext, error) {
	cwd, _ := os.Getwd()
	ctx := &EnvContext{
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		PackageManager: detectPkgMgr(),
		Shell:          detectShell(),
		CWD:            cwd,
	}
	if runtime.GOOS == "linux" {
		ctx.Distro = detectDistro()
	}
	return ctx, nil
}

func detectDistro() string {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return ""
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
		}
	}
	return ""
}

func detectPkgMgr() string {
	// check in priority order per OS
	var candidates []string
	switch runtime.GOOS {
	case "linux":
		candidates = []string{"apt", "dnf", "yum", "pacman", "zypper", "apk"}
	case "darwin":
		candidates = []string{"brew"}
	case "windows":
		candidates = []string{"winget", "choco", "scoop"}
	}
	for _, p := range candidates {
		if _, err := exec.LookPath(p); err == nil {
			return p
		}
	}
	return ""
}

func detectShell() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("COMSPEC")
	}
	return os.Getenv("SHELL")
}

func (e *EnvContext) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "OS:   %s\n", e.OS)
	if e.Distro != "" {
		fmt.Fprintf(&sb, "Distro: %s\n", e.Distro)
	}
	fmt.Fprintf(&sb, "Arch: %s\n", e.Arch)
	if e.PackageManager != "" {
		fmt.Fprintf(&sb, "Package manager: %s\n", e.PackageManager)
	}
	if e.Shell != "" {
		fmt.Fprintf(&sb, "Shell: %s\n", e.Shell)
	}
	fmt.Fprintf(&sb, "CWD:  %s\n", e.CWD)
	return sb.String()
}
