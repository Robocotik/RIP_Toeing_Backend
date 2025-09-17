package main

import (
	"backend/internal/app/ds"
	"backend/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&ds.Rumb{},
		&ds.FlyRequest{},
		&ds.FlyRequest_Rumb{},
		&ds.User{},
	)
	if err != nil {
		panic("cant migrate db")
	}
	print("successful migration")
}
