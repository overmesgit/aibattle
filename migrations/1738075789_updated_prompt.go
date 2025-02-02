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
				"CREATE UNIQUE INDEX `+"`"+`idx_kUTN3bSnlR`+"`"+` ON `+"`"+`prompt`+"`"+` (\n  `+"`"+`active`+"`"+`,\n  `+"`"+`user`+"`"+`\n) WHERE `+"`"+`active`+"`"+`=TRUE",
				"CREATE INDEX `+"`"+`idx_0JksSzSHuZ`+"`"+` ON `+"`"+`prompt`+"`"+` (`+"`"+`status`+"`"+`)",
				"CREATE INDEX `+"`"+`idx_3zsa0R4Oku`+"`"+` ON `+"`"+`prompt`+"`"+` (`+"`"+`created`+"`"+`)",
				"CREATE INDEX `+"`"+`idx_HVAr1tGMvD`+"`"+` ON `+"`"+`prompt`+"`"+` (`+"`"+`user`+"`"+`)"
			]
		}`), &collection); err != nil {
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
			"indexes": [
				"CREATE UNIQUE INDEX `+"`"+`idx_kUTN3bSnlR`+"`"+` ON `+"`"+`prompt`+"`"+` (`+"`"+`active`+"`"+`) WHERE `+"`"+`active`+"`"+`=TRUE",
				"CREATE INDEX `+"`"+`idx_0JksSzSHuZ`+"`"+` ON `+"`"+`prompt`+"`"+` (`+"`"+`status`+"`"+`)",
				"CREATE INDEX `+"`"+`idx_3zsa0R4Oku`+"`"+` ON `+"`"+`prompt`+"`"+` (`+"`"+`created`+"`"+`)",
				"CREATE INDEX `+"`"+`idx_HVAr1tGMvD`+"`"+` ON `+"`"+`prompt`+"`"+` (`+"`"+`user`+"`"+`)"
			]
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
