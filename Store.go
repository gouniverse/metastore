package metastore

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlserver"
	"github.com/georgysavva/scany/sqlscan"
	"github.com/gouniverse/uid"
)

// Store defines a session store
type Store struct {
	metaTableName      string
	db                 *sql.DB
	dbDriverName       string
	debugEnabled       bool
	automigrateEnabled bool
}

type NewStoreOptions struct {
	MetaTableName      string
	DB                 *sql.DB
	DbDriverName       string
	AutomigrateEnabled bool
	DebugEnabled       bool
}

// StoreOption options for the cache store
type StoreOption func(*Store)

// NewStore creates a new entity store
func NewStore(opts NewStoreOptions) (*Store, error) {
	store := &Store{
		metaTableName:      opts.MetaTableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 opts.DB,
		dbDriverName:       opts.DbDriverName,
		debugEnabled:       opts.DebugEnabled,
	}

	if store.metaTableName == "" {
		return nil, errors.New("meta store: metaTableName is required")
	}

	if store.db == nil {
		return nil, errors.New("meta store: DB is required")
	}

	if store.dbDriverName == "" {
		store.dbDriverName = store.DriverName(store.db)
	}

	if store.automigrateEnabled {
		store.AutoMigrate()
	}

	return store, nil
}

// AutoMigrate auto migrate
func (st *Store) AutoMigrate() {
	sql := st.SqlCreateTable()

	if st.debugEnabled {
		log.Println(sql)
	}

	_, err := st.db.Exec(sql)

	if err != nil {
		if st.debugEnabled {
			log.Println(err)
		}
	}
}

// DriverName finds the driver name from database
func (st *Store) DriverName(db *sql.DB) string {
	dv := reflect.ValueOf(db.Driver())
	driverFullName := dv.Type().String()
	if strings.Contains(driverFullName, "mysql") {
		return "mysql"
	}
	if strings.Contains(driverFullName, "postgres") || strings.Contains(driverFullName, "pq") {
		return "postgres"
	}
	if strings.Contains(driverFullName, "sqlite") {
		return "sqlite"
	}
	if strings.Contains(driverFullName, "mssql") {
		return "mssql"
	}
	return driverFullName
}

// EnableDebug - enables the debug option
func (st *Store) EnableDebug(debug bool) {
	st.debugEnabled = debug
}

// FindByKey finds a cache by key
func (st *Store) FindByKey(objectType string, objectID string, key string) (*Meta, error) {
	sqlStr, _, _ := goqu.Dialect(st.dbDriverName).
		From(st.metaTableName).
		Where(goqu.C("object_type").Eq(objectType), goqu.C("object_id").Eq(objectID), goqu.C("meta_key").Eq(key)).
		Limit(1).
		ToSQL()

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	var meta Meta
	err := sqlscan.Get(context.Background(), st.db, &meta, sqlStr)

	if err != nil {
		if err == sql.ErrNoRows {
			// Looks like this is now outdated for sqlscan
			return nil, nil // not really an error, no such row
		}

		if sqlscan.NotFound(err) {
			return nil, nil // not really an error, no such row
		}

		return nil, err
	}

	return &meta, nil
}

// Get gets a key from cache
func (st *Store) Get(objectType string, objectID string, key string, valueDefault string) (string, error) {
	meta, err := st.FindByKey(objectType, objectID, key)

	if err != nil {
		return "", err
	}

	if meta != nil {
		return meta.Value, nil
	}

	return valueDefault, nil
}

// GetJSON gets a JSON key from cache
func (st *Store) GetJSON(objectType string, objectID string, key string, valueDefault interface{}) (interface{}, error) {
	meta, err := st.FindByKey(objectType, objectID, key)

	if err != nil {
		return nil, err
	}

	if meta != nil {
		jsonValue := meta.Value
		var intrfc interface{}
		jsonError := json.Unmarshal([]byte(jsonValue), &intrfc)
		if jsonError != nil {
			return valueDefault, jsonError
		}

		return intrfc, nil
	}

	return valueDefault, nil
}

// Remove deletes a meta key
func (st *Store) Remove(objectType string, objectID string, key string) error {
	sqlStr, _, _ := goqu.Dialect(st.dbDriverName).
		From(st.metaTableName).
		Where(goqu.C("object_type").Eq(objectType), goqu.C("object_id").Eq(objectID), goqu.C("meta_key").Eq(key)).
		Delete().
		ToSQL()

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := st.db.Exec(sqlStr)
	if err != nil {
		if err == sql.ErrNoRows {
			// Looks like this is now outdated for sqlscan
			return nil // not really an error, already not there
		}

		if sqlscan.NotFound(err) {
			return nil
		}

		return err
	}

	return nil
}

// Set sets new key value pair
func (st *Store) Set(objectType string, objectID string, key string, value string) error {
	meta, err := st.FindByKey(objectType, objectID, key)

	if err != nil {
		return err
	}

	var sqlStr string
	var sqlErr error
	if meta == nil {
		var newMeta = &Meta{
			ID:         uid.HumanUid(),
			ObjectType: objectType,
			ObjectID:   objectID,
			Key:        key,
			Value:      value,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		sqlStr, _, sqlErr = goqu.Dialect(st.dbDriverName).
			Insert(st.metaTableName).
			Rows(newMeta).
			ToSQL()
	} else {
		fields := map[string]interface{}{}
		fields["meta_value"] = value
		fields["updated_at"] = time.Now()
		sqlStr, _, sqlErr = goqu.Dialect(st.dbDriverName).
			Update(st.metaTableName).
			Where(goqu.C("object_type").Eq(objectType), goqu.C("object_id").Eq(objectID), goqu.C("meta_key").Eq(key)).
			Set(fields).
			ToSQL()
	}

	if sqlErr != nil {
		return sqlErr
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	_, err = st.db.Exec(sqlStr)

	if err != nil {
		return err
	}

	return nil
}

// SetJSON sets new key value pair
func (st *Store) SetJSON(objectType string, objectID string, key string, value interface{}) error {
	jsonValue, jsonError := json.Marshal(value)

	if jsonError != nil {
		return jsonError
	}

	return st.Set(objectType, objectID, key, string(jsonValue))
}
