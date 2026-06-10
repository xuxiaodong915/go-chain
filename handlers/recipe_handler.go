package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"go-chain/database"
	"go-chain/models"

	"github.com/gin-gonic/gin"
)

// GetDailyRecommend returns the daily recommended recipe + backup list
func GetDailyRecommend(c *gin.Context) {
	db := database.GetDB()
	today := time.Now().Format("2006-01-02")

	// Use today's date as a random seed so the same recipe lasts all day
	r := rand.New(rand.NewSource(time.Now().Unix() / 86400))

	// Count total recipes
	var total int64
	db.Model(&models.Recipe{}).Count(&total)
	if total == 0 {
		c.JSON(http.StatusOK, gin.H{"date": today, "recipe": nil, "backup": []models.RecipeBrief{}})
		return
	}

	// Pick a random recipe for today
	offset := r.Intn(int(total))
	var recipe models.Recipe
	db.Preload("Category").Offset(offset).First(&recipe)

	// Get backup list (other recipes, up to 5)
	var backups []models.Recipe
	db.Preload("Category").Where("id != ?", recipe.ID).Order("RANDOM()").Limit(5).Find(&backups)

	backupBriefs := make([]models.RecipeBrief, len(backups))
	for i, b := range backups {
		backupBriefs[i] = b.ToBrief()
	}

	c.JSON(http.StatusOK, models.DailyRecommendResponse{
		Recipe: recipe.ToDetail(),
		Backup: backupBriefs,
		Date:   today,
	})
}

// GetRecipes returns paginated recipes, optionally filtered by category
func GetRecipes(c *gin.Context) {
	db := database.GetDB()

	categoryIDStr := c.Query("category_id")
	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(sizeStr)
	if size < 1 || size > 50 {
		size = 20
	}

	query := db.Model(&models.Recipe{}).Preload("Category")
	if categoryIDStr != "" {
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err == nil && categoryID > 0 {
			query = query.Where("category_id = ?", categoryID)
		}
	}

	var total int64
	query.Count(&total)

	var recipes []models.Recipe
	query.Offset((page - 1) * size).Limit(size).Find(&recipes)

	briefs := make([]models.RecipeBrief, len(recipes))
	for i, r := range recipes {
		briefs[i] = r.ToBrief()
	}

	c.JSON(http.StatusOK, models.RecipeListResponse{
		Recipes: briefs,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// GetRecipeByID returns full recipe detail
func GetRecipeByID(c *gin.Context) {
	db := database.GetDB()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipe id"})
		return
	}

	var recipe models.Recipe
	result := db.Preload("Category").First(&recipe, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}

	c.JSON(http.StatusOK, recipe.ToDetail())
}

// GetCategories returns all categories
func GetCategories(c *gin.Context) {
	db := database.GetDB()
	var categories []models.Category
	db.Order("name ASC").Find(&categories)
	c.JSON(http.StatusOK, categories)
}

// AddShoppingFromRecipe adds recipe ingredients to shopping list
func AddShoppingFromRecipe(c *gin.Context) {
	db := database.GetDB()

	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		RecipeID uint   `json:"recipe_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the recipe
	var recipe models.Recipe
	if err := db.First(&recipe, req.RecipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}

	// Parse ingredients
	var ingredients []string
	if err := json.Unmarshal([]byte(recipe.Ingredients), &ingredients); err != nil {
		// Fallback to simple parsing
		ingredients = parseSimpleList(recipe.Ingredients)
	}

	// Add each ingredient as a shopping item
	items := make([]models.ShoppingItem, len(ingredients))
	for i, ing := range ingredients {
		items[i] = models.ShoppingItem{
			UserID:   req.UserID,
			RecipeID: &req.RecipeID,
			ItemName: ing,
			Checked:  false,
		}
	}

	if len(items) > 0 {
		db.Create(&items)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "ingredients added to shopping list",
		"count":   len(items),
	})
}

func parseSimpleList(s string) []string {
	var result []string
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return []string{s}
	}
	return result
}
