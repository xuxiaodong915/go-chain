package handlers

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-chain/database"
	"go-chain/models"

	"github.com/gin-gonic/gin"
)

const uploadDir = "uploads"
const maxUploadSize = 5 << 20 // 5MB

// UploadImage handles image file upload
func UploadImage(c *gin.Context) {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large, max 5MB"})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type, allowed: jpg, jpeg, png, gif, webp"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(header.Filename, ext), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write file"})
		return
	}

	imageURL := "/uploads/" + filename

	// If recipe_id provided, auto-assign
	recipeIDStr := c.PostForm("recipe_id")
	if recipeIDStr != "" {
		recipeID, err := strconv.Atoi(recipeIDStr)
		if err == nil && recipeID > 0 {
			db := database.GetDB()
			db.Model(&models.Recipe{}).Where("id = ?", recipeID).Update("image_url", imageURL)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "upload successful",
		"image_url": imageURL,
		"filename":  filename,
	})
}

// UpdateRecipeImage updates a recipe's image URL
func UpdateRecipeImage(c *gin.Context) {
	db := database.GetDB()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipe id"})
		return
	}

	var req struct {
		ImageURL string `json:"image_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := db.Model(&models.Recipe{}).Where("id = ?", id).Update("image_url", req.ImageURL)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "image updated", "image_url": req.ImageURL})
}

// ListImages returns all uploaded images
func ListImages(c *gin.Context) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read upload directory"})
		return
	}

	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read upload directory"})
		return
	}

	type ImageInfo struct {
		Filename string `json:"filename"`
		URL      string `json:"url"`
		Size     int64  `json:"size"`
	}

	var images []ImageInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		images = append(images, ImageInfo{
			Filename: entry.Name(),
			URL:      "/uploads/" + entry.Name(),
			Size:     info.Size(),
		})
	}

	c.JSON(http.StatusOK, images)
}

// categoryImages maps category names to available image files
var categoryImages = map[string][]string{
	"川菜": {
		"https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1555939594-58d7cb561ad1?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1563379926898-05f4575a45d8?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1540189549336-e6e99c3679fe?w=400&h=300&fit=crop",
	},
	"粤菜": {
		"https://images.unsplash.com/photo-1555939594-58d7cb561ad1?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1512058564366-18510be2db19?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1565299624946-b28f40a0ae38?w=400&h=300&fit=crop",
	},
	"湘菜": {
		"https://images.unsplash.com/photo-1504674900247-0877df9cc836?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1565299624946-b28f40a0ae38?w=400&h=300&fit=crop",
	},
	"甜点": {
		"https://images.unsplash.com/photo-1551024506-0bccd828d307?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1540189549336-e6e99c3679fe?w=400&h=300&fit=crop",
	},
	"早餐": {
		"https://images.unsplash.com/photo-1490645935967-10de6ba17061?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1567620905732-2d1ec7ab7445?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1504754524776-8f4f37790ca0?w=400&h=300&fit=crop",
	},
	"汤羹": {
		"https://images.unsplash.com/photo-1547592166-23ac45744acd?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1484723091739-30a097e8f929?w=400&h=300&fit=crop",
	},
	"素菜": {
		"https://images.unsplash.com/photo-1512621776951-a57141f2eefd?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1540189549336-e6e99c3679fe?w=400&h=300&fit=crop",
	},
	"面食": {
		"https://images.unsplash.com/photo-1476224203421-9ac39bcb3327?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1569718212165-3a8278d5f624?w=400&h=300&fit=crop",
		"https://images.unsplash.com/photo-1484723091739-30a097e8f929?w=400&h=300&fit=crop",
	},
}

var fallbackImgs = []string{
	"https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=400&h=300&fit=crop",
	"https://images.unsplash.com/photo-1504674900247-0877df9cc836?w=400&h=300&fit=crop",
	"https://images.unsplash.com/photo-1551024506-0bccd828d307?w=400&h=300&fit=crop",
}

// AssignImages assigns placeholder images to all recipes based on category
func AssignImages(c *gin.Context) {
	db := database.GetDB()
	var recipes []models.Recipe
	db.Preload("Category").Find(&recipes)

	count := 0
	for _, r := range recipes {
		images, ok := categoryImages[r.Category.Name]
		if !ok || len(images) == 0 {
			images = fallbackImgs
		}
		img := images[rand.Intn(len(images))]
		db.Model(&r).Update("image_url", img)
		count++
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("assigned images to %d recipes", count),
		"count":   count,
	})
}
