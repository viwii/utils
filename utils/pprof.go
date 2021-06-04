package utils

import (
	"github.com/arl/statsviz"
	"net/http"
)

func StartPporfService(ipstr string) {
	go func(ipstr string) {
		statsviz.RegisterDefault()
		http.ListenAndServe(ipstr, nil)
	}(ipstr)
}
