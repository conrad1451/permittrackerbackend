// --- main.go ---

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	// The `pq` package is a pure Go PostgreSQL driver for `database/sql`.

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)
 
 type RecordStore struct {
	Id                   *int    `json:"id"`
	Available_date           *string    `json:"available_date"`
	Table_name         *string    `json:"table_name"`
	Processed_at                *string    `json:"processed_at"`
 	Record_count                     *int       `json:"record_count"`    
}

// CHQ: Gemini AI edited struct to correct types
type MyPermitRecord struct {
   	PermitID                   *string    `json:"permit_id"`
	PermitNumber               *string    `json:"permit_number"`
	PermitType         		   *string    `json:"permit_type"`
	PermitSubtype              *string    `json:"permit_subtype"`
	Status          		   *string    `json:"status"`
	FileDate                   *time.Time `json:"file_date"`
	IssueDate                  *time.Time `json:"issue_date"`
	FinalDate                  *time.Time `json:"final_date"`
	ApprovalDuration           *int64     `json:"approval_duration"`
	ConstructionDuration       *int64     `json:"construction_duration"`
	TotalDuration              *int64     `json:"total_duration"`
	ApprovalRatio              *float64   `json:"approval_ratio"`
	ConstructionRatio          *float64   `json:"construction_ratio"`
	DurationCategory           *string    `json:"duration_category"`
	BottleneckPhase            *string    `json:"bottleneck_phase"`
	PropertyType               *string    `json:"property_type"`
	JobValue                   *float64   `json:"job_value"`
}

var db *sql.DB
 

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("FATAL: required environment variable %s is not set", key)
	}
	return val
}

	 

// CHQ: Gemini AI generated function
// helloHandler is the function that will be executed for requests to the "/" route.
func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "This is the server for the monarch butterflies app. It's written in Go (aka GoLang).")
}


// faviconHandler serves the favicon.ico file.
func faviconHandler(w http.ResponseWriter, r *http.Request) {
    // Open the favicon file
    favicon, err := os.ReadFile("./static/butterfly_net.ico")
    if err != nil {
        http.NotFound(w, r)
        return
    }

    // Set the Content-Type header
    w.Header().Set("Content-Type", "image/x-icon")
    
    // Write the file content to the response
    w.Write(favicon)
}
 
func main() { 
	connStr := mustGetEnv("XATA_DB_CONSTRUCTION")
	
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	// Set up the HTTP router.
		// Initialize the router
	router := mux.NewRouter()

	router.HandleFunc("/", helloHandler)
	router.HandleFunc("/favicon.ico", faviconHandler)
	router.HandleFunc("/health/db", healthDBHandler)
	// Protected routes (require session validation)
	// protectedRoutes := router.PathPrefix("/api").Subrouter()
	// protectedRoutes.Use(sessionValidationMiddleware) // Apply middleware to all routes in this subrouter

	router.HandleFunc("/permittracker/scanner/{startDate}/{endDate}", scanDateRange).Methods("GET")
	
	// --- CORS Setup ---
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	corsRouter := handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(router)
 
 	// Start the HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	fmt.Printf("Server listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, corsRouter))
}
 

