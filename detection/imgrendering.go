package detection

import (
	"fmt"
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

	detectedfilename = "screendetected.png"
	screenshotname   = "screencapture.png"
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

// Save current image in destination file
func (r *RenderedImage) Save() {

	var savefn func(string) error

	switch r.RenderingMode {
	case datastore.NORMALRENDERING:
		savefn = func(filepath string) error {
			opencv.SaveImage(filepath, r.cvimg, 0)
			return nil
		}

	case datastore.FUNRENDERING:
		savefn = func(filepath string) error {
			f, err := os.Create(filepath)
			if err != nil {
				return err
			}
			defer f.Close()
			return png.Encode(f, r.img)
		}
	}

	if err := saveatomic(datadir, detectedfilename, savefn); err != nil {
		fmt.Println(err)
	}
}

func saveatomic(dir string, filename string, savefn func(string) error) error {
	tempfilen := path.Join(dir, "new"+filename)
	dstfilen := path.Join(dir, filename)

	if err := savefn(tempfilen); err != nil {
		return fmt.Errorf("Couldn't save image to %s: %s", tempfilen, err)
	}
	defer os.Remove(tempfilen)

	if err := os.Rename(tempfilen, dstfilen); err != nil {
		return fmt.Errorf("Couldn't save temp image %s to %s: %s", tempfilen, dstfilen, err)
	}
	return nil
}
