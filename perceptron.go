package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/gosuri/uiprogress"
	chart "github.com/wcharczuk/go-chart"
	"gonum.org/v1/gonum/mat"
)

func runPerceptron() {
	if len(os.Args) < 3 {
		fmt.Println("Please enter a histogram file as an argument after 'run'.")
		os.Exit(1)
	}
	histograms := shuffleHistograms(loadHistograms(os.Args[2]))
	trainingSplit := int64(float64(len(histograms)) * 0.8)
	trainingData := histograms[:trainingSplit]
	validationData := histograms[trainingSplit:]
	fmt.Printf("Out of %d histograms: %d are for training, %d are for validation.\n", len(histograms), len(trainingData), len(validationData))
	histograms = nil

	featureLength := len(trainingData[0].Data)

	weights := make([]float64, featureLength)
	for i := 0; i < len(weights); i++ {
		weights[i] = rand.Float64()
	}

	learningRate, _ := strconv.ParseFloat(os.Args[4], 64)
	epochs, _ := strconv.Atoi(os.Args[3])
	errors := make([]float64, epochs)
	accuracy := make([]float64, epochs)

	xAxis := make([]float64, epochs)
	for i := 1; i < epochs; i++ {
		xAxis[i] = float64(i + 1)
	}

	uiprogress.Start()
	bar := uiprogress.AddBar(epochs).AppendCompleted().PrependElapsed()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Epoch (%d/%d)", b.Current(), epochs)
	})

	for epoch := 0; epoch < epochs; epoch++ {
		errorSet := 0.0
		for _, histogram := range trainingData {
			u := mat.NewVecDense(featureLength, histogram.Data)
			v := mat.NewVecDense(featureLength, weights)
			output := activation(mat.Dot(u, v))
			error := histogram.Classification - output
			errorSet += float64(error)
			for index := range weights {
				weights[index] += learningRate * float64(error) * float64(histogram.Data[index])
			}
		}
		errorSet /= float64(len(trainingData))
		errors[epoch] = errorSet

		validationTotal := len(validationData)
		validationSuccesses := 0

		for _, histogram := range validationData {
			u := mat.NewVecDense(featureLength, histogram.Data)
			v := mat.NewVecDense(featureLength, weights)
			output := activation(mat.Dot(u, v))
			if histogram.Classification == output {
				validationSuccesses++
			}
		}

		accuracy[epoch] = 100.0 * (float64(validationSuccesses) / float64(validationTotal))

		bar.Incr()
	}

	uiprogress.Stop()

	validationTotal := len(validationData)
	validationSuccesses := 0

	for _, histogram := range validationData {
		u := mat.NewVecDense(featureLength, histogram.Data)
		v := mat.NewVecDense(featureLength, weights)
		output := activation(mat.Dot(u, v))
		if histogram.Classification == output {
			validationSuccesses++
		}
	}

	fmt.Printf("Validation successes: %d/%d (%0.2f%% accuracy)\n", validationSuccesses, validationTotal, 100.0*(float64(validationSuccesses)/float64(validationTotal)))

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name:      "Epochs",
			NameStyle: chart.StyleShow(),
			Style: chart.Style{
				Show: true,
			},
			TickPosition: chart.TickPositionBetweenTicks,
		},
		YAxis: chart.YAxis{
			Name:      "Accuracy (Percent)",
			NameStyle: chart.StyleShow(),
			Style: chart.Style{
				Show: true,
			},
		},
		YAxisSecondary: chart.YAxis{
			Name:      "Error",
			NameStyle: chart.StyleShow(),
			Style: chart.Style{
				Show: true,
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name:    "Accuracy",
				XValues: xAxis,
				YValues: accuracy,
			},
			chart.ContinuousSeries{
				Name:    "Error",
				XValues: xAxis,
				YValues: errors,
				YAxis:   chart.YAxisSecondary,
			},
		},
	}
	graph.Elements = []chart.Renderable{
		chart.LegendThin(&graph),
	}

	f, err := os.Create("rate_" + os.Args[4] + "_over_" + os.Args[3] + "_epochs.png")
	defer f.Close()
	if err == nil {
		buffer := bytes.NewBuffer([]byte{})
		graph.Render(chart.PNG, buffer)
		f.Write(buffer.Bytes())
	}

	results, err := os.Create("rate_" + os.Args[4] + "_over_" + os.Args[3] + "_epochs.dat")
	defer results.Close()
	if err != nil {
		panic("could not write output file.")
	}

	resultString := ""
	for _, weight := range weights {
		resultString += fmt.Sprintf("%0.10f", weight) + " "
	}
	results.WriteString(resultString)

}

func activation(num float64) int {
	if num < 0.5 {
		return 0
	}
	return 1
}
