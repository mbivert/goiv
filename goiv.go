package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

var argv0 string

var paths []string
var npaths = 0

var random = flag.Bool("r", false, "randomize input files")
var doHelp = flag.Bool("h", false, "show help")

var window *sdl.Window
var image *sdl.Surface

var defWinW = int32(800)
var defWinH = int32(600)

// Black
var defBGRed = uint8(0)
var defBGGreen = uint8(0)
var defBGBlue = uint8(0)
var defBGAlpha = uint8(0)

var clickX = int32(0)
var clickY = int32(0)

var zoom = 1.

// XXX clumsy
var isImgs = map[string]bool{
	".jpg": true,
	".JPG": true,

	".jpeg": true,
	".JPEG": true,

	".png": true,
	".PNG": true,

	".webp": true,
	".WEBP": true,

	".avif": true,
	".AVIF": true,

	// XXX guess we should reload the image when we zoom to
	// leverage "vectoriality"
	".svg": true,
	".SVG": true,

	// XXX
	".gif": true,
	".GIF": true,
}

func help(n int) {
	fmt.Fprintf(os.Stderr, "%s [-r] <path> [paths...]\n", argv0)
	os.Exit(n)
}

func fails(err error) {
	log.Fatal(argv0, ": ", err)
	cleanup()
}

func init() {
	argv0 = path.Base(os.Args[0])

	flag.Parse()

	if *doHelp {
		help(0)
	}

	args := flag.Args()

	var err error
	paths, err = lsPaths(args)
	if err != nil {
		fails(err)
	}
	if *random {
		paths = shuffle(paths)
	}

	if len(paths) == 0 {
		help(1)
	}
}

// XXX Readdirnames?
func lsDir(dpath string) ([]string, error) {
	xs := make([]string, 0)
	err := filepath.Walk(dpath, func(path string, _ fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if _, ok := isImgs[filepath.Ext(path)]; ok {
			xs = append(xs, path)
		}
		return nil
	})

	return xs, err
}

func isDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), err
}

func lsPaths(paths []string) ([]string, error) {
	xs := make([]string, 0)
	for _, path := range paths {
		dir, err := isDir(path)
		if err != nil {
			return xs, err
		}
		if dir {
			ys, err := lsDir(path)
			if err != nil {
				return xs, err
			}
			xs = append(xs, ys...)
		} else {
			xs = append(xs, path)
		}
	}
	return xs, nil
}

func shuffle(xs []string) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(xs), func(i, j int) { xs[i], xs[j] = xs[j], xs[i] })
	return xs
}

func mkWindow() {
	var err error

	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		fails(err)
	}

	window, err = sdl.CreateWindow(
		argv0, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		defWinW, defWinH, sdl.WINDOW_SHOWN,
	)
	if err != nil {
		fails(err)
	}

	window.SetResizable(true)
}

func cleanup() {
	sdl.Quit()
	window.Destroy()
	if image != nil {
		image.Free()
	}
}

func loadImg() {
	if image != nil {
		image.Free()
		image = nil
	}

	// Okay here for now
	clickX = 0
	clickY = 0

	// TODO: resiliency
	var err error
	if image, err = img.Load(paths[npaths]); err != nil {
		fails(err)
	}

	window.SetTitle(argv0 + " - " + paths[npaths])
}

func loadAndDrawImg() { loadImg(); drawImg() }

func nextImg() {
	npaths++
	if npaths == len(paths) {
		npaths = 0
	}
	loadAndDrawImg()
}

func prevImg() {
	npaths--
	if npaths == -1 {
		npaths = len(paths) - 1
	}
	loadAndDrawImg()
}

func drawBG(surface *sdl.Surface, ww, wh int32) {
	all := sdl.Rect{X: 0, Y: 0, W: ww, H: wh}
	bg := sdl.MapRGBA(surface.Format, defBGRed, defBGGreen, defBGBlue, defBGAlpha)
	surface.FillRect(&all, bg)
}

