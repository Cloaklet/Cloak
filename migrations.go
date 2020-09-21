package main

import (
	"database/sql"
	"fmt"
	"github.com/lopezator/migrator"
)

var migrations []interface{}

func init() {
	migrations = []interface{}{
		&migrator.Migration{
			Name: "Create vaults table",
			Func: func(tx *sql.Tx) error {
				_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS vaults (
    id INTEGER PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    mountpoint TEXT UNIQUE
);`)
				return err
			},
		},
		&migrator.Migration{
			Name: "Add autoreveal & readonly column",
			Func: func(tx *sql.Tx) error {
				for _, column := range []string{"autoreveal", "readonly"} {
					_, err := tx.Exec(fmt.Sprintf(`ALTER TABLE vaults ADD COLUMN %s BOOLEAN DEFAULT false;`, column))
					if err != nil {
						return err
					}
				}
				return nil
			},
		},
		&migrator.Migration{
			Name: "Change NULL mountpoints to empty strings",
			Func: func(tx *sql.Tx) error {
				_, err := tx.Exec(`ALTER TABLE vaults RENAME TO vaults_backup;
CREATE TABLE vaults(
    id INTEGER PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    mountpoint TEXT,
    autoreveal BOOLEAN DEFAULT false,
    readonly BOOLEAN DEFAULT false
);
INSERT INTO vaults (id, path, mountpoint, autoreveal, readonly)
	SELECT id, path, mountpoint, autoreveal, readonly FROM vaults_backup;
DROP TABLE vaults_backup;
UPDATE vaults SET mountpoint = "" WHERE mountpoint IS NULL;`)
				return err
			},
		},
	}
}
