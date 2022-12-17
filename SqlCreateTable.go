package metastore

// SqlCreateTable returns a SQL string for creating the setting table
func (st *Store) SqlCreateTable() string {
	sqlMysql := `
	CREATE TABLE IF NOT EXISTS ` + st.metaTableName + ` (
	  id varchar(40) NOT NULL PRIMARY KEY,
	  object_type longtext NOT NULL,
	  object_id longtext NOT NULL,
	  key longtext NOT NULL,
	  value longtext NOT NULL,
	  created_at datetime NOT NULL,
	  updated_at datetime,
	  deleted_at datetime
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