/*
 * Simple, bug-prone code: let's make it clear.
 *
 * 0) We have a drawing surface (~window)
 * (0,0) ------------------------------------- (ww,0)
 *   |
 *   |
 *   |
 *   |
 * (0,wh) ------------------------------------ (ww,wh)
 *
 * "x" coordinates range on the width.
 * "y" coordinates range on the height.
 *
 * 1) We want to draw an image on this drawing surface. The image
 * has dimensions (iw,ih).
 *
 * Take the image's greater side (height or width; default to width),
 * and compare it to the corresponding window's side (height or width).
 *
 * If the window's side is large enough to accomodate the corresponding,
 * image side, compute margins.
 *
 * Otherwise, that is, if the image's side is greater than the window's
 * side, compute a scaling factor (sh or sw) between those two sides:
 * we'll need to apply this scaling factor to the other side.
 *
 * We're now ready to manage the other side, for which we just have
 * to apply the scaling factor, and to compute the margins.
 *
 * All the measure can be wrapped into a rectangle
 *
 * 2) We now want to introduce a zooming factor: we simply
 * stretch the previous rectangle's width/height, and adjust
 * its left corner accordingly. We can do it "easily" because
 * SDL allows us to plot to negative points, so we don't have
 * clever computations to perform.
 *
 * TODO: zoom should be centered on the cursor's position
 */
func drawImg() {
	// XXX we could probably get it once?
	surface, err := window.GetSurface()
	if err != nil {
		fails(err)
	}

	ww := surface.W
	wh := surface.H
	iw := image.W
	ih := image.H

	x := int32(0)
	y := int32(0)
	w := int32(0)
	h := int32(0)

	if iw >= ih {
		s := 1.

		if ww >= iw {
			// margins
			x = (ww - iw) / 2
			w = iw // - (ww-iw)/2
		} else {
			// no margins, but scaling factor for the other side
			x = 0
			w = ww
			s = float64(ww) / float64(iw)
		}

		// adjust the other side
		h = int32(float64(ih) * s)
		y = (wh - h) / 2
	} else {
		s := 1.

		if wh >= ih {
			// margins
			y = (wh - ih) / 2
			h = ih // - (wh-ih)/2
		} else {
			// no margins, but scaling factor for the other side
			y = 0
			h = wh
			s = float64(wh) / float64(ih)
		}

		// adjust the other side
		w = int32(float64(iw) * s)
		x = (ww - w) / 2
	}

	zw := int32(zoom * float64(w))
	zh := int32(zoom * float64(h))

	// NOTE: we can "fortunately" plot to negative x/y
	x = (ww - zw) / 2
	y = (wh - zh) / 2

	// Shift to take into account moving around
	x += clickX
	y += clickY

	drawBG(surface, ww, wh)

	image.BlitScaled(nil, surface, &sdl.Rect{X: x, Y: y, W: zw, H: zh})

	window.UpdateSurface()
}

func mainLoop() {
	running := true
	ctrl := false
	clicking := false

	for event := sdl.WaitEvent(); event != nil && running; event = sdl.WaitEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			running = false
			break
		case *sdl.WindowEvent:
			switch t.Event {
			case sdl.WINDOWEVENT_MOVED:
				fallthrough
			case sdl.WINDOWEVENT_SIZE_CHANGED:
				fallthrough
			case sdl.WINDOWEVENT_RESIZED:
				drawImg()
			}
		case *sdl.KeyboardEvent:
			keyCode := t.Keysym.Sym
			//			println("keyboard:", string(keyCode), ctrl, t.State == sdl.PRESSED)
			ctrl = false
			switch t.Keysym.Mod {
			case sdl.KMOD_LCTRL:
				fallthrough
			case sdl.KMOD_RCTRL:
				ctrl = t.State == sdl.PRESSED
			}

			switch string(keyCode) {
			case "n":
				fallthrough
			case " ":
				if t.State == sdl.RELEASED {
					nextImg()
				}
			case "p":
				if t.State == sdl.RELEASED {
					prevImg()
				}
			case "q":
				running = !ctrl

			// meh; we could keyCode == sdl.K_ENTER (~)
			// (NOTE: github doesn't display the rune)
			case "":
				zoom = 1.
				clickX = 0
				clickY = 0
				drawImg()
			}

		case *sdl.MouseButtonEvent:
			if t.Type == sdl.MOUSEBUTTONDOWN {
				clicking = true
			} else if t.Type == sdl.MOUSEBUTTONUP {
				clicking = false
			}

		case *sdl.MouseMotionEvent:
			if clicking {
				clickX += t.XRel
				clickY += t.YRel
				// TODO: be smarter (there's some lag)
				drawImg()
			}

		case *sdl.MouseWheelEvent:
			if t.X == 0 && ctrl {
				if t.Y >= 0 {
					zoom /= 0.9
				} else {
					zoom *= 0.9
				}
				drawImg()
			} else if !ctrl {
				if t.Y >= 0 {
					prevImg()
				} else {
					nextImg()
				}
			}
		}
	}
}

func main() {
	mkWindow()
	loadAndDrawImg()
	mainLoop()
	cleanup()
}
