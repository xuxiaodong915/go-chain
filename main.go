package main

import (
	"log"

	"go-chain/config"
	"go-chain/database"
	"go-chain/handlers"
	"go-chain/seed"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Initialize database
	if err := database.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	// Seed data
	seed.Seed()

	// Setup router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes
	api := r.Group("/api")
	{
		// Recipe routes
		api.GET("/recipes/daily", handlers.GetDailyRecommend)
		api.GET("/recipes", handlers.GetRecipes)
		api.GET("/recipes/:id", handlers.GetRecipeByID)

		// Category routes
		api.GET("/categories", handlers.GetCategories)

		// Favorite routes
		api.POST("/favorites", handlers.AddFavorite)
		api.DELETE("/favorites/:recipe_id", handlers.RemoveFavorite)
		api.GET("/favorites", handlers.GetFavorites)

		// Shopping list routes
		api.GET("/shopping-list", handlers.GetShoppingList)
		api.POST("/shopping-list", handlers.AddShoppingItem)
		api.POST("/shopping-list/from-recipe", handlers.AddShoppingFromRecipe)
		api.PUT("/shopping-list/:id", handlers.UpdateShoppingItem)
		api.DELETE("/shopping-list/:id", handlers.DeleteShoppingItem)
		api.DELETE("/shopping-list/checked/clear", handlers.ClearCheckedItems)

		// Image upload routes
		api.POST("/upload", handlers.UploadImage)
		api.GET("/images", handlers.ListImages)
		api.PUT("/recipes/:id/image", handlers.UpdateRecipeImage)
		api.POST("/recipes/assign-images", handlers.AssignImages)
	}

	// Serve uploaded files as static resources
	r.Static("/uploads", "./uploads")
	// Serve built-in recipe cover images
	r.Static("/images", "./images")

	log.Printf("Server starting on %s", cfg.ServerPort)
	if err := r.Run(cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
