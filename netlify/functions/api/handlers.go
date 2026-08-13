// --- netlify/functions/api/handlers.go ---

package main

import (

	// The `pq` package is a pure Go PostgreSQL driver for `database/sql`.

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

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