package models

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type vaultTestSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
	repo *VaultRepo
}

func (s *vaultTestSuite) SetupSuite() {
	var db *sql.DB
	var err error
	db, s.mock, err = sqlmock.New()
	s.Require().NoError(err)
	s.repo = NewVaultRepo(db)
	s.Require().NotNil(s.repo)
}

func (s *vaultTestSuite) TearDownSuite() {
	s.repo.db.Close()
}

func (s *vaultTestSuite) AfterTest(_, _ string) {
	s.Require().NoError(s.mock.ExpectationsWereMet())
}

func (s *vaultTestSuite) Test_01_Vault() {
	path := "/test"
	mountpoint := ""

	// Create
	s.mock.ExpectExec(`INSERT INTO vaults(.+)`).
		WithArgs(path, nil).
		WillReturnResult(sqlmock.NewResult(1, 0))
	v, err := s.repo.Create(map[string]interface{}{
		"path":       path,
		"mountpoint": mountpoint,
	}, nil)
	s.Require().NoError(err)
	s.Require().IsType(Vault{}, v)
	s.Require().Greater(v.ID, int64(0))

	// Update
	newPath := "/test_new"
	v.Path = newPath
	s.mock.ExpectExec(`UPDATE vaults SET (.+)`).
		WithArgs(newPath, nil, v.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	err = s.repo.Update(&v, nil)
	s.Require().NoError(err)
	s.Require().EqualValues(newPath, v.Path)
	s.Require().EqualValues(v.ID, 1)

	// List
	s.mock.ExpectExec(`INSERT INTO vaults(.+)`).
		WithArgs(path, "/123").
		WillReturnResult(sqlmock.NewResult(2, 0))
	v2, err := s.repo.Create(map[string]interface{}{
		"path":       path,
		"mountpoint": "/123",
	}, nil)
	s.Require().NoError(err)
	s.Require().IsType(Vault{}, v2)
	s.Require().Greater(v2.ID, int64(0))
	s.Require().NotEqualValues(v.ID, v2.ID)

	s.mock.ExpectQuery(`SELECT \* FROM vaults(.+)`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "path", "mountpoint"}).
				AddRow(1, newPath, nil).
				AddRow(2, path, nil),
		)
	vaults, err := s.repo.List(nil)
	s.Require().NoError(err)
	s.Require().Len(vaults, 2)

	// Get
	s.mock.ExpectQuery(`SELECT \* FROM vaults WHERE id = \?(.+)`).
		WithArgs(v.ID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "path", "mountpoint"}).
				AddRow(1, newPath, nil),
		)
	vault, err := s.repo.Get(v.ID, nil)
	s.Require().NoError(err)
	s.Require().IsType(v, vault)
	s.Require().EqualValues(v.ID, vault.ID)
	s.Require().EqualValues(v.Path, vault.Path)

	// Delete
	s.mock.ExpectExec(`DELETE FROM vaults WHERE id = \?(.+)`).
		WithArgs(v.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	err = s.repo.Delete(&v, nil)
	s.Require().NoError(err)
	s.Require().EqualValues(0, v.ID)
}

func Test_VaultRepo(t *testing.T) {
	suite.Run(t, new(vaultTestSuite))
}
