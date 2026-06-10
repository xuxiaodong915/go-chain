package handlers

import (
	"net/http"
	"strconv"

	"go-chain/database"
	"go-chain/models"

	"github.com/gin-gonic/gin"
)

// GetShoppingList returns user's shopping list
func GetShoppingList(c *gin.Context) {
	db := database.GetDB()
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var items []models.ShoppingItem
	db.Where("user_id = ?", userID).Order("checked ASC, created_at DESC").Find(&items)

	// Attach recipe names
	for i, item := range items {
		if item.RecipeID != nil {
			var recipe models.Recipe
			if err := db.Select("name").First(&recipe, *item.RecipeID).Error; err == nil {
				items[i].RecipeName = recipe.Name
			}
		}
	}

	c.JSON(http.StatusOK, items)
}

// AddShoppingItem adds an item to shopping list
func AddShoppingItem(c *gin.Context) {
	db := database.GetDB()

	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		ItemName string `json:"item_name" binding:"required"`
		Quantity string `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := models.ShoppingItem{
		UserID:   req.UserID,
		ItemName: req.ItemName,
		Quantity: req.Quantity,
		Checked:  false,
	}

	if err := db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdateShoppingItem updates a shopping item (check/uncheck)
func UpdateShoppingItem(c *gin.Context) {
	db := database.GetDB()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	var req struct {
		Checked *bool `json:"checked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var item models.ShoppingItem
	if err := db.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	if req.Checked != nil {
		item.Checked = *req.Checked
	}
	db.Save(&item)

	c.JSON(http.StatusOK, item)
}

// DeleteShoppingItem deletes a shopping item
func DeleteShoppingItem(c *gin.Context) {
	db := database.GetDB()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	result := db.Delete(&models.ShoppingItem{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item deleted"})
}

// ClearCheckedItems deletes all checked items for a user
func ClearCheckedItems(c *gin.Context) {
	db := database.GetDB()
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	result := db.Where("user_id = ? AND checked = ?", userID, true).Delete(&models.ShoppingItem{})
	c.JSON(http.StatusOK, gin.H{
		"message": "checked items cleared",
		"count":   result.RowsAffected,
	})
}
