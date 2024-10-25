package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jamespearly/loggly"
	"github.com/tomasen/realip"
)

func getTime(w http.ResponseWriter, r *http.Request) {
	now := time.Now().Format("01-02-2006 3:4:5 PM")

	Time, _ := json.Marshal(now)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(Time)
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewStatusResponseWriter returns pointer to a new statusResponseWriter object
func NewStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}
func (sw *statusResponseWriter) WriteHeader(statusCode int) {
	sw.statusCode = statusCode
	sw.ResponseWriter.WriteHeader(statusCode)
}

func RequestLoggerMiddleware(r *mux.Router, logger *loggly.ClientType) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			sw := NewStatusResponseWriter(w)
			defer func() {
				message := fmt.Sprintf("[%s] %s %s [%d] %s \nip: %s",
					req.Method,
					req.Host,
					req.URL.Path,
					sw.statusCode,
					req.URL.RawQuery,
					realip.FromRequest(req))
				logger.EchoSend("info", message)
			}()
			next.ServeHTTP(sw, req)
		})
	}
}

func main() {
	router := mux.NewRouter()
	logger := loggly.New("mmahone5")
	router.HandleFunc("/mmahone5/status", getTime)
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	})
	router.Use(RequestLoggerMiddleware(router, logger))

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		fmt.Print("status", "fatal", "err", err)
		os.Exit(1)
	}
}
