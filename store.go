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

	if store.automigrateEnabled == true {
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
		log.Println(err)
		return
	}

	return
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

// SqlCreateTable returns a SQL string for creating the setting table
func (st *Store) SqlCreateTable() string {
	sqlMysql := `
	CREATE TABLE IF NOT EXISTS ` + st.metaTableName + ` (
	  ID varchar(40) NOT NULL PRIMARY KEY,
	  ObjectType longtext NOT NULL,
	  ObjectID longtext NOT NULL,
	  Key longtext NOT NULL,
	  Value longtext NOT NULL,
	  CreatedAt datetime NOT NULL,
	  UpdatedAt datetime,
	  DeletedAt datetime
	);
	`

	sqlPostgres := `
	CREATE TABLE IF NOT EXISTS "` + st.metaTableName + `" (
	  "ID" varchar(40) NOT NULL PRIMARY KEY,
	  "ObjectType" longtext NOT NULL,
	  "ObjectID" longtext NOT NULL,
	  "Key" longtext NOT NULL,
	  "Value" longtext NOT NULL,
	  "CreatedAt" timestamptz(6) NOT NULL,
	  "UpdatedAt" datetime,
	  "DeletedAt" timestamptz(6) 
	)
	`

	sqlSqlite := `
	CREATE TABLE IF NOT EXISTS "` + st.metaTableName + `" (
	  "ID" varchar(40) NOT NULL PRIMARY KEY,
	  "ObjectType" longtext NOT NULL,
	  "ObjectID" longtext NOT NULL,
	  "Key" longtext NOT NULL,
	  "Value" longtext NOT NULL,
	  "CreatedAt" datetime NOT NULL,
	  "UpdatedAt" datetime,
	  "DeletedAt" datetime 
	)
	`

	sql := "unsupported driver '" + st.dbDriverName + "'"

	if st.dbDriverName == "mysql" {
		sql = sqlMysql
	}
	if st.dbDriverName == "postgres" {
		sql = sqlPostgres
	}
	if st.dbDriverName == "sqlite" {
		sql = sqlSqlite
	}

	return sql
}

// FindByKey finds a cache by key
func (st *Store) FindByKey(objectType string, objectID string, key string) *Meta {
	sqlStr, _, _ := goqu.Dialect(st.dbDriverName).From(st.metaTableName).Where(goqu.C("ObjectType").Eq(objectType), goqu.C("ObjectID").Eq(objectID), goqu.C("Key").Eq(key)).Limit(1).ToSQL()

	if st.debug {
		log.Println(sqlStr)
	}

	var meta Meta
	err := sqlscan.Get(context.Background(), st.db, &meta, sqlStr)

	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			return nil
		}
		log.Fatal("Failed to execute query: ", err)
		return nil
	}
	return &meta
}

// Get gets a key from cache
func (st *Store) Get(objectType string, objectID string, key string, valueDefault string) string {
	cache := st.FindByKey(objectType, objectID, key)

	if cache != nil {
		return cache.Value
	}

	return valueDefault
}

// GetJSON gets a JSON key from cache
func (st *Store) GetJSON(objectType string, objectID string, key string, valueDefault interface{}) interface{} {
	meta := st.FindByKey(objectType, objectID, key)

	if meta != nil {
		jsonValue := meta.Value
		var e interface{}
		jsonError := json.Unmarshal([]byte(jsonValue), &e)
		if jsonError != nil {
			return valueDefault
		}

		return e
	}

	return valueDefault
}

// Remove deletes a meta key
func (st *Store) Remove(objectType string, objectID string, key string) (bool, error) {
	sqlStr, _, _ := goqu.Dialect(st.dbDriverName).From(st.metaTableName).Where(goqu.C("objecttype").Eq(objectType), goqu.C("objectid").Eq(objectID), goqu.C("key").Eq(key)).Delete().ToSQL()

	if st.debug {
		log.Println(sqlStr)
	}

	_, err := st.db.Exec(sqlStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, err
		}

		return false, err
	}
	return true, nil
}

// Set sets new key value pair
func (st *Store) Set(objectType string, objectID string, key string, value string, seconds int64) bool {
	meta := st.FindByKey(objectType, objectID, key)

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

	_, err := st.db.Exec(sqlStr)

	if err != nil {
		return false
	}

	return true
}

// SetJSON sets new key value pair
func (st *Store) SetJSON(objectType string, objectID string, key string, value interface{}, seconds int64) bool {
	jsonValue, jsonError := json.Marshal(value)
	if jsonError != nil {
		return false
	}

	return st.Set(objectType, objectID, key, string(jsonValue), seconds)
}
