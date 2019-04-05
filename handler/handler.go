package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gusaul/grpcox/core"
)

// Handler hold all handler methods
type Handler struct {
	g *core.GrpCox
}

// InitHandler Constructor
func InitHandler() *Handler {
	return &Handler{
		g: core.InitGrpCox(),
	}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) getActiveConns(w http.ResponseWriter, r *http.Request) {
	response(w, h.g.GetActiveConns(context.TODO()))
}

func (h *Handler) closeActiveConns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host := vars["host"]
	if host == "" {
		writeError(w, fmt.Errorf("Invalid Host"))
		return
	}

	err := h.g.CloseActiveConns(strings.Trim(host, " "))
	if err != nil {
		writeError(w, err)
		return
	}
	response(w, map[string]bool{"success": true})
}

func (h *Handler) getLists(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host := vars["host"]
	if host == "" {
		writeError(w, fmt.Errorf("Invalid Host"))
		return
	}

	service := vars["serv_name"]

	useTLS, _ := strconv.ParseBool(r.Header.Get("use_tls"))
	restart, _ := strconv.ParseBool(r.FormValue("restart"))

	res, err := h.g.GetResource(context.Background(), host, !useTLS, restart)
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := res.List(service)
	if err != nil {
		writeError(w, err)
		return
	}

	h.g.Extend(host)
	response(w, result)
}

func (h *Handler) describeFunction(w http.ResponseWriter, r *http.Request) {
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

	useTLS, _ := strconv.ParseBool(r.Header.Get("use_tls"))

	res, err := h.g.GetResource(context.Background(), host, !useTLS, false)
	if err != nil {
		writeError(w, err)
		return
	}

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

	h.g.Extend(host)
	response(w, desc{
		Schema:   result,
		Template: template,
	})

}

func (h *Handler) invokeFunction(w http.ResponseWriter, r *http.Request) {
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

	useTLS, _ := strconv.ParseBool(r.Header.Get("use_tls"))

	res, err := h.g.GetResource(context.Background(), host, !useTLS, false)
	if err != nil {
		writeError(w, err)
		return
	}

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

	h.g.Extend(host)
	response(w, invRes{
		Time:   timer.String(),
		Result: result,
	})
}
