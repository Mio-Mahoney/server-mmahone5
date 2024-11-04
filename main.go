package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
	"github.com/jamespearly/loggly"
	"github.com/tomasen/realip"
)

// Global Variables
var tableName = "mmahone5_test"
var svc *dynamodb.Client

type invalidQuery struct {
	Error   int32  `json:"error"`
	Message string `json:"message"`
}

// End Point Functions
// Returns the status of the server
func status(w http.ResponseWriter, r *http.Request) {
	//get write TableName to response JSON.

	res, err := svc.Scan(context.TODO(), &dynamodb.ScanInput{TableName: aws.String(tableName),
		Select: "COUNT",
	})
	if err != nil {
		panic(err)
	}

	recordCount := res.ScannedCount //Get Record count
	rjson := map[string]string{"Table Name": tableName, "RecordCount": fmt.Sprint(recordCount)}

	daJson, _ := json.Marshal(rjson)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write(daJson)

}

// Returns all the data in the table
func getAll(w http.ResponseWriter, r *http.Request) {
	//get data from DynamoDB
	// Write DynamoDB data to response JSON.
	res, err := svc.Scan(context.TODO(), &dynamodb.ScanInput{TableName: aws.String(tableName)})
	if err != nil {
		panic(err)
	}

	resJ, _ := json.Marshal(res.Items)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resJ)
}

func search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//get qurry
	q := r.URL.Query().Get("q")
	fmt.Print(q)
	possibleInputs := []string{"SNOW_BLOCK", "SNOW_BALL", "ENCHANTED_SNOW_BLOCK"}

	//santize query
	resultFound := false
	for _, input := range possibleInputs {
		if q == input { // Case-sensitive comparison
			resultFound = true
			break
		}
	}

	if !resultFound {
		invalidQueryJ, _ := json.Marshal(invalidQuery{Error: 400, Message: "Invalid Query"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(invalidQueryJ)
		return
	}

	input := &dynamodb.ScanInput{
		TableName:        aws.String(tableName),
		FilterExpression: aws.String("attribute_exists(Products.#productKey)"),
		ExpressionAttributeNames: map[string]string{
			"#productKey": q,
		},
	}

	res, err := svc.Scan(context.TODO(), input)
	if err != nil {
		fmt.Printf("Failed to scan items: %v", err)
		invalidScanJ, _ := json.Marshal(invalidQuery{Error: http.StatusInternalServerError, Message: "Failed to scan items"})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(invalidScanJ)
		return
	}

	var filteredResults []map[string]interface{}
	for _, item := range res.Items {
		if productInfo, ok := item["Products"].(*types.AttributeValueMemberM).Value[q]; ok {
			// Extract the specific product info
			filteredResults = append(filteredResults, map[string]interface{}{
				"product": productInfo.(*types.AttributeValueMemberM).Value,
			})
		}
	}

	resJ, _ := json.Marshal(filteredResults)
	w.WriteHeader(http.StatusOK)

	w.Write(resJ)
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
	//AWS Config
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = "us-east-1"
		return nil
	})
	if err != nil {
		panic(err)
	}
	svc = dynamodb.NewFromConfig(cfg)

	// Create a new router
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
