// Заготовка сервера. Закрывайте этапы по одному, пока не позеленеет go test ./tests/ -v
//
// Запуск: go run ./cmd/server — порт берётся из PORT, по умолчанию 8080.
// Панель уже раздаётся: откройте http://localhost:8080/ и смотрите, как этапы
// зеленеют по ходу работы.
package main

import (
	"context"
	"entrytest/internal/config"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	port := os.Getenv(config.PORT_ENV)
	if port == "" {
		port = config.PORT_DEFAULT
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("некорректный PORT: %v", err)
		os.Exit(1)
	}

	g, shutdownCtx := errgroup.WithContext(ctx)

	mux := http.NewServeMux()

	// Панель из frontend/. Каталог берётся относительно рабочего, поэтому
	// запускайте из корня модуля: go run ./cmd/server
	mux.Handle("/", http.FileServer(http.Dir("frontend")))

	// TODO Этап 1: GET /health           -> 200, тело "ok"
	// TODO Этап 2: POST /echo            -> тело запроса без изменений
	// TODO Этап 3: POST /echo            -> на application/json разобрать {"message": "..."} и вернуть JSON
	// TODO Этап 4: POST /messages        -> сохранить в памяти, 201
	// TODO Этап 5: GET /messages         -> все сообщения, новые сверху
	// TODO Этап 6: DELETE /messages/{id} -> 204, либо 404 если такого нет

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(portNum),
		Handler:           mux,
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

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Println("shutting down...")
		return srv.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
	log.Println("shutdown complete")
}
