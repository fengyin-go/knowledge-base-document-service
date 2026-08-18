package handler

import (
	"net/http"

	"wiki/internal/model"
	"wiki/pkg/httpx"
)

func (s *Server) registerDirectoryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/directories", s.createDirectory)
	mux.HandleFunc("GET /api/directories", s.listDirectories)
	mux.HandleFunc("GET /api/directories/{id}", s.getDirectory)
	mux.HandleFunc("PUT /api/directories/{id}", s.updateDirectory)
	mux.HandleFunc("DELETE /api/directories/{id}", s.deleteDirectory)
	mux.HandleFunc("GET /api/directories/tree", s.directoryTree)
}

type directoryRequest struct {
	Name        string `json:"name"`
	ParentID    string `json:"parent_id"`
	Description string `json:"description"`
}

func (s *Server) createDirectory(w http.ResponseWriter, r *http.Request) {
	var req directoryRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	d, err := s.svc.CreateDirectory(model.Directory{Name: req.Name, ParentID: req.ParentID, Description: req.Description})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, d)
}

func (s *Server) listDirectories(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.DirectoryFilter{
		ParentID: r.URL.Query().Get("parent_id"),
		Keyword:  r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListDirectories(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getDirectory(w http.ResponseWriter, r *http.Request) {
	d, err := s.svc.GetDirectory(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, d)
}

func (s *Server) updateDirectory(w http.ResponseWriter, r *http.Request) {
	var req directoryRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	d, err := s.svc.UpdateDirectory(r.PathValue("id"), model.Directory{Name: req.Name, Description: req.Description})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, d)
}

func (s *Server) deleteDirectory(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteDirectory(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) directoryTree(w http.ResponseWriter, r *http.Request) {
	tree, err := s.svc.DirectoryTree()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, tree)
}
