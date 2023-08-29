package main

import (
	"chatbot/src/routes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := map[string]interface{}{
		"Websocket": "WebSocket with OpenAi",
		"Version":   "1.0.0",
		"Author":    "Joshar Cordova",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/", homePage)

	routes.RoutesWebSocket()
	fmt.Println("Websocket Server on port:2000")
	log.Fatal(http.ListenAndServe(":2000", nil))
}
