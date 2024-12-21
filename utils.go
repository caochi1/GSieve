package main

import (
	"image/color"
	"log"
	"math/rand"
	"reflect"
	"sort"
	"unsafe"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// 按正态分布排序
func normalSort(arr []int) []int {
	sort.Ints(arr)
	n := len(arr)
	result := make([]int, n)
	p := n / 2
	right, left, state := p+1, p-1, true //left = true

	result[p] = arr[n-1]
	for i := n - 2; i >= 0; i-- {
		if state {
			result[left] = arr[i]
			left--
			state = false
		} else {
			result[right] = arr[i]
			right++
			state = true
		}
	}
	return result
}

// 0：正态分布，1：scandata
func generateAndDraw(n, mode int, draw bool) ([]int, int) {
	var (
		m           = make(map[int]int)
		requestList = make([]int, n)
		points      = make(plotter.XYs, 0, n)
		v           int
		objects     []int
	)
	if mode == 0 {
		for i := 0; i < n; i++ {
			v = int(rand.NormFloat64() * 10000)
			requestList[i] = v
			m[v]++
		}
		objects = allObjects(m)
		if !draw {
			return requestList, len(objects)
		}
		for i, v := range normalSort(objects) {
			points = append(points, plotter.XY{X: float64(i), Y: float64(v)})
		}
	} else {
		p := n / 2
		for i := 0; i < n-p; i++ {
			requestList[i] = i
			m[i]++
		}
		for i := n - p; i < n; i++ {
			v = int(rand.NormFloat64() * 10000)
			requestList[i] = v
			m[v]++
		}
		objects = allObjects(m)
		rand.Shuffle(len(requestList), func(i, j int) {
			requestList[i] = requestList[j]
		})
		if !draw {
			return requestList, len(objects)
		}
		for i, v := range normalSort(objects) {
			points = append(points, plotter.XY{X: float64(i), Y: float64(v)})
		}
	}
	p := plot.New()
	p.Title.Text = "Mode B"
	p.X.Label.Text = "objects"
	p.Y.Label.Text = "frequency"
	// p.X.Tick.Label.Color = color.White
	// p.Y.Tick.Label.Color = color.White
	line, _ := plotter.NewLine(points)
	p.Add(line)

	p.Save(12*vg.Inch, 8*vg.Inch, "example.png")
	return requestList, len(objects)
}

func darwMissRatioByTime(line1Data, line2Data XYs) {
	p := plot.New()
	p.Title.Text = "Multiple Lines on One Plot"
	p.X.Label.Text = "X-axis"
	p.Y.Label.Text = "Y-axis"
	p.X.Tick.Label.Color = color.White
	p.Y.Tick.Label.Color = color.White

	line1, _ := plotter.NewLine(line1Data)
	line1.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255} // 红色

	line2, _ := plotter.NewLine(line2Data)
	line2.Color = color.RGBA{B: 255, A: 255} // 蓝色

	p.Add(line1, line2)

	p.Legend.Add(line1Data.name, line1)
	p.Legend.Add(line2Data.name, line2)
	p.Legend.Top = true // 将图例放在顶部

	if err := p.Save(6*vg.Inch, 4*vg.Inch, "two_lines_plot.png"); err != nil {
		log.Fatalf("failed to save plot: %v", err)
	}

	log.Println("图表已保存为 two_lines_plot.png")
}

func allObjects(m map[int]int) []int {
	var i int = 0
	objects := make([]int, len(m))
	for _, v := range m {
		objects[i] = v
		i++
	}
	return objects
}

func i2f(a, b int) float64 {
	return float64(a)/float64(b)
}

func Of(v interface{}) int {
	cache := make(map[uintptr]bool)
	return sizeOf(reflect.Indirect(reflect.ValueOf(v)), cache)
}

func sizeOf(v reflect.Value, cache map[uintptr]bool) int {
	switch v.Kind() {

	case reflect.Array:
		sum := 0
		for i := 0; i < v.Len(); i++ {
			s := sizeOf(v.Index(i), cache)
			if s < 0 {
				return -1
			}
			sum += s
		}

		return sum + (v.Cap()-v.Len())*int(v.Type().Elem().Size())

	case reflect.Slice:
		if cache[v.Pointer()] {
			return 0
		}
		cache[v.Pointer()] = true

		sum := 0
		for i := 0; i < v.Len(); i++ {
			s := sizeOf(v.Index(i), cache)
			if s < 0 {
				return -1
			}
			sum += s
		}

		sum += (v.Cap() - v.Len()) * int(v.Type().Elem().Size())

		return sum + int(v.Type().Size())

	case reflect.Struct:
		sum := 0
		for i, n := 0, v.NumField(); i < n; i++ {
			s := sizeOf(v.Field(i), cache)
			if s < 0 {
				return -1
			}
			sum += s
		}
		padding := int(v.Type().Size())
		for i, n := 0, v.NumField(); i < n; i++ {
			padding -= int(v.Field(i).Type().Size())
		}

		return sum + padding

	case reflect.String:
		s := v.String()
		hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))

		if cache[hdr.Data] {
			return int(v.Type().Size())
		}
		cache[hdr.Data] = true
		return len(s) + int(v.Type().Size())

	case reflect.Ptr:
		if cache[v.Pointer()] {
			return int(v.Type().Size())
		}
		cache[v.Pointer()] = true
		if v.IsNil() {
			return int(reflect.New(v.Type()).Type().Size())
		}
		s := sizeOf(reflect.Indirect(v), cache)
		if s < 0 {
			return -1
		}
		return s + int(v.Type().Size())

	case reflect.Bool,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Int, reflect.Uint,
		reflect.Chan,
		reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Func:
		return int(v.Type().Size())

	case reflect.Map:
		if cache[v.Pointer()] {
			return 0
		}
		cache[v.Pointer()] = true
		sum := 0
		keys := v.MapKeys()
		for i := range keys {
			val := v.MapIndex(keys[i])
			sv := sizeOf(val, cache)
			if sv < 0 {
				return -1
			}
			sum += sv
			sk := sizeOf(keys[i], cache)
			if sk < 0 {
				return -1
			}
			sum += sk
		}

		return sum + int(v.Type().Size()) + int(float64(len(keys))*10.79)

	case reflect.Interface:
		return sizeOf(v.Elem(), cache) + int(v.Type().Size())

	}

	return -1
}

