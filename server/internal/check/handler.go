package check

import (
	"net/http"

	gendb "github.com/trnahnh/katana-id/internal/db/generated"
)

type Handler struct {
	Queries     *gendb.Queries
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	// This bitch is the handler
}