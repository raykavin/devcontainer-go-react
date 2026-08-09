package cli

import (
	"fmt"
	"time"

	"github.com/raykavin/gobox/cli"
)

// DisplayBanner prints the application banner to the terminal.
func DisplayBanner(name, description, version string) {
	_ = cli.PrintBanner(name)
	cli.PrintText(description)
	cli.PrintText("Header 1")
	cli.PrintText(fmt.Sprintf("Copyright (c) %d Your Company, All rights reserved.", time.Now().Year()))
	cli.PrintHeader("Version: " + version)
}
