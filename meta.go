package metastore

import (
	"time"

	"github.com/gouniverse/uid"
	"gorm.io/gorm"
)

const ()

// Meta type
type Meta struct {
	ID         string     `gorm:"type:varchar(40);column:id;primary_key;"`
	ObjectType string     `gorm:"type:varchar(50);column:object_type;"`
	ObjectID   string     `gorm:"type:varchar(40);column:object_id;"`
	Key        string     `gorm:"type:varchar(510);column:meta_key;"`
	Value      string     `gorm:"type:longtext;column:meta_value;"`
	CreatedAt  time.Time  `json:"created_at" gorm:"type:datetime;column:created_at;DEFAULT NULL;"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"type:datetime;column:updated_at;DEFAULT NULL;"`
	DeletedAt  *time.Time `json:"deleted_at" gorm:"type:datetime;olumn:deleted_at;DEFAULT NULL;"`
}

// BeforeCreate adds UID to model
func (m *Meta) BeforeCreate(tx *gorm.DB) (err error) {
	uuid := uid.HumanUid()
	m.ID = uuid
	return nil
}
