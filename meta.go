package metastore

import (
	"time"
)

const ()

// Meta type
type Meta struct {
	ID         string     `db:"ID"`
	ObjectType string     `db:"ObjectType"`
	ObjectID   string     `db:"ObjectID"`
	Key        string     `db:"Key"`
	Value      string     `db:"Value"`
	CreatedAt  time.Time  `db:"CreatedAt"`
	UpdatedAt  time.Time  `db:"UpdatedAt"`
	DeletedAt  *time.Time `db:"DeletedAt"`
}
