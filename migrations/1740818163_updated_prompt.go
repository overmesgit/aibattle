package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1442582902")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(7, []byte(`{
			"hidden": false,
			"id": "select3571151285",
			"maxSelect": 1,
			"name": "language",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "select",
			"values": [
				"go",
				"py",
				"js"
			]
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1442582902")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(7, []byte(`{
			"hidden": false,
			"id": "select3571151285",
			"maxSelect": 1,
			"name": "language",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "select",
			"values": [
				"go",
				"py"
			]
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
