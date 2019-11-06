package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func findGo() string {
	path, _ := exec.LookPath("go")
	return path
}

var gobin = flag.String("go", findGo(), "Go binary")

var targets = []struct{ os, arch string }{
	{"linux", "amd64"},
	{"linux", "386"},
	{"linux", "arm64"},
	{"darwin", "amd64"},
	{"windows", "amd64"},
	{"windows", "386"},
	{"openbsd", "amd64"},
	{"freebsd", "amd64"},
}

const ldflags = `-buildid= ` +
	`-X github.com/decred/dcrd/internal/version.BuildMetadata=release ` +
	`-X github.com/decred/dcrd/internal/version.PreRelease=rc1 ` +
	`-X github.com/decred/dcrwallet/version.BuildMetadata=release ` +
	`-X github.com/decred/dcrwallet/version.PreRelease=rc1`

const tags = "safe"

var tools = []struct{ tool, builddir string }{
	{"decred.org/dcrwallet", "./dcrwallet"},
	{"github.com/decred/dcrd", "./dcrd"},
	{"github.com/decred/dcrd/cmd/dcrctl", "./dcrd"},
	{"github.com/decred/dcrd/cmd/promptsecret", "./dcrd"},
	{"github.com/decred/dcrlnd/cmd/dcrlnd", "./dcrlnd"},
}

func main() {
	flag.Parse()
	logvers()
	for i := range targets {
		for j := range tools {
			build(tools[j].tool, targets[i].os, targets[i].arch, tools[j].builddir)
		}
	}
}

func logvers() {
	output, err := exec.Command(*gobin, "version").CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("releasing with %s %s", *gobin, output)
}

func build(tool, goos, arch, builddir string) {
	exe := path.Base(tool) // TODO: fix for /v2+
	if goos == "windows" {
		exe += ".exe"
	}
	out := filepath.Join("..", "bin", goos+"-"+arch, exe)
	log.Printf("build: %s", out[3:]) // trim off leading "../"
	gocmd(goos, arch, builddir, "build", "-trimpath", "-tags", tags, "-o", out, "-ldflags", ldflags, tool)
}

func gocmd(goos, arch, builddir string, args ...string) {
	os.Setenv("GOOS", goos)
	os.Setenv("GOARCH", arch)
	os.Setenv("CGO_ENABLED", "0")
	os.Setenv("GOFLAGS", "")
	cmd := exec.Command(*gobin, args...)
	cmd.Dir = builddir
	output, err := cmd.CombinedOutput()
	if len(output) != 0 {
		log.Printf("go '%s'\n%s", strings.Join(args, `' '`), output)
	}
	if err != nil {
		log.Fatal(err)
	}
}
