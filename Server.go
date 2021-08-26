package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

var (
	adminToken *jwtauth.JWTAuth
	userToken  *jwtauth.JWTAuth
)

func init() {
	adminToken = jwtauth.New("HS256", []byte("papel"), []byte("papel"))
	userToken = jwtauth.New("HS256", []byte("manana"), []byte("manana"))
}

func main() {
	r := chi.NewRouter()

	fmt.Println("server running...")

	r.Post("/createUser", createUser)
	r.Post("/loginAsAdmin", loginAdmin)
	r.Post("/loginAsUser", loginUser)

	//r.Group(admin)
	//r.Group(user)
	r.Route("/admin", admin)
	r.Route("/user", user)

	log.Fatal(http.ListenAndServe(":8081", r))
}

func admin(r chi.Router) {

	r.Use(jwtauth.Verifier(adminToken))
	r.Use(jwtauth.Authenticator)

	r.Route("/products", func(r chi.Router) {
		r.Get("/", showAll)
		r.Post("/", addNewProduct)
		r.Put("/{c}", updateProduct)
		r.Delete("/", deleteProduct)
	})
}

func user(r chi.Router) {

	r.Use(jwtauth.Verifier(userToken))
	r.Use(jwtauth.Authenticator)

	r.Get("/showInvoice", getInvoice)
	r.Route("/product", func(r chi.Router) {
		r.Get("/search", searchProduct)
		r.Post("/buy", buyProduct)

	})
}
