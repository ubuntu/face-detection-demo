package detection

import (
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/lazywei/go-opencv/opencv"
	"github.com/nfnt/resize"
)

func drawFace(img *opencv.IplImage, face *opencv.Rect, num int) {

	opencv.Circle(img,
		opencv.Point{
			X: face.X() + (face.Width() / 2),
			Y: face.Y() + (face.Height() / 2),
		},
		face.Width()/2,
		opencv.ScalarAll(255.0), 1, 1, 0)

	infile, err := os.Open("/tmp/logo.png")
	if err != nil {
		// replace this with real error handling
		log.Fatal(err)
	}
	defer infile.Close()

	logosrc, _, err := image.Decode(infile)
	if err != nil {
		// replace this with real error handling
		log.Fatal(err)
	}
	logo := resize.Resize(0, uint(face.Height()), logosrc, resize.NearestNeighbor)

	logorect := image.Rect(face.X()+face.Width()/2-logo.Bounds().Dx()/2,
		face.Y()+face.Height()/2-logo.Bounds().Dy()/2,
		face.X()+logo.Bounds().Dx(),
		face.Y()+logo.Bounds().Dy())

	source := img.ToImage()

	m := image.NewRGBA(source.Bounds())
	draw.Draw(m, m.Bounds(), source, image.ZP, draw.Src)

	draw.Draw(m, logorect, logo, image.ZP, draw.Over)

	w, _ := os.Create("/tmp/result.png")
	defer w.Close()
	png.Encode(w, m)
}
