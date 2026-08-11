package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png" // Registers the PNG decoder with image.Decode
	"os"

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

//go:embed 	_assets/*
var fs embed.FS

func getMapData(mapName string) Map {
	file, err := fs.Open(fmt.Sprintf("_assets/metadata/%s.txt", mapName))
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

func getMapFile(mapName string) string {
	return fmt.Sprintf("C:/Users/allison/Codespaces/CS/Scout/src/_assets/radar/%s_radar_psd.png", mapName)
}

func getImage(path string) image.Image {
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
