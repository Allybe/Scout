package demo

import (
	"Scout/src/utils"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	_ "image/png" // Registers the PNG decoder with image.Decode
	"os"
	"path/filepath"

	"github.com/andygrunwald/vdf"
)

type Map struct {
	PosX  float64 `json:"pos_x,string"`
	PosY  float64 `json:"pos_y,string"`
	Scale float64 `json:"scale,string"`
}

type Position struct {
	PosX float64
	PosY float64
}

func (m Map) Translate(x, y float64) (float64, float64) {
	return x - m.PosX, m.PosY - y
}

func (m Map) TranslateScale(x, y float64) (float64, float64) {
	x, y = m.Translate(x, y)
	return x / m.Scale, y / m.Scale
}

func GetMapData(mapName string) Map {
	path, err := filepath.Abs("./src/demo/_assets/metadata/%s.txt")
	utils.CheckError(err)
	file, err := os.Open(fmt.Sprintf(path, mapName))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	valveData, err := vdf.NewParser(file).Parse()
	if err != nil {
		panic(err)
	}

	b, err := json.Marshal(valveData)
	if err != nil {
		panic(err)
	}

	var data map[string]Map

	err = json.Unmarshal(b, &data)
	if err != nil {
		panic(err)
	}

	mapInfo, ok := data[mapName]
	if !ok {
		panic(fmt.Sprintf("failed to get map info.json entry for %q", mapName))
	}

	return mapInfo

}

func GetMapFile(mapName string) string {
	path, err := filepath.Abs("./src/demo/_assets/radar/%s_radar_psd.png")
	utils.CheckErrorm(err, "Failed to get absolute path of radar image.")
	return fmt.Sprintf(path, mapName)
}

func GetImage(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return img
}

func DrawPointsAndSave(img image.Image, pts []Position, outPath string) error {
	// Create a new RGBA canvas matching the bounds of the original image
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	// Draw the original radar image onto the new canvas
	draw.Draw(rgba, bounds, img, image.Point{0, 0}, draw.Src)

	// Define a color for the dots (e.g., solid Red)
	dotColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	dotRadius := 4

	// Iterate through the points and draw a circle for each
	for _, p := range pts {
		x := int(p.PosX)
		y := int(p.PosY)

		for dx := -dotRadius; dx <= dotRadius; dx++ {
			for dy := -dotRadius; dy <= dotRadius; dy++ {
				// Use the Pythagorean theorem to draw a rough circle
				if dx*dx+dy*dy <= dotRadius*dotRadius {
					rgba.Set(x+dx, y+dy, dotColor)
				}
			}
		}
	}

	// Create the output file
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Encode and save the image as a PNG
	return png.Encode(outFile, rgba)
}
