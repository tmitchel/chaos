package main

////////////////////////////////////////////////////
// Purpose: To solve the Rossler equations with a //
// given set of initial conditions and parameter  // 
// values                                         //
// Return: A single pdf with 3 plots overlayed.   //
// The pdf contains plots of each permutation of  //
// the three space coordinates                    //
////////////////////////////////////////////////////

import (
    // "fmt"
    "flag"
    "math"
    "strconv"
    "image/color"
    "gonum.org/v1/plot"
    "gonum.org/v1/plot/plotter"
)

////////////////////////////////////////////////////
// Purpose: Simple struct to hold a point for     //
// plotting                                       //
// Variables: X, Y, Z coordinates and an int to   //
// signal when a result blows up and can't be     //
// plotted                                        //
////////////////////////////////////////////////////
type point struct {
    x, y, z float64
    breaker int
}

////////////////////////////////////////////////////
// Purpose: Do the iterative calculations         //
// Returns: A buffered channel packed with point  //
// structs                                        //
////////////////////////////////////////////////////
func iter(x0, y0, z0, a, b, c float64) chan point {
    channel := make(chan point, 400)
    sign := 0

    go func () {
        for x, y, z := x0, y0, z0; ; x, y, z = (-(y+z))/1000.+x, (x+a*y)/1000.+y, (b+z*(x-c))/1000.+z {
            if math.IsNaN(x) || math.IsNaN(y) || math.IsNaN(z) || math.IsInf(x, sign) || math.IsInf(y, sign) || math.IsInf(z, sign) {
                channel <- point{x: x, y: y, z: z, breaker: -1}
                break
            }
            channel <- point{x: x, y: y, z: z}
        }
    } ()
    return channel
}

func main() {

    // Command-line options
    a  := flag.Float64("a" , 0.2    , "Parameter a")
    b  := flag.Float64("b" , 0.2    , "Parameter b")
    c  := flag.Float64("c" , 5.7    , "Parameter c")
    x0 := flag.Float64("x0", -1.0   , "Initial Condition x0")
    y0 := flag.Float64("y0", 0.0    , "Initial Condition y0")
    z0 := flag.Float64("z0", 0.0    , "Initial Condition z0")
    t  := flag.Int(    "t" , 100000 , "Number of time steps")
    flag.Parse()

    // channel holding the results
    results := iter(*x0, *y0, *z0, *a, *b, *c)
    
    p, err := plot.New()
    if err != nil {
        panic(err)
    }

    pts_xy := make(plotter.XYs, *t)
    pts_xz := make(plotter.XYs, *t)
    pts_yz := make(plotter.XYs, *t)

    // read from channel to fill plots
    for i := 0; i < *t; i++ {
        pt := <-results
        if pt.breaker < 0 {
            break
        }
        pts_xy[i].X, pts_xy[i].Y = pt.x, pt.y
        pts_xz[i].X, pts_xz[i].Y = pt.x, pt.z
        pts_yz[i].X, pts_yz[i].Y = pt.y, pt.z
    }

    // make the plots pretty
    lxy, _ := plotter.NewLine(pts_xy)
    lxy.Color = color.RGBA{G: 255}
    lxz, _ := plotter.NewLine(pts_xz)
    lxz.Color = color.RGBA{R: 255}
    lyz, _ := plotter.NewLine(pts_yz)
    lyz.Color = color.RGBA{B: 255}

    p.Add(lxy, lxz, lyz)

    // convert values to strings for title/file name
    a_val  := strconv.FormatFloat(*a, 'f', -1, 64)
    b_val  := strconv.FormatFloat(*b, 'f', -1, 64)
    c_val  := strconv.FormatFloat(*c, 'f', -1, 64)
    x0_val := strconv.FormatFloat(*x0, 'f', -1, 64)
    y0_val := strconv.FormatFloat(*y0, 'f', -1, 64)
    z0_val := strconv.FormatFloat(*z0, 'f', -1, 64)

    p.Title.Text = "Rossler System:\n\t a="+a_val+" b="+b_val+" c="+c_val+"\n\t x0="+x0_val+" y0="+y0_val+" z0="+z0_val
    p.Legend.Add("X(t) vs Y(T)", lxy)
    p.Legend.Add("X(t) vs Z(T)", lxz)
    p.Legend.Add("Y(t) vs Z(T)", lyz)

    // save as pdf
    if err := p.Save(600, 400, "rossler_c"+c_val+".pdf"); err != nil {
        panic(err)
    }
}    

