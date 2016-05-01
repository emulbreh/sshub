package libsshub

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func InstallHttpHandlers(hub *Hub) {
	linksEndpoint := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(hub.serializeLinks())
			return
		}
		if r.Method != "POST" {
			w.Header().Add("Allow", "POST")
			http.Error(w, "Method Not Allowed", 405)
			return
		}
		var link Link
		err := json.NewDecoder(r.Body).Decode(&link)
		if err != nil {
			log.Infof("json decoding error %v", err)
			http.Error(w, "Bad Request", 400)
			json.NewEncoder(w).Encode(&APIError{Code: "invalid-json", Message: err.Error()})
			return
		}
		err = hub.addLink(&link)
		if err != nil {
			http.Error(w, "Bad Request", 400)
			json.NewEncoder(w).Encode(&APIError{Code: "invalid-tunnel", Message: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(struct{}{})
	}

	http.HandleFunc("/links/", linksEndpoint)
}
