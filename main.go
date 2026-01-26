// --- main.go ---

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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

	// protectedRoutes.HandleFunc("/monarchbutterlies/dayscan/{calendarDate}", getSingleDayScan).Methods("GET")
 
	router.HandleFunc("/monarchbutterlies/dayscan/{calendarDate}", getSingleDayScan).Methods("GET")
  
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

func getAllMonarchsAsAdmin2(w http.ResponseWriter, _ *http.Request) {
	var monarchButterflies []MyMonarchRecord
	query := `SELECT * FROM "2025_M06_JUN_2025_butterflies_CT" ORDER BY "date_only"`
	rows, err := db.Query(query)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving butterflies: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var monarchButterfly MyMonarchRecord
		err := rows.Scan(&monarchButterfly.DateOnly, &monarchButterfly.TimeOnly, &monarchButterfly.CityOrTown, &monarchButterfly.County, &monarchButterfly.StateProvince)
		if err != nil {
			log.Printf("Error scanning monarch butterfly row: %v", err)
			continue
		}
		monarchButterflies = append(monarchButterflies, monarchButterfly)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Error iterating over monarch butterfly rows: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(monarchButterflies)
}

// CHQ: Gemini AI corrected function
// Corrected getAllMonarchsAsAdmin to ignore the 'r' parameter
// func getMonarchButterfliesSingleDayAsAdmin(theTablename string, w http.ResponseWriter, r *http.Request) {

func getMonarchButterfliesSingleDayAsAdmin(theTablename string, w http.ResponseWriter, _ *http.Request) {
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
	
	var monarchButterflies []MyMonarchRecord
	tableName := theTablename
	
	// Explicitly listing all 35 columns to match the struct fields.
	query := fmt.Sprintf(`SELECT "gbifID", "datasetKey", "publishingOrgKey", "eventDate", "eventDateParsed", "year", "month", "day", "day_of_week", "week_of_year", "date_only", "scientificName", "vernacularName", "taxonKey", "kingdom", "phylum", "class", "order", "family", "genus", "species", "decimalLatitude", "decimalLongitude", "coordinateUncertaintyInMeters", "countryCode", "stateProvince", "individualCount", "basisOfRecord", "recordedBy", "occurrenceID", "collectionCode", "catalogNumber", "county", "cityOrTown", "time_only" FROM "%s" ORDER BY "date_only"`, tableName)
	
	// 3. Execute Query
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving butterflies: %v", err), http.StatusInternalServerError)
		log.Printf("Query failed for table %s: %v", tableName, err) // Log 3: Query failure
		return
	}
	defer rows.Close()

	// 4. Iterate and Scan Rows
	for rows.Next() {
		var record MyMonarchRecord
		err := rows.Scan(
			&record.GBIFID,
			&record.DatasetKey,
			&record.PublishingOrgKey,
			&record.EventDate,
			&record.EventDateParsed,
			&record.Year,
			&record.Month,
			&record.Day,
			&record.DayOfWeek,
			&record.WeekOfYear,
			&record.DateOnly,
			&record.ScientificName,
			&record.VernacularName,
			&record.TaxonKey,
			&record.Kingdom,
			&record.Phylum,
			&record.Class,
			&record.Order,
			&record.Family,
			&record.Genus,
			&record.Species,
			&record.DecimalLatitude,
			&record.DecimalLongitude,
			&record.CoordinateUncertaintyInMeters,
			&record.CountryCode,
			&record.StateProvince,
			&record.IndividualCount,
			&record.BasisOfRecord,
			&record.RecordedBy,
			&record.OccurrenceID,
			&record.CollectionCode,
			&record.CatalogNumber,
			&record.County,
			&record.CityOrTown,
			&record.TimeOnly,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to scan row: %v", err), http.StatusInternalServerError)
			log.Printf("Failed to scan row from table %s: %v", tableName, err) // Log 4: Scan failure
			return
		}
		monarchButterflies = append(monarchButterflies, record)
	}

	// 5. Check for Row Iteration Errors
	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Error iterating over monarch butterfly rows: %v", err), http.StatusInternalServerError)
		log.Printf("Error iterating over rows from table %s: %v", tableName, err) // Log 5: Row iteration error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(monarchButterflies)
}

// Corrected getAllMonarchsAsAdmin to ignore the 'r' parameter
func getAllMonarchsAsAdmin(w http.ResponseWriter, _ *http.Request) {
	// connStr := os.Getenv("IBM_DOCKER_PSQL_MONARCH")
	// db, err := sql.Open("postgres", connStr)
	// if err != nil {
	// 	http.Error(w, fmt.Sprintf("Failed to connect to database: %v", err), http.StatusInternalServerError)
	// 	log.Printf("Failed to connect to database: %v", err)
	// 	return
	// }
	// defer db.Close()

	// err = db.Ping()
	// if err != nil {
	// 	http.Error(w, fmt.Sprintf("Database ping failed: %v", err), http.StatusInternalServerError)
	// 	log.Printf("Database ping failed: %v", err)
	// 	return
	// }

	// getMonarchButterfliesSingleDayAsAdmin("june212025", w, nil)
}
 
