package handler

import (
	"net/http"

	"wiki/internal/model"
	"wiki/pkg/httpx"
)

func (s *Server) registerDocumentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/documents", s.createDocument)
	mux.HandleFunc("GET /api/documents", s.listDocuments)
	mux.HandleFunc("GET /api/documents/{id}", s.getDocument)
	mux.HandleFunc("PUT /api/documents/{id}", s.updateDocument)
	mux.HandleFunc("DELETE /api/documents/{id}", s.deleteDocument)
	mux.HandleFunc("POST /api/documents/{id}/publish", s.publishDocument)
	mux.HandleFunc("POST /api/documents/{id}/archive", s.archiveDocument)
	mux.HandleFunc("GET /api/documents/{id}/versions", s.listDocumentVersions)
	mux.HandleFunc("POST /api/documents/{id}/rollback", s.rollbackDocument)
	mux.HandleFunc("GET /api/search/documents", s.searchDocuments)
	mux.HandleFunc("POST /api/documents/batch-archive", s.batchArchive)
	mux.HandleFunc("POST /api/documents/batch-delete", s.batchDelete)
}

type documentRequest struct {
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	DirectoryID string   `json:"directory_id"`
	AuthorID    string   `json:"author_id"`
	Tags        []string `json:"tags"`
}

func (s *Server) createDocument(w http.ResponseWriter, r *http.Request) {
	var req documentRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	d, err := s.svc.CreateDocument(model.Document{
		Title: req.Title, Content: req.Content, DirectoryID: req.DirectoryID,
		AuthorID: req.AuthorID, Tags: req.Tags,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, d)
}

func (s *Server) listDocuments(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.DocumentFilter{
		DirectoryID: r.URL.Query().Get("directory_id"),
		AuthorID:    r.URL.Query().Get("author_id"),
		Status:      r.URL.Query().Get("status"),
		Tag:         r.URL.Query().Get("tag"),
		Keyword:     r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListDocuments(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getDocument(w http.ResponseWriter, r *http.Request) {
	d, err := s.svc.GetDocument(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, d)
}

func (s *Server) updateDocument(w http.ResponseWriter, r *http.Request) {
	var req documentRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	d, err := s.svc.UpdateDocument(r.PathValue("id"), model.Document{
		Title: req.Title, Content: req.Content, Tags: req.Tags,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, d)
}

func (s *Server) deleteDocument(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteDocument(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) publishDocument(w http.ResponseWriter, r *http.Request) {
	d, err := s.svc.PublishDocument(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, d)
}

func (s *Server) archiveDocument(w http.ResponseWriter, r *http.Request) {
	d, err := s.svc.ArchiveDocument(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, d)
}

func (s *Server) listDocumentVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := s.svc.DocumentVersions(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, versions)
}

type rollbackRequest struct {
	VersionID string `json:"version_id"`
}

func (s *Server) rollbackDocument(w http.ResponseWriter, r *http.Request) {
	var req rollbackRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	d, err := s.svc.RollbackDocument(r.PathValue("id"), req.VersionID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, d)
}

func (s *Server) searchDocuments(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	items, total, err := s.svc.SearchDocuments(r.URL.Query().Get("q"), pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

type batchIDsRequest struct {
	IDs []string `json:"ids"`
}

func (s *Server) batchArchive(w http.ResponseWriter, r *http.Request) {
	var req batchIDsRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	n, err := s.svc.BatchArchiveDocuments(req.IDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]int{"archived": n})
}

func (s *Server) batchDelete(w http.ResponseWriter, r *http.Request) {
	var req batchIDsRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	n, err := s.svc.BatchDeleteDocuments(req.IDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]int{"deleted": n})
}
