package handler

import (
	"net/http"

	"wiki/internal/model"
	"wiki/pkg/httpx"
)

func (s *Server) registerFavoriteRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/favorites", s.createFavorite)
	mux.HandleFunc("GET /api/favorites", s.listFavorites)
	mux.HandleFunc("DELETE /api/favorites", s.deleteFavorite)
}

type favoriteRequest struct {
	DocumentID string `json:"document_id"`
	UserID     string `json:"user_id"`
}

func (s *Server) createFavorite(w http.ResponseWriter, r *http.Request) {
	var req favoriteRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	f, err := s.svc.Favorite(req.UserID, req.DocumentID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, f)
}

func (s *Server) listFavorites(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.FavoriteFilter{
		UserID:     r.URL.Query().Get("user_id"),
		DocumentID: r.URL.Query().Get("document_id"),
	}
	items, total, err := s.svc.ListFavorites(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) deleteFavorite(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	documentID := r.URL.Query().Get("document_id")
	if err := s.svc.Unfavorite(userID, documentID); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
