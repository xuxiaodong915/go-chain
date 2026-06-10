package models

import "time"

type Favorite struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"size:100;not null;index"`
	RecipeID  uint      `json:"recipe_id" gorm:"not null;index"`
	Recipe    Recipe    `json:"recipe" gorm:"foreignKey:RecipeID"`
	CreatedAt time.Time `json:"created_at"`
}
