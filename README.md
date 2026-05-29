# 🤖 Cassandra - Edna/Bender 3D com Gemini Live & Face Tracking

O **Cassandra** é um projeto interativo em tempo real que une **computação gráfica 3D (Three.js)**, **rastreamento facial inteligente (GoCV/OpenCV)**, **inteligência artificial por voz (Gemini Live API da Google)** e um **sistema local de detecção de palavra de ativação (openWakeWord e ONNX Runtime)**.

O resultado é um modelo 3D estilizado do robô que acompanha fisicamente o movimento do seu rosto e responde a conversas por voz em tempo real, gerando uma experiência de interação imersiva e responsiva.

<p align="center">
  <img src="web/assets/demo.gif" alt="Cassandra Bender 3D Demo" width="700px" />
</p>

---

## 🚀 Funcionalidades Principais

*   **Edna/Bender 3D Interativo**: Modelo do robô inteiramente modelado e animado em Three.js.
*   **Gemini Live API (WebSockets)**: Comunicação bidirecional por voz em baixíssima latência (streaming de áudio PCM de entrada e saída).
*   **Rastreamento Facial em Tempo Real**: O servidor escrito em Go capta o stream de vídeo, processa o rastreamento usando OpenCV (`CascadeClassifier`) e move a perspectiva tridimensional (olhos e câmera) para alinhar-se ao rosto do usuário.
*   **Detecção de Palavra de Ativação Real (Wake Word)**: Ponte Go-Python de alto desempenho rodando modelos ONNX locais (`edna.onnx`, `ok_bender.onnx`, etc.) por meio de um subprocesso otimizado e isolado por conexão. Não há ativações falsas por volume ou palmas.
*   **Centralização do Prompt (`system_prompt.md`)**: A personalidade e as instruções do sistema que regem o robô residem em um arquivo markdown dedicado na raiz do projeto, carregado de forma dinâmica durante as chamadas.
*   **Saudação Inteligente Automática**: Ao iniciar a conexão, o robô toma a iniciativa e começa a falar imediatamente a frase `"Como posso ajudar?"`, tornando a experiência de conversação natural.
*   **Sinalização Acústica e Visual (Chime)**: Uso da Web Audio API nativa para reproduzir efeitos sonoros de ativação (ascendente) e desativação (descendente) no navegador, além de alteração da cor do botão do microfone para vermelho indicando gravação ativa.

---

## 📁 Estrutura de Arquivos

```text
├── main.go                       # Servidor central em Go (WebSockets, gerenciador de sub-processos e Servidor Estático)
├── tracker.go                    # Lógica de processamento de imagem e rastreamento facial (GoCV)
├── tracker_mock.go               # Fallback mock do FaceTracker para compilação sem GoCV/OpenCV
├── wakeword_detector.py          # Script de inferência local em Python usando openwakeword e ONNX Runtime
├── system_prompt.md              # Onde vivem as instruções de comportamento do modelo
├── models/                       # Diretório centralizado de arquivos de redes neurais do Wake Word
│   ├── edna.onnx                 # Modelo ONNX ativo de ativação
│   ├── ok_bender.onnx            # Modelo ONNX alternativo
│   └── *.tflite                  # Modelos adicionais compactos
├── web/                          # Diretório contendo os arquivos frontend da aplicação web
│   ├── index.html                # Interface web do usuário
│   ├── main.js                   # Lógica 3D Three.js, animação facial e sintetizador de chimes acústicos
│   ├── live_client.js            # Cliente Web Audio e gerador de PCM para a conexão bidirecional
│   ├── style.css                 # Estilização moderna da interface com foco em design responsivo
│   └── ai_studio.js              # Lógica de fallback para chat HTTP
├── haarcascade_frontalface_default.xml # Modelo treinado do OpenCV para detecção de rostos
├── .env.example                  # Template das variáveis de ambiente globais
└── .gitignore                    # Arquivos ignorados pelo Git
```

---

## 🛠️ Pré-requisitos

### Servidor Go (Golang)
*   Recomendado: Go 1.22 ou superior.
*   **GoCV / OpenCV (Opcional)**: Para habilitar o rastreamento facial real pela câmera. Se você preferir rodar sem instalar OpenCV/GoCV, o projeto funcionará no modo mock nativo (utilizando o movimento do mouse/toque no 3D) de forma imediata!

### Detector de Voz (Python)
*   Recomendado: Python 3.13 ou superior.
*   Instale as bibliotecas necessárias para rodar o motor local de Wake Word:
    ```bash
    pip install openwakeword onnxruntime numpy
    ```

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
    Abra o `.env` e configure sua chave da API do Gemini e o modelo desejado:
    ```env
    GEMINI_API_KEY=AIzaSy...sua_chave_real
    WAKEWORD_MODEL_PATH=models/edna.onnx
    WAKEWORD_THRESHOLD=0.8
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

---

## 🛡️ Segurança e Limpeza de Histórico

Este repositório foi higienizado profissionalmente de ponta a ponta:
*   **Sem Vazamentos**: As chaves de API estão **completamente isoladas** nos arquivos `.env`.
*   **Histórico Higienizado**: O repositório foi totalmente purgado via reescrita de commits (`git filter-branch`) para garantir que **antigos arquivos de configuração com chaves vazadas tenham sido permanentemente apagados de toda a história do Git**, garantindo total conformidade de segurança e privacidade antes do push para o GitHub.

---

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

---

*Desenvolvido com 🤖 e ⚡ por Marcelo.*
