// Resolves the shortened URL to the main URL
package routes

import (
	"github.com/Saswata28/url_shortener/database"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(c *fiber.Ctx) error {
	id := c.Params("id")
	r := database.CreateClient(0) //CreateClient func in database dir
	defer r.Close()

	value, err := r.Get(database.Ctx, id).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shorten URL not found in the database"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not connect to the database"})
	}

	rInt := database.CreateClient(1)
	defer rInt.Close()

	_ = rInt.Incr(database.Ctx, "counter") //The first time this line is executed, Redis will create the "counter" key with a value of 1, and then each subsequent call to INCR will increment its value by 1.
	return c.Redirect(value, 301)
}
