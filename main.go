package main

import (
	"fmt"
	"main/gsieve"

	"gonum.org/v1/plot/plotter"
)

type cache interface {
	Get(key any) (any, bool)
	Insert(key, value any)
	Miss() int
}

type XYs struct {
	plotter.XYs
	name string
}

func main() {
	missRatio(10000000, 1, 2, true)
	for i := 2; i < 11; i += 2 {
		fmt.Println(i)
		missRatio(10000000, 1, i, false)
	}
}

func missRatio(n, mode, c int, draw bool) {
	requests, count := generateAndDraw(n, mode, draw)
	capacity := count / c
	sieve, gsieve := gsieve.NewSieveCache(capacity), gsieve.NewGSieve(capacity)
	ch := make(chan *XYs)
	go request(sieve, requests, "sieve", ch)
	go request(gsieve, requests, "gsieve", ch)
	if !draw {
		return
	}
	var xys []*XYs
	for i := 0; i < 2; i++ {
		v := <-ch
		xys = append(xys, v)
	}
	close(ch)
	darwMissRatioByTime(*xys[0], *xys[1])
}

func request(c cache, list []int, name string, ch chan *XYs) {
	points := make(plotter.XYs, len(list))
	for i, v := range list {
		if _, ok := c.Get(v); !ok {
			c.Insert(v, v)
		}
		points[i] = plotter.XY{X: float64(i), Y: i2f(c.Miss(), i+1)}
	}
	ch <- &XYs{points, name}
	fmt.Println(name, i2f(c.Miss(), len(list)))

}

// gsieve.ListenHTTP("0.0.0.0:8080", "127.0.0.1:6379")
// gsieve.Connect("127.0.0.1:6379")
// gsieve.Run("127.0.0.1:6379", 5, 4, 500)
