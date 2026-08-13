// --- netlify/functions/api/handlers.go ---

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// CHQ: Claude AI (Sonnet): getPermitsInDateRange queries permit_durations 
// for records whose file_date falls in [startDate, endDate).
func getPermitsInDateRange(startDate string, endDate string, w http.ResponseWriter) {
	dbConn, err := getDB()
	if err != nil {
		log.Printf("db unavailable: %v", err)
		http.Error(w, "Database unavailable", http.StatusInternalServerError)
		return
	}

	var constructionPermits []MyPermitRecord

	query := `SELECT
        "permit_id", "permit_number", "permit_type", "permit_subtype",
        "file_date", "issue_date", "final_date",
        "approval_duration", "construction_duration", "total_duration",
        "approval_ratio", "construction_ratio", "duration_category",
        "bottleneck_phase", "property_type", "job_value"
        FROM permit_durations
        WHERE "file_date" >= $1 AND "file_date" < $2
        ORDER BY "issue_date"`

	rows, err := dbConn.Query(query, startDate, endDate)
	if err != nil {
		log.Printf("Query failed: %v", err)
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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
			log.Printf("Failed to scan row from permit_durations: %v", err)
			http.Error(w, fmt.Sprintf("Failed to scan row: %v", err), http.StatusInternalServerError)
			return
		}
		constructionPermits = append(constructionPermits, record)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over permit_durations rows: %v", err)
		http.Error(w, fmt.Sprintf("Error iterating over rows: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(constructionPermits)
}

// CHQ: Claude AI (Sonnet): getValidDates returns the contents of 
// data_inventory. Columns are listed explicitly (rather than 
// SELECT *) so they always line up 1:1 with the Scan() 
// destinations below - SELECT * silently breaks this if the table
// gains or loses a column.
func getValidDates(w http.ResponseWriter, _ *http.Request) {
	dbConn, err := getDB()
	if err != nil {
		log.Printf("db unavailable: %v", err)
		http.Error(w, "Database unavailable", http.StatusInternalServerError)
		return
	}

	var theRecords []RecordStore

	query := `SELECT "id", "available_date", "table_name", "processed_at", "record_count" FROM data_inventory`

	rows, err := dbConn.Query(query)
	if err != nil {
		log.Printf("Query failed for data_inventory: %v", err)
		http.Error(w, fmt.Sprintf("Error retrieving records: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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
			log.Printf("Failed to scan row from data_inventory: %v", err)
			http.Error(w, fmt.Sprintf("Failed to scan row: %v", err), http.StatusInternalServerError)
			return
		}
		theRecords = append(theRecords, record)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over data_inventory rows: %v", err)
		http.Error(w, fmt.Sprintf("Error iterating over rows: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(theRecords)
}



// CHQ: Claude AI (Sonnet): scanDateRange validates the {startDate} 
// and {endDate} path params and delegates to getPermitsInDateRange.
func scanDateRange(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startDate := vars["startDate"]
	endDate := vars["endDate"]

	if !dateRegex.MatchString(startDate) || !dateRegex.MatchString(endDate) {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	getPermitsInDateRange(startDate, endDate, w)
}

// CHQ: Claude AI (Sonnet): healthDBHandler reports whether the database is reachable.
func healthDBHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbConn, err := getDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "error",
			"database": "unreachable",
			"error":    err.Error(),
		})
		return
	}

	if err := dbConn.Ping(); err != nil {
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