package main

import (
    "log"
    "fmt"
    "time"
    "math"
    "image/color"
    "gonum.org/v1/plot/vg/draw"
    "gonum.org/v1/plot"
    "gonum.org/v1/plot/vg"
    "gonum.org/v1/plot/plotter"
)

////////////////////////////////////////////////////////
// Purpose: Continuously iterate for a given (r, x0)  //
// Return: Channel containing the result of each      //
// iteration (channel will start blocking at          //
//   buffer = 400 entries)                            //
////////////////////////////////////////////////////////
func feig_gen(r float64, x0 float64) chan float64 {

    // channel to return
    c := make(chan float64, 400)

    // start goroutine to fill channel
    go func() {
        for i := x0; ; i = r * i * (1 - i) {
            // place result in the channel
            c <- i
        } 
    } ()
    return c
}

///////////////////////////////////////////////////////
// Purpose: Generate Feigenbaum Diagram and assess   //
// the convergence                                   //
// Return: A (large) pdf of the diagram and prints   //
// useful numbers for convergence to the console     //
/////////////////////////////////////////////////////// 
func main() {
    start := time.Now()
    
    // 2d array of channels to hold results
    comps := make([][]chan float64, 0)

    for r := 0.; r < 3.999; r += 0.001 {
        temp := make([]chan float64, 0)
        for x0 := 0.; x0 < 1.; x0 += 0.01 { 
            // start the iteration and place return channel in array
            temp = append(temp, feig_gen(r, x0))
        }
        comps = append(comps, temp)
    }

    calc_time := time.Since(start)

    // format data to be plotted
    fillXY := func(input *[][]chan float64) plotter.XYs {
        total_len := len(*input) * len((*input)[0])
        points := make(plotter.XYs, total_len)
        for indr, compr := range *input {
            for indx, compx := range compr {
                // want result after 300 iteration, so pop off first 299 objects in the channel
                printed := false
                for i := 0; i < 300; i++ {
                    // check convergence in 1 < r < 3 range (should conv to 0.5 in 4 iterations)
                    if curr_r := 4*float64(indr)/float64(len(*input)); curr_r == 2. {
                        temp := <- compx
                        if math.Abs(temp - 0.5) < 0.001 && !printed && indx == 30 {
                            fmt.Println("Convergence to Xn", temp, "with accuracy", math.Abs(temp - 0.5), "after", i, "iterations.")
                            printed = true
                        }
                    } else if curr_r == 3.2 {
                        // Should oscillate between (0.51304 and 0.79946) and does after ~30 iterations
                        temp := <- compx
                        if indx == 30 {
                            fmt.Println("r=3.2 and Xo=0.3 has Xn", temp, "after", i, "iterations.")
                        } 
                    } else {
                        <- compx
                    }
                }
                // convert 2d array indices into a 1d array index
                index := indr*indx + indx
                points[index].X = 4 * float64(indr) / float64(len(*input))
                points[index].Y = <- compx
            }
        }
        return points
    }

    // define a plot
    p, err := plot.New()
    if err != nil {
        log.Fatal(err)
    }

    // do some nice formatting
    p.Title.Text = "Feigenbaum Diagram"
    p.X.Label.Text = "r"
    p.Y.Label.Text = "x"
    p.Add(plotter.NewGrid())

    // fill the scatter plot with points
    s, err := plotter.NewScatter(fillXY(&comps))
    if err != nil {
        log.Fatal(err)
    }

    // scatter plot related formatting
    s.GlyphStyle.Color = color.RGBA{R: 128, B: 0, G: 0}
    s.GlyphStyle.Radius = vg.Points(0.5)
    s.GlyphStyle.Shape = draw.CircleGlyph{}

    // add it and save
    p.Add(s)
    p.Save(600, 400, "fullRange_feigenbaum.pdf")

    elapsed := time.Since(start)
    log.Printf("Time spent on calculation: %s", calc_time)
    log.Printf("Time spent on plotting: %s", elapsed - calc_time)
    log.Printf("Processing completed in: %s", elapsed)

}




