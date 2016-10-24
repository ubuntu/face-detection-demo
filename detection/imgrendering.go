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
	"github.com/ubuntu/face-detection-demo/appstate"
	"github.com/ubuntu/face-detection-demo/datastore"
)

var (
	logos     []image.Image
	logosPath = []string{"ubuntu.png", "archlinux.png", "debian.png", "gentoo.png",
		"fedora.png", "opensuse.png", "yocto.png", "smiley.png"}
	datadir string

	detectedfilename = "screendetected.png"
	screenshotname   = "screencapture.png"
)

// RenderedImage abstract if we are using opencv or direct image blending
type RenderedImage struct {
	cvimg         *opencvImg
	img           *rgbaImg
	RenderingMode datastore.RenderMode
}

type opencvImg opencv.IplImage
type rgbaImg struct {
	*image.RGBA
}
type saver interface {
	Save(string) error
}

// load logo images. Ignore unreachable or undecodable ones.
func init() {
	datadir = appstate.Datadir

	logos = make([]image.Image, len(logosPath))
	i := 0

	for _, p := range logosPath {
		imgPath := path.Join(appstate.Rootdir, "images", p)
		f, err := os.Open(imgPath)
		if err != nil {
			log.Println("Couldn't open", imgPath)
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

// Save opencv images
func (i *opencvImg) Save(filepath string) error {
	opencv.SaveImage(filepath, (*opencv.IplImage)(i), 0)
	return nil
}

// Save rgba images in png
func (i *rgbaImg) Save(filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, i)
}

// DrawFace renders a new face on top of image depending on rendering type
func (r *RenderedImage) DrawFace(face *opencv.Rect, num int, cvimage *opencv.IplImage) {

	if appstate.BrokenMode {
		// force drawing smileys instead of people
		r.drawFunFace(face, len(logos)-1, cvimage)
		return
	}

	switch r.RenderingMode {
	case datastore.NORMALRENDERING:
		if r.cvimg == nil {
			r.cvimg = (*opencvImg)(cvimage.Clone())
		}

		opencv.Circle((*opencv.IplImage)(r.cvimg),
			opencv.Point{
				X: face.X() + (face.Width() / 2),
				Y: face.Y() + (face.Height() / 2),
			},
			face.Width()/2,
			opencv.ScalarAll(255.0), 1, 1, 0)

	case datastore.FUNRENDERING:
		// TODO: logo needs to be randomized depending on num
		r.drawFunFace(face, num%(len(logos)-1), cvimage)
	}
}

func (r *RenderedImage) drawFunFace(face *opencv.Rect, num int, cvimage *opencv.IplImage) {

	if r.img == nil {
		source := cvimage.ToImage()
		r.img = &rgbaImg{image.NewRGBA(source.Bounds())}
		draw.Draw(r.img, r.img.Bounds(), source, image.ZP, draw.Src)
	}

	// resize logo to match face
	logo := resize.Resize(0, uint(face.Height()), logos[num], resize.NearestNeighbor)
	logorect := image.Rect(face.X()+face.Width()/2-logo.Bounds().Dx()/2,
		face.Y()+face.Height()/2-logo.Bounds().Dy()/2,
		face.X()+logo.Bounds().Dx(),
		face.Y()+logo.Bounds().Dy())

	draw.Draw(r.img, logorect, logo, image.ZP, draw.Over)
}

// Save current image in destination file
func (r *RenderedImage) Save() {

	var i saver
	if r.cvimg != nil {
		i = r.cvimg
	} else {
		i = r.img
	}

	if err := saveatomic(datadir, detectedfilename, i); err != nil {
		fmt.Println(err)
	}
}

func saveatomic(dir string, filename string, s saver) error {
	tempfilen := path.Join(dir, "new"+filename)
	dstfilen := path.Join(dir, filename)

	if err := s.Save(tempfilen); err != nil {
		return fmt.Errorf("Couldn't save image to %s: %s", tempfilen, err)
	}
	defer os.Remove(tempfilen)

	if err := os.Rename(tempfilen, dstfilen); err != nil {
		return fmt.Errorf("Couldn't save temp image %s to %s: %s", tempfilen, dstfilen, err)
	}
	return nil
}

// WipeScreenshots removes screenshots in dir unconditionally (existing or not)
func WipeScreenshots(dir string) {
	os.Remove(path.Join(dir, detectedfilename))
	os.Remove(path.Join(dir, screenshotname))
}
