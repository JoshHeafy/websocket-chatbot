package middleware

import (
	"chatbot/src/auth"
	"chatbot/src/controller"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func Autentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := controller.NewResponseManager()
		token := r.Header.Get("Access-Token")
		_, err := auth.ValidateToken(token)

		if token == "" || err != nil {
			response.Msg = "Inicio de Sesión Inválido"
			response.Status = "error"
			response.StatusCode = 401
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func MiddlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Auth-Date, Auth-Periodo, Access-Token")
			next.ServeHTTP(w, req)
		})
}

func EnableCORS(router *mux.Router) {
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}).Methods(http.MethodOptions)
	router.Use(MiddlewareCors)
}
