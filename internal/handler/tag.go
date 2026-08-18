package handler

import (
	"net/http"

	"wiki/internal/model"
	"wiki/pkg/httpx"
)

func (s *Server) registerTagRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/tags", s.createTag)
	mux.HandleFunc("GET /api/tags", s.listTags)
	mux.HandleFunc("GET /api/tags/{id}", s.getTag)
	mux.HandleFunc("DELETE /api/tags/{id}", s.deleteTag)
}

func (s *Server) createTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	t, err := s.svc.CreateTag(model.Tag{Name: req.Name})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, t)
}

func (s *Server) listTags(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 50, s.maxPageSize())
	items, total, err := s.svc.ListTags(pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getTag(w http.ResponseWriter, r *http.Request) {
	t, err := s.svc.GetTag(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, t)
}

func (s *Server) deleteTag(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteTag(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
