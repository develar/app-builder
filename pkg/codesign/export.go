// +build darwin

package codesign

import (
	"fmt"

	"github.com/alecthomas/kingpin"
	"github.com/develar/go-keychain"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("export-identity", "Export code sign identities (certificate + private key)")
	types := command.Flag("type", "The required types (targets).").Short('t').Enums("mas", "pkg", "app")

	command.Action(func(context *kingpin.ParseContext) error {
		return export(types)
	})
}

func export(types *[]string) error {
	query := keychain.NewItem()
	query.SetSecClass(keychain.KSecClassKey)
	//query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return err
	}

	fmt.Print(results)
	return nil
}