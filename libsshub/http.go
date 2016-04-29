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
	tunnelsEndpoint := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(hub.serializeTunnels())
			return
		}
		if r.Method != "POST" {
			w.Header().Add("Allow", "POST")
			http.Error(w, "Method Not Allowed", 405)
			return
		}
		var tunnel Tunnel
		err := json.NewDecoder(r.Body).Decode(&tunnel)
		if err != nil {
			log.Infof("json decoding error %v", err)
			http.Error(w, "Bad Request", 400)
			json.NewEncoder(w).Encode(&APIError{Code: "invalid-json", Message: err.Error()})
			return
		}
		err = hub.addTunnel(&tunnel)
		if err != nil {
			http.Error(w, "Bad Request", 400)
			json.NewEncoder(w).Encode(&APIError{Code: "invalid-tunnel", Message: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(struct{}{})
	}

	http.HandleFunc("/tunnels/", tunnelsEndpoint)
}
