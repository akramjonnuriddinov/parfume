package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"parfume/backend/internal/assistant"
	"parfume/backend/internal/store"
)

func (h *Handler) chatStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": h.assistant != nil})
}

func (h *Handler) chat(w http.ResponseWriter, r *http.Request) {
	var in assistant.Request
	if !decodeJSON(w, r, &in) {
		return
	}
	in.Locale = normalizeLocale(in.Locale)
	in.Message = clean(in.Message)

	if !lengthBetween(in.Message, 1, 2000) {
		writeError(w, http.StatusUnprocessableEntity, "message must be 1-2000 characters")
		return
	}
	if len(in.Cart) > 100 {
		in.Cart = in.Cart[:100]
	}

	reply, err := h.assistant.Chat(r.Context(), in)
	switch {
	case errors.Is(err, assistant.ErrDisabled):
		writeError(w, http.StatusServiceUnavailable, "assistant is disabled")
	case err != nil:
		log.Printf("assistant: %v", err)
		writeError(w, http.StatusBadGateway, "assistant is unavailable")
	default:
		writeJSON(w, http.StatusOK, reply)
	}
}

// chatConfirmOrder places the order the assistant proposed, after the
// customer pressed "Confirm" on the summary card.
func (h *Handler) chatConfirmOrder(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SessionID string `json:"session_id"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	order, err := h.assistant.ConfirmOrder(r.Context(), in.SessionID)
	switch {
	case errors.Is(err, assistant.ErrDisabled):
		writeError(w, http.StatusServiceUnavailable, "assistant is disabled")
	case errors.Is(err, assistant.ErrNoPendingOrder):
		writeError(w, http.StatusConflict, err.Error())
	case err != nil:
		writeStoreError(w, err)
	default:
		writeJSON(w, http.StatusCreated, order)
	}
}

// --- Admin ---

func (h *Handler) adminListChats(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := h.store.ListChatSessions(r.Context(), store.ChatFilter{
		Outcome: q.Get("outcome"),
		Query:   clean(q.Get("q")),
		Page:    queryInt(r, "page", 1),
		Limit:   queryInt(r, "limit", 20),
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *Handler) adminGetChat(w http.ResponseWriter, r *http.Request) {
	chat, err := h.store.GetChatSessionDetail(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chat)
}

func (h *Handler) adminDeleteChat(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DeleteChatSession(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) adminMissingProducts(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListMissingProducts(r.Context(), queryInt(r, "days", 0))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) adminDeleteMissingProduct(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := h.store.DeleteMissingProduct(r.Context(), in.Name); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
