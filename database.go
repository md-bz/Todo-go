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
	ID          uint
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
	Token    string
}

func database() *gorm.DB {
	var db, err = gorm.Open(sqlite.Open("db.sqlite"), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Todo{}, &User{})
	return db
}
