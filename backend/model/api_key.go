package model

import "time"

type APIKey struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(64);uniqueIndex;not null" json:"name"`
	KeyHash   string    `gorm:"column:key_hash;type:varchar(128);uniqueIndex;not null" json:"-"`
	Scope     string    `gorm:"column:scope;type:varchar(10);not null;default:'read'" json:"scope"`
	Status    int       `gorm:"column:status;type:tinyint;default:1;not null" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (APIKey) TableName() string { return "api_keys" }
