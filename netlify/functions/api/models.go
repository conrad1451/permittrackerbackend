// --- netlify/functions/api/models.go ---

package main

import (
	"time"

	// The `pq` package is a pure Go PostgreSQL driver for `database/sql`.

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