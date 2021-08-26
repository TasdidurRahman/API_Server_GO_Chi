package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type User struct {
	Name string `json:"name"`
	Pass string `json:"pass,omitempty"`
}

type Product struct {
	Category string `json:"category,omitempty"`
	Brand    string `json:"brand,omitempty"`
	Model    string `json:"model,omitempty"`
	Price    int    `json:"price,omitempty"`
	Count    int    `json:"count,omitempty"`
}

type Invoice struct {
	UserName     string    `json:"user_name"`
	Product      Product   `json:"product"`
	PurchaseDate time.Time `json:"purchase_date,omitempty"`
}

var (
	users = []User{
		{"a", "b"},
		{"c", "d"},
		{"e", "f"},
	}
	admins = []User{
		{"g", "h"},
		{"i", "j"},
	}

	products = []Product{
		{"monitor", "DELL", "d22", 15000, 5},
		{"monitor", "LG", "g94", 16000, 1},
		{"mouse", "A4tech", "super", 350, 10},
		{"keyboard", "Delux", "soft", 400, 3},
		{"headphone", "bits", "mux35", 22000, 2},
		{"Router", "TpLink", "Archer10", 3000, 3},
	}

	invoices []Invoice
)

func matchIncompleteCompleteProducts(p Product, q Product) bool { // p : incomplete , q : complete
	if p.Category == "" && p.Brand == "" && p.Model == "" {
		return false
	}
	if (p.Category == q.Category || p.Category == "") &&
		(p.Model == q.Model || p.Model == "") &&
		(p.Brand == q.Brand || p.Brand == "") {
		return true
	}
	return false
}

func (p *Product) search() (ps []Product) {
	for _, j := range products {
		if matchIncompleteCompleteProducts(*p, j) {
			ps = append(ps, j)
		}
	}
	return
}

func (p *Product) changeCount(c int) {
	for i, j := range products {
		if matchIncompleteCompleteProducts(*p, j) {
			products[i].Count += c
			return
		}
	}
}

func (p *Product) deleteProduct() {
	for i, j := range products {
		if matchIncompleteCompleteProducts(*p, j) {
			products = append(products[:i], products[i+1:]...)
			return
		}
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, "invalid request")
		return
	}
	for _, i := range users {
		if i.Name == u.Name {
			writeResponse(w, http.StatusBadRequest, "username not available")
			return
		}
	}
	users = append(users, u)
	w.Write([]byte("successfully added user"))
	return
}

func loginAdmin(w http.ResponseWriter, r *http.Request) {
	var u User
	json.NewDecoder(r.Body).Decode(&u)
	for _, i := range admins {
		if i.Name == u.Name && i.Pass == u.Pass {
			//gen token
			_, tokenString, _ := adminToken.Encode(map[string]interface{}{"aud": i.Name})
			//set cookie
			http.SetCookie(w, &http.Cookie{
				Name:    "jwt",
				Value:   tokenString,
				Expires: time.Now().Add(time.Second * 50000),
			})
			w.Write([]byte("login as admin successfull"))
			return
		}
	}
	w.Write([]byte("no such credential"))
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	var u User
	json.NewDecoder(r.Body).Decode(&u)
	for _, i := range users {
		if i.Name == u.Name && i.Pass == u.Pass {
			//gen token
			_, tokenString, _ := userToken.Encode(map[string]interface{}{"aud": i.Name})
			//set cookie
			http.SetCookie(w, &http.Cookie{
				Name:    "jwt",
				Value:   tokenString,
				Expires: time.Now().Add(time.Second * 50000),
			})
			w.Write([]byte("login as user successfull"))
			return
		}
	}
	w.Write([]byte("no such credential"))
}

func showAll(w http.ResponseWriter, r *http.Request) {
	if len(products) == 0 {
		writeResponse(w, http.StatusNoContent, "no products available")
	}
	json.NewEncoder(w).Encode(products)

}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	c := chi.URLParam(r, "c")
	cc, _ := strconv.Atoi(c)
	json.NewDecoder(r.Body).Decode(&p)
	for i, j := range products {
		fmt.Println(j, p)
		if matchIncompleteCompleteProducts(p, j) {
			products[i].changeCount(cc)
			json.NewEncoder(w).Encode(products)
			return
		}
	}

}
func addNewProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	json.NewDecoder(r.Body).Decode(&p)
	products = append(products, p)
}
func deleteProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	json.NewDecoder(r.Body).Decode(&p)
	p.deleteProduct()
}
func getInvoice(w http.ResponseWriter, r *http.Request) {
	name := getUserNameFromCookie(r, "jwt", userToken)
	var ret []Invoice
	for i, j := range invoices {
		if j.UserName == name {
			ret = append(ret, invoices[i])
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ret)
}
func searchProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.Println(err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(p.search()); err != nil {
		writeResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func buyProduct(w http.ResponseWriter, r *http.Request) {
	var I Invoice
	var pro Product
	json.NewDecoder(r.Body).Decode(&pro)
	uName := getUserNameFromCookie(r, "jwt", userToken)
	I.UserName = uName
	I.Product = pro
	pro.changeCount(-1 * pro.Count)
	invoices = append(invoices, I)
}

func writeResponse(w http.ResponseWriter, code int, errMsg string) {
	w.WriteHeader(code)
	if _, err := w.Write([]byte(errMsg)); err != nil {
		log.Println(err)
	}
}
func getUserNameFromCookie(r *http.Request, ckName string, tokenAuth *jwtauth.JWTAuth) (name string) {
	cookie, _ := r.Cookie(ckName)
	token, _ := tokenAuth.Decode(cookie.String()[len(ckName)+1:])
	name = token.Audience()[0]
	return
}
