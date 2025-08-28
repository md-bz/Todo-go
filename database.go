package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Todo struct {
	gorm.Model
	UserId      uint
	Description string `json:"description"`
	Done        bool
}
type APITodo struct {
	Description string
	Done        bool
}
type User struct {
	gorm.Model
	Username string
	Password string
	Todo     []Todo
}

type APIUser struct {
	Username string
	Pas      string
}

func database() *gorm.DB {
	var db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Todo{}, &User{})
	return db
}
