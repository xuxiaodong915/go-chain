package models

import "time"

type ShoppingItem struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"size:100;not null;index"`
	RecipeID  *uint     `json:"recipe_id" gorm:"index"`
	ItemName  string    `json:"item_name" gorm:"size:100;not null"`
	Quantity  string    `json:"quantity" gorm:"size:50"`
	Checked   bool      `json:"checked" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	// Populated from recipe when listing
	RecipeName string `json:"recipe_name,omitempty" gorm:"-"`
}
