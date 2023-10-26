package model

import "time"

type Subkey struct {
	ID        int64     `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"not null"`
	Name      string    `gorm:"not null"`

	Avatar string `gorm:"not null"`
	Algorithm string `gorm:"algo"`
}
