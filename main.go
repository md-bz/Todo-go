package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type header struct {
	Authorization string `reqHeader:"Authorization"`
}

type updateTodo struct {
	OldDescription string `json:"oldDescription"`
	NewDescription string `json:"newDescription"`
}

var user string

func main() {
	app := fiber.New()

	db := database()

	app.Post("/signup", func(c *fiber.Ctx) error {
		user := new(User)
		err := c.BodyParser(user)
		if err != nil || user.Username == "" || user.Password == "" {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
				"error": "please provide Username and Password in json",
			})
		}

		db.Model(&User{}).Create(&User{Username: user.Username, Password: user.Password})
		return c.JSON(APIUser{user.Username, user.Password})
	})

	app.Use(func(c *fiber.Ctx) error {
		// placeholder for jwt
		h := new(header)
		err := c.ReqHeaderParser(h)

		auth := strings.Split(h.Authorization, " ")

		if err != nil || h.Authorization == "" || len(auth) != 2 || auth[0] != "Bearer" || auth[1] == "" {
			return c.Status(403).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var user User
		res := db.Model(&User{}).
			Where(&User{Username: strings.Trim(auth[1], " ")}).
			First(&user)

		if res.Error != nil {
			fmt.Print(err)
			return c.Status(403).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
		c.Locals("user", &user)

		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		var todos []APITodo
		user := c.Locals("user").(*User)

		db.Model(&Todo{}).Where(&Todo{UserId: user.ID}).Find(&todos)

		return c.JSON(todos)
	})

	app.Post("/", func(c *fiber.Ctx) error {
		t := new(Todo)
		err := c.BodyParser(t)

		if err != nil || t.Description == "" {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
				"error": "please provide description in json",
			})
		}

		user := c.Locals("user").(*User)

		db.Create(&Todo{UserId: user.ID, Description: t.Description})

		return c.JSON(APITodo{t.Description, t.Done})
	})

	app.Delete("/", func(c *fiber.Ctx) error {
		t := new(Todo)
		if err := c.BodyParser(t); err != nil || t.Description == "" {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
				"error": "please provide description in json",
			})
		}

		user := c.Locals("user").(*User)

		res := db.Where("description = ? AND user_id = ?", t.Description, user.ID).Delete(&Todo{})
		if res.Error != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to delete todo"})
		}
		if res.RowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "todo not found"})
		}

		return c.JSON(APITodo{t.Description, t.Done})
	})

	app.Patch("/", func(c *fiber.Ctx) error {
		t := new(updateTodo)
		err := c.BodyParser(t)

		if err != nil || t.NewDescription == "" || t.OldDescription == "" {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
				"error": "please provide OldDescription,NewDescription in json",
			})
		}

		user := c.Locals("user").(*User)

		var todo Todo
		res := db.Where("description = ? AND user_id = ?", t.OldDescription, user.ID).First(&todo)
		if res.Error != nil {
			return c.Status(404).JSON(fiber.Map{"error": "todo not found"})
		}

		todo.Description = t.NewDescription
		if err := db.Save(&todo).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to update todo"})
		}

		return c.JSON(APITodo{todo.Description, todo.Done})
	})

	app.Post("/toggle", func(c *fiber.Ctx) error {
		t := new(Todo)
		if err := c.BodyParser(t); err != nil || t.Description == "" {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
				"error": "please provide description in json",
			})
		}

		user := c.Locals("user").(*User)

		var todo Todo
		res := db.Where("description = ? AND user_id = ?", t.Description, user.ID).First(&todo)
		if res.Error != nil {
			return c.Status(404).JSON(fiber.Map{"error": "No todo found with that description"})
		}

		todo.Done = !todo.Done
		if err := db.Save(&todo).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to toggle todo"})
		}

		return c.JSON(APITodo{todo.Description, todo.Done})
	})

	log.Fatal(app.Listen(":3000"))
}
