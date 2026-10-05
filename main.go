package main

import (
	"os"
	"path/filepath"

	"github.com/Sn0wo2/caelum"
	"github.com/Sn0wo2/hmcl-token-manager/internal/cli"
)

func main() {
	if err := cli.Run(filepath.Join(os.Getenv("APPDATA"), ".hmcl")); err != nil {
		caelum.New(caelum.Config{
			Targets: []caelum.Target{{Writer: os.Stderr}},
		}).Error("HMCL Token Manager failed", "error", err)
		os.Exit(1)
	}
}
