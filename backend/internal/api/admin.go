package api

import (
	"errors"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"parfume/backend/internal/auth"
	"parfume/backend/internal/store"
)

// dummyHash keeps login timing similar for unknown usernames.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}

	admin, err := h.store.AdminByUsername(r.Context(), clean(in.Username))
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		writeStoreError(w, err)
		return
	}
	hash := dummyHash
	if err == nil {
		hash = []byte(admin.PasswordHash)
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(in.Password)) != nil || err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, expires, err := h.auth.Issue(admin.ID, admin.Username)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": expires,
		"admin":      admin,
	})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	id, err := auth.AdminID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	admin, err := h.store.AdminByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, admin)
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	id, err := auth.AdminID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	admin, err := h.store.AdminByID(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(in.CurrentPassword)) != nil {
		writeError(w, http.StatusUnprocessableEntity, "current password is incorrect")
		return
	}
	if !lengthBetween(in.NewPassword, 6, 72) {
		writeError(w, http.StatusUnprocessableEntity, "new password must be 6-72 characters")
		return
	}
	if err := h.store.UpdateAdminPassword(r.Context(), id, in.NewPassword); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	st, err := h.store.GetStats(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// --- Categories ---

func (h *Handler) adminListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListCategories(r.Context(), false)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func validateCategory(in *store.CategoryInput) string {
	in.Slug = clean(in.Slug)
	in.NameUz, in.NameRu, in.NameEn = clean(in.NameUz), clean(in.NameRu), clean(in.NameEn)
	in.DescriptionUz, in.DescriptionRu, in.DescriptionEn = clean(in.DescriptionUz), clean(in.DescriptionRu), clean(in.DescriptionEn)
	switch {
	case !slugRe.MatchString(in.Slug) || len(in.Slug) > 80:
		return "slug must contain only lowercase latin letters, digits and dashes"
	case !lengthBetween(in.NameUz, 1, 120):
		return "name_uz is required"
	}
	return ""
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var in store.CategoryInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if msg := validateCategory(&in); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}
	c, err := h.store.CreateCategory(r.Context(), in)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	var in store.CategoryInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if msg := validateCategory(&in); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}
	c, err := h.store.UpdateCategory(r.Context(), id, in)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteCategory(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Products ---

func (h *Handler) adminListProducts(w http.ResponseWriter, r *http.Request) {
	f := productFilterFromQuery(r)
	if s := r.URL.Query().Get("sort"); s == "" {
		f.Sort = "newest"
	}
	page, err := h.store.ListProducts(r.Context(), f)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *Handler) adminGetProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	p, err := h.store.GetProduct(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func validateProduct(in *store.ProductInput) string {
	in.NameUz, in.NameRu, in.NameEn = clean(in.NameUz), clean(in.NameRu), clean(in.NameEn)
	in.DescriptionUz, in.DescriptionRu, in.DescriptionEn = clean(in.DescriptionUz), clean(in.DescriptionRu), clean(in.DescriptionEn)
	in.Brand, in.ImageURL = clean(in.Brand), clean(in.ImageURL)
	in.NotesTop, in.NotesHeart, in.NotesBase = clean(in.NotesTop), clean(in.NotesHeart), clean(in.NotesBase)
	if in.CategoryID != nil && *in.CategoryID <= 0 {
		in.CategoryID = nil
	}
	switch {
	case !lengthBetween(in.NameUz, 1, 200):
		return "name_uz is required"
	case !lengthBetween(in.NameRu, 0, 200) || !lengthBetween(in.NameEn, 0, 200):
		return "name is too long"
	case !lengthBetween(in.Brand, 0, 120):
		return "brand is too long"
	case !store.ValidGender(in.Gender):
		return "invalid gender"
	case !store.ValidConcentration(in.Concentration):
		return "invalid concentration"
	case !lengthBetween(in.NotesTop, 0, 500) || !lengthBetween(in.NotesHeart, 0, 500) || !lengthBetween(in.NotesBase, 0, 500):
		return "notes are too long"
	case len(in.ImageURL) > 1000:
		return "image_url is too long"
	case len(in.Variants) == 0 || len(in.Variants) > 20:
		return "add 1-20 volumes"
	}
	for _, v := range in.Variants {
		switch {
		case v.VolumeML < 1 || v.VolumeML > 1000:
			return "volume must be 1-1000 ml"
		case !store.ValidVariantKind(v.Kind):
			return "invalid volume kind"
		case v.Price < 0 || v.Price > 1_000_000_000_000:
			return "invalid price"
		}
	}
	return ""
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var in store.ProductInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if msg := validateProduct(&in); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}
	p, err := h.store.CreateProduct(r.Context(), in)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	var in store.ProductInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if msg := validateProduct(&in); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}
	p, err := h.store.UpdateProduct(r.Context(), id, in)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteProduct(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Orders ---

func (h *Handler) adminListOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := h.store.ListOrders(r.Context(), store.OrderFilter{
		Status: q.Get("status"),
		Query:  q.Get("q"),
		Page:   queryInt(r, "page", 1),
		Limit:  queryInt(r, "limit", 20),
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *Handler) adminGetOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	o, err := h.store.GetOrder(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (h *Handler) adminUpdateOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	var in struct {
		Status    string `json:"status"`
		AdminNote string `json:"admin_note"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if !store.ValidOrderStatus(in.Status) {
		writeError(w, http.StatusUnprocessableEntity, "invalid status")
		return
	}
	if !lengthBetween(in.AdminNote, 0, 2000) {
		writeError(w, http.StatusUnprocessableEntity, "admin_note is too long")
		return
	}
	o, err := h.store.UpdateOrder(r.Context(), id, in.Status, clean(in.AdminNote))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (h *Handler) adminDeleteOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteOrder(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Messages ---

func (h *Handler) adminListMessages(w http.ResponseWriter, r *http.Request) {
	unread, _ := strconv.ParseBool(r.URL.Query().Get("unread"))
	page, err := h.store.ListMessages(r.Context(), unread, queryInt(r, "page", 1), queryInt(r, "limit", 20))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *Handler) adminUpdateMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	var in struct {
		IsRead bool `json:"is_read"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := h.store.SetMessageRead(r.Context(), id, in.IsRead); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) adminDeleteMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	if err := h.store.DeleteMessage(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Settings ---

func (h *Handler) adminSaveSettings(w http.ResponseWriter, r *http.Request) {
	var in map[string]string
	if !decodeJSON(w, r, &in) {
		return
	}
	for k, v := range in {
		in[k] = clean(v)
		if !lengthBetween(in[k], 0, 500) {
			writeError(w, http.StatusUnprocessableEntity, k+" is too long")
			return
		}
	}
	if err := h.store.SaveSettings(r.Context(), in); err != nil {
		writeStoreError(w, err)
		return
	}
	h.getSettings(w, r)
}
