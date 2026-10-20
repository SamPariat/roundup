package http

import (
	"github.com/SamPariat/roundup/shared/api"
	"github.com/gofiber/fiber/v3"
)

// swaggerUIVersion pins the Swagger UI build loaded from the CDN.
const swaggerUIVersion = "5.17.14"

// docsPage loads Swagger UI from jsDelivr and points it at the spec route.
const docsPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Roundup API</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({ url: "/docs/openapi.yaml", dom_id: "#swagger-ui", persistAuthorization: true });
  </script>
</body>
</html>`

// RegisterDocs serves Swagger UI at /docs and the raw spec at
// /docs/openapi.yaml. Mounted only when config.DocsEnabled is set.
func RegisterDocs(app *fiber.App) {
	app.Get("/docs", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)

		return c.SendString(docsPage)
	})

	app.Get("/docs/openapi.yaml", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "application/yaml")

		return c.Send(api.Spec)
	})
}
