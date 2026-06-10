package handlers

import (
	"net/http"
	"strconv"

	"go-chain/database"
	"go-chain/models"

	"github.com/gin-gonic/gin"
)

// AddFavorite adds a recipe to user's favorites
func AddFavorite(c *gin.Context) {
	db := database.GetDB()

	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		RecipeID uint   `json:"recipe_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if already favorited
	var existing models.Favorite
	result := db.Where("user_id = ? AND recipe_id = ?", req.UserID, req.RecipeID).First(&existing)
	if result.RowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "already favorited", "favorite": existing})
		return
	}

	favorite := models.Favorite{
		UserID:   req.UserID,
		RecipeID: req.RecipeID,
	}
	if err := db.Create(&favorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload with recipe data
	db.Preload("Recipe.Category").First(&favorite, favorite.ID)

	c.JSON(http.StatusCreated, gin.H{"message": "favorited", "favorite": favorite})
}

// RemoveFavorite removes a recipe from user's favorites
func RemoveFavorite(c *gin.Context) {
	db := database.GetDB()

	userID := c.Query("user_id")
	recipeIDStr := c.Param("recipe_id")
	recipeID, err := strconv.Atoi(recipeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipe id"})
		return
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	result := db.Where("user_id = ? AND recipe_id = ?", userID, recipeID).Delete(&models.Favorite{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "favorite not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "favorite removed"})
}

// GetFavorites returns user's favorite recipes
func GetFavorites(c *gin.Context) {
	db := database.GetDB()
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var favorites []models.Favorite
	db.Preload("Recipe.Category").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&favorites)

	// Convert to brief format
	type FavoriteItem struct {
		ID        uint               `json:"id"`
		Recipe    models.RecipeBrief `json:"recipe"`
		CreatedAt string             `json:"created_at"`
	}

	items := make([]FavoriteItem, len(favorites))
	for i, f := range favorites {
		items[i] = FavoriteItem{
			ID:        f.ID,
			Recipe:    f.Recipe.ToBrief(),
			CreatedAt: f.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	c.JSON(http.StatusOK, items)
}
