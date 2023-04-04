package cloudbuild

import "github.com/alecthomas/kingpin/v2"

func Register(app *kingpin.Application) {
	cmd := app.Command("cloudbuild", "migrate google cloud build data")
	//registerMigrate(cmd)
	//registerExport(cmd)
	//registerImport(cmd)
	registerConvert(cmd)
}
