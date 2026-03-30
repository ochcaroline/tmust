package main

import "github.com/ochcaroline/tmust/cmd"

// version is stamped at build time:
//
//	go build -ldflags "-X main.version=1.2.3"
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
