package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"regexp"

	"github.com/gusaul/grpcox/index"
)

var (
	reGetFuncArg *regexp.Regexp
	indexHTML    *template.Template

	fs = index.FS
)

func init() {
	reGetFuncArg = regexp.MustCompile("\\( (.*) \\) returns")
	indexHTML = template.Must(template.New("index.html").Delims("{[", "]}").ParseFS(fs, "index.html"))
}

// Response - Standar ajax Response
type Response struct {
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data"`
}

func writeError(w http.ResponseWriter, err error) {
	e, _ := json.Marshal(Response{
		Error: err.Error(),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(e)
}

func response(w http.ResponseWriter, data interface{}) {
	e, _ := json.Marshal(Response{
		Data: data,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(e)
}
