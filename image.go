package main

import (
	"os"
	"os/exec"
	"bytes"
	"regexp"
	"strconv"
	"container/heap"
	"image/png"
	"encoding/json"
	"sync"
	"io/ioutil"
)

var (
	images     [][][]int
	frames     int
	width      int
	height     int
	tempdir    string
)

func checkErr(e error){
	if e != nil {
		panic(e)
	}
}

func calcDev(i, x, y int) int {
	a := images[i-1][x+1][y]
	b := images[i-1][x][y]
	if a == 0 || b == 0 { 
		return 0
	}
	diff := a - b
	if diff < 0 {
		return -diff
	}
	return diff
}

func imageDev(id string, i, minx, maxx, miny, maxy int) int{
	pq := &PriorityQueue{}
	heap.Init(pq)
	maxes := 0
	for y := miny; y < maxy; y += 2 {
		for x := minx; x < maxx; x++ {
			dev := calcDev(i, x, y)
			item := &Item{
				priority: dev,
			}
			heap.Push(pq, item)
		}
	}
	for i := 0; i < 12; i++ {
		m := heap.Pop(pq).(*Item)
		maxes += m.priority * m.priority * m.priority
	}
	return maxes
}

func bestFit(id string, minx, maxx, miny, maxy int) int {
	maxdev := 0
	maximage := 1
	for i := 1; i <= frames; i++ {
		dev := imageDev(id, i, minx, maxx, miny, maxy)
		if dev > maxdev {
			maxdev = dev
			maximage = i - 1
		}
	}
	return maximage
}


func getDevJSON(id string) []byte {
	// TODO: check if file exists at requested ID

	// get dimensions of video from output of ffmpeg
	cmd := exec.Command("ffmpeg", "-i", "www/images/" + id + ".m4v")
	var dim bytes.Buffer
	cmd.Stderr = &dim
	err := cmd.Run()

	// find dimensions in output w/ regex
	rx, err := regexp.Compile(`\d{2,4}x\d{2,4}`)
	matches := rx.FindAllStringSubmatch(dim.String(), -1)
	if len(matches) == 0 || len(matches[0]) == 0 {
		panic("No matches on dimension found for FFMPEG output")
	}
	dimensions := matches[0][0]

	rx2, err := regexp.Compile(`(\d{2,4})x(\d{2,4})`)
	matches = rx2.FindAllStringSubmatch(dimensions, -1)
	width, err = strconv.Atoi(matches[0][1])
	height, err = strconv.Atoi(matches[0][2])

	// create a temporary directory for the frames
	tempdir, err = ioutil.TempDir("temp", id)

	// fill temp directory w/ frames
	cmd = exec.Command("ffmpeg", "-y", "-i", "www/images/" + id + ".m4v", "-r", "24",
		                "-vcodec", "png", "-s", dimensions, tempdir + "/" + id + "-%02d.png")
	var out bytes.Buffer
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	
	// find number of frames generated
	rx, err = regexp.Compile(`frame=   (\d{1,3})`)
	matches = rx.FindAllStringSubmatch(out.String(), -1)
	if len(matches) == 0 || len(matches[0]) == 0 {
		panic("Couldn't find number of generated frames")
	}
	frames, err = strconv.Atoi(matches[0][1])

	// load file descriptors for all frame files
	images = make([][][]int, frames)
	max := uint32(0xFFFF)
	for i := 1; i <= frames; i++ {
		istr := strconv.FormatInt(int64(i), 10)
		if i < 10 {
			istr = "0" + istr
		}	
		reader, err := os.Open(tempdir + "/" + id + "-" + istr + ".png")
		checkErr(err)		
		img, err := png.Decode(reader)
		images[i-1] = make([][]int, width)
		for x := 0; x < width; x++ {
			images[i-1][x] = make([]int, height)
			for y := 0; y < height; y++ {
				r, g, b, _ := img.At(x, y).RGBA()
				r = 255 * r / max
				g = 255 * g / max
				b = 255 * b / max
				images[i-1][x][y] = int((r + g + b) / 3)
			}
		}
		reader.Close()
	}

	// find JS output table from frames
	
	c := width / 50
	r := height / 50
	w := float32(width) / float32(c)
	h := float32(height) / float32(r)
	grid := make([][]int, r)

	// run each bestFit call in its own thread and 
	// wait for them all to finish
	var wg sync.WaitGroup
	for rr := 0; rr < r; rr++ {
		grid[rr] = make([]int, c)
		for cc := 0; cc < c; cc++ {
			wg.Add(1)
			go func(rr, cc int){
				defer wg.Done()
				grid[rr][cc] = bestFit(id, int(float32(cc) * w), int(float32(cc) * w + w - 1), 
					                   int(float32(rr) * h), int(float32(rr) * h + w - 1))
			}(rr, cc)
		}
	}

	wg.Wait()

	// delete temporary directory
	os.RemoveAll(tempdir)

	jarr, err := json.Marshal(grid)
	return jarr
}

