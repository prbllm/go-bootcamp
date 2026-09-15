// Заготовка сервера. Закрывайте этапы по одному, пока не позеленеет go test ./tests/ -v
//
// Запуск: go run ./cmd/server — порт берётся из PORT, по умолчанию 8080.
// Панель уже раздаётся: откройте http://localhost:8080/ и смотрите, как этапы
// зеленеют по ходу работы.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"entrytest/internal/config"
	"entrytest/internal/handlers"
	"entrytest/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	port := os.Getenv(config.PortEnv)
	if port == "" {
		port = config.PortDefault
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("некорректный PORT: %v", err)

		return
	}

	g, shutdownCtx := errgroup.WithContext(ctx)

	store := storage.New()
	h := handlers.New(store)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get(config.HealthPath, h.HealthHandler)
	router.Post(config.EchoPath, h.EchoHandler)
	router.Get(config.MessagesPath, h.MessagesListHandler)
	router.Post(config.MessagesPath, h.MessagesHandler)
	router.Delete(config.MessagesPath+"/{id}", h.MessagesDeleteHandler)

	// Панель из frontend/. Каталог берётся относительно рабочего, поэтому
	// запускайте из корня модуля: go run ./cmd/server
	router.Handle("/", http.FileServer(http.Dir("frontend")))

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(portNum),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	g.Go(func() error {
		log.Printf("сервер слушает http://localhost:%d", portNum)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	g.Go(func() error {
		<-shutdownCtx.Done()

		timeoutCtx, cancel := context.WithTimeout(context.WithoutCancel(shutdownCtx), 10*time.Second)
		defer cancel()

		log.Println("shutting down...")

		return srv.Shutdown(timeoutCtx)
	})

	if err := g.Wait(); err != nil {
		log.Printf("server stopped with error: %v", err)

		return
	}

	log.Println("shutdown complete")
}
