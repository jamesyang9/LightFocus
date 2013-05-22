package main

import (
	"net/http"
	"html/template"
	"fmt"
)

type Video struct {
	ID string
	Code string
}

func handleUpload(writer http.ResponseWriter, req *http.Request) {
	GET := req.URL.Query()
	paramList, fail := GET["id"]
	if !fail {
		return
	}

	jsarr := getDevJSON(paramList[0])
	output := "var grid" + paramList[0] + " = " + string(jsarr)

	// load in upload template and serve to user
	vid := Video{ID: paramList[0], Code: output}
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

func main(){
	fmt.Println("Starting server...")
	http.Handle("/", http.FileServer(http.Dir("www")))
	http.Handle("/upload", http.HandlerFunc(handleUpload))
	http.ListenAndServe(":8080", nil)
}

