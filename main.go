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

// Private fucntions
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func (r *Reloader) log(v ...any) {
	if r.logsEnabled {
		log.Println(v...)
	}
}

// createReloadScript generates the WebSocket client code
func createReloadScript(port int) template.HTML {
	script := fmt.Sprintf(`
		<script>
		(function() {
			let ws = null;
			
			// Simple function to update stylesheets
			function updateStylesheets(newDoc) {
				// Remove all existing stylesheets
				document.querySelectorAll('link[rel="stylesheet"]').forEach(sheet => {
					sheet.remove();
				});

				// Add all new stylesheets
				newDoc.querySelectorAll('link[rel="stylesheet"]').forEach(newSheet => {
					document.head.appendChild(newSheet.cloneNode(true));
				});

				// Handle inline styles
				document.querySelectorAll('style').forEach(style => style.remove());
				newDoc.querySelectorAll('style').forEach(newStyle => {
					document.head.appendChild(newStyle.cloneNode(true));
				});
			}

			// Helper function to compare and update elements
			function updateElement(oldEl, newEl) {
			    if (oldEl.hasAttribute('data-client-state') || oldEl.classList.contains('client-state')) {
        			return;
    			}

				// Skip if either element is undefined/null
				if (!oldEl || !newEl) {
					console.warn('[Go-Again] Skipping update for undefined element');
					return;
				}

				// Special handling for images
				// if (oldEl.tagName === 'IMG') {
				// 	if (oldEl.src !== newEl.src) {
				// 		oldEl.src = newEl.src;
				// 	}
				// 	return;
				// }


				// Update attributes
				Array.from(newEl.attributes).forEach(attr => {
					if (oldEl.getAttribute(attr.name) !== attr.value) {
						oldEl.setAttribute(attr.name, attr.value);
					}
				});
				
				// Remove old attributes that don't exist in new element
				Array.from(oldEl.attributes).forEach(attr => {
					if (!newEl.hasAttribute(attr.name)) {
						oldEl.removeAttribute(attr.name);
					}
				});

				// Compare text content if it's a text node
				if (newEl.childNodes.length === 1 && newEl.firstChild.nodeType === 3) {
					if (oldEl.textContent !== newEl.textContent) {
						oldEl.textContent = newEl.textContent;
					}
					return;
				}

				// Compare children
				const oldChildren = oldEl.children;
				const newChildren = newEl.children;
				
				// Update existing children and add new ones
				const maxLength = Math.max(oldChildren.length, newChildren.length);
				for (let i = 0; i < maxLength; i++) {
					if (i >= oldChildren.length) {
						// Add new child
						oldEl.appendChild(newChildren[i].cloneNode(true));
					} else if (i >= newChildren.length) {
						// Remove extra old child
						oldEl.removeChild(oldChildren[i]);
					} else {
						// Update existing child
						if (oldChildren[i].tagName === newChildren[i].tagName) {
							updateElement(oldChildren[i], newChildren[i]);
						} else {
							oldEl.replaceChild(newChildren[i].cloneNode(true), oldChildren[i]);
						}
					}
				}
			}

			// Function to fetch and update content
			async function updateContent() {
				try {
					const response = await fetch(window.location.href);
					const text = await response.text();
					
					// Create a temporary container to parse the HTML
					const parser = new DOMParser();
					const newDoc = parser.parseFromString(text, 'text/html');
					
					// Update stylesheets
					updateStylesheets(newDoc);
					
					// Update title if changed
					if (document.title !== newDoc.title) {
						document.title = newDoc.title;
					}
					
					// Update body content
					updateElement(document.body, newDoc.body);
					
					console.log('[Go-Again] DOM and styles updated successfully');
				} catch (error) {
					console.error('[Go-Again] Error updating content:', error);
				}
			}

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
					console.log("[Go-Again] Updating content due to: " + evt.data);
					updateContent();
				};
			}
			
			connect();
		})();
		</script>
	`, port)
	return template.HTML(script)
}

// Public functions

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

// Add path to folder or file to be watched
func (r *Reloader) Add(path string) error {
	r.log("Watching path: ", path)
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
				callback()

				if strings.HasSuffix(event.Name, ".html") || strings.HasSuffix(event.Name, ".tmpl") {
					r.log("Template modified:", event.Name)
				} else if strings.HasSuffix(event.Name, ".css") {
					r.log("Styles modified:", event.Name)
				} else {
					r.log("File modified:", event.Name)
				}
				r.notifyClients(event.Name)

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
