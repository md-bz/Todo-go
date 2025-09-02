package main

import (
	. "api/database"
	"api/helpers"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type header struct {
	Authorization string `reqHeader:"Authorization"`
}

type updateTodo struct {
	Description string `json:"description"`
}

var user string

func main() {

	app := fiber.New()

	app.Post("/signup", func(c *fiber.Ctx) error {
		user := new(User)
		err := c.BodyParser(user)
		if err != nil || user.Username == "" || user.Password == "" {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
				"error": "please provide Username and Password in json",
			})
		}

		dbErr := DB.Model(&User{}).Create(&User{Username: user.Username, Password: user.Password})

		if dbErr.Error != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "failed to sign up",
			})
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "login to get token",
		})
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		user := new(User)
		err := c.BodyParser(user)
		if err != nil || user.Username == "" || user.Password == "" {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
				"error": "please provide Username and Password in json",
			})
		}

		res := DB.Where(&User{Username: user.Username}).First(&user)

		if res.Error != nil {
			return c.Status(403).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		if user.Password != user.Password {
			return c.Status(403).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
		tokenString, err := helpers.CreateJWT(user.ID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(APIUser{Username: user.Username, Token: tokenString})
	})

	app.Use(func(c *fiber.Ctx) error {
		h := new(header)
		err := c.ReqHeaderParser(h)

		auth := strings.Split(h.Authorization, " ")

		if err != nil || h.Authorization == "" || len(auth) != 2 || auth[0] != "Bearer" || auth[1] == "" {
			return c.Status(403).JSON(fiber.Map{
				"error": "No Authorization header",
			})
		}

		var user User
		userId, err := helpers.VerifyJWT(auth[1])
		if err != nil {
			return c.Status(403).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		res := DB.
			Where("ID = ?", userId).
			First(&user)

		if res.Error != nil {
			fmt.Print(err)
			return c.Status(403).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		c.Locals("user", &user)

		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		var todos []APITodo
		user := c.Locals("user").(*User)

		DB.Model(&Todo{}).Where(&Todo{UserId: user.ID}).Find(&todos)

		return c.JSON(todos)
	})

	app.Get("/:id", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*User)
		id := c.Params("id")
		var todo APITodo

		DB.Model(&Todo{}).Where("user_id = ? AND id = ?", user.ID, id).Find(&todo)

		return c.JSON(todo)
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

		t.UserId = user.ID
		DB.Create(&t)

		return c.JSON(APITodo{ID: t.ID, Description: t.Description, Done: t.Done})
	})

	app.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		if id == "" {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{"error": "please provide id in url"})
		}
		user := c.Locals("user").(*User)

		res := DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&Todo{})

		if res.Error != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to delete todo"})
		}
		if res.RowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "todo not found"})
		}

		return c.Status(204).JSON(fiber.Map{"message": "todo deleted"})
	})

	app.Patch("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		t := new(updateTodo)
		err := c.BodyParser(t)

		if id == "" {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{"error": "please provide id in url"})
		}

		if err != nil || t.Description == "" {
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
				"error": "please provide Description in json",
			})
		}

		user := c.Locals("user").(*User)

		var todo Todo
		res := DB.Where("ID = ? AND user_id = ?", id, user.ID).First(&todo)
		if res.Error != nil {
			return c.Status(404).JSON(fiber.Map{"error": "todo not found"})
		}

		todo.Description = t.Description
		if err := DB.Save(&todo).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to update todo"})
		}

		return c.JSON(APITodo{ID: todo.ID, Description: todo.Description, Done: todo.Done})
	})

	app.Post("/toggle/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		user := c.Locals("user").(*User)

		var todo Todo
		res := DB.Where("ID = ? AND user_id = ?", id, user.ID).First(&todo)
		if res.Error != nil {
			return c.Status(404).JSON(fiber.Map{"error": "todo not found"})
		}

		todo.Done = !todo.Done
		if err := DB.Save(&todo).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to toggle todo"})
		}

		return c.JSON(APITodo{ID: todo.ID, Description: todo.Description, Done: todo.Done})
	})

	log.Fatal(app.Listen(":3000"))
}
