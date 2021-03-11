package metastore

import (
	"encoding/json"
	"errors"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Store defines a session store
type Store struct {
	metaTableName      string
	db                 *gorm.DB
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

// WithDriverAndDNS sets the driver and the DNS for the database for the cache store
func WithDriverAndDNS(driverName string, dsn string) StoreOption {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	return func(s *Store) {
		s.db = db
	}
}

// WithGormDb sets the GORM database for the cache store
func WithGormDb(db *gorm.DB) StoreOption {
	return func(s *Store) {
		s.db = db
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
	st.db.Table(st.metaTableName).AutoMigrate(&Meta{})
}

// FindByKey finds a cache by key
func (st *Store) FindByKey(objectType string, objectID string, key string) *Meta {
	// log.Println(key)

	meta := &Meta{}

	result := st.db.Table(st.metaTableName).Where("`object_type` = ?", objectType).Where("`object_id` = ?", objectID).Where("`meta_key` = ?", key).First(&meta)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}

		log.Panic(result.Error)
	}

	return meta
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
		jsonError := json.Unmarshal([]byte(jsonValue), e)
		if jsonError != nil {
			return valueDefault
		}

		return e
	}

	return valueDefault
}

// Remove deletes a meta key
func (st *Store) Remove(objectType string, objectID string, key string) {
	st.db.Table(st.metaTableName).Where("`object_type` = ?", objectType).Where("`object_id` = ?", objectID).Where("`meta_key` = ?", key).Delete(Meta{})
}

// Set sets new key value pair
func (st *Store) Set(objectType string, objectID string, key string, value string) bool {
	meta := st.FindByKey(objectType, objectID, key)

	if meta != nil {
		meta.Value = value
		dbResult := st.db.Table(st.metaTableName).Save(&meta)
		if dbResult != nil {
			return false
		}
		return true
	}

	var newMeta = Meta{ObjectType: objectType, ObjectID: objectID, Key: key, Value: value}

	dbResult := st.db.Table(st.metaTableName).Create(&newMeta)

	if dbResult.Error != nil {
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

	return st.Set(objectType, objectID, key, string(jsonValue))
}
