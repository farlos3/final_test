package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var stats = []fiber.Map{}
var nextID = 1

var playHistory = struct {
	Count      int
	TotalScore int
	BestScore  int
}{
	Count:      0,
	TotalScore: 0,
	BestScore:  0,
}

func main() {
	app := fiber.New()

	// Add request logging middleware
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} (${latency})\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	// Allow CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	app.Get("/api/getStats", func(c *fiber.Ctx) error {
		log.Println("Fetching stats...")
		return c.JSON(stats)
	})

	app.Post("/api/saveScore", func(c *fiber.Ctx) error {
		var newScore struct {
			Score     int    `json:"score"`
			Moves     int    `json:"moves"`
			Time      int    `json:"time"`
			CreatedAt string `json:"created_at"`
		}

		if err := c.BodyParser(&newScore); err != nil {
			log.Printf("Error parsing score JSON: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid JSON",
			})
		}

		// Initial tracking stats if empty
		if len(stats) == 0 {
			newEntry := fiber.Map{
				"id":            nextID,
				"best_score":    newScore.Score,
				"average_score": float64(newScore.Score),
				"play_count":    1,
				"latest_score":  newScore.Score,
				"moves":         newScore.Moves,
				"time":          newScore.Time,
				"created_at":    newScore.CreatedAt,
			}
			stats = append(stats, newEntry)
			log.Printf("First score saved - ID: %d, Score: %d", nextID, newScore.Score)
			nextID++
			return c.Status(fiber.StatusCreated).JSON(newEntry)
		}

		// Update existing stats (assume stats[0] is global score stat)
		existing := stats[0]

		playCount := existing["play_count"].(int)
		totalScore := existing["average_score"].(float64) * float64(playCount)
		playCount += 1
		totalScore += float64(newScore.Score)
		avgScore := totalScore / float64(playCount)

		bestScore := existing["best_score"].(int)
		if newScore.Score > bestScore {
			bestScore = newScore.Score
		}

		existing["play_count"] = playCount
		existing["average_score"] = avgScore
		existing["best_score"] = bestScore
		existing["latest_score"] = newScore.Score
		existing["moves"] = newScore.Moves
		existing["time"] = newScore.Time
		existing["created_at"] = newScore.CreatedAt

		log.Printf("Score updated - Total Plays: %d, Best Score: %d, Avg Score: %.2f", playCount, bestScore, avgScore)

		return c.Status(fiber.StatusCreated).JSON(existing)
	})

	app.Delete("/api/clearScores", func(c *fiber.Ctx) error {
		stats = []fiber.Map{}
		nextID = 1
		playHistory = struct {
			Count      int
			TotalScore int
			BestScore  int
		}{}
		log.Println("All scores cleared")
		return c.JSON(fiber.Map{
			"message": "All scores cleared",
		})
	})

	log.Fatal(app.Listen(":5000"))
}
