//go:build !gocv

package main

import (
	"fmt"
	"time"
)

// FaceTracker é uma implementação de fallback que permite compilar o servidor Go
// sem a necessidade de instalar o OpenCV e o GoCV na máquina de desenvolvimento.
type FaceTracker struct {
	streamURL string
	cascade   string
	OnFace    func(x, y float64)
	Processed chan []byte
}

// NewFaceTracker inicializa a versão de fallback (sem GoCV) do FaceTracker
func NewFaceTracker(url, cascade string) *FaceTracker {
	return &FaceTracker{
		streamURL: url,
		cascade:   cascade,
		Processed: make(chan []byte, 1),
	}
}

// Start inicia o loop mock (não faz captura de vídeo real)
func (ft *FaceTracker) Start() {
	fmt.Println("INFO: Rastreamento Facial em modo simulado (GoCV desativado).")
	fmt.Println("Dica: Para compilar com suporte real a rastreamento por câmera, use: go run -tags gocv .")
	
	for {
		// Mantém o loop ativo sem consumir CPU
		time.Sleep(1 * time.Hour)
	}
}
