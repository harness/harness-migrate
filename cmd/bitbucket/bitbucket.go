package bitbucket

import "github.com/alecthomas/kingpin/v2"

func Register(app *kingpin.Application) {
	cmd := app.Command("bitbucket", "migrate bitbucket data")
	//registerImport(cmd)
	registerConvert(cmd)
}
