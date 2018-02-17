package main

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// // NoAurora = -1, Unsure = 0, Aurora = 1
// on closer rereading: NoAurora = 0, Aurora = 1
const (
	NoAurora int = iota
	// Unsure
	Aurora
)

// Directory holds a directory and a classification to assign to the images in that directory
type Directory struct {
	DirectoryName  string
	Classification int
}

// Histogram holds the image color data and classification, as well as filename for debugging purposes
type Histogram struct {
	FileName       string
	Data           []float64
	Min            float64
	Max            float64
	Classification int
}

// Stringify returns a string representation of the histogram as specified in the assignment
func (hist Histogram) Stringify() string {
	return strconv.Itoa(hist.Classification) + " " + strings.Trim(fmt.Sprintf("%0.10f", hist.Data), "[]") + " " + hist.FileName
}

func runHistogram() {

	histograms := []Histogram{}

	directories := []Directory{
		{
			DirectoryName:  "./images/aurora",
			Classification: Aurora,
		},
		{
			DirectoryName:  "./images/noaurora",
			Classification: NoAurora,
		},
		// {
		// 	DirectoryName:  "./images/unsure",
		// 	Classification: Unsure,
		// },
	}

	for _, directory := range directories {
		files, err := ioutil.ReadDir(directory.DirectoryName)
		if err != nil {
			log.Fatal(err)
		}
		tempHistograms := make([]Histogram, len(files))
		badImports := make([]int, len(files))

		var wg sync.WaitGroup
		wg.Add(len(files))
		for index, file := range files {
			go func(index int, fileName string) {
				values, min, max, err := generateHistogram(directory.DirectoryName + "/" + fileName)
				if err == nil {
					tempHistograms[index] = Histogram{
						FileName:       fileName,
						Classification: directory.Classification,
						Data:           values,
						Min:            min,
						Max:            max,
					}
					badImports[index] = -1
				} else {
					badImports[index] = 1
				}
				wg.Done()
			}(index, file.Name())
		}
		// resync threads
		wg.Wait()
		// get rid of files in the directory that didn't import correctly
		for index, value := range badImports {
			if value == 1 {
				tempHistograms = append(tempHistograms[:index], tempHistograms[index+1:]...)
			}
		}
		histograms = append(histograms, tempHistograms...)
	}

	globalMin := 999999999.0
	globalMax := 0.0

	for _, histogram := range histograms {
		if histogram.Min < globalMin {
			globalMin = histogram.Min
		}
		if histogram.Max > globalMax {
			globalMax = histogram.Max
		}
	}

	histogramsSerialized := ""

	for _, histogram := range histograms {
		normalizedData := make([]float64, len(histogram.Data))
		for i := 0; i < len(histogram.Data); i++ {
			normalizedData[i] = (histogram.Data[i] - globalMin) / (globalMax - globalMin)
		}
		histogram.Data = normalizedData
		histogramsSerialized += histogram.Stringify() + "\n"
	}

	f, err := os.Create("histograms.dat")
	defer f.Close()
	if err != nil {
		panic("could not write output file.")
	}
	f.WriteString(histogramsSerialized)
	fmt.Println("Finished processing histograms. Saved as histograms.dat")

}

//generateHistogram returns the histogram data, the min value, and the max value (for normalization)
func generateHistogram(fileName string) ([]float64, float64, float64, error) {
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		panic("I just saw this file and now I can't open it")
	}
	img, _, err := image.Decode(file)
	if err != nil {
		// panic("couldn't decode this image! " + err.Error())
		return nil, 0.0, 0.0, errors.New("decode error")
	}
	rect := img.Bounds()
	r := make([]float64, 256)
	g := make([]float64, 256)
	b := make([]float64, 256)
	for i := 0; i < rect.Max.X; i++ {
		for j := 0; j < rect.Max.Y; j++ {
			rCurrent, gCurrent, bCurrent, _ := img.At(i, j).RGBA()
			r[rCurrent/256]++
			g[gCurrent/256]++
			b[bCurrent/256]++
		}
	}
	result := append(append(r, g...), b...)
	tempSorted := make([]float64, len(result))
	copy(tempSorted, result)
	sort.Float64s(tempSorted)
	return result, tempSorted[0], tempSorted[len(tempSorted)-1], nil
}

func loadHistograms(fileName string) []Histogram {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	histogramImport := []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		histogramImport = append(histogramImport, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	histograms := make([]Histogram, len(histogramImport))

	// let's put those strings back into Histogram structs...
	var wg sync.WaitGroup
	wg.Add(len(histogramImport)) // expect len(histogramImport) number of threads to complete data load
	for index, hist := range histogramImport {
		go func(index int, hist string) {
			splittingMeSoftly := strings.Split(hist, " ")
			class, _ := strconv.Atoi(splittingMeSoftly[0])
			data := []float64{}
			for _, str := range splittingMeSoftly[1:768] {
				parseley, _ := strconv.ParseFloat(str, 64)
				data = append(data, parseley)
			}
			newHistogram := Histogram{
				Classification: class,
				Data:           data,
				FileName:       strings.Join(splittingMeSoftly[769:], " "),
			}
			histograms[index] = newHistogram
			wg.Done()
		}(index, hist)
	}

	// resync all loader threads
	wg.Wait()
	fmt.Printf("%s loaded and deserialized into %d histograms.\n", os.Args[2], len(histograms))

	return histograms
}

func shuffleHistograms(input []Histogram) []Histogram {
	output := make([]Histogram, len(input))
	perm := rand.Perm(len(input))
	for index, rand := range perm {
		output[rand] = input[index]
	}
	return output
}
