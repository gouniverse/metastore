package metastore

import (
	"os"
	"testing"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) *sql.DB {
	os.Remove(filepath) // remove database
	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		panic(err)
	}

	return db
}

func InitStore() *Store {
	db := InitDB("test_metastore_automigrate.db")
	return &Store{
		metaTableName:      "test_metastore_automigrate.db",
		db:                 db,
		automigrateEnabled: false,
	}
}

func TestWithAutoMigrate(t *testing.T) {
	db := InitDB("test_metastore_automigrate.db")

	s := Store{
		metaTableName:      "log_with_automigrate_false",
		db:                 db,
		automigrateEnabled: false,
	}

	f := WithAutoMigrate(true)
	f(&s)

	if s.automigrateEnabled != true {
		t.Fatalf("automigrateEnabled: Expected [true] received [%v]", s.automigrateEnabled)
	}

	s = Store{
		metaTableName:      "log_with_automigrate_true",
		db:                 db,
		automigrateEnabled: true,
	}

	f = WithAutoMigrate(false)
	f(&s)

	if s.automigrateEnabled == true {
		t.Fatalf("automigrateEnabled: Expected [true] received [%v]", s.automigrateEnabled)
	}
}

func TestWithDb(t *testing.T) {
	s := Store{
		metaTableName:      "log_with_automigrate_true",
		db:                 nil,
		automigrateEnabled: true,
	}

	db := InitDB("test")
	f := WithDb(db)
	f(&s)

	if s.db == nil {
		t.Fatalf("DB: Expected Initialized DB, received [%v]", s.db)
	}

}

func TestWithTableName(t *testing.T) {
	s := Store{
		metaTableName:      "",
		db:                 nil,
		automigrateEnabled: false,
	}
	table_name := "Table1"
	f := WithTableName(table_name)
	f(&s)
	if s.metaTableName != table_name {
		t.Fatalf("Expected logTableName [%v], received [%v]", table_name, s.metaTableName)
	}
	table_name = "Table2"
	f = WithTableName(table_name)
	f(&s)
	if s.metaTableName != table_name {
		t.Fatalf("Expected logTableName [%v], received [%v]", table_name, s.metaTableName)
	}
}

func Test_Store_AutoMigrate(t *testing.T) {
	db := InitDB("test_metastore_automigrate.db")

	s := NewStore(WithDb(db), WithTableName("log_with_automigrate"), WithAutoMigrate(true))

	s.AutoMigrate()

	if s.metaTableName != "log_with_automigrate" {
		t.Fatalf("Expected logTableName [log_with_automigrate] received [%v]", s.metaTableName)
	}
	if s.db == nil {
		t.Fatalf("DB Init Failure")
	}
	if s.automigrateEnabled != true {
		t.Fatalf("Failure:  WithAutoMigrate")
	}
}

func Test_Store_Set(t *testing.T) {
	db := InitDB("test_metastore_automigrate.db")
	s := NewStore(WithDb(db), WithTableName("log_with_automigrate"), WithAutoMigrate(true))

	objType := "Test_Obj"
	objID := "12345"
	key := "1234z"
	val := "123zx"
	ok := s.Set(objType, objID, key, val, 0)

	if !ok {
		t.Fatalf("Failure: Set")
	}
}

func Test_Store_SetJSON(t *testing.T) {
	db := InitDB("test_metastore_automigrate.db")
	s := NewStore(WithDb(db), WithTableName("log_with_automigrate"), WithAutoMigrate(true))

	objType := "Test_Obj"
	objID := "12345"
	key := "1234z"
	val := `{"a" : "b", "c" : "d"}`
	ok := s.SetJSON(objType, objID, key, val, 0)

	if !ok {
		t.Fatalf("Failure: Set")
	}
}

func Test_Store_Remove(t *testing.T) {
	db := InitDB("test_metastore_automigrate.db")
	s := NewStore(WithDb(db), WithTableName("log_with_automigrate"), WithAutoMigrate(true))

	objType := "Test_Obj"
	objID := "12345"
	key := "1234z"
	val := "123zx"
	ok := s.Set(objType, objID, key, val, 0)

	if !ok {
		t.Fatalf("Failure: Set")
	}

	s.Remove(objType, objID, key)
	ret := s.Get(objType, objID, key, "default")
	if ret != "default" {
		t.Fatalf("Unable to delete!!! Entry Persists")
	}
}

func Test_Store_Get(t *testing.T) {
	db := InitDB("test_metastore_automigrate.db")
	s := NewStore(WithDb(db), WithTableName("log_with_automigrate"), WithAutoMigrate(true))

	objType := "Test_Obj"
	objID := "12345"
	key := "1234z"
	val := "123zx"
	ok := s.Set(objType, objID, key, val, 0)

	if !ok {
		t.Fatalf("Failure: Set")
	}

	ret := s.Get(objType, objID, key, "default")
	if ret != val {
		t.Fatalf("Unable to Get: Expected [%v] Received [%v]", val, ret)
	}
}

func Test_Store_FindByKey(t *testing.T) {
	db := InitDB("test_metastore_automigrate.db")
	s := NewStore(WithDb(db), WithTableName("log_with_automigrate"), WithAutoMigrate(true))

	objType := "Test_Obj"
	objID := "12345"
	key := "1234z"
	val := "123zx"
	ok := s.Set(objType, objID, key, val, 0)

	if !ok {
		t.Fatalf("Failure: Set")
	}
	meta := s.FindByKey(objType, objID, key)
	if meta.ObjectID != objID {
		t.Fatalf("Incorrect Record Received [%v]", meta)
	}
}
func Test_Store_GetJSON(t *testing.T) {
	db := InitDB("test_metastore_automigrate.db")
	s := NewStore(WithDb(db), WithTableName("log_with_automigrate"), WithAutoMigrate(true))

	objType := "Test_Obj"
	objID := "12345"
	key := "1234z"
	val := `{"a" : "b", "c" : "d"}`
	ok := s.SetJSON(objType, objID, key, val, 10)

	if !ok {
		t.Fatalf("Failure: Set")
	}
	ret := s.GetJSON(objType, objID, key, nil)
	if ret == nil {
		t.Fatalf("Failure getting JSON value")
	}
}
