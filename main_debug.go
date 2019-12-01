//+build go1.12,debug

package main

import (
	"net/http"
	_ "net/http/pprof"
)

func init() {
	go http.ListenAndServe("localhost:8080", nil)
}
