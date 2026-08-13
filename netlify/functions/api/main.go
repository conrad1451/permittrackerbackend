// --- netlify/functions/api/main.go ---

package main

import (

	// The `pq` package is a pure Go PostgreSQL driver for `database/sql`.

	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"
)

// router dispatches based on HTTP method and whether the path ends in an id.
// Routes:
//
//	GET    /api/items       -> list
//	POST   /api/items       -> create
//	GET    /api/items/{id}  -> get one
//	PUT    /api/items/{id}  -> update
//	DELETE /api/items/{id}  -> delete
func router(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{StatusCode: 204, Headers: corsHeaders}, nil
	}

	id, hasID := pathID(req.Path)

	switch req.HTTPMethod {
	case "GET":
		if hasID {
			return getItem(ctx, id)
		}
		return listItems(ctx)
	case "POST":
		return createItem(ctx, req.Body)
	case "PUT":
		if !hasID {
			return errorResponse(400, "id required in path for update"), nil
		}
		return updateItem(ctx, id, req.Body)
	case "DELETE":
		if !hasID {
			return errorResponse(400, "id required in path for delete"), nil
		}
		return deleteItem(ctx, id)
	default:
		return errorResponse(405, "method not allowed"), nil
	}
}

func main() {
	lambda.Start(router)
}
