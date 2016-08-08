package detection

import "github.com/lazywei/go-opencv/opencv"

func drawFace(img *opencv.IplImage, face *opencv.Rect, num int) {
	opencv.Circle(img,
		opencv.Point{
			X: face.X() + (face.Width() / 2),
			Y: face.Y() + (face.Height() / 2),
		},
		face.Width()/2,
		opencv.ScalarAll(255.0), 1, 1, 0)
}
