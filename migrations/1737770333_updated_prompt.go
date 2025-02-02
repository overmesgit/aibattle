package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1442582902")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"indexes": [
				"CREATE UNIQUE INDEX `+"`"+`idx_kUTN3bSnlR`+"`"+` ON `+"`"+`prompt`+"`"+` (`+"`"+`active`+"`"+`) WHERE `+"`"+`active`+"`"+`=TRUE",
				"CREATE INDEX `+"`"+`idx_0JksSzSHuZ`+"`"+` ON `+"`"+`prompt`+"`"+` (`+"`"+`status`+"`"+`)",
				"CREATE INDEX `+"`"+`idx_3zsa0R4Oku`+"`"+` ON `+"`"+`prompt`+"`"+` (`+"`"+`created`+"`"+`)",
				"CREATE INDEX `+"`"+`idx_HVAr1tGMvD`+"`"+` ON `+"`"+`prompt`+"`"+` (`+"`"+`user`+"`"+`)"
			]
		}`), &collection); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(5, []byte(`{
			"hidden": false,
			"id": "bool1260321794",
			"name": "active",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "bool"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1442582902")
		if err != nil {
			return err
		}

		// update collection data
		if err := json.Unmarshal([]byte(`{
			"indexes": []
		}`), &collection); err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("bool1260321794")

		return app.Save(collection)
	})
}
