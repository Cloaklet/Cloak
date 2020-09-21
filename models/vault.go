package models

import "database/sql"

type Vault struct {
	ID         int64   `db:"column:id;" json:"id"`
	Path       string  `db:"column:path;" json:"path"`
	MountPoint *string `db:"column:mountpoint;" json:"mountpoint"`
	AutoReveal bool    `db:"column:autoreveal;" json:"autoreveal"`
	ReadOnly   bool    `db:"column:readonly;" json:"readonly"`
}

type VaultRepo struct {
	*BaseRepo
}

func NewVaultRepo(db *sql.DB) *VaultRepo {
	return &VaultRepo{&BaseRepo{db}}
}

// Create creates a new vault record
func (r *VaultRepo) Create(values map[string]interface{}, tx Transactional) (vault Vault, err error) {
	if tx == nil {
		tx = r.db
	}
	vault.Path = values["path"].(string)
	if v, ok := values["mountpoint"]; ok && (v != "") {
		mountpoint := v.(string)
		vault.MountPoint = &mountpoint
	}
	if v, ok := values["autoreveal"].(bool); ok {
		vault.AutoReveal = v
	}
	if v, ok := values["readonly"].(bool); ok {
		vault.ReadOnly = v
	}

	var result sql.Result
	result, err = tx.Exec(
		`INSERT INTO vaults (path, mountpoint, autoreveal, readonly) VALUES (?, ?, ?, ?);`,
		vault.Path, vault.MountPoint, vault.AutoReveal, vault.ReadOnly,
	)
	if err != nil {
		return
	} else {
		vault.ID, err = result.LastInsertId()
		return
	}
}

// Get gets a vault by ID
func (r *VaultRepo) Get(id int64, tx Transactional) (vault Vault, err error) {
	if tx == nil {
		tx = r.db
	}
	err = tx.QueryRow(`SELECT * FROM vaults WHERE id = ?;`, id).Scan(r.FieldPointers(&vault)...)
	return
}

// Update updates fields for given vault record
func (r *VaultRepo) Update(v *Vault, tx Transactional) error {
	if tx == nil {
		tx = r.db
	}
	_, err := tx.Exec(
		`UPDATE vaults SET path = ?, mountpoint = ?, autoreveal = ?, readonly = ? WHERE id = ?;`,
		v.Path, v.MountPoint, v.AutoReveal, v.ReadOnly, v.ID,
	)
	return err
}

// Delete permanently deletes given vault record.
// Its ID will be zero after the deletion.
func (r *VaultRepo) Delete(v *Vault, tx Transactional) error {
	if tx == nil {
		tx = r.db
	}
	if result, err := tx.Exec(`DELETE FROM vaults WHERE id = ?;`, v.ID); err != nil {
		return err
	} else {
		if _, err = result.RowsAffected(); err != nil {
			return err
		} else {
			v.ID = 0
			return nil
		}
	}
}

// List lists all existing vault records
func (r *VaultRepo) List(tx Transactional) (vaults []Vault, err error) {
	if tx == nil {
		tx = r.db
	}
	var rows *sql.Rows
	if rows, err = tx.Query(`SELECT * FROM vaults ORDER BY id ASC;`); err != nil {
		return
	} else {
		defer rows.Close()
		for rows.Next() {
			var vault Vault
			err = rows.Scan(r.FieldPointers(&vault)...)
			if err != nil {
				return
			}
			vaults = append(vaults, vault)
		}
	}
	return
}
