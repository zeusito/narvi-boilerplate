package router

import (
	"backend-api/pkg/terrors"
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// RenderJSON is a helper function to write a JSON response
func RenderJSON(ctx context.Context, w http.ResponseWriter, httpStatusCode int, payload any) {
	// Headers
	w.Header().Set(middleware.RequestIDHeader, middleware.GetReqID(ctx))
	w.Header().Set("Content-Type", "application/json")

	js, err := json.Marshal(payload)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(httpStatusCode)
	_, _ = w.Write(js)
}

// RenderError Renders an error with some sane defaults.
func RenderError(ctx context.Context, w http.ResponseWriter, err error) {
	var terrorToRender *terrors.Terror

	if terr, ok := err.(*terrors.Terror); ok {
		terrorToRender = terr
	} else {
		terrorToRender = terrors.Unknown(err.Error())
	}

	RenderJSON(ctx, w, terrorToRender.HttpStatusCode, terrorToRender)
}

func DefaultSuccessResponseBody() map[string]string {
	return map[string]string{
		"code":    "Success",
		"message": "action completed successfully",
	}
}

// GetRequestID retrieves the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	return middleware.GetReqID(ctx)
}
