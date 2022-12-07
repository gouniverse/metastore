# Meta Store


[![Tests Status](https://github.com/gouniverse/metastore/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/gouniverse/metastore/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gouniverse/metastore)](https://goreportcard.com/report/github.com/gouniverse/metastore)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/gouniverse/metastore)](https://pkg.go.dev/github.com/gouniverse/metastore)

Meta stores meta information for any object to a database table.

Store to database additional information to anything using metas (key - value) pairs

## Installation
```
go get -u github.com/gouniverse/metastore
```

## Table Schema ##

The following schema is used for the database.

| meta        |                  |
|-------------|------------------|
| id          | String, UniqueId |
| object_type | String (100)     |
| object_id   | String (40)     |
| meta_key    | String (255)     |
| meta_value  | Long Text        |
| created_at  | DateTime         |
| updated_at  | DateTime         |
| deleted_at  | DateTime         |

## Setup

```
metaStore = metastore.NewStore(metastore.WithGormDb(databaseInstance), metastore.WithTableName("my_meta"))
```



## Usage

- Set a meta values (for user with ID 1)
```
metaStore.Set("user", "1", "verified", "yes")
metaStore.Set("user", "1", "verified_at", "2021-03-12")
```

- Get meta values (for user with ID 1), if not found a default value is returned
```
log.Println(metaStore.Get("user", "1", "verified", ""))
log.Println(metaStore.Get("user", "1", "verified_at", ""))
```

## Changelog

2021.12.29 - Added tests badge

2021.12.29 - Added tests
