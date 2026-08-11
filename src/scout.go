package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/msg"
)

var (
	mapData  Map
	mapName  string
	radarImg image.Image
	points   []Position
)

func drawPointsAndSave(img image.Image, pts []Position, outPath string) error {
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

func onSmokeStart(event events.SmokeStart) {
	playerName := "Unknown"
	if event.Grenade != nil && event.Grenade.Owner != nil {
		playerName = event.Grenade.Owner.Name
	}

	if playerName == "Selection55" || playerName == "Cheesenipps" || playerName == "__Byron" || playerName == "Conray13" || playerName == "Hunterr" {
		if event.Thrower.Team == 3 { // 2 t 3 ct
			return
		}
		x, y := mapData.TranslateScale(event.Position.X, event.Position.Y)
		points = append(points, Position{PosX: x, PosY: y})
		fmt.Printf("%s's threw smoke \n", playerName)
	}
}

func main() {
	filePath := "C:/Users/allison/Codespaces/CS/DemoFiles/2.dem"

	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	parser := demoinfocs.NewParser(f)

	parser.RegisterNetMessageHandler(func(message *msg.CSVCMsg_ServerInfo) {
		mapName = message.GetMapName()
		mapFile := getMapFile(mapName)
		radarImg = getImage(mapFile)
		mapData = getMapData(mapName)
		points = make([]Position, 0)
	})

	parser.RegisterEventHandler(onSmokeStart)

	err = parser.ParseToEnd()
	//parser.Close()

	if err != nil {
		panic(err)
	}

	outFilePath := "C:/Users/allison/Codespaces/CS/DemoFiles/radar_with_smokes.png"
	fmt.Printf("Drawing %d smoke points to %s...\n", len(points), outFilePath)

	err = drawPointsAndSave(radarImg, points, outFilePath)
	if err != nil {
		fmt.Printf("Failed to save image: %v\n", err)
	} else {
		fmt.Println("Successfully saved radar image!")
	}
}
