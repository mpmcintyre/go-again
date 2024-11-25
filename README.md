# Go-Again 🔄

Go-Again is a lightweight live-update solution for Go web applications. It automatically updates your browser content when template files change, making development faster and more efficient. Unlike server hot-reloading tools (like Air) that require restarting the entire Go server, Go-Again only updates the affected content while keeping your server running and preserving client-side state.

Perfect for rapid development of:

- Template modifications
- CSS styling changes
- Content updates
- Layout adjustments

## Features

- 🔥 Hot reloading of HTML templates, css stylesheets, and static files
- 🎯 Minimal setup required
- 🌐 WebSocket-based DOM element replacement
- 📁 Multiple directory watching
- 🛠️ Framework agnostic (with growing compatibility list)
- 🪶 Lightweight with minimal dependencies
- 💾 State Preservation: Maintains client-side state (form inputs, counters, etc.) during updates

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

You can also tell the HMR script to ignore client-side state values to preserve the view of a value using the `data-client-state` tag:

```html
<span data-client-state data-bind="clientCount">0</span>
```

## Complete Example

Checkout the full examles in the [example](./examples/) directory

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
    rel.Add("static")

    // Register LiveReload function
    app.SetFuncMap(template.FuncMap{
        "LiveReload": rel.TemplateFunc()["LiveReload"],
    })

    // Load templates
    app.LoadHTMLGlob("templates/**/*")

    // Routes
    count := 0
    app.GET("/", func(g *gin.Context) {
        count += 1
        g.HTML(http.StatusOK, "index.html", gin.H{
            "title": "Go Again",
            "count": count,
        })
    })

    app.Run(":8080")
}
```

With your template as:

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    {{LiveReload}}
    <title>{{ .title }}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta charset="UTF-8" />
    <script>
      var clientCount = 0;
      // Function to update all elements that display the counter
      function updateCounterDisplays() {
        document
          .querySelectorAll('[data-bind="clientCount"]')
          .forEach((element) => {
            element.textContent = clientCount;
          });
      }

      // Function to increment counter
      function incrementCounter() {
        console.log("Increment");
        clientCount++;
        updateCounterDisplays();
      }

      // Initialize displays when page loads
      document.addEventListener("DOMContentLoaded", updateCounterDisplays);
    </script>
  </head>
  <body>
    <div>
      <div>Server side count: {{.count}}</div>
      <div>
        Client counter value:
        <span data-client-state data-bind="clientCount">0</span>
      </div>
      <button onclick="incrementCounter()">Increment Counter</button>
    </div>
  </body>
</html>
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
