// --- netlify/functions/api/main.go ---

package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"
)

var muxAdapter *gorillamux.GorillaMuxAdapter

// CHQ: Claude AI (Sonnet): corsMiddleware adds permissive 
// CORS headers and short-circuits preflight requests. 
// Tighten Access-Control-Allow-Origin before shipping to prod.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CHQ: Claude AI (Sonnet): newRouter wires up the routes backed by the 
// handlers that actually exist in handlers.go. Netlify strips the 
// /.netlify/functions/api prefix (or whatever redirect you configure) 
// before this router sees the path, so routes are declared relative 
// to /api.
func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(corsMiddleware)

	r.HandleFunc("/api/health", healthDBHandler).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/valid-dates", getValidDates).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/permits/{startDate}/{endDate}", scanDateRange).Methods(http.MethodGet, http.MethodOptions)

	return r
}

// CHQ: Claude AI (Sonnet): router adapts the classic API Gateway 
// (REST API / v1 proxy) event format, which is what Netlify 
// Functions sends, into a gorilla/mux request and back again.
func router(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switchableReq := core.NewSwitchableAPIGatewayRequestV1(&req)

	switchableResp, err := muxAdapter.ProxyWithContext(ctx, *switchableReq)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return *switchableResp.Version1(), nil
}

func main() {
	muxAdapter = gorillamux.New(newRouter())
	lambda.Start(router)
}