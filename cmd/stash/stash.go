package stash

import "github.com/alecthomas/kingpin/v2"

func Register(app *kingpin.Application) {
	cmd := app.Command("stash", "migrate stash data")
	registerGit(cmd)
}
