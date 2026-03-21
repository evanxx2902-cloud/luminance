package server

import (
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer creates a new HTTP server (scaffold only — no routes yet).
func NewHTTPServer() *kratoshttp.Server {
	return kratoshttp.NewServer(
		kratoshttp.Address(":8000"),
	)
}
