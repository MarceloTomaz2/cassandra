# 🤖 Cassandra - Bender 3D com Gemini Live & Face Tracking

O **Cassandra** é um projeto interativo em tempo real que une **computação gráfica 3D (Three.js)**, **rastreamento facial inteligente (GoCV/OpenCV)** e **inteligência artificial por voz (Gemini Live API da Google)**. 

O resultado é um modelo 3D estilizado do icônico robô **Bender** que acompanha fisicamente o movimento do seu rosto e responde a conversas por voz em tempo real, gerando uma experiência de interação imersiva e responsiva.

<p align="center">
  <img src="web/assets/demo.gif" alt="Cassandra Bender 3D Demo" width="700px" />
</p>

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
├── tracker_mock.go               # Fallback mock do FaceTracker para compilação sem GoCV/OpenCV
├── web/                          # Diretório contendo os arquivos frontend da aplicação web
│   ├── index.html                # Interface web do usuário
│   ├── main.js                   # Lógica de renderização 3D, controle de câmera e eventos Three.js
│   ├── live_client.js            # Cliente Web Audio e gerador de PCM para a conexão bidirecional
│   ├── style.css                 # Estilização da interface com foco em design moderno e responsivo
│   └── ai_studio.js              # Lógica de fallback para chat HTTP
├── haarcascade_frontalface_default.xml # Modelo treinado do OpenCV para detecção de rostos
├── .env.example                  # Template das variáveis de ambiente globais
└── .gitignore                    # Arquivos ignorados pelo Git
```

---

## 🛠️ Pré-requisitos

### Go (Golang)
*   Recomendado: Go 1.20 ou superior.
*   **GoCV / OpenCV (Opcional)**: Para habilitar o rastreamento facial real pela câmera, é necessário instalar as dependências do OpenCV em seu sistema operacional (Windows, macOS ou Linux). Consulte o guia oficial do [GoCV Roadmap](https://gocv.io/).
*   **Modo de Compatibilidade**: Se você preferir rodar sem instalar OpenCV/GoCV, o projeto funcionará no modo mock nativo (utilizando o movimento do mouse/toque no 3D) de forma imediata!

---

## ⚙️ Configuração e Instalação

1.  **Clone o Repositório**:
    ```bash
    git clone https://github.com/MarceloTomaz2/cassandra.git
    cd cassandra
    ```

2.  **Configure as Variáveis de Ambiente**:
    Copie o arquivo de exemplo `.env.example` para `.env`:
    ```bash
    cp .env.example .env
    ```
    Abra o `.env` e adicione sua chave ativa do **Google AI Studio (Gemini)** e demais configurações:
    ```env
    GEMINI_API_KEY=AIzaSy...sua_chave_real
    PORT=8080
    ```

---

## 🏃 Como Executar

### Executando o Servidor Principal (Go)

1.  Garanta que as dependências do Go estão instaladas:
    ```bash
    go mod tidy
    ```

2.  Inicie o servidor Go no modo padrão (Sem dependência do OpenCV/GoCV):
    ```bash
    go run .
    ```
    *Caso possua o OpenCV configurado e queira ativar o rastreamento por câmera real:*
    ```bash
    go run -tags gocv .
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
