package main

import (
	"log"
	"net/http"

	reloader "github.com/MPMcIntyre/go-again"
	"github.com/gin-gonic/gin"
)

func main() {
	app := gin.Default()

	count := 0

	// Add templates, this is the same function used to reload the assets
	app.LoadHTMLGlob("templates/**/*")

	app.GET("/", func(g *gin.Context) {
		count += 1
		g.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Go Again",
			"count": count,
		})
	})

	// Add hot reloading, listen on port 9000, enable logging
	rel, err := reloader.New(func() { app.LoadHTMLGlob("templates/**/*") }, 9000, true)
	if err != nil {
		log.Fatal(err)
	}
	defer rel.Close()

	// Listen for changes on the components
	err = rel.Add("templates/components")
	if err != nil {
		log.Fatal(err)
	}
	// Listen for changes on the views
	err = rel.Add("templates/views")
	if err != nil {
		log.Fatal(err)
	}

	app.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}
