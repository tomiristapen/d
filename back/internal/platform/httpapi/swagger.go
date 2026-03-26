package httpapi

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed swagger_assets/*
var swaggerAssets embed.FS

func swaggerUIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := swaggerAssets.ReadFile("swagger_assets/swagger.html")
		if err != nil {
			http.Error(w, "swagger ui not available", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	}
}

func swaggerSpecHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := swaggerAssets.ReadFile("swagger_assets/openapi.yaml")
		if err != nil {
			http.Error(w, "openapi spec not available", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(data)
	}
}

func swaggerAssetsFS() fs.FS {
	sub, err := fs.Sub(swaggerAssets, "swagger_assets")
	if err != nil {
		return nil
	}
	return sub
}