// CHQ: Gemini AI corrected parameters to ignore the r
func getAllMonarchs(w http.ResponseWriter, _ *http.Request) {
	// if (isAnAdmin) {
		getAllMonarchsAsAdmin2(w, nil) // You can pass nil as the request since the function doesn't use it

    // getAllMonarchsAsAdmin(w, nil) // You can pass nil as the request since the function doesn't use it
	// } else {
		// getAllMonarchsAsTeacher(w, r)
	// }
}

func generateTableName(day int, monthInt int, year int) string {
	// 1. Define the equivalent of my_calendar (a map in Go)
	// You would typically define this map globally or pass it in,
	// but defining it here works for a direct conversion.

	monthIntToStr := ""

	if(monthInt < 10){
		monthIntToStr = ("0" + strconv.Itoa(monthInt))
	} else {
		monthIntToStr = strconv.Itoa(monthInt)
	}
  
	myCalendar := map[string]string{
		"01":   "january",
		"02":   "february",
		"03":   "march",
		"04":   "april",
		"05":   "may",
		"06":   "june",
		"07":   "july",
		"08":    "august",
		"09": "september",
		"10":   "october",
		"11":  "november",
		"12":  "december",
	}

	// Retrieve the month string from the map
	// monthStr, ok := myCalendar[month]
	monthStr, ok := myCalendar[monthIntToStr]
 
	if !ok {
		// Handle case where the month is not found (optional, but good practice)
		return "Error: Invalid month"
	}

	// 2. Implement the conditional logic and string concatenation
	var tableName string
	yearStr := strconv.Itoa(year) // Convert int year to string

  	// FIX: Use %02d to ensure single-digit days (like 8) are formatted with a leading zero ("08").
	// This makes it compatible with table names like june082025.
	dayStr := fmt.Sprintf("%02d", day)
	tableName = fmt.Sprintf("%s%s%s", monthStr, dayStr, yearStr)

	return tableName
}


// func getValidDates(w http.ResponseWriter, r *http.Request) {
func getValidDates(w http.ResponseWriter, _ *http.Request) {

	
// }
// func getMonarchButterfliesSingleDayAsAdmin(theTablename string, w http.ResponseWriter, _ *http.Request) {
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
func getSingleDayScan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Get the date as a string (MMDDYYYY)
	dateStr := vars["calendarDate"]

	// **********************************************
	// DEBUG LOGGING ADDED HERE TO SEE THE RECEIVED DATE STRING AND LENGTH
	// **********************************************
	log.Printf("Received calendarDate: %s (Length: %d)", dateStr, len(dateStr))

	// 1. String length check (must be exactly 8 characters)
	if len(dateStr) != 8 {
		http.Error(w, "Invalid date given - expected 8 digits in MMDDYYYY format", http.StatusBadRequest)
		log.Printf("Invalid date string length: %s, expected 8", dateStr)
		return
	}

	// 2. Extract components using string slicing (MMDDYYYY)
	monthStr := dateStr[0:2] // MM (e.g., "06")
	dayStr := dateStr[2:4]   // DD (e.g., "30")
	yearStr := dateStr[4:8]  // YYYY (e.g., "2025")

	// 3. Convert day and year to integers for table name generation
	dayInt, err := strconv.Atoi(dayStr)
	if err != nil {
		http.Error(w, "Invalid day format in date", http.StatusBadRequest)
		log.Printf("Invalid day format: %s", dayStr)
		return
	}

	yearInt, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year format in date", http.StatusBadRequest)
		log.Printf("Invalid year format: %s", yearStr)
		return
	}
	
	// Check if monthStr is a valid two-digit number
	// Although we use monthStr in generateTableName, we need to ensure it's a number
	if _, err := strconv.Atoi(monthStr); err != nil {
		http.Error(w, "Invalid month format in date: not a number", http.StatusBadRequest)
		log.Printf("Invalid month format: %s", monthStr)
		return
	}

	monthInt, err := strconv.Atoi(monthStr)
	if err != nil {
		http.Error(w, "Invalid month format in date", http.StatusBadRequest)
		log.Printf("Invalid month format: %s", monthStr)
		return	
	}
	
	// The `useVariable` flag is preserved from your original code
	useVariable := false 

	// Generate the dynamic table name using the string-based month
	myChoice := generateTableName(dayInt, monthInt, yearInt)

	// If useVariable is false, override with the hardcoded test table name
	if useVariable {
		myChoice = "june212025"
		log.Printf("Using hardcoded table name: %s", myChoice)
	} else {
		log.Printf("Using generated table name: %s", myChoice)
	}

	// Call the function to fetch data from the determined table
	getMonarchButterfliesSingleDayAsAdmin(myChoice, w, nil)
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