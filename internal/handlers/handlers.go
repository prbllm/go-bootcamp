package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"entrytest/internal/config"
)

func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	writeResponse(w, http.StatusOK, []byte(config.HealthOK))
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
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

func writeResponse(w http.ResponseWriter, status int, body []byte) {
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		log.Printf("write response: %v", err)
	}
}
