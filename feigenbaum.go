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
    // "gonum.org/v1/plot/plotutil"
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
    find_liapunov := flag.Bool("liap", false, "Find Liapunov exponent")
    find_bifunction := flag.Bool("bi", false, "Find the bifunction points (1 and 2 only)")
    r_print := flag.Float64("r", 2., "Value of r to print (must be less than 4)")
    x0_print := flag.Float64("x0", 0.5, "Value of x0 to print")
    n_iter := flag.Int("n", 300, "Number of iterations to complete")

    flag.Parse()
    start := time.Now()
    
    // ugly way to deal with errors. fix later
    if *r_print > 4 || *r_print < 0 {
        fmt.Println("r must be in the range [0., 4.]")
        os.Exit(1)
    } else if *x0_print > 1 || *x0_print < 0 {
        fmt.Println("x0 must be in the range [0., 1.]")
        os.Exit(1)
    }

    data := make([][]chan float64, 0)
    results := data_holder{r: *r_print, x0: *x0_print, input: &data}

    for r := 0.; r < 4.; r += 0.001 {
        temp := make([]chan float64, 0)
        for x0 := 0.; x0 < 1.; x0+= 0.01 {
            // start the iteration and place return channel in data array
            temp = append(temp, feig_gen(r, x0))
        }
        *results.input = append(*results.input, temp)
    }

    calc_time := time.Since(start)

    if *make_plot {
        results.do_plotting(n_iter)
    } else if *find_liapunov {
        results.plot_liapunov(n_iter, x0_print, 0.001, 4.0)
    } else if *find_bifunction {
        results.bifunction(x0_print)
    } else if *r_print > 0 && *r_print < 3.569455 {
        results.conv_print(n_iter, r_print, x0_print)
    } else {
        results.chaos_print(n_iter)
    }

    elapsed := time.Since(start)
    log.Printf("Time spent on calculation: %s", calc_time)
    log.Printf("Time spent on other: %s", elapsed - calc_time)
    log.Printf("Processing completed in: %s", elapsed)

}



/////////////////////////////////////////////////////
// Purpose: Hold 2d channel array and provide some // 
// nice functions for accessing elements           //
// (I'd probably put this all in a different file, //
// but that would make the grading difficult)      //
/////////////////////////////////////////////////////
type data_holder struct {
    r, x0 float64
    input *[][] chan float64
}

func (d data_holder) bifunction(x0 *float64) {
    x_ind := d.get_idx(*x0)
    found_one, found_two := false, false
    for i, arr := range (*d.input) {
        if i == 0 {
            continue
        }
        r := float64(i)/1000.
        for j := 0; j < 297; j++ {
            <-arr[x_ind]
        }
        comp_r3 := <-arr[x_ind]
        comp_r2 := <-arr[x_ind]
        comp_r1 := <-arr[x_ind]
        curr := <-arr[x_ind]
        if !found_one {
            if math.Abs(curr - comp_r1) > 0.0001 {
                found_one = true
                fmt.Println("1->2 period bifurcation point:", r)
            }
        } else if found_one && !found_two {
            if math.Abs(curr - comp_r1) > 0.0001 && math.Abs(curr - comp_r2) > 0.0001 {
                fmt.Println("2->4 period bifurcation point:", r)
                found_two = true    
            }
        } else {
            if math.Abs(curr - comp_r1) > 0.0001 && math.Abs(curr - comp_r2) > 0.0001 && math.Abs(curr - comp_r3) > 0.0001 {
                fmt.Println("4->8 period bifurcation point:", r)
                found_two = true                
                break
            }
        }
    }
}

func (d data_holder) liapunov(n *int, x0 *float64) []float64 {
    expos := make([]float64, d.r_length())
    x_ind := d.get_idx(*x0)
    for i, arr := range (*d.input) {
        r := float64(i)/1000.
        expo := 0.
        <-arr[x_ind]
        for j := 0; j < *n; j++ {
            expo += math.Log(math.Abs(r - 2*r*(<- arr[x_ind])))
        }
        expos[i] = expo/float64(*n)
    }
    return expos
}

func (d data_holder) plot_liapunov(n *int, x0 *float64, r_init, r_fin float64) {
    expos := d.liapunov(n, x0)

    p, err := plot.New()
    if err != nil {
        log.Fatal(err)
    }

    pts := make(plotter.XYs, d.r_length())
    for i, _ := range pts {
        if i == 0 {
            continue
        }
        pts[i].X = float64(i)/1000.
        pts[i].Y = expos[i]
    }

    l, err := plotter.NewLine(pts)
    if err != nil {
        log.Fatal(err)
    }

    p.Add(l)
    // p.X.Min = 3.4
    p.Y.Min = -1
    p.Y.Max = 1

    if err := p.Save(600, 400, "liapunov.pdf"); err != nil {
        log.Fatal(err)
    }
}

