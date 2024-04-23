package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	
	"github.com/rs/cors"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("meow"))
	})

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5000"},
		AllowedMethods: []string{http.MethodGet},
	})

	handler := c.Handler(r)

	http.ListenAndServe(":3000", handler)

}

func JSONMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w,r)
	})
}
