package api

import (
	"net/http"
	"strconv"
	"strings"

	"parfume/backend/internal/store"
)

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListCategories(r.Context(), true)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func productFilterFromQuery(r *http.Request) store.ProductFilter {
	q := r.URL.Query()
	f := store.ProductFilter{
		CategorySlug: clean(q.Get("category")),
		Gender:       q.Get("gender"),
		Brand:        clean(q.Get("brand")),
		Tier:         q.Get("tier"),
		Query:        clean(q.Get("q")),
		Featured:     q.Get("featured") == "1" || q.Get("featured") == "true",
		Sort:         q.Get("sort"),
		Page:         queryInt(r, "page", 1),
		Limit:        queryInt(r, "limit", 12),
	}
	if id, err := strconv.ParseInt(q.Get("category_id"), 10, 64); err == nil {
		f.CategoryID = id
	}
	if raw := q.Get("ids"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			if id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil && id > 0 {
				f.IDs = append(f.IDs, id)
			}
		}
	}
	return f
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	f := productFilterFromQuery(r)
	f.ActiveOnly = true
	page, err := h.store.ListProducts(r.Context(), f)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	p, err := h.store.GetProduct(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !p.IsActive {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var in store.OrderInput
	if !decodeJSON(w, r, &in) {
		return
	}

	in.FullName = clean(in.FullName)
	in.Comment = clean(in.Comment)
	in.Locale = normalizeLocale(in.Locale)
	if len(in.ChatSessionID) > 64 {
		in.ChatSessionID = ""
	}
	phone, phoneOK := normalizePhone(in.Phone)
	in.Phone = phone

	switch {
	case !lengthBetween(in.FullName, 2, 120):
		writeError(w, http.StatusUnprocessableEntity, "full_name must be 2-120 characters")
		return
	case !phoneOK:
		writeError(w, http.StatusUnprocessableEntity, "invalid phone number")
		return
	case !lengthBetween(in.Comment, 0, 1000):
		writeError(w, http.StatusUnprocessableEntity, "comment is too long")
		return
	case len(in.Items) == 0 || len(in.Items) > 50:
		writeError(w, http.StatusUnprocessableEntity, "order must contain 1-50 items")
		return
	}
	for _, it := range in.Items {
		if it.VariantID <= 0 || it.Quantity < 1 || it.Quantity > 99 {
			writeError(w, http.StatusUnprocessableEntity, "invalid order item")
			return
		}
	}

	order, err := h.store.CreateOrder(r.Context(), in)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":         order.ID,
		"total":      order.Total,
		"created_at": order.CreatedAt,
	})
}

func (h *Handler) createMessage(w http.ResponseWriter, r *http.Request) {
	var in store.MessageInput
	if !decodeJSON(w, r, &in) {
		return
	}
	in.FullName = clean(in.FullName)
	in.Message = clean(in.Message)
	phone, phoneOK := normalizePhone(in.Phone)
	in.Phone = phone

	switch {
	case !lengthBetween(in.FullName, 2, 120):
		writeError(w, http.StatusUnprocessableEntity, "full_name must be 2-120 characters")
		return
	case !phoneOK:
		writeError(w, http.StatusUnprocessableEntity, "invalid phone number")
		return
	case !lengthBetween(in.Message, 0, 2000):
		writeError(w, http.StatusUnprocessableEntity, "message is too long")
		return
	}

	m, err := h.store.CreateMessage(r.Context(), in)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": m.ID})
}

func (h *Handler) listBrands(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListBrands(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	s, err := h.store.GetSettings(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}
