package underpass

import (
	"database/sql"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	_ "github.com/lib/pq"
)

type Underpass struct {
	db     *sql.DB
	router *fiber.App
}

func (u *Underpass) InitDB() error {

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	u.db = db

	return nil
}

func (u *Underpass) InitRouter(w io.Writer) {
	router := fiber.New()
	SetupRoutes(router, u.db, w)
	u.router = router
}

func (u *Underpass) StartRouter() error {
	addr := net.JoinHostPort("localhost", os.Getenv("UNDERPASS_PORT"))
	return u.router.Listen(addr)
}

func SetupRoutes(router *fiber.App, db *sql.DB, w io.Writer) {

	handlers := NewHandlers(db)

	router.Use(logger.New(logger.Config{
		Output:     w,
		Format:     "${date} ${time} UTC | ${path} | ${status} | ${latency} | ${bytesSent} B\n",
		TimeFormat: "02-Jan-2006 15:04:05",
		TimeZone:   "UTC",
	}))

	// highways
	router.Get("/api/highways/id/:id", handlers.QueryHighwaysByID)
	router.Get("/api/highways/bbox/:bbox", handlers.QueryHighwaysByBoundingBox)

	// buildings
	router.Get("/api/buildings/id/:id", handlers.QueryBuildingsByID)
	router.Get("/api/buildings/bbox/:bbox", handlers.QueryBuildingsByBoundingBox)

}
