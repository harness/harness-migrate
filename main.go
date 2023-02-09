// Copyright 2023 Harness Inc. All rights reserved.

package main

import (
	"github.com/harness/harness-migrate/cmd"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cmd.Command()
}
