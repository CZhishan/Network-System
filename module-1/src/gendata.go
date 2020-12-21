// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 21.

// Server3 is an "echo" server that displays request parameters.
package main

import (
	"strconv"
	"strings"
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/gendata", gendata)
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}

//!+handler
// handler echoes the HTTP request.
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Proto)
	for k, v := range r.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
	fmt.Fprintf(w, "Host = %q\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr = %q\n", r.RemoteAddr)
	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}
	for k, v := range r.Form {
		fmt.Fprintf(w, "Form[%q] = %q\n", k, v)
	}
}
//!-handler
//!-gendata
// handler to genenrate numBytes of string
func gendata(w http.ResponseWriter, r *http.Request){
	numBytes,er:=r.URL.Query()["numByte"]
	if!er||len(numBytes[0])<1 {
		log.Println("numByte is missing")
	}
	numByteString :=numBytes[0]
	s:="."
	numByte,err:=strconv.Atoi(numByteString)
	if err!=nil{
		log.Println(err)
	}
	fmt.Fprintf(w,strings.Repeat(s,numByte))
}
