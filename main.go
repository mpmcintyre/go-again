package reloader

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type Reloader struct {
	watcher       *fsnotify.Watcher
	logsEnabled   bool
	wsConnections []*websocket.Conn
	wsPort        int
	reloadScript  template.HTML
}

// ReloaderOption allows for functional options pattern
type ReloaderOption func(*Reloader)

// WithLogs enables or disables logging
func WithLogs(enabled bool) ReloaderOption {
	return func(r *Reloader) {
		r.logsEnabled = enabled
	}
}

func (r *Reloader) log(v ...any) {
	if r.logsEnabled {
		log.Println(v...)
	}
}

// Add path to folder or file to be watched
func (r *Reloader) Add(path string) error {
	return r.watcher.Add(path)
}

// Close ensures all resources are properly cleaned up
func (r *Reloader) Close() {
	if r.watcher != nil {
		r.watcher.Close()
	}
	for _, conn := range r.wsConnections {
		conn.Close()
	}
}

// createReloadScript generates the WebSocket client code
func createReloadScript(port int) template.HTML {
	script := fmt.Sprintf(`
		<script>
		(function() {
			let ws = null;
			function connect() {
				if (ws) {
					return;
				}
				ws = new WebSocket("ws://localhost:%d/ws");
				ws.onopen = function() {
					console.log("[Go-Again] Connected to reload server");
				};
				ws.onclose = function() {
					console.log("[Go-Again] Disconnected from reload server");
					ws = null;
					setTimeout(connect, 1000);
				};
				ws.onmessage = function(evt) {
					console.log("[Go-Again] Reloading page due to: " + evt.data);
					window.location.reload();
				};
			}
			connect();
		})();
		</script>
	`, port)
	return template.HTML(script)
}

// wsHandler handles WebSocket connections
func (r *Reloader) wsHandler(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		r.log("WebSocket upgrade error:", err)
		return
	}

	r.wsConnections = append(r.wsConnections, conn)
	r.log("New WebSocket connection established")

	// Keep the connection alive and remove it when closed
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				r.log("WebSocket read error:", err)
				r.removeConnection(conn)
				return
			}
		}
	}()
}

func (r *Reloader) removeConnection(conn *websocket.Conn) {
	for i, c := range r.wsConnections {
		if c == conn {
			r.wsConnections = append(r.wsConnections[:i], r.wsConnections[i+1:]...)
			break
		}
	}
}

// TemplateFunc returns a template function that injects the reload script
func (r *Reloader) TemplateFunc() template.FuncMap {
	return template.FuncMap{
		"LiveReload": func() template.HTML {
			return r.reloadScript
		},
	}
}

// New creates a new Reloader instance
func New(callback func(), wsPort int, opts ...ReloaderOption) (*Reloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	r := &Reloader{
		watcher:       watcher,
		wsPort:        wsPort,
		wsConnections: make([]*websocket.Conn, 0),
		reloadScript:  createReloadScript(wsPort),
	}

	// Apply options
	for _, opt := range opts {
		opt(r)
	}

	// Start file watcher
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if strings.HasSuffix(event.Name, ".html") || strings.HasSuffix(event.Name, ".tmpl") {
					r.log("Template modified:", event.Name)
					callback()
					r.notifyClients(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				r.log("Watcher error:", err)
			}
		}
	}()

	// Start WebSocket server
	http.HandleFunc("/ws", r.wsHandler)
	go func() {
		addr := fmt.Sprintf(":%d", wsPort)
		if err := http.ListenAndServe(addr, nil); err != nil {
			r.log("WebSocket server error:", err)
		}
	}()

	return r, nil
}

func (r *Reloader) notifyClients(filename string) {
	for _, conn := range r.wsConnections {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(filename)); err != nil {
			r.log("Error notifying client:", err)
			r.removeConnection(conn)
		}
	}
}
