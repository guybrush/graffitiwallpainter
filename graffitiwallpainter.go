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
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var interval time.Duration
var xOffset int
var yOffset int
var imageFile string
var graffitiFile string
var explorerURL string
var runOnce bool
var max int

// Build information. Populated at build-time
var (
	Version   = "undefined"
	GitDate   = "undefined"
	GitCommit = "undefined"
	BuildDate = "undefined"
	GoVersion = runtime.Version()
)

func main() {
	flag.StringVar(&explorerURL, "url", "https://beaconcha.in/api/v1/graffitiwall", "url to graffitiwall page of explorer")
	flag.DurationVar(&interval, "interval", time.Duration(32*12)*time.Second, "interval in which the graffiti-file will be updated")
	flag.StringVar(&imageFile, "image", "image.png", "path to image")
	flag.IntVar(&xOffset, "x", 0, "offset x")
	flag.IntVar(&yOffset, "y", 0, "offset y")
	flag.StringVar(&graffitiFile, "graffiti", "graffiti.txt", "path to graffiti-file")
	flag.BoolVar(&runOnce, "once", false, "if set, write all pixels to file and exit")
	flag.IntVar(&max, "max", 0, "if set to >0, only N pixels will be writen to file")
	flag.Parse()

	if max < 0 {
		logrus.Fatal("max must be >=0")
	}
	if interval < time.Second*12 {
		interval = time.Second * 12
		logrus.Warnf("setting interval to %v, lower value does not make sense\n", interval)
	}

	logrus.WithFields(logrus.Fields{
		"version":      Version,
		"gitDate":      GitDate,
		"gitCommit":    GitCommit,
		"buildDate":    BuildDate,
		"goVersion":    GoVersion,
		"explorerURL":  explorerURL,
		"interval":     interval,
		"imageFile":    imageFile,
		"xOffset":      xOffset,
		"yOffset":      yOffset,
		"graffitiFile": graffitiFile,
	}).Info("starting graffitiwallpainter")

	gwWant, err := readImage(imageFile, xOffset, yOffset)
	if err != nil {
		logrus.WithError(err).Fatal("couldnt read image")
	}

	for {
		err := run(explorerURL, gwWant)
		if err != nil {
			logrus.WithError(err).Error("run error")
		}
		if runOnce {
			return
		}
		time.Sleep(interval)
	}
}

func run(explorerURL string, gwWant map[string]string) error {
	gwIs, err := getGraffitiwall(explorerURL)
	if err != nil {
		return err
	}
	res := []string{}
	for xy, color := range gwWant {
		colorIs, exists := gwIs[xy]
		if color == "ffffff" && !exists {
			continue
		}
		if !exists || colorIs != color {
			res = append(res, fmt.Sprintf("  - \"graffitiwall:%s:#%s\"\n", xy, color))
		}
	}
	if len(res) == 0 {
		logrus.Infof("all pixels match the image!")
		return nil
	}
	if max > 0 {
		currmax := max
		if currmax > len(res) {
			currmax = len(res)
		}
		res = res[:currmax]
		rand.Shuffle(currmax, func(i, j int) { res[i], res[j] = res[j], res[i] })
	}
	err = ioutil.WriteFile(graffitiFile, []byte("random:\n"+strings.Join(res, "")), 0644)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{"pixelsLeft": len(res)}).Infof("updated graffiti")
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
	if w+offsetX > 1000 || h+offsetY > 1000 {
		return nil, fmt.Errorf("image or offset is too big (%v, %v, %v, %v)", w, h, offsetX, offsetY)
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

type APIResponse struct {
	Status string `json:"status"`
	Data   []struct {
		X         uint32 `json:"x"`
		Y         uint32 `json:"y"`
		Color     string `json:"color"`
		Validator uint64 `json:"validator"`
		Slot      uint64 `json:"slot"`
	}
}

func getGraffitiwall(explorerURL string) (map[string]string, error) {
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Get(explorerURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	dataBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data APIResponse
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		return nil, err
	}
	if data.Status != "OK" {
		return nil, fmt.Errorf("api status not ok: %s", data.Status)
	}
	res := make(map[string]string)
	for _, g := range data.Data {
		res[fmt.Sprintf("%d:%d", g.X, g.Y)] = g.Color
	}
	return res, nil
}
