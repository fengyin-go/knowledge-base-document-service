package handler

import (
	"net/http"

	"wiki/internal/service"
	"wiki/pkg/httpx"
)

func (s *Server) registerExportRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/export/documents", s.exportDocuments)
	mux.HandleFunc("POST /api/import/documents", s.importDocuments)
}

func (s *Server) exportDocuments(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.ExportDocuments()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

func (s *Server) importDocuments(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Documents []service.DocumentImportItem `json:"documents"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	result, err := s.svc.ImportDocuments(req.Documents)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, result)
}
