package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Init - routes initialization
func Init(router *mux.Router) {
	router.HandleFunc("/", index)

	ajaxRoute := router.PathPrefix("/server/{host}").Subrouter()
	ajaxRoute.HandleFunc("/services", corsHandler(getLists)).Methods(http.MethodGet, http.MethodOptions)
	ajaxRoute.HandleFunc("/service/{serv_name}/functions", corsHandler(getLists)).Methods(http.MethodGet, http.MethodOptions)
	ajaxRoute.HandleFunc("/function/{func_name}/describe", corsHandler(describeFunction)).Methods(http.MethodGet, http.MethodOptions)
	ajaxRoute.HandleFunc("/function/{func_name}/invoke", corsHandler(invokeFunction)).Methods(http.MethodPost, http.MethodOptions)

	assetsPath := "index"
	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir(assetsPath+"/css/"))))
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir(assetsPath+"/js/"))))
	router.PathPrefix("/font/").Handler(http.StripPrefix("/font/", http.FileServer(http.Dir(assetsPath+"/font/"))))
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
