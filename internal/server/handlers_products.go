package server

import (
	"database/sql"
	"errors"
	"net/http"

	"shopadmin/internal/store"
)

// GET /api/products?page=1&page_size=20
func (s *Server) handleListProducts(w http.ResponseWriter, r *http.Request) {
	page, size, err := parsePage(r.URL.Query())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	items, err := s.q.ListProducts(ctx, store.ListProductsParams{
		Limit:  size,
		Offset: (page - 1) * size,
	})
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	total, err := s.q.CountProducts(ctx)
	if err != nil {
		s.serverError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, Page[store.ListProductsRow]{
		Items: items, Total: total, Page: page, PageSize: size,
	})
}

// GET /api/products/{id}
func (s *Server) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	p, err := s.q.GetProduct(r.Context(), id)
	// errors.Is는 감싸진(wrapped) 에러 체인까지 따라가며 비교한다.
	// :one 쿼리에서 행이 없으면 sql.ErrNoRows가 반환된다.
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}
