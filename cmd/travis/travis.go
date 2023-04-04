package travis

import "github.com/alecthomas/kingpin/v2"

func Register(app *kingpin.Application) {
	cmd := app.Command("travis", "migrate travis data")
	//registerMigrate(cmd)
	//registerExport(cmd)
	//registerImport(cmd)
	registerConvert(cmd)
}
