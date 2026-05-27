package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex

	videoClients   map[chan []byte]bool
	videoBroadcast chan []byte
	videoMu        sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		clients:        make(map[*websocket.Conn]bool),
		broadcast:      make(chan []byte),
		register:       make(chan *websocket.Conn),
		unregister:     make(chan *websocket.Conn),
		videoClients:   make(map[chan []byte]bool),
		videoBroadcast: make(chan []byte),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		case frame := <-h.videoBroadcast:
			h.videoMu.Lock()
			for clientChan := range h.videoClients {
				select {
				case clientChan <- frame:
				default:
				}
			}
			h.videoMu.Unlock()
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Aviso: Arquivo .env não encontrado. Tentando ler das variáveis de ambiente do sistema.")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		fmt.Println("ERRO: A variável GEMINI_API_KEY ou GOOGLE_API_KEY não foi encontrada no .env ou no sistema.")
		fmt.Println("Certifique-se de que o arquivo .env existe e contém: GEMINI_API_KEY=sua_chave")
		os.Exit(1)
	} else {
		fmt.Printf("Sucesso: Chave de API carregada (%d caracteres).\n", len(strings.TrimSpace(apiKey)))
	}

	hub := newHub()
	go hub.run()

	streamURL := os.Getenv("TRACKER_STREAM_URL")
	if streamURL == "" {
		streamURL = "http://192.168.1.109:81/stream"
	}
	cascadeFile := os.Getenv("TRACKER_CASCADE_FILE")
	if cascadeFile == "" {
		cascadeFile = "haarcascade_frontalface_default.xml"
	}
	tracker := NewFaceTracker(streamURL, cascadeFile)
	tracker.OnFace = func(x, y float64) {
		msg, _ := json.Marshal(map[string]interface{}{
			"type": "face",
			"x":    x,
			"y":    y,
		})
		hub.broadcast <- msg
	}
	go func() {
		for frame := range tracker.Processed {
			hub.videoBroadcast <- frame
		}
	}()
	go tracker.Start()

	http.HandleFunc("/video", func(w http.ResponseWriter, r *http.Request) {
		frameChan := make(chan []byte, 1)
		hub.videoMu.Lock()
		hub.videoClients[frameChan] = true
		hub.videoMu.Unlock()

		defer func() {
			hub.videoMu.Lock()
			delete(hub.videoClients, frameChan)
			hub.videoMu.Unlock()
		}()

		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		for frame := range frameChan {
			fmt.Fprintf(w, "--frame\r\n")
			fmt.Fprintf(w, "Content-Type: image/jpeg\r\n")
			fmt.Fprintf(w, "Content-Length: %d\r\n", len(frame))
			fmt.Fprintf(w, "\r\n")
			w.Write(frame)
			fmt.Fprintf(w, "\r\n")
		}
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		clientConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Events WebSocket upgrade failed: %v", err)
			return
		}
		defer func() {
			hub.unregister <- clientConn
			clientConn.Close()
		}()
		hub.register <- clientConn

		for {
			if _, _, err := clientConn.ReadMessage(); err != nil {
				break
			}
		}
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, apiKey, hub)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "web/index.html")
			return
		}
		path := "web/" + r.URL.Path[1:]
		if _, err := os.Stat(path); err == nil {
			http.ServeFile(w, r, path)
			return
		}
		http.NotFound(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Servidor rodando em: http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, apiKey string, hub *Hub) {
	apiKey = strings.TrimSpace(apiKey)
	apiKey = strings.Trim(apiKey, "\r\n")

	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer func() {
		hub.unregister <- clientConn
		clientConn.Close()
	}()
	hub.register <- clientConn

	liveURL := os.Getenv("GEMINI_LIVE_URL")
	if liveURL == "" {
		liveURL = "wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1alpha.GenerativeService.BidiGenerateContent"
	}
	geminiURL := liveURL + "?key=" + apiKey

	header := http.Header{}
	header.Add("Origin", "http://localhost:8080")

	geminiConn, _, err := websocket.DefaultDialer.Dial(geminiURL, header)
	if err != nil {
		log.Printf("Failed to connect to Gemini: %v", err)
		return
	}
	defer geminiConn.Close()
	fmt.Println("Conectado ao Gemini!")

	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "models/gemini-3.1-flash-live-preview"
	}
	voiceName := os.Getenv("GEMINI_VOICE_NAME")
	if voiceName == "" {
		voiceName = "Puck"
	}
	systemInstruction := os.Getenv("GEMINI_SYSTEM_INSTRUCTION")
	if systemInstruction == "" {
		systemInstruction = "Você é o RoboEd, um robô divertido que explica tudo sobre a relação sobre inglês, matemática e tecnologia para alunos do ensino fundamental."
	}

	setupMsg := map[string]interface{}{
		"setup": map[string]interface{}{
			"model": modelName,
			"generationConfig": map[string]interface{}{
				"responseModalities": []string{"AUDIO"},
				"speechConfig": map[string]interface{}{
					"voiceConfig": map[string]interface{}{
						"prebuiltVoiceConfig": map[string]interface{}{
							"voiceName": voiceName,
						},
					},
				},
			},
			"systemInstruction": map[string]interface{}{
				"parts": []map[string]interface{}{
					{"text": systemInstruction},
				},
			},
		},
	}
	if err := geminiConn.WriteJSON(setupMsg); err != nil {
		log.Printf("Failed to send setup message: %v", err)
		return
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := geminiConn.ReadMessage()
			if err != nil {
				log.Printf("Gemini receive error: %v", err)
				return
			}

			var geminiResp map[string]interface{}
			if err := json.Unmarshal(message, &geminiResp); err != nil {
				continue
			}

			if serverContent, ok := geminiResp["serverContent"].(map[string]interface{}); ok {
				if modelTurn, ok := serverContent["modelTurn"].(map[string]interface{}); ok {
					if parts, ok := modelTurn["parts"].([]interface{}); ok {
						for _, part := range parts {
							p := part.(map[string]interface{})
							if inlineData, ok := p["inlineData"].(map[string]interface{}); ok {
								clientConn.WriteJSON(map[string]string{
									"type": "audio",
									"data": inlineData["data"].(string),
								})
							}
							if text, ok := p["text"].(string); ok && text != "" {
								clientConn.WriteJSON(map[string]string{
									"type": "text",
									"data": text,
								})
							}
						}
					}
				}
			}
		}
	}()

	for {
		_, message, err := clientConn.ReadMessage()
		if err != nil {
			break
		}

		var clientMsg map[string]string
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			continue
		}

		if clientMsg["type"] == "audio" {
			geminiInput := map[string]interface{}{
				"realtimeInput": map[string]interface{}{
					"audio": map[string]interface{}{
						"data":     clientMsg["data"],
						"mimeType": "audio/pcm;rate=16000",
					},
				},
			}
			if err := geminiConn.WriteJSON(geminiInput); err != nil {
				break
			}
		}
	}
}
