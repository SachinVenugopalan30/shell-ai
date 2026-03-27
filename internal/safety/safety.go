package safety

import (
	"regexp"
	"strings"
)

type CheckResult struct {
	IsDestructive      bool
	DestructiveReason  string
	PkgManagerMismatch bool
	MismatchDetail     string
}

// known package managers for mismatch detection
var pkgMgrs = []string{"apt", "dnf", "yum", "pacman", "zypper", "apk", "brew", "winget", "choco", "scoop"}

type pattern struct {
	re     *regexp.Regexp
	reason string
}

var destructive = []pattern{
	{regexp.MustCompile(`rm\s+(-[a-zA-Z]*r[a-zA-Z]*|-r)(\s|$)`), "recursive deletion"},
	{regexp.MustCompile(`dd\s+if=`), "raw disk write with dd"},
	{regexp.MustCompile(`mkfs`), "filesystem format"},
	{regexp.MustCompile(`:\(\)\s*\{`), "fork bomb"},
	{regexp.MustCompile(`chmod\s+(-R\s+)?777\s+/`), "chmod 777 on root"},
	{regexp.MustCompile(`(?i)DROP\s+(TABLE|DATABASE)`), "SQL DROP statement"},
	{regexp.MustCompile(`\btruncate\b`), "file truncation"},
	{regexp.MustCompile(`\bfdisk\b`), "disk partitioning"},
	{regexp.MustCompile(`\bshred\b`), "file shredding"},
	{regexp.MustCompile(`>\s*/dev/sd`), "raw device write"},
	{regexp.MustCompile(`mv\s+/[\s*]`), "moving root filesystem"},
	{regexp.MustCompile(`\|\s*(ba)?sh`), "piped to shell"},
	{regexp.MustCompile(`\|\s*zsh`), "piped to shell"},
	{regexp.MustCompile(`(bash|sh|zsh)\s+-c\s+`), "nested shell execution"},
	{regexp.MustCompile(`\beval\s+`), "eval execution"},
	{regexp.MustCompile(`\bexec\s+`), "exec execution"},
	{regexp.MustCompile(`base64.*\|\s*(ba)?sh`), "encoded shell execution"},
	{regexp.MustCompile(`/bin/rm\s`), "deletion via absolute path"},
	{regexp.MustCompile(`/sbin/mkfs`), "format via absolute path"},
	{regexp.MustCompile(`rm\s+--recursive`), "recursive deletion (long flag)"},
	{regexp.MustCompile(`rm\s+-r\s+-f`), "recursive force deletion (split flags)"},
	{regexp.MustCompile(`(curl|wget)\s+.*\|`), "download piped to command"},
}

func Check(cmd, detectedPkgMgr string) *CheckResult {
	res := &CheckResult{}

	// check destructive patterns
	for _, p := range destructive {
		if p.re.MatchString(cmd) {
			res.IsDestructive = true
			res.DestructiveReason = p.reason
			break
		}
	}

	// check pkg manager mismatch
	if detectedPkgMgr != "" {
		tok := firstToken(cmd)
		if tok != detectedPkgMgr && isKnownPkgMgr(tok) {
			res.PkgManagerMismatch = true
			res.MismatchDetail = "command uses '" + tok + "' but system uses '" + detectedPkgMgr + "'"
		}
	}

	return res
}

// firstToken returns the first real command token, skipping sudo
func firstToken(cmd string) string {
	parts := strings.Fields(cmd)
	for _, p := range parts {
		if p != "sudo" {
			return p
		}
	}
	return ""
}

func isKnownPkgMgr(s string) bool {
	for _, m := range pkgMgrs {
		if s == m {
			return true
		}
	}
	return false
}
