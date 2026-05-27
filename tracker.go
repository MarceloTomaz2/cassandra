//go:build gocv

package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"gocv.io/x/gocv"
)

type FaceTracker struct {
	streamURL string
	cascade   string
	OnFace    func(x, y float64)
	Processed chan []byte
}

func NewFaceTracker(url, cascade string) *FaceTracker {
	return &FaceTracker{
		streamURL: url,
		cascade:   cascade,
		Processed: make(chan []byte, 1),
	}
}

func (ft *FaceTracker) Start() {
	webcam, err := gocv.OpenVideoCapture(ft.streamURL)
	if err != nil {
		log.Printf("Error opening video stream: %v", err)
		return
	}
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(ft.cascade) {
		log.Printf("Error reading cascade file: %v", ft.cascade)
		return
	}

	fmt.Printf("Face tracking started on %s\n", ft.streamURL)

	green := color.RGBA{0, 255, 0, 0}

	for {
		if ok := webcam.Read(&img); !ok {
			log.Printf("Device closed on %s", ft.streamURL)
			return
		}
		if img.Empty() {
			continue
		}

		rects := classifier.DetectMultiScale(img)
		if len(rects) > 0 {
			// Find the largest face
			var largestRect image.Rectangle
			maxArea := 0
			for _, r := range rects {
				// Draw rectangle on face
				gocv.Rectangle(&img, r, green, 3)

				area := r.Dx() * r.Dy()
				if area > maxArea {
					maxArea = area
					largestRect = r
				}
			}

			// Calculate center and normalize
			width := float64(img.Cols())
			height := float64(img.Rows())

			centerX := float64(largestRect.Min.X + largestRect.Dx()/2)
			centerY := float64(largestRect.Min.Y + largestRect.Dy()/2)

			normX := (centerX/width)*2 - 1
			normY := (centerY/height)*2 - 1

			if ft.OnFace != nil {
				// OpenCV coordinates: 0,0 is top-left.
				// Mouse coordinates in Three.js: -1,-1 is bottom-left, 1,1 is top-right.
				ft.OnFace(normX, -normY)
			}
		}

		// Encode to JPEG and send to channel
		buf, err := gocv.IMEncode(".jpg", img)
		if err == nil {
			select {
			case ft.Processed <- buf.GetBytes():
			default:
				// Skip if channel is full
			}
			buf.Close()
		}
	}
}
