package reloader

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

// Private variables
var upgrader = websocket.Upgrader{} // use default options

type Reloader struct {
	watcher        *fsnotify.Watcher
	logs_enabled   bool
	ws_connections []*websocket.Conn
	reloadScript   string
}

// Private functions
func createReloadScript(port int) string {
	return fmt.Sprintf(`
    <script>
    (function() {
        let ws = new WebSocket("ws://localhost:%d");
        ws.onopen = () => console.log("WebSocket connection opened.");
        ws.onmessage = (evt) => {
            console.log("Reload triggered:", evt.data);
            location.reload();
        };
        ws.onclose = () => console.log("WebSocket connection closed.");
        ws.onerror = (evt) => console.error("WebSocket error:", evt);
    })();
    </script>
    `, port)
}

// Private functions
// Logs the desired statement depending if logging is enabled or not
func (r *Reloader) log(v ...any) {
	if r.logs_enabled {
		log.Println(v...)
	}
}

// Handler for connecting to websockets
func (rel *Reloader) wshome(w http.ResponseWriter, r *http.Request) {

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	rel.ws_connections = append(rel.ws_connections, c)

	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

// Public functions

// Add path to folder or file to be watched
func (r *Reloader) Add(path string) error {
	return r.watcher.Add(path)
}

// Defer this in order to ensure all resources are closed
func (r *Reloader) Close() {
	// Close watcher
	r.watcher.Close()

	// Close ws connections
	if len(r.ws_connections) > 0 {
		for _, conn := range r.ws_connections {
			conn.Close()
		}
	}
}

// Public interfaces
// Inject the reload script for users
func (r *Reloader) GetReloadScript() template.HTML {
	return template.HTML(r.reloadScript)
}

func (r *Reloader) AddReloadToTemplates(t *template.Template) *template.Template {
	return t.Funcs(template.FuncMap{
		"go_again_reload": func() template.HTML {
			return r.GetReloadScript()
		},
	})
}

func New(cb func(), wsport int, logs_enabled bool) (*Reloader, error) {

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	reloadScript := createReloadScript(wsport)

	if err != nil {
		return nil, err
	}

	r := &Reloader{
		watcher:        watcher,
		logs_enabled:   logs_enabled,
		ws_connections: make([]*websocket.Conn, 0),
		reloadScript:   reloadScript,
	}

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				r.log("event:", event)
				r.log("modified file:", event.Name)
				cb()
				if len(r.ws_connections) > 0 {
					for _, conn := range r.ws_connections {
						conn.WriteMessage(websocket.TextMessage, []byte(event.Name))
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				r.log("error:", err)
			}
		}
	}()

	// // Set up websocket to reload webpage
	// var addr = flag.String("addr", fmt.Sprintf("%s%d", "localhost:", wsport), "http service address")

	// http.HandleFunc("/", r.wshome)
	// go http.ListenAndServe(*addr, nil)

	// // Add template

	// return r, err

	// Set up WebSocket handler
	addr := fmt.Sprintf(":%d", wsport)
	http.HandleFunc("/", r.wshome)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("WebSocket server failed: %v", err)
		}
	}()

	return r, nil
}
