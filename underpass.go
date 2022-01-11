package underpass

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Underpass struct {
	db     *pgxpool.Pool
	router *fiber.App
	Log    io.Writer
}

func New() *Underpass {
	return &Underpass{Log: os.Stderr}
}

func (u *Underpass) Start() error {
	u.initDB()
	u.initRouter()
	return u.startRouter()
}

func (u *Underpass) initDB() {

	if u.db != nil {
		return
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	ctx := context.Background()

	db, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err = db.Ping(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	u.db = db
}

func (u *Underpass) initRouter() {
	if u.router != nil {
		return
	}
	router := fiber.New()
	u.router = router
	u.setupRoutes()
}

func (u *Underpass) startRouter() error {
	addr := net.JoinHostPort("localhost", os.Getenv("UNDERPASS_PORT"))
	return u.router.Listen(addr)
}

func (u *Underpass) setupRoutes() {

	if u.router == nil {
		fmt.Println("router not initialized")
		os.Exit(1)
	}
	if u.db == nil {
		fmt.Println("db not initialized")
		os.Exit(1)
	}

	handlers := newHandlers(u.db)

	u.router.Use(logger.New(logger.Config{
		Output:     u.Log,
		Format:     "${date} ${time} UTC | ${path} | ${status} | ${latency} | ${bytesSent} B\n",
		TimeFormat: "02-Jan-2006 15:04:05",
		TimeZone:   "UTC",
	}))

	u.router.Get("/api/node/:id", handlers.QueryNodesByID)
	u.router.Get("/api/way/:id", handlers.QueryWaysByID)
	u.router.Get("/api/relation/:id", handlers.QueryRelationsByID)
	u.router.Get("/api/bbox/:bbox", handlers.QueryBboxByBbox)
}

func Start() error {
	return New().Start()
}
