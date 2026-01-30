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

// Define the structure for the request body
// type ViewCreationRequest struct {
//     PYear     int    `json:"p_year"`
//     PMonth    string `json:"p_month"`
//     PStartDay int    `json:"p_start_day"`
//     PEndDay   int    `json:"p_end_day"`
//     PState    string `json:"p_state"`
// }

// MonarchRecord represents a row from the database table.
// Using a map[string]interface{} is flexible since the schema
// might change, similar to the dynamic dictionary creation in the Python version.
type MonarchRecord map[string]interface{}

// MonarchRecord represents a row from the database table 'june212025'. 
type RecordStore struct {
	Id                   *int    `json:"id"`
	Available_date           *string    `json:"available_date"`
	Table_name         *string    `json:"table_name"`
	Processed_at                *string    `json:"processed_at"`
 	Record_count                     *int       `json:"record_count"`    
	// TimeOnly                 *string    `json:"time_only"` // Storing as string to handle "time without time zone"
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
   	TimeOnly                   *string    `json:"time_only"` // Storing as string to handle "time without time zone"
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

	// protectedRoutes.HandleFunc("/monarchbutterlies/dayscan/{calendarDate}", scanDateRange).Methods("GET")
 
	router.HandleFunc("/permittracker/scanner/from={startDate}to={endDate}", scanDateRange).Methods("GET")
  
	router.HandleFunc("/monarchbutterlies/scanneddates", getValidDates).Methods("GET")

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
// func getPermitsInDateRange(theTablename string, w http.ResponseWriter, r *http.Request) {

func getPermitsInDateRange(theTablename string, w http.ResponseWriter, _ *http.Request) {
	// // 1. Establish DB Connection
	// connStr := os.Getenv("IBM_DOCKER_PSQL_MONARCH")
	// db, err := sql.Open("postgres", connStr)
	// if err != nil {
	// 	log.Printf("Failed to connect to database: %v", err) // Log 1: Connection failure
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
	
	var constructionPermits []MyPermitRecord
	tableName := theTablename
	
	// Explicitly listing all columns to match the struct fields.
	query := fmt.Sprintf(`SELECT
	"PermitID",
	"PermitNumber",
	"PermitType",
	"PermitSubtype",
			"FileDate",
			"IssueDate",
			"FinalDate",
			"ApprovalDuration",
			"ConstructionDuration",
			"TotalDuration",
			"ApprovalRatio",
			"ConstructionRatio",
			"DurationCategory",
			"BottleneckPhase",
			"PropertyType",
			"JobValue",
			"TimeOnly", 	
		
		FROM "%s" ORDER BY "IssueDate"`, tableName)
	 

	// 3. Execute Query
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving permits: %v", err), http.StatusInternalServerError)
		log.Printf("Query failed for table %s: %v", tableName, err) // Log 3: Query failure
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
			&record.TimeOnly,   
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to scan row: %v", err), http.StatusInternalServerError)
			log.Printf("Failed to scan row from table %s: %v", tableName, err) // Log 4: Scan failure
			return
		}
		constructionPermits = append(constructionPermits, record)
	}

	// 5. Check for Row Iteration Errors
	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Error iterating over monarch butterfly rows: %v", err), http.StatusInternalServerError)
		log.Printf("Error iterating over rows from table %s: %v", tableName, err) // Log 5: Row iteration error
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

	// Get the date as a string (YYYY-MM-DD)
	startDate := vars["startDate"]
	endDate := vars["endDate"]

	// **********************************************
	// DEBUG LOGGING ADDED HERE TO SEE THE RECEIVED DATE STRING AND LENGTH
	// **********************************************
	log.Printf("Received startDate: %s (Length: %d)", startDate, len(startDate))
	log.Printf("Received endDate: %s (Length: %d)", startDate, len(endDate))

	theRegex := `^\\d{4}-\\d{2}-\\d{2}$` //CHQ: source - Gemini (via Google Search)

	startDayMatch, err1 := regexp.MatchString(theRegex, startDate)
	if err1 != nil {
		fmt.Println("Error compiling regex:", err1)
		return
	}

	endDayMatch, err2 := regexp.MatchString(theRegex, endDate)
	if err2 != nil {
		fmt.Println("Error compiling regex:", err2)
		return
	}

	// 1. String length check (must be exactly 8 characters)
	if (!(startDayMatch && endDayMatch) ){
	// if len(startDate) != 10 || len(endDate) != 10 {
		http.Error(w, "Invalid date given - expected format of YYYY-MM-DD format", http.StatusBadRequest)
		log.Printf("Invalid date given - expected format of YYYY-MM-DD format")
		return
	} 
	
	// i should make the ETL so that instead of permit_durations, it titles each table it makes 
	// permit_durations_date1_date2
	
	// The `useVariable` flag is preserved from your original code
	// useVariable := false 

	// Generate the dynamic table name using the string-based month
	myChoice := generateTableName(startDate, endDate)

	// // If useVariable is false, override with the hardcoded test table name
	// if useVariable {
	// 	myChoice = "permit_durations_2025-06-30_to_2026-01-24"
	// 	log.Printf("Using hardcoded table name: %s", myChoice)
	// } else {
	// 	log.Printf("Using generated table name: %s", myChoice)
	// }

	// Call the function to fetch data from the determined table
	getPermitsInDateRange(myChoice, w, nil)
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