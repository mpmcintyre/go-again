# Go-Again 🔄

Go-Again is a lightweight live-reload solution for Go web applications. It automatically refreshes your browser when template files change, making development faster and more efficient. Unlike server hot-reloading tools (like Air) that require restarting the entire Go server, Go-Again only refreshes the browser while keeping your server running, allowing for rapid template development without server restarts.

**Note: Go-Again is a live-reload (full page refresh) tool, not a hot-module-replacement solution. This means it's perfect for template development but will reset client-side application state on changes. The server-side state is preserved as no server restart is required.**

Quick Reference:

- Live-Reload (Go-Again): Refreshes the browser page when templates change, server keeps running
- Hot-Module-Replacement: Updates specific components without page refresh
- Server Hot-Reload (Air): Rebuilds and restarts entire Go server when code changes

## Features

- 🔥 Hot reloading of HTML templates
- 🎯 Minimal setup required
- 🌐 WebSocket-based browser refresh
- 📁 Multiple directory watching
- 🛠️ Framework agnostic (with growing compatibility list)
- 🪶 Lightweight with minimal dependencies
-

## Contents

- [Go-Again 🔄](#go-again-)
  - [Features](#features)
  - [Contents](#contents)
  - [Getting Started](#getting-started)
  - [Complete Example](#complete-example)
  - [Framework Compatibility](#framework-compatibility)
  - [Configuration Options](#configuration-options)
    - [WithLogs](#withlogs)
  - [Contributing](#contributing)
  - [License](#license)
  - [Support](#support)

## Getting Started

To install the module, run the following command:

```bash
go get github.com/MPMcIntyre/go-again
```

Then import the package in your main routing component:

```go
import reloader "github.com/MPMcIntyre/go-again"
```

Next, initialize the reloader:

```go
rel, err := reloader.New(
    func() { app.LoadHTMLGlob("templates/**/*") },
    9000,
    reloader.WithLogs(true),
)
if err != nil {
    log.Fatal(err)
}
defer rel.Close()
```

Now that the reloader is set up, you can add directories to watch:

```go
rel.Add("templates/components")
rel.Add("templates/views")
```

**Note: If you are using a tool like [air](https://github.com/air-verse/air) you must ensure that the templates directories are ignored in the .air.toml configuration files.**

Then add the LiveReload function to your templates:

```go
app.SetFuncMap(template.FuncMap{
    "LiveReload": rel.TemplateFunc()["LiveReload"],
})
```

Finally, include the LiveReload tag in your base template:

```html
{{ LiveReload }}
```

## Complete Example

```go
package main

import (
    "html/template"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    reloader "github.com/MPMcIntyre/go-again"
)

func main() {
    app := gin.Default()

    // Initialize reloader
    rel, err := reloader.New(
        func() { app.LoadHTMLGlob("templates/**/*") },
        9000,
        reloader.WithLogs(true),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer rel.Close()

    // Watch template directories
    rel.Add("templates/components")
    rel.Add("templates/views")

    // Register LiveReload function
    app.SetFuncMap(template.FuncMap{
        "LiveReload": rel.TemplateFunc()["LiveReload"],
    })

    // Load templates
    app.LoadHTMLGlob("templates/**/*")

    // Routes
    app.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", gin.H{
            "title": "Go Again Demo",
        })
    })

    app.Run(":8080")
}
```

## Framework Compatibility

| Framework   | Compatibility |
| ----------- | ------------- |
| Gin         | ✅            |
| Echo        | ❓            |
| Fiber       | ❓            |
| Chi         | ❓            |
| Buffalo     | ❓            |
| Beego       | ❓            |
| Gorilla Mux | ❓            |
| Iris        | ❓            |
| Revel       | ❓            |
| Martini     | ❓            |

✅ - Compatible
❓ - Not tested yet
❌ - Not compatible

## Configuration Options

### WithLogs

Enable or disable logging:

```go
reloader.New(callback, 9000, reloader.WithLogs(true))
```

## Contributing

Contributions are welcome! Feel free to:

1. Fork the repository
2. Create a feature branch
3. Submit a Pull Request

Please ensure you test your changes and update the framework compatibility table if you've verified compatibility with additional frameworks.

## License

MIT License - see LICENSE file for details.

## Support

If you encounter any issues or have questions, please file an issue on GitHub.
