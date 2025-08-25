package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"futuna/internal/analyzer"
)

// Router returns an HTTP handler for the web API.
func Router(svc *analyzer.Service) http.Handler {
	r := chi.NewRouter()

	// simple CORS for dev front-end
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

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

	r.Get("/api/dates", func(w http.ResponseWriter, r *http.Request) {
		rows, err := svc.ListAnalysisDates(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// format as YYYY-MM-DD strings
		out := make([]string, 0, len(rows))
		for _, d := range rows {
			out = append(out, d.Format("2006-01-02"))
		}
		json.NewEncoder(w).Encode(out)
	})

	return r
}
