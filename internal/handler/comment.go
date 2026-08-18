package handler

import (
	"net/http"

	"wiki/internal/model"
	"wiki/pkg/httpx"
)

func (s *Server) registerCommentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/comments", s.createComment)
	mux.HandleFunc("GET /api/comments", s.listComments)
	mux.HandleFunc("GET /api/comments/{id}", s.getComment)
	mux.HandleFunc("DELETE /api/comments/{id}", s.deleteComment)
}

type commentRequest struct {
	DocumentID string `json:"document_id"`
	AuthorID   string `json:"author_id"`
	Content    string `json:"content"`
}

func (s *Server) createComment(w http.ResponseWriter, r *http.Request) {
	var req commentRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	c, err := s.svc.CreateComment(model.Comment{
		DocumentID: req.DocumentID, AuthorID: req.AuthorID, Content: req.Content,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, c)
}

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.CommentFilter{
		DocumentID: r.URL.Query().Get("document_id"),
		AuthorID:   r.URL.Query().Get("author_id"),
	}
	items, total, err := s.svc.ListComments(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getComment(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.GetComment(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

func (s *Server) deleteComment(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComment(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
