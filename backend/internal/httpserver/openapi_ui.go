package httpserver

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/AminN77/senju/backend/openapi"
	"github.com/gin-gonic/gin"
)

//go:embed swagger_ui.html
var swaggerUIPage []byte

//go:embed swagger_ui_static/swagger-ui.css swagger_ui_static/swagger-ui-bundle.js swagger_ui_static/swagger-ui-standalone-preset.js
var swaggerUIAssets embed.FS

// registerOpenAPISpecRoute exposes the embedded OpenAPI document (always available for tooling/clients).
func registerOpenAPISpecRoute(r *gin.Engine) {
	r.GET("/openapi.yaml", handleOpenAPISpec)
}

// registerSwaggerUIRoute exposes interactive docs; use only in non-release Gin mode.
// Swagger UI JS/CSS are embedded (swagger-ui-dist) so /docs works offline and avoids CDN supply-chain risk.
func registerSwaggerUIRoute(r *gin.Engine) {
	sub, err := fs.Sub(swaggerUIAssets, "swagger_ui_static")
	if err != nil {
		panic("httpserver: swagger UI static assets: " + err.Error())
	}
	r.StaticFS("/docs/swagger-ui", http.FS(sub))
	r.GET("/docs", handleSwaggerUI)
}

func handleOpenAPISpec(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapi.SpecYAML)
}

func handleSwaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerUIPage)
}
