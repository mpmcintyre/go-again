package main

import (
	"html/template"
	"log"
	"net/http"

	reloader "github.com/MPMcIntyre/go-again"
	"github.com/gin-gonic/gin"
)

func main() {
	debug := true
	gin.SetMode(gin.DebugMode)
	app := gin.New()

	if debug {
		app.Use(gin.Logger())
	}
	app.Use(gin.Recovery())

	// Create reloader with logging enabled, ws reload port on 9000
	rel, err := reloader.New(
		func() { app.LoadHTMLGlob("templates/**/*") },
		9000,
		reloader.WithLogs(debug),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rel.Close()

	// Watch template directories, this will trigger a reload on change
	rel.Add("./templates/components")
	rel.Add("./templates/views")
	err = rel.Add("./static")

	if err != nil {
		log.Fatal(err)
	}

	// Register the LiveReload template function
	app.SetFuncMap(template.FuncMap{
		"LiveReload": rel.TemplateFunc()["LiveReload"],
	})

	app.Static("/static", "./static")

	count := 0
	app.GET("/", func(g *gin.Context) {
		count += 1
		g.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Go Again",
			"count": count,
		})
	})
	app.LoadHTMLGlob("templates/**/*")
	app.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}
