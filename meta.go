package metastore

import (
	"time"

	"github.com/gouniverse/uid"
	"gorm.io/gorm"
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

// BeforeCreate adds UID to model
func (m *Meta) BeforeCreate(tx *gorm.DB) (err error) {
	uuid := uid.HumanUid()
	m.ID = uuid
	return nil
}
