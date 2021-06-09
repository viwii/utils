package utils

import (
	"net/http"

	"github.com/arl/statsviz"
)

func StartPporfService(ipstr string) {
	go func(ipstr string) {
		statsviz.RegisterDefault()
		http.ListenAndServe(ipstr, nil)
	}(ipstr)
}
