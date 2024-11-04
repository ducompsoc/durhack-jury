package main

import (
	"server/config"
	"server/database"
	"server/router"
)

// todo: run go tidy and go fmt to get replace tabs with 4 spaces
func main() {
	// Check for all necessary env variables
	config.CheckEnv()

	// Connect to the database
	db := database.InitDb()

	// Create the router and attach variables
	r := router.NewRouter(db)

	// Start the Gin server!
	r.Run(":" + config.Port)
}
