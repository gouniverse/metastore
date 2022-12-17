package metastore

import (
	"time"
)

const ()

// Meta type
type Meta struct {
	ID         string     `db:"id"`
	ObjectType string     `db:"object_type"`
	ObjectID   string     `db:"object_id"`
	Key        string     `db:"meta_key"`
	Value      string     `db:"meta_value"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}
