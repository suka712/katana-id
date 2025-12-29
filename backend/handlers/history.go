package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"katanaid/database"
)

// AnalysisRecord represents a single analysis from the database
type AnalysisRecord struct {
	ID         int       `json:"id"`
	FileID     string    `json:"file_id"`
	Filename   string    `json:"filename"`
	FileType   string    `json:"file_type"`
	Result     string    `json:"result"`
	Confidence float64   `json:"confidence"`
	Details    string    `json:"details"`
	CreatedAt  time.Time `json:"created_at"`
}

// HistoryResponse is the success response
type HistoryResponse struct {
	Analyses []AnalysisRecord `json:"analyses"`
	Count    int              `json:"count"`
	Message  string           `json:"message"`
}

// History handles GET /api/history
func History(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user_id from JWT token and filter by user
	// For now, return all analyses

	rows, err := database.DB.Query(
		context.Background(),
		`SELECT id, file_id, filename, file_type, result, confidence, details, created_at
		 FROM analyses
		 ORDER BY created_at DESC
		 LIMIT 50`,
	)

	if err != nil {
		log.Print("Error fetching history:", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch history"})
		return
	}
	defer rows.Close()

	var analyses []AnalysisRecord

	for rows.Next() {
		var record AnalysisRecord
		var details *string // Handle NULL details

		err := rows.Scan(
			&record.ID,
			&record.FileID,
			&record.Filename,
			&record.FileType,
			&record.Result,
			&record.Confidence,
			&details,
			&record.CreatedAt,
		)

		if err != nil {
			log.Print("Error scanning row:", err)
			continue
		}

		if details != nil {
			record.Details = *details
		}

		analyses = append(analyses, record)
	}

	if err := rows.Err(); err != nil {
		log.Print("Error iterating rows:", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch history"})
		return
	}

	// Handle empty results
	if analyses == nil {
		analyses = []AnalysisRecord{}
	}

	log.Printf("Returning %d analysis records", len(analyses))

	writeJSON(w, http.StatusOK, HistoryResponse{
		Analyses: analyses,
		Count:    len(analyses),
		Message:  "History retrieved successfully",
	})
}
