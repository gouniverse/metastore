# Meta Store

Store to database additional information to anything using metas (key - value) pairs

## Installation
```
go get -u github.com/gouniverse/metastore
```

## Setup

```
metaStore = metastore.NewStore(metastore.WithGormDb(databaseInstance), metastore.WithTableName("my_meta"))
```



## Usage

```
// Set a meta key with value
models.GetMetaStore().Set("user", "1", "verified", "yes")
models.GetMetaStore().Set("user", "1", "verified_at", "2021-03-12")
  
// Getting the value with default if not found
log.Println(models.GetMetaStore().Get("user", "1", "verified", ""))
log.Println(models.GetMetaStore().Get("user", "1", "verified_at", ""))
```
