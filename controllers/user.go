package controllers

import (
	"github.com/cploutarchou/go-kafka-rest/models"
	"github.com/gofiber/fiber/v2"
)

type UserController struct {
}

func (u *UserController) GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": user}})
}
