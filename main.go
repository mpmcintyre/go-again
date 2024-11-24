package reloader

import (
	"flag"
	"fmt"
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
}

// Private functions

func (r *Reloader) log(v ...any) {
	if r.logs_enabled {
		log.Println(v...)
	}
}

func (r *Reloader) fatal(v ...any) {
	if r.logs_enabled {
		log.Fatal(v...)
	}
}

// Public functions

// Add path to folder or file to be watched
func (r *Reloader) Add(path string) {

	err := r.watcher.Add(path)
	if err != nil {
		r.fatal(err)
	}
}

// The template required to allow live reloading for HTML templates
func (r *Reloader) Template() string {
	return `
	<script>
	(){
		if (ws) {
			return false;
		}
	  	ws = new WebSocket("{{.}}");
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
	`
}

// Defer this in order to ensure all resources are closed
func (r *Reloader) Close(path string) {
	// Close watcher
	r.watcher.Close()

	// Close ws connections
	for _, conn := range r.ws_connections {
		conn.Close()
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

func New(cb func(), wsport int, logs_enabled bool) Reloader {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	r := Reloader{
		watcher:        watcher,
		logs_enabled:   logs_enabled,
		ws_connections: make([]*websocket.Conn, 1),
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
				for _, conn := range r.ws_connections {
					conn.WriteMessage(websocket.TextMessage, []byte(event.Name))
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

	return r
}
