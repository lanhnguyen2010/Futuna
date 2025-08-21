package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"futuna/internal/analyzer"
)

// Router returns an HTTP handler for the web API and static files.
func Router(svc *analyzer.Service) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/analysis", func(w http.ResponseWriter, r *http.Request) {
		dateStr := r.URL.Query().Get("date")
		var date time.Time
		var err error
		if dateStr == "" {
			date = time.Now().Truncate(24 * time.Hour)
		} else {
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				http.Error(w, "invalid date", http.StatusBadRequest)
				return
			}
		}
		rows, err := svc.ListAnalyses(r.Context(), date)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(rows)
	})
	r.Get("/api/tickers", func(w http.ResponseWriter, r *http.Request) {
		rows, err := svc.ListTickers(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(rows)
	})
	fileServer := http.FileServer(http.Dir("web/static"))
	r.Handle("/*", fileServer)
	return r
}
