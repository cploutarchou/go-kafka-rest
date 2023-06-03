package controllers

import (
	"fmt"
	"github.com/cploutarchou/go-kafka-rest/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserController struct {
	Model models.User
	DB    *gorm.DB
}

func (u *UserController) GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": user}})
}

func (u *UserController) RefreshToken(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.UserResponse)

	fmt.Println(user.Email)
	// get user from db by id

	user_, err := u.Model.GetByEmail(user.Email)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to retrieve user"})
	}
	token, err := CreateJWT(user_)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Something went wrong"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"token": token}})
}
