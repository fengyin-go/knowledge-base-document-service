package handler

import (
	"net/http"

	"wiki/internal/model"
	"wiki/pkg/httpx"
)

func (s *Server) registerAuthorRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/authors", s.createAuthor)
	mux.HandleFunc("GET /api/authors", s.listAuthors)
	mux.HandleFunc("GET /api/authors/{id}", s.getAuthor)
	mux.HandleFunc("PUT /api/authors/{id}", s.updateAuthor)
	mux.HandleFunc("DELETE /api/authors/{id}", s.deleteAuthor)
}

type authorRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Bio   string `json:"bio"`
}

func (s *Server) createAuthor(w http.ResponseWriter, r *http.Request) {
	var req authorRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.CreateAuthor(model.Author{Name: req.Name, Email: req.Email, Bio: req.Bio})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, a)
}

func (s *Server) listAuthors(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.AuthorFilter{Keyword: r.URL.Query().Get("keyword")}
	items, total, err := s.svc.ListAuthors(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getAuthor(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetAuthor(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) updateAuthor(w http.ResponseWriter, r *http.Request) {
	var req authorRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.UpdateAuthor(r.PathValue("id"), model.Author{Name: req.Name, Email: req.Email, Bio: req.Bio})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) deleteAuthor(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteAuthor(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
