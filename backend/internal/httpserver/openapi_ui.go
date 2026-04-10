package httpserver

import (
	"net/http"

	_ "embed"

	"github.com/AminN77/senju/backend/openapi"
	"github.com/gin-gonic/gin"
)

//go:embed swagger_ui.html
var swaggerUIPage []byte

func registerOpenAPIRoutes(r *gin.Engine) {
	r.GET("/openapi.yaml", handleOpenAPISpec)
	r.GET("/docs", handleSwaggerUI)
}

func handleOpenAPISpec(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapi.SpecYAML)
}

func handleSwaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerUIPage)
}
