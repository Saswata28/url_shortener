package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/Saswata28/url_shortener/database"
	"github.com/Saswata28/url_shortener/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {

	body := new(request)

	//Parsing the request in json
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Can't Parse json"})
	}

	//implement rate limiting
	r2 := database.CreateClient(1)
	defer r2.Close()

	apiQuota, err := r2.Get(database.Ctx, c.IP()).Result() //gets the API_QUOTA for the IP

	if err == redis.Nil {
		err = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), time.Hour).Err() //Set is to create a key:value pair in Redis, 2nd parameter is the key, 3rd parameter is the value, 4th parameter is key expiration time(key won't expire if this field is set to 0)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not set initial quota"})
		}
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not connect to the database"})
	} else {
		// val, err := r2.Get(database.Ctx, c.IP()).Result()
		valInt, err := strconv.Atoi(apiQuota) //making the API_QUOTA to int
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not parse the quota"})
		}
		if valInt <= 0 { //If API_QUOTA is reached
			limit, err := r2.TTL(database.Ctx, c.IP()).Result() //Get the expiration time of the IP key
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve rate limit reset time"})
			}

			limitMinutes := limit.Minutes() //The remaining lifetime time(time before expiration) of a key is by default in `time.Duration` type in golang,  which can hold different units of time (nanoseconds, seconds, minutes, etc.). Minutes(): This is a method of the time.Duration type in Go. It converts the time.Duration into a floating-point number representing the duration in minutes. For example, if the limit is 1800000000000 nanoseconds (30 minutes), limit.Minutes() would return 30.
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Your API quota per hour is reached", "rate_limit_rest": limitMinutes})
		}
	}

	//URL validation Checking
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	//check for domain error(If someone tries to enter localhost as the url to shorten)
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Can't Input localhost"})
	}

	//enforce https
	body.URL = helpers.EnforceHTTPS(body.URL)

	//Taking random string or user given custom string for generating the shortened URL
	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	userCustomStrings, _ := r.Get(database.Ctx, id).Result()
	// if err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Couldn't get id"})
	// }

	if userCustomStrings != "" { //If custom string or url is not empty, so r.Get for id key above found a value in database, So user passed custom url or string is in the database
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "provided url is already in use"})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry*time.Hour).Err() //Set is to create a key:value pair in Redis,2nd parameter is the key, 3rd parameter is the value, 4th parameter is key expiration time(key won't expire if this field is set to 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not set initial quota"})
	}

	//response
	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 60,
	}

	//Decrementing API_QUOTA value in IP key in Redis
	r2.Decr(database.Ctx, c.IP())

	//Taking XRateRemaining value for response
	apiQuota, err = r2.Get(database.Ctx, c.IP()).Result() //taking the API_QUOTA again
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not get the api quota for the user in response"})
	}

	resp.XRateRemaining, err = strconv.Atoi(apiQuota)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not convert the api quota to string"})
	}

	//Taking XRateLimitReset value for response
	timeLimit, err := r2.TTL(database.Ctx, c.IP()).Result()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve rate limit reset time for response"})
	}

	resp.XRateLimitReset = time.Duration(timeLimit.Minutes())

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
