package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"
)

var interval time.Duration
var xOffset int
var yOffset int
var imageFile string
var graffitiFile string

func main() {
	flag.DurationVar(&interval, "interval", time.Minute, "interval in which the graffiti-file will be updated")
	flag.StringVar(&imageFile, "image", "image.png", "path to image")
	flag.IntVar(&xOffset, "x", 0, "offset x")
	flag.IntVar(&yOffset, "y", 0, "offset y")
	flag.StringVar(&graffitiFile, "graffiti", "graffiti.txt", "path to graffiti-file")
	flag.Parse()
	for {
		err := run()
		if err != nil {
			fmt.Println("error", err)
		}
		time.Sleep(interval)
	}
}

func run() error {
	gwIs, err := getGraffitiwall()
	if err != nil {
		return err
	}
	gwWant, err := readImage(imageFile, xOffset, yOffset)
	if err != nil {
		return err
	}
	res := []string{}
	for xy, color := range gwWant {
		colorIs, exists := gwIs[xy]
		// fmt.Println(color, colorIs)
		if color == "ffffff" && !exists {
			continue
		}
		if !exists || colorIs != color {
			res = append(res, fmt.Sprintf("graffitiwall:%s:#%s", xy, color))
		}
	}
	g := res[rand.Intn(len(res))]
	err = ioutil.WriteFile(graffitiFile, []byte(g), 0644)
	if err != nil {
		return err
	}
	fmt.Printf("current graffiti: %v, pixels left: %v\n", g, len(res))
	return nil
}

func readImage(file string, offsetX, offsetY int) (map[string]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	res := make(map[string]string)

	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	// fmt.Println(w, h)
	if w+offsetX > 1000 || h+offsetY > 1000 {
		return nil, fmt.Errorf("image or offset is too big")
	}
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			c := src.At(x, y)
			r, g, b, _ := c.RGBA()
			r, g, b = r/0x101, g/0x101, b/0x101
			res[fmt.Sprintf("%d:%d", x+offsetX, y+offsetY)] = fmt.Sprintf("%02x%02x%02x", r, g, b)
		}
	}
	return res, nil
}

func getGraffitiwall() (map[string]string, error) {
	url := "https://beaconcha.in/graffitiwall"
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	dataRE := regexp.MustCompile(`var pixels = (.*)\n`)
	m := dataRE.FindAllStringSubmatch(string(data), -1)
	if len(m) < 1 || len(m[0]) < 2 {
		return nil, fmt.Errorf("could not read wall from beaconchain")
	}

	type graffitiJSONT []struct {
		X     uint32
		Y     uint32
		Color string
	}
	var graffitiJSON graffitiJSONT
	err = json.Unmarshal([]byte(m[0][1]), &graffitiJSON)
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	for _, g := range graffitiJSON {
		res[fmt.Sprintf("%d:%d", g.X, g.Y)] = g.Color
	}
	return res, nil
}
