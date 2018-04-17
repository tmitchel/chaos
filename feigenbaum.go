package main

import (
    "os"
    "log"
    "fmt"
    "time"
    "math"
    "flag"
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

    make_plot := flag.Bool("plot", false, "Create pdf of Feigenbaum Diagram")
    r_print := flag.Float64("r", 2., "Value of r to print (must be less than 4)")
    x0_print := flag.Float64("x0", 0.5, "Value of x0 to print")
    n_iter := flag.Int("n", 300, "Number of iterations to complete")

    flag.Parse()
    start := time.Now()
    
    // ugly way to deal with errors
    if *r_print > 4 || *r_print < 0 {
        fmt.Println("r must be in the range [0., 4.]")
        os.Exit(1)
    } else if *x0_print > 1 || *x0_print < 0 {
        fmt.Println("x0 must be in the range [0., 1.]")
        os.Exit(1)
    }

    // 2d array of channels to hold results
    comps := make([][]chan float64, 0)

    for r := 0.; r < 4.; r += 0.001 { 
        temp := make([]chan float64, 0)
        for x0 := 0.; x0 < 1.; x0 += 0.01 { 
            // start the iteration and place return channel in array
            temp = append(temp, feig_gen(r, x0))
        }
        comps = append(comps, temp)
    }

    calc_time := time.Since(start)

    op_plot := do_plotting
    if *make_plot {
        op_plot(&comps, n_iter)
    } else if *r_print > 0 && *r_print < 3.569455 {
        conv_print(&comps, n_iter, r_print, x0_print)
    } else {
        chaos_print(&comps, n_iter, r_print, x0_print)
    }

    elapsed := time.Since(start)
    log.Printf("Time spent on calculation: %s", calc_time)
    log.Printf("Time spent on other: %s", elapsed - calc_time)
    log.Printf("Processing completed in: %s", elapsed)

}

/////////////////////////////////////////////////////////////////
// Purpose: Handle producing the pdf of the Feigenbaum Diagram //
// Return: Nothing (pdf saved to system)                       //
/////////////////////////////////////////////////////////////////
func do_plotting(input *[][]chan float64, n *int) {
    // format data to be plotted
    total_len := len(*input) * len((*input)[0])
    points := make(plotter.XYs, total_len)
    for indr, compr := range *input {
        for indx, compx := range compr {
            // want result after 300 iteration, so pop off first 299 objects in the channel
            for i := 0; i < *n; i++ {
                <- compx
            }
            // convert 2d array indices into a 1d array index
            index := indr*indx + indx
            points[index].X = 4 * float64(indr) / float64(len(*input))
            points[index].Y = <- compx
        }
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
    // s, err := plotter.NewScatter(fillXY(&comps))
    s, err := plotter.NewScatter(points)
    if err != nil {
        log.Fatal(err)
    }

    // scatter plot related formatting
    s.GlyphStyle.Color = color.RGBA{R: 128, B: 0, G: 0}
    s.GlyphStyle.Radius = vg.Points(0.5)
    s.GlyphStyle.Shape = draw.CircleGlyph{}

    // add it and save
    p.Add(s)
    p.Save(600, 400, "feigenbaum.pdf")
}

////////////////////////////////////////////////////////////
// Purpose: Handle the printing of values/convergence for //
// r not such that the system is in the chaotic region    //
// Return: Nothing (Printing to console)                  //
////////////////////////////////////////////////////////////
func conv_print(input *[][]chan float64, n_iter *int, r_print, x0_print *float64) {
    r := int(*r_print * 1000)
    x0 := int((*x0_print * 100))

    // hard-code to compare with the book when r=2, x0=0.3
    if *r_print == 2. {
        for i := 0; i < *n_iter; i++ {
            current := <-(*input)[r][x0]
            fmt.Printf("Asympt: %6.4f; Xn = %6.4f with diff %6.4f after %v iterations\n", 0.5, current, math.Abs(current-0.5), i+1)
        }
    } else if *r_print == 3.2 {
        // hard-code to compare with the book when r=3.2, x0=0.3
        for i := 0; i < *n_iter; i++ {
            current := <-(*input)[r][x0]
            if math.Abs(current-0.51304) < math.Abs(current-0.79946) {
                fmt.Printf("Asympt: %6.4f; Xn = %6.4f with diff %6.5f after %v iterations\n", 0.51304, current, math.Abs(current-0.51304), i+1)
            } else {
                fmt.Printf("Asympt: %6.4f; Xn = %6.4f with diff %6.5f after %v iterations\n", 0.79946, current, math.Abs(current-0.79946), i+1)
            }
        }
    } else {
        // not sure where all the attractors are so just print values
        for i := 0; i < *n_iter; i++ {
            fmt.Printf("Xn = %6.4f after %v iterations", <-(*input)[r][x0], i+1)
        }
    }
}

//////////////////////////////////////////////////////////
// Purpose: Handle the printing when r is such that the //
// system is not in the chaotic region                  //
// Return: Nothing (Printing to console)                //
//////////////////////////////////////////////////////////
func chaos_print(input *[][]chan float64, n_iter *int, r_print, x0_print *float64) {
    r := int(*r_print * 1000)
    x0 := int((*x0_print * 100))

    // print the difference between requested x0 and one nearby
    for i := 1; i < *n_iter; i++ {
        Xn := <-(*input)[r][x0]
        Xnp := <-(*input)[r][x0+1]
        fmt.Printf("Xn = %6.4f; Xn' = %6.4f; delta(X) = %6.4f\n", Xn, Xnp, math.Abs(Xn-Xnp))
    }
}


// TODO: put all of data into this struct to make accessing cleaner
// TODO: make all current functions into methods of data_holder struct
type data_holder struct {
    r, x0 int
    input *[][] chan float64
}

func (d data_holder) r_length() int {
    return len(*d.input)
}

func (d data_holder) x_length() int {
    return len((*d.input)[0])
}

func (d data_holder) get_r(r int) []chan float64 {
    return (*d.input)[r*1000]
}

func (d data_holder) get(r, x int) chan float64 {
    return (*d.input)[r*1000][x*1000]
}

func (d data_holder) get_iteration_n(n ...int) chan float64 {
    var pop_chan chan float64
    if len(n) == 1 {
        pop_chan = (*d.input)[d.r*1000][d.x0*1000]
    } else if len(n) == 3 {
        pop_chan = (*d.input)[n[1]*1000][n[2]*1000]
    } else {
        fmt.Println("Wrong number of parameters. (Will add better handling later.)")
        os.Exit(1)
    }
    for i := 0; i < n[0]; i++ {
        <-pop_chan
    }
    return pop_chan
}
