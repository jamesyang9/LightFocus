package main

import (
	"net/http"
	"html/template"
	"fmt"
	"time"
	"os"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
)

type Video struct {
	ID string
	Code string
	Uploaded bool
}

func execTemplate(writer http.ResponseWriter, vid Video) {
	t := template.New("upload.html")
	templ, err := t.ParseFiles("tmpl/upload.html")
	if err != nil {
		panic(err)
	}
	err = templ.Execute(writer, vid)
	if err != nil {
		panic(err)
	}
}


func handleUpload(writer http.ResponseWriter, req *http.Request) {
	// retrieve ID from GET
	var id string
	val := req.FormValue("post")
	dir, err := ioutil.ReadDir("www/images")
	checkErr(err)
	rx, err := regexp.Compile(`\d{1,3}`)
	checkErr(err)
	if val != "upload" {
		GET := req.URL.Query()
		paramList, fail := GET["id"]
		if !fail {
			execTemplate(writer, Video{ID: "", Code: "", Uploaded: false})
			return
		}
		id = paramList[0]
		output, err := ioutil.ReadFile("www/images/" + id + ".js")
		if err == nil {
			execTemplate(writer, Video{
				ID: id,
				Code: string(output),
				Uploaded: true,
			})
			return
		}
	} else {
		movie, _, err := req.FormFile("movie")
		defer movie.Close()
		checkErr(err)
		
		// scan directory to find next ID
		max := 0
		for _, stat := range dir {
			matches := rx.FindAllStringSubmatch(stat.Name(), -1)
			if(len(matches) == 0){
				continue
			}
			n, err := strconv.Atoi(matches[0][0])
			checkErr(err)
			if n > max {
				max = n
			}
		}

		id = strconv.Itoa(max + 1)
		
		file, err := os.Create("www/images/" + id + ".m4v")
		checkErr(err)
		io.Copy(file, movie)
		file.Close()
	}

	
	fmt.Printf("Processing request for movie %s...", id)
	start := float64(time.Now().UnixNano())

	// get grid for given $id.m4v
	jsarr := getDevJSON(id)
	output := "var grid" + id + " = " + string(jsarr)

	end := (float64(time.Now().UnixNano()) - start) / float64(1e9)
	fmt.Printf(" finished after %f seconds.\n", end)

	// load in upload template and serve to user
	save, err := os.Create("www/images/" + id + ".js")
	checkErr(err)
	io.WriteString(save, output)
	save.Close()
	execTemplate(writer, Video{ID: id, Code: output, Uploaded: true})
}

func main(){
	fmt.Println("Starting server...")
	http.Handle("/", http.FileServer(http.Dir("www")))
	http.Handle("/upload", http.HandlerFunc(handleUpload))
	http.ListenAndServe(":8000", nil)
}

