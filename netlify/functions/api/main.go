// --- netlify/functions/api/main.go ---

package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"
)

// functionPrefix is the path segment Netlify puts in front of every request
// to this function. Unlike AWS API Gateway with a custom base path mapping,
// Netlify does NOT strip this before invoking the function - req.Path
// arrives as e.g. "/.netlify/functions/api/api/health", so we trim it
// ourselves before handing the request to mux.
const functionPrefix = "/.netlify/functions/api"

var muxAdapter *gorillamux.GorillaMuxAdapter

// corsMiddleware adds permissive CORS headers and short-circuits preflight
// requests. Tighten Access-Control-Allow-Origin before shipping to prod.
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

// newRouter wires up the routes backed by the handlers that actually exist
// in handlers.go. Routes are declared relative to /api - the
// /.netlify/functions/api prefix is stripped in router() below before mux
// ever sees the path.
func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(corsMiddleware)

	r.HandleFunc("/api/health", healthDBHandler).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/valid-dates", getValidDates).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/permits/{startDate}/{endDate}", scanDateRange).Methods(http.MethodGet, http.MethodOptions)

	return r
}

// router adapts the classic API Gateway (REST API / v1 proxy) event format,
// which is what Netlify Functions sends, into a gorilla/mux request and
// back again. It strips Netlify's function-name prefix from the path first,
// since Netlify (unlike API Gateway) doesn't do that for you.
func router(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	trimmed := strings.TrimPrefix(req.Path, functionPrefix)
	if trimmed == "" {
		trimmed = "/"
	}
	req.Path = trimmed

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