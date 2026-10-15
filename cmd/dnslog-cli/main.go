package main

import (
	"os"

	"github.com/lanyi1998/DNSlog-GO/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
