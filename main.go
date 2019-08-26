package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/gorilla/mux"
)

const (
	x                = 749
	y                = 743
	topLineAdjust    = 8
	bottomLineAdjust = 1.2
	lineSpacing      = 1.2
	maxWidth         = 746
	strokeSize       = 6
	savePath         = "memes/"
)

// Req is the JSON Request body.
type Req struct {
	Text string `json:"text"`
}

// Res is the JSON Response body.
type Res struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

func createMeme(w http.ResponseWriter, r *http.Request) {
	var req Req
	json.NewDecoder(r.Body).Decode(&req)

	if req.Text != "" {
		var title, body string
		rand.Seed(time.Now().UnixNano())
		w.Header().Set("Content-Type", "application/json")

		im, err := gg.LoadImage("img/bob.png")
		if err != nil {
			log.Fatal(err)
		}

		dc := gg.NewContext(x, y)
		if err := dc.LoadFontFace("font/impact.ttf", 84); err != nil {
			panic(err)
		}

		if strings.Contains(req.Text, ";") {
			title = strings.Split(req.Text, ";")[0]
			body = strings.Split(req.Text, ";")[1]
		} else {
			body = req.Text
		}

		dc.DrawImage(im, 0, 0)
		if title != "" {
			drawTextLine(dc, randomizeCapitalization(title), topLineAdjust)
		}
		drawTextLine(dc, randomizeCapitalization(body), bottomLineAdjust)
		var fileName = "sarcasm_bob_" + time.Now().Format("20060102150405") + ".png"
		dc.SavePNG(savePath + fileName)

		res := Res{
			ResponseType: "in_channel",
			Text:         r.Host + "/memes/" + fileName}
		json.NewEncoder(w).Encode(res)
	}
}

func randomizeCapitalization(label string) string {
	label = strings.ToLower(label)
	var newLabel string
	for _, letter := range strings.Split(label, "") {
		if len(letter) > 0 {
			if rand.Float32() < 0.5 {
				newLabel += strings.ToUpper(letter)
			} else {
				newLabel += letter
			}
		}
	}
	return newLabel
}

func drawTextLine(dc *gg.Context, label string, adjust float64) {
	dc.SetRGB(0, 0, 0)
	for dy := -strokeSize; dy <= strokeSize; dy++ {
		for dx := -strokeSize; dx <= strokeSize; dx++ {
			if dx*dx+dy*dy >= strokeSize*strokeSize {
				// give it rounded corners
				continue
			}
			x := x/2 + float64(dx)
			y := y/adjust + float64(dy)
			dc.DrawStringWrapped(label, x, y, float64(0.5), float64(0.5), float64(maxWidth), float64(lineSpacing), gg.AlignCenter)
		}
	}

	dc.SetRGB(1, 1, 1)
	dc.DrawStringWrapped(label, float64(x/2), float64(y/adjust), float64(0.5), float64(0.5), float64(maxWidth), float64(lineSpacing), gg.AlignCenter)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/sarcasm", createMeme).Methods("POST")
	router.PathPrefix("/memes/").Handler(http.StripPrefix("/memes/", http.FileServer(http.Dir("./memes"))))

	if os.Getenv("PORT") == "" {
		log.Fatal(http.ListenAndServe(":8000", router))
	} else {
		log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), router))
	}
}
