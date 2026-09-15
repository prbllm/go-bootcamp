package handlers

import (
	"cmp"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"slices"

	"entrytest/internal/config"
	"entrytest/internal/storage"
)

type Handlers struct {
	store storage.Storage
}

func New(store storage.Storage) *Handlers {
	return &Handlers{store: store}
}

func (h *Handlers) HealthHandler(w http.ResponseWriter, _ *http.Request) {
	writeResponse(w, http.StatusOK, []byte(config.HealthOK))
}

func (h *Handlers) EchoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		log.Printf("read body: %v", err)
		writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))

		return
	}

	if r.Header.Get(config.ContentTypeHeader) == config.ContentTypeJSON {
		w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)

		var data map[string]any

		if err = json.Unmarshal(body, &data); err != nil {
			writeResponse(w, http.StatusBadRequest, []byte(err.Error()))

			return
		}

		if err = json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("encode response: %v", err)
		}

		return
	}

	writeResponse(w, http.StatusOK, body)
}

func (h *Handlers) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(config.ContentTypeHeader) != config.ContentTypeJSON {
		writeResponse(w, http.StatusBadRequest, []byte("unsupported content type: "+r.Header.Get(config.ContentTypeHeader)))

		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		log.Printf("read body: %v", err)
		writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))

		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)

	var data storage.Message

	if err = json.Unmarshal(body, &data); err != nil {
		writeResponse(w, http.StatusBadRequest, []byte(err.Error()))

		return
	}

	if data.Message == "" {
		writeResponse(w, http.StatusBadRequest, []byte("message is required"))

		return
	}

	message, err := h.store.Save(&data)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))

		return
	}

	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(message); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func (h *Handlers) MessagesListHandler(w http.ResponseWriter, _ *http.Request) {
	messages, err := h.store.GetAll()
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))

		return
	}

	slices.SortFunc(messages, func(a, b *storage.Message) int {
		return cmp.Compare(b.ID, a.ID)
	})

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)

	if err = json.NewEncoder(w).Encode(messages); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeResponse(w http.ResponseWriter, status int, body []byte) {
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil { //nolint:gosec // G705: intentional non-HTML response body
		log.Printf("write response: %v", err)
	}
}
