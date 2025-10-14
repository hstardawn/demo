package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID    int64 `gorm:"primary_key" json:"id"`
	Ctime int64 `json:"ctime"`
	Utime int64 `json:"utime"`
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	t := time.Now().UnixNano() / 1e6
	tx.Statement.SetColumn("ctime", t)
	tx.Statement.SetColumn("utime", t)
	return nil
}

func (m *BaseModel) BeforeUpdate(tx *gorm.DB) (err error) {
	t := time.Now().UnixNano() / 1e6
	tx.Statement.SetColumn("utime", t)
	return nil
}
