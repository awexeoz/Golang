package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var log = logrus.New()
var limiter = rate.NewLimiter(1, 3) // Rate limit of 1 request per second with a burst of 3 requests

func init() {
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	log.WithFields(logrus.Fields{
		"action": "start",
		"status": "success",
	}).Info("Application starting...")

	initMongoDB()
	router := mux.NewRouter()

	// Serve static files (HTML, CSS, JS) from the "static" directory
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Handle requests for the HTML page
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	}).Methods("GET")

	// Обработчик GET-запроса для эндпоинта "/laptops"
	router.HandleFunc("/laptops", rateLimitMiddleware(getLaptops)).Methods("GET") // Add rate limiting middleware

	// Подготовка к приему сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Запуск веб-сервера в отдельной горутине
	srv := &http.Server{Addr: ":5000", Handler: router} // Добавлен Handler: router
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Обработка сигнала завершения
	<-quit
	log.Println("Server is shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}

// rateLimitMiddleware is a middleware function to enforce rate limiting
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			// Return a JSON response when the rate limit is exceeded
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			response := map[string]string{"error": "Rate limit exceeded. Please try again later."}
			json.NewEncoder(w).Encode(response)
			return
		}
		next(w, r)
	}
}
