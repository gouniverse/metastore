package metastore

// SqlCreateTable returns a SQL string for creating the setting table
func (st *Store) SqlCreateTable() string {
	sqlMysql := `
	CREATE TABLE IF NOT EXISTS ` + st.metaTableName + ` (
	  id 			varchar(40)		NOT NULL PRIMARY KEY,
	  object_type	varchar(100) 	NOT NULL,
	  object_id		varchar(40) 	NOT NULL,
	  meta_key  	varchar(255)	NOT NULL,
	  meta_value 	longtext,
	  created_at 	datetime		NOT NULL,
	  updated_at 	datetime		NOT NULL,
	  deleted_at 	datetime
	);
	`

	sqlPostgres := `
	CREATE TABLE IF NOT EXISTS "` + st.metaTableName + `" (
	  "id"			varchar(40)		NOT NULL PRIMARY KEY,
	  "object_type"	varchar(100) 	NOT NULL,
	  "object_id"	varchar(40) 	NOT NULL,
	  "meta_key"  	varchar(255)	NOT NULL,
	  "meta_value" 	longtext,
	  "created_at"	timestamptz(6)	NOT NULL,
	  "updated_at"	timestamptz(6)	NOT NULL,
	  "deleted_at"	timestamptz(6) 
	)
	`

	sqlSqlite := `
	CREATE TABLE IF NOT EXISTS "` + st.metaTableName + `" (
	  "id"			varchar(40)		NOT NULL PRIMARY KEY,
	  "object_type"	varchar(100) 	NOT NULL,
	  "object_id"	varchar(40) 	NOT NULL,
	  "meta_key"  	varchar(255)	NOT NULL,
	  "meta_value" 	longtext,
	  "created_at"	datetime		NOT NULL,
	  "updated_at"	datetime		NOT NULL,
	  "deleted_at"	datetime 
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