/////////////////////////////////////////////////////////////////
// Purpose: Handle producing the pdf of the Feigenbaum Diagram //
// Return: Nothing (pdf saved to system)                       //
/////////////////////////////////////////////////////////////////
func (d data_holder) do_plotting(n *int) {
    // format data to be plotted
    total_len := len(*d.input) * len((*d.input)[0])
    points := make(plotter.XYs, total_len)
    for indr, compr := range *d.input {
        for indx, _ := range compr {
            // convert 2d array indices into a 1d array index
            index := indr*indx + indx
            points[index].X = 4 * float64(indr) / float64(len(*d.input))
            points[index].Y = <-d.get_iteration_n(*n, indr, indx)
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
func (d data_holder) conv_print(n *int, opt ...*float64) {
    var r, x0 int
    var r_print, x0_print float64
    if len(opt) == 0 {
        r = d.get_idr(d.r)
        x0 = d.get_idx(d.x0)
        r_print = d.r
        x0_print = d.x0
    } else {
        r_print = *opt[0]
        x0_print = *opt[1]
        r = d.get_idr(r_print)
        x0 = d.get_idx(x0_print)
    }

    // hard-code to compare with the book when r=2, x0=0.3
    if r_print == 2. {
        for i := 0; i < *n; i++ {
            current := <-(*d.input)[r][x0]
            fmt.Printf("Asympt: %6.4f; Xn = %6.4f with diff %6.4f after %v iterations\n", 0.5, current, math.Abs(current-0.5), i+1)
        }
    } else if r_print == 3.2 {
        // hard-code to compare with the book when r=3.2, x0=0.3
        for i := 0; i < *n; i++ {
            current := <-(*d.input)[r][x0]
            if math.Abs(current-0.51304) < math.Abs(current-0.79946) {
                fmt.Printf("Asympt: %6.4f; Xn = %6.4f with diff %6.5f after %v iterations\n", 0.51304, current, math.Abs(current-0.51304), i+1)
            } else {
                fmt.Printf("Asympt: %6.4f; Xn = %6.4f with diff %6.5f after %v iterations\n", 0.79946, current, math.Abs(current-0.79946), i+1)
            }
        }
    } else {
        // not sure where all the attractors are so just print values
        for i := 0; i < *n; i++ {
            fmt.Printf("Xn = %6.4f after %v iterations", <-(*d.input)[r][x0], i+1)
        }
    }
}

//////////////////////////////////////////////////////////
// Purpose: Handle the printing when r is such that the //
// system is not in the chaotic region                  //
// Return: Nothing (Printing to console)                //
//////////////////////////////////////////////////////////
func (d data_holder) chaos_print(n *int, opt ...float64) {
    var r, x0 int
    if len(opt) == 0 {
        r = d.get_idr(d.r)
        x0 = d.get_idx(d.x0)
    } else {
        r = d.get_idr(opt[0])
        x0 = d.get_idx(opt[1])
    }

    // print the difference between requested x0 and one nearby
    for i := 1; i < *n; i++ {
        Xn := <-(*d.input)[r][x0]
        Xnp := <-(*d.input)[r][x0+1]
        fmt.Printf("Xn = %6.4f; Xn' = %6.4f; delta(X) = %6.4f\n", Xn, Xnp, math.Abs(Xn-Xnp))
    }
}

/////////////////////////////////////////////////////////
// Purpose: Pop off a given number of entries from the //
// channel queue                                       //
// Return: channel after a given number of iterations  //
/////////////////////////////////////////////////////////
func (d data_holder) get_iteration_n(n int, opt ...int) chan float64 {
    var pop_chan chan float64
    if len(opt) == 0 {
        pop_chan = (*d.input)[d.get_idr(d.r)][d.get_idx(d.x0)]
    } else if len(opt) == 2 {
        pop_chan = (*d.input)[opt[0]][opt[1]]
    } else {
        fmt.Println("Wrong number of parameters. (Will add better handling later.)")
        os.Exit(1)
    }
    for i := 0; i < n; i++ {
        <-pop_chan
    }
    return pop_chan
}

// other nifty functions for later. Not important for homework assignment
func (d data_holder) get_idr(r float64) int {
    return int(r*1000)
}

func (d data_holder) get_idx(x0 float64) int {
    return int(x0*100)
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