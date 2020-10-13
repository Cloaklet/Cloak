package models

import (
	"database/sql"
	"reflect"
	"strings"
)

// BaseRepo implements transactional function execution and other useful utilities
type BaseRepo struct {
	db *sql.DB
}

// TxFunc stands for a function which decides whether a transaction should be aborted or committed.
//
// If any error is returned from this function, the transaction aborts; Otherwise it commits.
type TxFunc func(tx Transactional) error

// Transactional represents a type which can execute SQL just like `sql.Tx`.
type Transactional interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// WithTransaction executes `TxFunc` within a transaction.
func (r *BaseRepo) WithTransaction(handled TxFunc) error {
	var (
		err error
		tx  *sql.Tx
	)
	panicked := true

	tx, err = r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil || panicked {
			tx.Rollback()
		}
	}()

	// The closure will be given a `Transactional` interface on which only a handful SQL methods are available
	// It cannot be manually rolled back or committed.
	// To roll back, the closure must return an error.
	// To commit, the closure must return `nil`.
	err = handled(tx)

	if err == nil {
		err = tx.Commit()
	}

	panicked = false
	return err
}

// Field represents a struct field of a model instance
type Field struct {
	Name    string      // Field name
	Pointer interface{} // This is the pointer you will use for SQL scanning
}

// Fields returns all defined columns of given struct pointer `m`
//
// Notice: `m` must be a pointer, otherwise this function panics.
func (r *BaseRepo) Fields(m interface{}) []*Field {
	if !reflect.ValueOf(m).IsValid() || reflect.ValueOf(m).Kind() != reflect.Ptr {
		panic(`*BaseRepo.Fields() takes only a pointer argument`)
	}
	s := reflect.ValueOf(m).Elem()
	typeOfM := s.Type()
	fields := make([]*Field, s.NumField())

	for i := 0; i < s.NumField(); i++ {
		for _, value := range strings.Split(typeOfM.Field(i).Tag.Get("db"), ";") {
			v := strings.Split(value, ":")
			if len(v) < 2 || strings.TrimSpace(v[0]) != "column" {
				continue
			}
			fields[i] = &Field{
				Name:    strings.TrimSpace(strings.Join(v[1:], ":")),
				Pointer: s.Field(i).Addr().Interface(),
			}
		}
	}

	return fields
}

// FieldPointers is just like `Fields` but returns pointers instead.
//
// The returned pointers can be used to Scan values into fields.
func (r *BaseRepo) FieldPointers(m interface{}) []interface{} {
	fields := r.Fields(m)
	pointers := make([]interface{}, len(fields))
	for i, v := range fields {
		pointers[i] = v.Pointer
	}
	return pointers
}
