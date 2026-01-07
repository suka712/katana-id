package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"google.golang.org/genai"
)

func GenerateUsername(w http.ResponseWriter, r *http.Request) {
	var req UsernameGenerationRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Print("Error decoding JSON:", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Something went wrong"})
		return
	}

	count, err := strconv.Atoi(req.Count)
	if err != nil {
		log.Print("Invalid value for count:", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Something went wrong"})
		return
	}

	if count > 10 || count < 0 {
		log.Print("Invalid username request")
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		genai.Text("Generate six random username and output in csv format. No additional text"),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	writeJSON(w, http.StatusOK, UsernameGenerationSuccessResponse{Usernames: result.Text()})
}
