package metastore

import (
	"context"
	"encoding/json"
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
	debug              bool
	automigrateEnabled bool
}

// StoreOption options for the cache store
type StoreOption func(*Store)

// WithAutoMigrate sets the table name for the cache store
func WithAutoMigrate(automigrateEnabled bool) StoreOption {
	return func(s *Store) {
		s.automigrateEnabled = automigrateEnabled
	}
}

// WithGormDb sets the GORM database for the cache store
func WithDb(db *sql.DB) StoreOption {
	return func(s *Store) {
		s.db = db
		s.dbDriverName = s.DriverName(s.db)
	}
}

// WithTableName sets the table name for the cache store
func WithTableName(metaTableName string) StoreOption {
	return func(s *Store) {
		s.metaTableName = metaTableName
	}
}

// NewStore creates a new entity store
func NewStore(opts ...StoreOption) *Store {
	store := &Store{}
	for _, opt := range opts {
		opt(store)
	}

	if store.db == nil {
		log.Panic("log store: db is required")
		return nil
	}

	if store.dbDriverName == "" {
		log.Panic("log store: dbDriverName is required")
		return nil
	}

	if store.metaTableName == "" {
		log.Panic("Meta store: metaTableName is required")
	}

	if store.automigrateEnabled {
		store.AutoMigrate()
	}

	return store
}

// AutoMigrate auto migrate
func (st *Store) AutoMigrate() {
	sql := st.SqlCreateTable()

	if st.debug {
		log.Println(sql)
	}

	_, err := st.db.Exec(sql)

	if err != nil {
		if st.debug {
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
	st.debug = debug
}

// FindByKey finds a cache by key
func (st *Store) FindByKey(objectType string, objectID string, key string) (*Meta, error) {
	sqlStr, _, _ := goqu.Dialect(st.dbDriverName).From(st.metaTableName).Where(goqu.C("ObjectType").Eq(objectType), goqu.C("ObjectID").Eq(objectID), goqu.C("Key").Eq(key)).Limit(1).ToSQL()

	if st.debug {
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
	sqlStr, _, _ := goqu.Dialect(st.dbDriverName).From(st.metaTableName).Where(goqu.C("objecttype").Eq(objectType), goqu.C("objectid").Eq(objectID), goqu.C("key").Eq(key)).Delete().ToSQL()

	if st.debug {
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
func (st *Store) Set(objectType string, objectID string, key string, value string, seconds int64) error {
	meta, err := st.FindByKey(objectType, objectID, key)

	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(time.Second * time.Duration(seconds))

	var newMeta = &Meta{ObjectType: objectType, ObjectID: objectID, Key: key, Value: value}
	if meta == nil {
		meta = newMeta
		meta.Value = value
		meta.ID = uid.HumanUid()
		meta.CreatedAt = time.Now()
		meta.DeletedAt = &expiresAt
		meta.UpdatedAt = time.Now()
	} else {
		meta.Value = value
		meta.DeletedAt = &expiresAt
		meta.UpdatedAt = time.Now()
	}

	sqlStr, _, _ := goqu.Dialect(st.dbDriverName).Insert(st.metaTableName).Rows(meta).ToSQL()
	log.Println(sqlStr)

	if st.debug {
		log.Println(sqlStr)
	}

	_, err = st.db.Exec(sqlStr)

	if err != nil {
		return err
	}

	return nil
}

// SetJSON sets new key value pair
func (st *Store) SetJSON(objectType string, objectID string, key string, value interface{}, seconds int64) error {
	jsonValue, jsonError := json.Marshal(value)

	if jsonError != nil {
		return jsonError
	}

	return st.Set(objectType, objectID, key, string(jsonValue), seconds)
}
