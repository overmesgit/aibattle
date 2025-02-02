package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `{
			"createRule": null,
			"deleteRule": null,
			"fields": [
				{
					"autogeneratePattern": "[a-z0-9]{15}",
					"hidden": false,
					"id": "text3208210256",
					"max": 15,
					"min": 15,
					"name": "id",
					"pattern": "^[a-z0-9]+$",
					"presentable": false,
					"primaryKey": true,
					"required": true,
					"system": true,
					"type": "text"
				},
				{
					"cascadeDelete": false,
					"collectionId": "pbc_1442582902",
					"hidden": false,
					"id": "relation2881271877",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "prompt1",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "relation"
				},
				{
					"cascadeDelete": false,
					"collectionId": "pbc_1442582902",
					"hidden": false,
					"id": "relation850782719",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "prompt2",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "relation"
				},
				{
					"hidden": false,
					"id": "select217473038",
					"maxSelect": 1,
					"name": "winner",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "select",
					"values": [
						"prompt1",
						"prompt2"
					]
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text3437106334",
					"max": 0,
					"min": 0,
					"name": "output",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": false,
					"system": false,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "autodate2990389176",
					"name": "created",
					"onCreate": true,
					"onUpdate": false,
					"presentable": false,
					"system": false,
					"type": "autodate"
				},
				{
					"hidden": false,
					"id": "autodate3332085495",
					"name": "updated",
					"onCreate": true,
					"onUpdate": true,
					"presentable": false,
					"system": false,
					"type": "autodate"
				}
			],
			"id": "pbc_613051002",
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_gbQ3aYEHfc` + "`" + ` ON ` + "`" + `battle` + "`" + ` (` + "`" + `prompt1` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_HGrn94kKOb` + "`" + ` ON ` + "`" + `battle` + "`" + ` (` + "`" + `prompt2` + "`" + `)"
			],
			"listRule": null,
			"name": "battle",
			"system": false,
			"type": "base",
			"updateRule": null,
			"viewRule": null
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_613051002")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
