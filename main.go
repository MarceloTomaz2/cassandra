package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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

var (
	wakewordModel     string  = "models/ok_bender.onnx"
	wakewordThreshold float64 = 0.8
)

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

	// Inicialização da configuração do wake word
	wakewordModel = os.Getenv("WAKEWORD_MODEL_PATH")
	if wakewordModel == "" {
		wakewordModel = "models/edna.onnx"
	}
	thresholdStr := os.Getenv("WAKEWORD_THRESHOLD")
	if thresholdStr != "" {
		if t, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
			wakewordThreshold = t
		}
	}
	fmt.Printf("Sucesso: Configuração de wake word carregada. Modelo: %s (limiar: %.2f)\n", wakewordModel, wakewordThreshold)

	hub := newHub()
	go hub.run()

	streamURL := os.Getenv("TRACKER_STREAM_URL")
	if streamURL == "" {
		streamURL = "http://localhost:81/stream"
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

	var (
		geminiConn     *websocket.Conn
		isGeminiActive bool
		mu             sync.Mutex
	)

	// Inicia o processo Python para detecção de Wake Word em tempo real para esta conexão
	cmd := exec.Command("python", "wakeword_detector.py", wakewordModel, strconv.FormatFloat(wakewordThreshold, 'f', 2, 64))
	pythonStdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("Falha ao criar stdin pipe para o detector de wake word: %v", err)
		clientConn.Close()
		return
	}
	pythonStdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Falha ao criar stdout pipe para o detector de wake word: %v", err)
		clientConn.Close()
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Falha ao iniciar o subprocesso Python de wake word: %v", err)
		clientConn.Close()
		return
	}

	defer func() {
		hub.unregister <- clientConn
		clientConn.Close()
		mu.Lock()
		if isGeminiActive && geminiConn != nil {
			geminiConn.Close()
		}
		mu.Unlock()

		// Encerra o subprocesso Python com segurança
		pythonStdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		fmt.Println("Conexão WebSocket encerrada. Subprocesso Python finalizado.")
	}()

	hub.register <- clientConn
	fmt.Println("Novo cliente conectado! Aguardando palavra de ativação...")

	// Inicia notificando o frontend
	clientConn.WriteJSON(map[string]string{
		"type": "status",
		"data": "Aguardando palavra de ativação...",
	})

	// Goroutine para ler a saída do detector Python e ativar a conexão Gemini Live
	go func() {
		scanner := bufio.NewScanner(pythonStdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("[WakeWord] %s\n", line)
			if strings.HasPrefix(line, "DETECTED:") {
				mu.Lock()
				if !isGeminiActive {
					fmt.Printf("🔥 WAKEWORD DETECTADA: %s\n", line)
					// Conecta ao Gemini Live em tempo real
					gConn, err := dialGemini(apiKey, clientConn, hub, &isGeminiActive, &mu)
					if err == nil {
						geminiConn = gConn
						isGeminiActive = true
						clientConn.WriteJSON(map[string]string{
							"type": "status",
							"data": "Bender Ativado! Conversando...",
						})
						clientConn.WriteJSON(map[string]string{
							"type": "wake_word_detected",
							"data": "active",
						})
					} else {
						log.Printf("Falha ao conectar ao Gemini pós WakeWord: %v", err)
					}
				}
				mu.Unlock()
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Erro ao ler do stdout do subprocesso Python: %v", err)
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
			audioBytes, err := base64.StdEncoding.DecodeString(clientMsg["data"])
			if err != nil {
				continue
			}

			mu.Lock()
			if !isGeminiActive {
				// Envia os bytes PCM brutos de 16kHz mono para o detector Python
				_, _ = pythonStdin.Write(audioBytes)
			} else {
				// Gemini ativo: envia PCM diretamente
				geminiInput := map[string]interface{}{
					"realtimeInput": map[string]interface{}{
						"audio": map[string]interface{}{
							"data":     clientMsg["data"],
							"mimeType": "audio/pcm;rate=16000",
						},
					},
				}
				if err := geminiConn.WriteJSON(geminiInput); err != nil {
					log.Printf("Falha ao enviar áudio ao Gemini: %v. Revertendo para Escuta Ativa...", err)
					geminiConn.Close()
					isGeminiActive = false
					clientConn.WriteJSON(map[string]string{
						"type": "status",
						"data": "Aguardando palavra de ativação...",
					})
				}
			}
			mu.Unlock()
		}
	}
}

func dialGemini(apiKey string, clientConn *websocket.Conn, hub *Hub, isGeminiActive *bool, mu *sync.Mutex) (*websocket.Conn, error) {
	liveURL := os.Getenv("GEMINI_LIVE_URL")
	if liveURL == "" {
		liveURL = "wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1alpha.GenerativeService.BidiGenerateContent"
	}
	geminiURL := liveURL + "?key=" + apiKey

	header := http.Header{}
	header.Add("Origin", "http://localhost:8080")

	geminiConn, _, err := websocket.DefaultDialer.Dial(geminiURL, header)
	if err != nil {
		return nil, err
	}

	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "models/gemini-3.1-flash-live-preview"
	}
	voiceName := os.Getenv("GEMINI_VOICE_NAME")
	if voiceName == "" {
		voiceName = "Puck"
	}
	var systemInstruction string
	promptBytes, err := os.ReadFile("system_prompt.md")
	if err == nil {
		systemInstruction = strings.TrimSpace(string(promptBytes))
		fmt.Println("Sucesso: Prompt do sistema carregado de system_prompt.md")
	} else {
		systemInstruction = os.Getenv("GEMINI_SYSTEM_INSTRUCTION")
		if systemInstruction == "" {
			systemInstruction = "Você é o Bender, um robô divertido que explica tudo sobre o mundo da inteligência artificial de forma amigável e descontraída."
		}
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
		geminiConn.Close()
		return nil, err
	}

	fmt.Println("Conectado com sucesso ao Google Gemini Live!")

	// Envia um "Olá!" automático solicitando a frase específica ao iniciar
	initGreeting := map[string]interface{}{
		"clientContent": map[string]interface{}{
			"turns": []map[string]interface{}{
				{
					"role": "user",
					"parts": []map[string]interface{}{
						{"text": "Diga exatamente: 'Como posso ajudar?'"},
					},
				},
			},
			"turnComplete": true,
		},
	}
	if err := geminiConn.WriteJSON(initGreeting); err != nil {
		log.Printf("Aviso: Falha ao enviar saudação inicial ao Gemini: %v", err)
	}

	// Goroutine de leitura assíncrona do Gemini -> Cliente
	go func() {
		for {
			_, message, err := geminiConn.ReadMessage()
			if err != nil {
				log.Printf("Conexão do Gemini encerrada ou erro: %v. Revertendo para Escuta Ativa...", err)
				mu.Lock()
				*isGeminiActive = false
				mu.Unlock()
				clientConn.WriteJSON(map[string]string{
					"type": "status",
					"data": "Aguardando palavra de ativação...",
				})
				clientConn.WriteJSON(map[string]string{
					"type": "wake_word_detected",
					"data": "idle",
				})
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

	return geminiConn, nil
}
