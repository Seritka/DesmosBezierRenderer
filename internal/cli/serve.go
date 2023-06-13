package cli

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"os"
	"sort"

	"github.com/dennwc/gotrace"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
	"gocv.io/x/gocv"
)

const FRAME_DIR = "./DesmosBezierRenderer_fork/frames"
const FILE_EXT = "png"
const COLOUR = "#2464b4"

var FRAME_LATEX = 0
var HEIGHT = 0
var WIDTH = 0

var LATEX = []string{}

func ServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Print the version number of DesmosBezierRenderer",
		RunE: func(cmd *cobra.Command, args []string) error {
			// opencv Canny, potrace, and latex list final fiber server
			dir, err := os.ReadDir(FRAME_DIR)
			if err != nil {
				panic(err)
			}
			FRAME_LATEX := len(dir)
			go get_expressionss(FRAME_LATEX)

			app := fiber.New()

			app.Get("/", func(c *fiber.Ctx) error {
				frame := int(c.QueryInt("frame", 0))
				if frame >= len(dir) {
					return c.SendString("Frame not found")
				} else {
					return c.SendString(fmt.Sprintf("%v", get_expressionss(frame)))
				}
			})

			app.Get("/calculator", func(c *fiber.Ctx) error {
				return c.Render("./DesmosBezierRenderer_fork/frontend/index.html", fiber.Map{"api_key": "dcb31709b452b1cf9dc26972add0fda6", "total_frames": len(dir),
					"download_images": false, "show_grid": true})
			})

			app.Listen(":3000")

			return nil
		},
	}
}

func reverse_int(n int) int {
	new_int := 0
	for n > 0 {
		remainder := n % 10
		new_int *= 10
		new_int += remainder
		n /= 10
	}
	return new_int
}

func get_contours(filename string) []float64 {
	frame := 0

	image := gocv.IMRead(filename, gocv.IMReadColor)
	gray := gocv.NewMat()
	edge := gocv.NewMat()
	gocv.CvtColor(image, &gray, gocv.ColorBGRToGray)

	gocv.Canny(gray, &edge, 30, 200)

	frame += 1
	img, err := image.ToImage()
	if err != nil {
		panic(err)
	}
	HEIGHT = img.Bounds().Size().Y
	WIDTH = img.Bounds().Size().X
	fmt.Printf("\r--> Frame %d/%d", frame, FRAME_LATEX)

	data, err := edge.DataPtrFloat64()
	if err != nil {
		panic(err)
	}
	sort.Reverse(sort.Float64Slice(data))

	return data
}

func Float64ArrayToByteArray(arr []float64) []byte {
	byteArr := make([]byte, len(arr)*8) // Each float64 occupies 8 bytes
	for i, val := range arr {
		bits := math.Float64bits(val)
		binary.LittleEndian.PutUint64(byteArr[i*8:], bits)
	}
	return byteArr
}

func get_trace(contours []float64) []gotrace.Path {
	for _, contour := range contours {
		if contour > 1 {
			contour = 1
		}
	}

	img, _, err := image.Decode(bytes.NewReader(Float64ArrayToByteArray(contours)))
	if err != nil {
		panic(err)
	}

	bm := gotrace.NewBitmapFromImage(img, nil)
	paths, err := gotrace.Trace(bm, nil)
	if err != nil {
		panic(err)
	}

	return paths
}

func get_latex(filename string) []string {
	latex := []string{}
	paths := get_trace(get_contours(filename))
	for _, curve := range paths {
		segments := curve.Curve
		l := len(segments) - 1
		start_point := segments[l].Pnt[2]
		for _, segment := range segments {
			if segment.Type == gotrace.TypeCorner {
				x1 := segment.Pnt[1]
				end_point := segment.Pnt[2]
				latex = append(latex, fmt.Sprintf("((1-t)%f+t%f,(1-t)%f+t%f)", start_point.X, x1.X, start_point.Y, x1.Y))
				latex = append(latex, fmt.Sprintf("((1-t)%f+t%f,(1-t)%f+t%f)", x1.X, end_point.X, x1.Y, end_point.Y))
			} else {
				x1 := segment.Pnt[0]
				x2 := segment.Pnt[1]
				x3 := segment.Pnt[2]
				latex = append(latex, fmt.Sprintf("((1-t)^3*%f+3*(1-t)^2*t*%f+3*(1-t)*t^2*%f+t^3*%f,(1-t)^3*%f+3*(1-t)^2*t*%f+3*(1-t)*t^2*%f+t^3*%f)", start_point.X, x1.X, x2.X, x3.X, start_point.Y, x1.Y, x2.Y, x3.Y))
			}
			start_point = segment.Pnt[2]
		}
	}
	return latex
}

type Exprs struct {
	id     string
	latex  string
	color  string
	secret bool
}

func get_expressionss(frame int) []Exprs {
	exprid := 0
	exprs := make([]Exprs, frame)
	for _, expr := range get_latex(fmt.Sprintf(FRAME_DIR+"/frame%d.%s", frame, FILE_EXT)) {
		exprid += 1
		exprs = append(exprs, Exprs{id: fmt.Sprintf("%d", exprid), latex: expr, color: COLOUR, secret: false})
	}
	return exprs
}

func init() {
	rootCmd.AddCommand(ServeCmd())
}
