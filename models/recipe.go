package models

import "time"

type Recipe struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Name           string    `json:"name" gorm:"size:100;not null"`
	ImageURL       string    `json:"image_url" gorm:"size:500"`
	CategoryID     uint      `json:"category_id" gorm:"not null;index"`
	Category       Category  `json:"category" gorm:"foreignKey:CategoryID"`
	Ingredients    string    `json:"ingredients" gorm:"type:text"`       // JSON string array
	Steps          string    `json:"steps" gorm:"type:text"`             // JSON string array
	CookTime       string    `json:"cook_time" gorm:"size:50"`
	Difficulty     string    `json:"difficulty" gorm:"size:20"`          // 简单/中等/困难
	DailyRecommend bool      `json:"daily_recommend" gorm:"default:false;index"`
	CreatedAt      time.Time `json:"created_at"`
}

// API response structs
type RecipeBrief struct {
	ID         uint     `json:"id"`
	Name       string   `json:"name"`
	ImageURL   string   `json:"image_url"`
	Category   string   `json:"category"`
	CookTime   string   `json:"cook_time"`
	Difficulty string   `json:"difficulty"`
}

type RecipeDetail struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	ImageURL    string   `json:"image_url"`
	Category    string   `json:"category"`
	CookTime    string   `json:"cook_time"`
	Difficulty  string   `json:"difficulty"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
}

type DailyRecommendResponse struct {
	Recipe   RecipeDetail   `json:"recipe"`
	Backup   []RecipeBrief  `json:"backup"`
	Date     string         `json:"date"`
}

type RecipeListResponse struct {
	Recipes []RecipeBrief `json:"recipes"`
	Total   int64         `json:"total"`
	Page    int           `json:"page"`
	Size    int           `json:"size"`
}

// Convert Recipe to RecipeDetail
func (r *Recipe) ToDetail() RecipeDetail {
	ingredients := parseStringArray(r.Ingredients)
	steps := parseStringArray(r.Steps)
	categoryName := ""
	if r.Category.ID != 0 {
		categoryName = r.Category.Name
	}
	return RecipeDetail{
		ID:          r.ID,
		Name:        r.Name,
		ImageURL:    r.ImageURL,
		Category:    categoryName,
		CookTime:    r.CookTime,
		Difficulty:  r.Difficulty,
		Ingredients: ingredients,
		Steps:       steps,
	}
}

// Convert Recipe to RecipeBrief
func (r *Recipe) ToBrief() RecipeBrief {
	categoryName := ""
	if r.Category.ID != 0 {
		categoryName = r.Category.Name
	}
	return RecipeBrief{
		ID:         r.ID,
		Name:       r.Name,
		ImageURL:   r.ImageURL,
		Category:   categoryName,
		CookTime:   r.CookTime,
		Difficulty: r.Difficulty,
	}
}

func parseStringArray(s string) []string {
	if s == "" {
		return []string{}
	}
	// Simple JSON array parsing: ["a","b","c"]
	result := []string{}
	current := ""
	inQuote := false
	for _, c := range s {
		if c == '"' {
			inQuote = !inQuote
		} else if inQuote {
			current += string(c)
		} else if c == ',' || c == ']' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		}
	}
	return result
}
