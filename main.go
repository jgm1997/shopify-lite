package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	sqldb "shopify-lite/internal/db"
	authm "shopify-lite/internal/middleware"
	"shopify-lite/internal/users"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func runMigrations(dsn string) error {
	m, err := migrate.New("file://db/migrations", dsn)
	if err != nil {
		log.Fatalf("failed to create migration: %v", err)
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration failed: %v", err)
		return err
	}
	return nil
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("env DATABASE_URL is not set")
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("env JWT_SECRET is not set")
		return
	}

	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
		return
	}
	defer conn.Close()

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		log.Fatalf("database connection error: %v", err)
		return
	}
	log.Println("database connected")

	if err := runMigrations(dsn); err != nil {
		log.Fatalf("migration error: %v", err)
		return
	}
	log.Println("migrations completed successfully")

	usersStore := users.NewPsqlHandler(sqldb.New(conn))
	usersHandler := users.NewHandler(usersStore, jwtSecret)
	authMiddleware := authm.NewAuthMiddleware(jwtSecret)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Post("/api/v1/auth/register", usersHandler.RegisterHandler)
	r.Post("/api/v1/auth/login", usersHandler.LoginHandler)

	// Protected routes
	r.With(authMiddleware.Protect).Get("/api/v1/me", usersHandler.GetMeHandler)

	srv := &http.Server{Addr: ":8080", Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exiting")
}