// CHQ: Gemini AI corrected function
// Corrected getAllMonarchsAsAdmin to ignore the 'r' parameter
func getPermitsInDateRange(startDate string, endDate string, w http.ResponseWriter) {
	var constructionPermits []MyPermitRecord

    // Querying the static table with a WHERE clause
    query := `SELECT
        "permit_id", "permit_number", "permit_type", "permit_subtype",
        "file_date", "issue_date", "final_date",
        "approval_duration", "construction_duration", "total_duration",
        "approval_ratio", "construction_ratio", "duration_category",
        "bottleneck_phase", "property_type", "job_value"
        FROM permit_durations 
        WHERE "file_date" >= $1 AND "file_date" < $2
        ORDER BY "issue_date"`

    rows, err := db.Query(query, startDate, endDate)
    if err != nil {
        log.Printf("Query failed: %v", err)
        http.Error(w, "Database query error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

	// 4. Iterate and Scan Rows
	for rows.Next() {
		var record MyPermitRecord
		err := rows.Scan(
			&record.PermitID,
			&record.PermitNumber,
			&record.PermitType,
			&record.PermitSubtype,
			&record.FileDate,
			&record.IssueDate,
			&record.FinalDate,
			&record.ApprovalDuration,
			&record.ConstructionDuration,
			&record.TotalDuration,
			&record.ApprovalRatio,
			&record.ConstructionRatio,
			&record.DurationCategory,
			&record.BottleneckPhase,
			&record.PropertyType,
			&record.JobValue,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to scan row: %v", err), http.StatusInternalServerError)
			log.Printf("Failed to scan row from permit_durations: %v", err) // Log 4: Scan failure
			return
		}
		constructionPermits = append(constructionPermits, record)
	}

	// 5. Check for Row Iteration Errors
	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Error iterating over monarch butterfly rows: %v", err), http.StatusInternalServerError)
		// log.Printf("Error iterating over rows from table %s: %v", tableName, err) // Log 5: Row iteration error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(constructionPermits)
}
 

func generateTableName(startDate string, endDate string) string {
   	var tableName string 
	tableName = "permit_durations_" + startDate + "_to_" + endDate 

	return tableName
}

// func getValidDates(w http.ResponseWriter, r *http.Request) {
func getValidDates(w http.ResponseWriter, _ *http.Request) {

	
// }
// func getPermitsInDateRange(theTablename string, w http.ResponseWriter, _ *http.Request) {
// 	// // 1. Establish DB Connection
// 	// connStr := os.Getenv("IBM_DOCKER_PSQL_MONARCH")
// 	// db, err := sql.Open("postgres", connStr)
// 	// if err != nil {
// 	// 	log.Printf("Failed to connect to database: %v", err) // Log 1: Connection failure
	// 	http.Error(w, fmt.Sprintf("Failed to connect to database: %v", err), http.StatusInternalServerError)
	// 	return
	// }
	// defer db.Close()

	// // 2. Ping DB
	// err = db.Ping()
	// if err != nil {
	// 	log.Printf("Database ping failed: %v", err) // Log 2: Ping failure
	// 	http.Error(w, fmt.Sprintf("Database ping failed: %v", err), http.StatusInternalServerError)
	// 	return
	// }
	
	var theRecords []RecordStore
 	
	// Explicitly listing all 35 columns to match the struct fields.
	query := `SELECT * FROM data_inventory`
	// query := `SELECT * FROM december012021`
	 
	// 3. Execute Query
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving butterflies: %v", err), http.StatusInternalServerError)
		log.Printf("Query failed for table:", err) // Log 3: Query failure
		return
	}
	defer rows.Close()

	// 4. Iterate and Scan Rows
	for rows.Next() {
		var record RecordStore
		err := rows.Scan(
			&record.Id,
			&record.Available_date,   
			&record.Table_name,
			&record.Processed_at,
			&record.Record_count,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to scan row: %v", err), http.StatusInternalServerError)
			// log.Printf("Failed to scan row from table %s: %v", tableName, err) // Log 4: Scan failure
			return
		}
		theRecords = append(theRecords, record)
	}

	// 5. Check for Row Iteration Errors
	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Error iterating over monarch butterfly rows: %v", err), http.StatusInternalServerError)
		// log.Printf("Error iterating over rows from table %s: %v", tableName, err) // Log 5: Row iteration error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(theRecords)
}
 

// CHQ: Gemini AI added log statements to debug
func scanDateRange(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    startDate := vars["startDate"]
    endDate := vars["endDate"]

    log.Printf("Received startDate: %s", startDate)
    log.Printf("Received endDate: %s", endDate)

    // Using the fixed Raw String Literal with backticks
    theRegex := `^\d{4}-\d{2}-\d{2}$`
    
    startMatch, _ := regexp.MatchString(theRegex, startDate)
    endMatch, _ := regexp.MatchString(theRegex, endDate)

    if !startMatch || !endMatch {
        http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
        return
    }

    // myChoice := generateTableName(startDate, endDate)
    // getPermitsInDateRange(myChoice, w, r)
	getPermitsInDateRange(startDate, endDate, w)
}

 


// CHQ: Gemini AI created function
// monitors health of database
func healthDBHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "error",
			"database": "unreachable",
			"error":    err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "ok",
		"database": "reachable",
	})
}

//     w.Header().Set("Content-Type", "application/json")
//     w.WriteHeader(http.StatusOK)
//     json.NewEncoder(w).Encode(map[string]string{
//         "status": "success",
//         "message": confirmationMessage,
//     })
// }
