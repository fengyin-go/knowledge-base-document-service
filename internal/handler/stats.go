package handler

import (
	"net/http"
	"strconv"

	"wiki/pkg/httpx"
)

func (s *Server) registerStatsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/stats/documents", s.documentStats)
	mux.HandleFunc("GET /api/stats/tags", s.tagStats)
	mux.HandleFunc("GET /api/stats/authors", s.authorStats)
	mux.HandleFunc("GET /api/stats/popular", s.popularDocuments)
	mux.HandleFunc("GET /api/stats/monthly", s.monthlyStats)
	mux.HandleFunc("GET /api/stats/overview", s.overviewStats)
}

func (s *Server) documentStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.svc.DocumentStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}

func (s *Server) tagStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.svc.TagStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}

func (s *Server) authorStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.svc.ListAuthorStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}

func (s *Server) popularDocuments(w http.ResponseWriter, r *http.Request) {
	n := 10
	if v := r.URL.Query().Get("n"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			n = parsed
		}
	}
	items, err := s.svc.PopularDocuments(n)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) monthlyStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.svc.MonthlyStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}

func (s *Server) overviewStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.svc.OverviewStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}
