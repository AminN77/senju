package httpserver

import (
	"net/http"

	_ "embed"

	"github.com/AminN77/senju/backend/openapi"
	"github.com/gin-gonic/gin"
)

//go:embed swagger_ui.html
var swaggerUIPage []byte

// registerOpenAPISpecRoute exposes the embedded OpenAPI document (always available for tooling/clients).
func registerOpenAPISpecRoute(r *gin.Engine) {
	r.GET("/openapi.yaml", handleOpenAPISpec)
}

// registerSwaggerUIRoute exposes interactive docs; use only in non-release Gin mode.
func registerSwaggerUIRoute(r *gin.Engine) {
	r.GET("/docs", handleSwaggerUI)
}

func handleOpenAPISpec(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapi.SpecYAML)
}

func handleSwaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerUIPage)
}
