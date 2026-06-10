package models

import "time"

type Category struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:50;not null;unique"`
	Icon      string    `json:"icon" gorm:"size:100"`
	CreatedAt time.Time `json:"created_at"`
}
