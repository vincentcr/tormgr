package main

import (
	"encoding/json"
	"net/http"

	"github.com/vincentcr/validator"
)

func parseAndValidate(r *http.Request, result interface{}) error {
	if err := parseBody(r, result); err != nil {
		return NewHttpError(http.StatusBadRequest)
	}

	if err := validator.Validate(result); err != nil {
		return NewHttpError(http.StatusBadRequest)
	}

	return nil
}

func parseBody(r *http.Request, result interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(result)
}

func jsonify(result interface{}, w http.ResponseWriter) error {
	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return writeAs(w, "application/json", bytes)
}

func writeCacheHeaders(r *http.Request, w http.ResponseWriter, etag ETag) bool {
	w.Header().Set("ETag", string(etag))
	w.Header().Set("Cache-Control", "public")
	reqEtag := ETag(r.Header.Get("If-None-Match"))
	if etag == reqEtag {
		w.WriteHeader(304)
		return true
	} else {
		return false
	}
}

func writeCacheable(r *http.Request, w http.ResponseWriter, contentType string, cacheable Cacheable) error {
	w.Header().Set("content-type", contentType)
	if writeCacheHeaders(r, w, cacheable.ETag) {
		return nil
	}
	return writeAs(w, contentType, cacheable.Bytes)
}

func writeAs(w http.ResponseWriter, contentType string, bytes []byte) error {
	w.Header().Set("content-type", contentType)
	_, err := w.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}
