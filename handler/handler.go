package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gusaul/grpcox/core"
)

func index(w http.ResponseWriter, r *http.Request) {
	body := new(bytes.Buffer)
	err := indexHTML.Execute(body, make(map[string]string))
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(body.Bytes())
}

func getLists(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host := vars["host"]
	if host == "" {
		writeError(w, fmt.Errorf("Invalid Host"))
		return
	}

	service := vars["serv_name"]

	g := new(core.GrpCox)
	useTLS, _ := strconv.ParseBool(r.Header.Get("use_tls"))
	g.PlainText = !useTLS

	res, err := g.GetResource(context.Background(), host)
	if err != nil {
		writeError(w, err)
		return
	}
	defer res.Close()

	result, err := res.List(service)
	if err != nil {
		writeError(w, err)
		return
	}

	response(w, result)
}

func describeFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host := vars["host"]
	if host == "" {
		writeError(w, fmt.Errorf("Invalid Host"))
		return
	}

	funcName := vars["func_name"]
	if host == "" {
		writeError(w, fmt.Errorf("Invalid Func Name"))
		return
	}

	g := new(core.GrpCox)
	useTLS, _ := strconv.ParseBool(r.Header.Get("use_tls"))
	g.PlainText = !useTLS

	res, err := g.GetResource(context.Background(), host)
	if err != nil {
		writeError(w, err)
		return
	}
	defer res.Close()

	// get param
	result, _, err := res.Describe(funcName)
	if err != nil {
		writeError(w, err)
		return
	}
	match := reGetFuncArg.FindStringSubmatch(result)
	if len(match) < 2 {
		writeError(w, fmt.Errorf("Invalid Func Type"))
		return
	}

	// describe func
	result, template, err := res.Describe(match[1])
	if err != nil {
		writeError(w, err)
		return
	}

	type desc struct {
		Schema   string `json:"schema"`
		Template string `json:"template"`
	}

	response(w, desc{
		Schema:   result,
		Template: template,
	})

}

func invokeFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host := vars["host"]
	if host == "" {
		writeError(w, fmt.Errorf("Invalid Host"))
		return
	}

	funcName := vars["func_name"]
	if host == "" {
		writeError(w, fmt.Errorf("Invalid Func Name"))
		return
	}

	g := new(core.GrpCox)
	useTLS, _ := strconv.ParseBool(r.Header.Get("use_tls"))
	g.PlainText = !useTLS

	res, err := g.GetResource(context.Background(), host)
	if err != nil {
		writeError(w, err)
		return
	}
	defer res.Close()

	// get param
	result, timer, err := res.Invoke(context.Background(), funcName, r.Body)
	if err != nil {
		writeError(w, err)
		return
	}

	type invRes struct {
		Time   string `json:"timer"`
		Result string `json:"result"`
	}

	response(w, invRes{
		Time:   timer.String(),
		Result: result,
	})
}
