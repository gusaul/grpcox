package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Init - routes initialization
func Init(router *mux.Router) {
	h := InitHandler()

	router.HandleFunc("/", h.index)

	ajaxRoute := router.PathPrefix("/server/{host}").Subrouter()
	ajaxRoute.HandleFunc("/services", corsHandler(h.getLists)).Methods(http.MethodGet, http.MethodOptions)
	ajaxRoute.HandleFunc("/services", corsHandler(h.getListsWithProto)).Methods(http.MethodPost)
	ajaxRoute.HandleFunc("/service/{serv_name}/functions", corsHandler(h.getLists)).Methods(http.MethodGet, http.MethodOptions)
	ajaxRoute.HandleFunc("/function/{func_name}/describe", corsHandler(h.describeFunction)).Methods(http.MethodGet, http.MethodOptions)
	ajaxRoute.HandleFunc("/function/{func_name}/invoke", corsHandler(h.invokeFunction)).Methods(http.MethodPost, http.MethodOptions)

	// get list of active connection
	router.HandleFunc("/active/get", corsHandler(h.getActiveConns)).Methods(http.MethodGet, http.MethodOptions)
	// close active connection
	router.HandleFunc("/active/close/{host}", corsHandler(h.closeActiveConns)).Methods(http.MethodDelete, http.MethodOptions)

	router.PathPrefix("/css/").Handler(http.FileServer(http.FS(fs)))
	router.PathPrefix("/js/").Handler(http.FileServer(http.FS(fs)))
	router.PathPrefix("/font/").Handler(http.FileServer(http.FS(fs)))
	router.PathPrefix("/img/").Handler(http.FileServer(http.FS(fs)))
}

func corsHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Headers", "use_tls")
			return
		}

		h.ServeHTTP(w, r)
	}
}
