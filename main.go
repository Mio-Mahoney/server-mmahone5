package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jamespearly/loggly"
	"github.com/tomasen/realip"
)

// End Point Functions

// Returns the status of the server
func status(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	//get write TableName to response JSON.
	w.Write([]byte("Status"))
}

// Returns all the data in the table
func getAll(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	//get data from DynamoDB

	// Write DynamoDB data to response JSON.
	w.Write([]byte("All"))
}

func search(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	//get qurry

	//get data from DynamoDB

	// Write DynamoDB data to response JSON.
	w.Write([]byte("Search"))
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
				message := fmt.Sprintf("[%s] [%s] [%s] [%d] [%s] %s",
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
	router.HandleFunc("/mmahone5/status", status)
	router.HandleFunc("/mmahone5/search", search)
	router.HandleFunc("/mmahone5/all", getAll)
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
