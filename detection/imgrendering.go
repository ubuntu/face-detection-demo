package detection

import (
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path"

	"github.com/lazywei/go-opencv/opencv"
	"github.com/nfnt/resize"
	"github.com/ubuntu/face-detection-demo/datastore"
)

var (
	logos     []image.Image
	logosPath = []string{"ubuntu.png", "archlinux.png", "debian.png", "gentoo.png",
		"fedora.png", "opensuse.png", "yocto.png"}
	datadir string
)

// RenderedImage abstract if we are using opencv or direct image blending
type RenderedImage struct {
	cvimg         *opencv.IplImage
	img           *image.RGBA
	RenderingMode datastore.RenderMode
}

// InitLogos and destination datadir. Will ignore unreachable logos
func InitLogos(logodir string, ddir string) {
	datadir = ddir

	logos = make([]image.Image, len(logosPath))
	i := 0

	for _, p := range logosPath {
		f, err := os.Open(path.Join(logodir, p))
		if err != nil {
			log.Println("Couldn't open", path.Join(logodir, p))
			continue
		}
		defer f.Close()

		logo, _, err := image.Decode(f)
		if err != nil {
			log.Println("Couldn't load image", p)
			continue
		}
		logos[i] = logo
		i++
	}
	// reslice to have current len() in case we couldn't load some logos
	logos = logos[:i]
}

// DrawFace renders a new face on top of image depending on rendering type
func (r *RenderedImage) DrawFace(face *opencv.Rect, num int, cvimage *opencv.IplImage) {

	switch r.RenderingMode {
	case datastore.NORMALRENDERING:
		if r.cvimg == nil {
			r.cvimg = cvimage.Clone()
		}

		opencv.Circle(r.cvimg,
			opencv.Point{
				X: face.X() + (face.Width() / 2),
				Y: face.Y() + (face.Height() / 2),
			},
			face.Width()/2,
			opencv.ScalarAll(255.0), 1, 1, 0)

	case datastore.FUNRENDERING:
		if r.img == nil {
			source := cvimage.ToImage()
			r.img = image.NewRGBA(source.Bounds())
			draw.Draw(r.img, r.img.Bounds(), source, image.ZP, draw.Src)
		}

		// resize logo to match face
		// TODO: logo needs to be randomized depending on num
		logo := resize.Resize(0, uint(face.Height()), logos[num], resize.NearestNeighbor)
		logorect := image.Rect(face.X()+face.Width()/2-logo.Bounds().Dx()/2,
			face.Y()+face.Height()/2-logo.Bounds().Dy()/2,
			face.X()+logo.Bounds().Dx(),
			face.Y()+logo.Bounds().Dy())

		draw.Draw(r.img, logorect, logo, image.ZP, draw.Over)

	}
}

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
