# 🤖 Cassandra - Bender 3D com Gemini Live & Face Tracking

O **Cassandra** é um projeto interativo em tempo real que une **computação gráfica 3D (Three.js)**, **rastreamento facial inteligente (GoCV/OpenCV)** e **inteligência artificial por voz (Gemini Live API da Google)**. 

O resultado é um modelo 3D estilizado do icônico robô **Bender** que acompanha fisicamente o movimento do seu rosto e responde a conversas por voz em tempo real, gerando uma experiência de interação imersiva e responsiva.

---

## 🚀 Funcionalidades Principais

*   **Bender 3D Interativo**: Modelo do robô inteiramente modelado e animado em Three.js.
*   **Gemini Live API (WebSockets)**: Comunicação bidirecional por voz em baixíssima latência (streaming de áudio PCM de entrada e saída).
*   **Rastreamento Facial em Tempo Real**: O servidor escrito em Go capta o stream de vídeo, processa o rastreamento usando OpenCV (`CascadeClassifier`) e move a perspectiva tridimensional (olhos e câmera) para alinhar-se ao rosto do usuário.
*   **Animação de Voz Dinâmica**: A boca do Bender e suas expressões respondem de forma dinâmica e oscilante enquanto ele está falando ou ouvindo.
*   **Fallback Sarcástico**: Instruções de personalidade customizadas para agir de forma bem-humorada e sarcástica como o Bender clássico.

---

## 📁 Estrutura de Arquivos

```text
├── main.go                       # Servidor central em Go (WebSockets, GoCV Tracker e Servidor Estático)
├── tracker.go                    # Lógica de processamento de imagem e rastreamento facial (GoCV)
├── web/                          # Diretório contendo os arquivos frontend da aplicação web
│   ├── index.html                # Interface web do usuário
│   ├── main.js                   # Lógica de renderização 3D, controle de câmera e eventos Three.js
│   ├── live_client.js            # Cliente Web Audio e gerador de PCM para a conexão bidirecional
│   ├── style.css                 # Estilização da interface com foco em design moderno e responsivo
│   └── ai_studio.js              # Lógica de fallback para chat HTTP
├── server.py                     # Servidor proxy opcional em Python para chat de fallback (HTTP)
├── ai_studio_code.py             # Script Python independente de demonstração para a API Live (PyAudio + GenAI)
├── haarcascade_frontalface_default.xml # Modelo treinado do OpenCV para detecção de rostos
├── .env.example                  # Template das variáveis de ambiente globais
├── config.ini.example            # Template de configuração para o proxy Python
└── .gitignore                    # Arquivos ignorados pelo Git
```

---

## 🛠️ Pré-requisitos

### 1. Go (Golang)
*   Recomendado: Go 1.20 ou superior.
*   Instalação do **GoCV**: Siga as instruções oficiais do [GoCV Roadmap](https://gocv.io/) para instalar o OpenCV em seu sistema operacional (Windows, macOS ou Linux).
    *   *No Windows*, o GoCV facilita a instalação por meio de scripts que automatizam a compilação do OpenCV.

### 2. Python (Opcional, para rodar scripts adicionais)
*   Recomendado: Python 3.10 ou superior.
*   Instale as dependências executando:
    ```bash
    pip install google-genai opencv-python pyaudio pillow mss
    ```

---

## ⚙️ Configuração e Instalação

1.  **Clone o Repositório**:
    ```bash
    git clone https://github.com/seu-usuario/cassandra.git
    cd cassandra
    ```

2.  **Configure as Variáveis de Ambiente (Go / Frontend)**:
    Copie o arquivo de exemplo `.env.example` para `.env`:
    ```bash
    cp .env.example .env
    ```
    Abra o `.env` e adicione sua chave do **Google AI Studio (Gemini)**:
    ```env
    GEMINI_API_KEY=AIzaSy...seu_token_aqui
    PORT=8080
    ```

3.  **Configure o Proxy Python (Opcional)**:
    Se for utilizar o servidor proxy Python (`server.py`), copie `config.ini.example` para `config.ini`:
    ```bash
    cp config.ini.example config.ini
    ```
    Insira sua chave de API do Gemini no `config.ini`:
    ```ini
    [GEMINI]
    api_key = sk-...sua_chave_aqui
    ```

---

## 🏃 Como Executar

### Executando o Servidor Principal (Go)

1.  Garanta que as dependências do Go estão instaladas:
    ```bash
    go mod tidy
    ```

2.  Inicie o servidor Go:
    ```bash
    go run main.go tracker.go
    ```
    *Ou compile o binário localmente:*
    ```bash
    go build -o cassandra.exe main.go tracker.go
    ./cassandra.exe
    ```

3.  Acesse em seu navegador:
    👉 **http://localhost:8080**

### Usando o Modo Rastreamento Facial
*   Por padrão, o `main.go` tenta se conectar a uma câmera de rede (IP stream) no endereço configurado na linha 101 de `main.go`.
*   Para utilizar sua **Webcam integrada local**, você pode modificar a inicialização em `main.go` ou ajustar o `tracker.go` para capturar do dispositivo `0` do OpenCV:
    ```go
    // Para webcam integrada padrão:
    tracker := NewFaceTracker("0", "haarcascade_frontalface_default.xml")
    ```

---

## 🛡️ Segurança e Boas Práticas

Este repositório foi higienizado e configurado seguindo rígidos critérios de segurança:
*   As chaves de API estão **completamente isoladas** nos arquivos `.env` e `config.ini`.
*   O arquivo `.gitignore` está configurado para garantir que você **nunca envie suas chaves privadas ou binários compilados** por engano ao GitHub.

---

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes (sinta-se livre para criá-lo).

---

*Desenvolvido com 🤖 e ⚡ por Marcelo.*
