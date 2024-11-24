package reloader

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

type Reloader struct {
	watcher        *fsnotify.Watcher
	logs_enabled   bool
	ws_connections []*websocket.Conn
	template       *template.Template
}

// Private functions

func (r *Reloader) log(v ...any) {
	if r.logs_enabled {
		log.Println(v...)
	}
}

// Public functions

// Add path to folder or file to be watched
func (r *Reloader) Add(path string) error {
	return r.watcher.Add(path)
}

// The template required to allow live reloading for HTML templates
func create_template(port int) string {
	return fmt.Sprintf(`
	<script>
	(){
		if (ws) {
			return false;
		}
	  	ws = new WebSocket("localhost:%d");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
			print("ERROR: " + evt.data);
        }
		ws.send("connect");
	}()
	</script>
	`, port)
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

// Handler for connecting to websockets
func (rel *Reloader) wshome(w http.ResponseWriter, r *http.Request) {

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	// rel.ws_connections = append(rel.ws_connections, c)

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

func New(cb func(), wsport int, logs_enabled bool) (*Reloader, error) {

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	template, err := template.New("go-again").Parse(create_template(wsport))

	if err != nil {
		return nil, err
	}

	r := &Reloader{
		watcher:        watcher,
		logs_enabled:   logs_enabled,
		ws_connections: make([]*websocket.Conn, 0),
		template:       template,
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

	// Set up websocket to reload webpage
	var addr = flag.String("addr", fmt.Sprintf("%s%d", "localhost:", wsport), "http service address")

	http.HandleFunc("/", r.wshome)
	go http.ListenAndServe(*addr, nil)

	// Add template

	return r, err
}
