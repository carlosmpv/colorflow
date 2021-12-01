package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"math"
	"os"
	"strconv"
)

type Area struct{ x1, x2, y1, y2 int }

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	imData, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return imData, nil
}

type Scan func()

func getColorMean(colors []color.Color) color.Color {
	rs, gs, bs := 0.0, 0.0, 0.0
	l := float64(len(colors))

	for _, c := range colors {
		r, g, b, _ := c.RGBA()

		rs += float64(r)
		gs += float64(g)
		bs += float64(b)
	}

	return color.RGBA{
		R: uint8(rs / l),
		G: uint8(gs / l),
		B: uint8(bs / l),
		A: 0xff,
	}
}

func getColorDistance(c1, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return math.Sqrt(math.Pow(float64(r2-r1), 2) + math.Pow(float64(g2-g1), 2) + math.Pow(float64(b2-b1), 2))
}

func getColorsInArea(img image.Image, a Area) []color.Color {
	colors := []color.Color{}

	for x := a.x1; x < a.x2; x++ {
		for y := a.y1; y < a.y2; y++ {
			colors = append(colors, img.At(x, y))
		}
	}

	return colors
}

func (a Area) Neighbors() []Area {
	xs, ys := (a.x2 - a.x1), (a.y2 - a.y1)

	return []Area{
		{a.x1 - xs, a.x1, a.y1 - ys, a.y1},
		{a.x1, a.x2, a.y1 - ys, a.y1},
		{a.x2, a.x2 + xs, a.y1 - ys, a.y1},
		{a.x2, a.x2 + xs, a.y1, a.y2},

		{a.x2, a.x2 + xs, a.y2, a.y2 + ys},
		{a.x1, a.x2, a.y2, a.y2 + ys},
		{a.x1 - xs, a.x1, a.y2, a.y2 + ys},
		{a.x1 - xs, a.x1, a.y1, a.y2},
	}
}

func getClosestNeighbor(img image.Image, a Area, revealed []Area) int {
	colors := getColorsInArea(img, a)
	mean := getColorMean(colors)

	minDst := -1.0
	index := -1

	for i, nb := range a.Neighbors() {
		c := getColorsInArea(img, nb)
		dst := getColorDistance(mean, getColorMean(c))

		if nb.x1 < 0 || nb.y1 < 0 || nb.x2 > img.Bounds().Dx() || nb.y2 > img.Bounds().Dy() {
			break
		}

		known := false
		for _, rva := range revealed {
			if nb.x1 == rva.x1 && nb.y1 == rva.y1 {
				known = true
				break
			}
		}

		if known {
			continue
		}

		if dst < minDst || minDst == -1 {
			minDst = dst
			index = i
		}
	}

	return index
}

func findPath(img image.Image, area Area, steps int) []Area {
	path := []Area{
		area,
	}

	for i := 0; i < steps; i++ {
		closest := getClosestNeighbor(img, path[i], path)

		if closest < 0 {
			last := path[len(path)-1]

			smallerSeed := Area{}
			smallerSeed.x1 = last.x1
			smallerSeed.y1 = last.y1
			smallerSeed.x2 = last.x1 + ((last.x2 - last.x1) / 2)
			smallerSeed.y2 = last.y1 + ((last.y2 - last.y1) / 2)

			if smallerSeed.x2 == smallerSeed.x1 || smallerSeed.y2 == smallerSeed.y1 {
				break
			}

			return append(path, findPath(img, smallerSeed, steps)...)
		}

		path = append(path, path[i].Neighbors()[closest])
	}

	return path
}

func main() {
	p := os.Args[1]

	x1, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	x2, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}

	y1, err := strconv.Atoi(os.Args[4])
	if err != nil {
		log.Fatal(err)
	}

	y2, err := strconv.Atoi(os.Args[5])
	if err != nil {
		log.Fatal(err)
	}

	steps, err := strconv.Atoi(os.Args[6])
	if err != nil {
		log.Fatal(err)
	}

	img, err := loadImage(p)
	if err != nil {
		panic(err)
	}

	path := findPath(img, Area{x1, x2, y1, y2}, steps)

	nImg := image.NewRGBA(img.Bounds())

	for _, a := range path {
		for x := a.x1; x < a.x2; x++ {
			for y := a.y1; y < a.y2; y++ {
				nImg.Set(x, y, img.At(x, y))
			}
		}
	}

	out, err := os.Create("out.jpeg")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	jpeg.Encode(out, nImg, &jpeg.Options{
		Quality: jpeg.DefaultQuality,
	})

}
