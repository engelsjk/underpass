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
	Log    io.Writer
}

func New() *Underpass {
	return &Underpass{Log: os.Stderr}
}

func (u *Underpass) Start() error {
	if err := u.initDB(); err != nil {
		return err
	}
	if err := u.initRouter(); err != nil {
		return err
	}
	return u.startRouter()
}

func (u *Underpass) initDB() error {

	if u.db != nil {
		return nil
	}

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

func (u *Underpass) initRouter() error {
	if u.router != nil {
		return nil
	}
	router := fiber.New()
	u.router = router
	return u.setupRoutes()
}

func (u *Underpass) startRouter() error {
	addr := net.JoinHostPort("localhost", os.Getenv("UNDERPASS_PORT"))
	return u.router.Listen(addr)
}

func (u *Underpass) setupRoutes() error {

	if u.router == nil {
		return fmt.Errorf("router not initialized")
	}
	if u.db == nil {
		return fmt.Errorf("db not initialized")
	}

	handlers := newHandlers(u.db)

	u.router.Use(logger.New(logger.Config{
		Output:     u.Log,
		Format:     "${date} ${time} UTC | ${path} | ${status} | ${latency} | ${bytesSent} B\n",
		TimeFormat: "02-Jan-2006 15:04:05",
		TimeZone:   "UTC",
	}))

	// highways
	u.router.Get("/api/highways/id/:id", handlers.QueryHighwaysByID)
	u.router.Get("/api/highways/bbox/:bbox", handlers.QueryHighwaysByBoundingBox)

	// buildings
	u.router.Get("/api/buildings/id/:id", handlers.QueryBuildingsByID)
	u.router.Get("/api/buildings/bbox/:bbox", handlers.QueryBuildingsByBoundingBox)

	return nil
}

func Start() error {
	return New().Start()
}
