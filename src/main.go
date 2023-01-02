package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const MaxLookups = 256
const IdChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"

var IdLength = 4

type Context struct {
	db *sql.DB
}

type CreateLinkData struct {
	Url *string `json:"url"`
}

type CreateLinkResponse struct {
	Id *string `json:"id"`
}

func GenerateId() string {
	var id strings.Builder
	id.Grow(IdLength)

	chars := len(IdChars)
	for i := 0; i < IdLength; i++ {
		n := rand.Intn(chars)
		id.WriteByte(IdChars[n])
	}

	return id.String()
}

func (ctx *Context) ReadLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	query := `SELECT destination FROM links WHERE id = $1`
	var destination string
	err := ctx.db.QueryRow(query, id).Scan(&destination)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Add("Location", destination)
	w.WriteHeader(http.StatusFound)
}

func (ctx *Context) CreateLink(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		fmt.Fprintln(w, "415 unsupported media type")
		return
	}

	var data CreateLinkData
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "400 bad request")
		return
	}
	if data.Url == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "missing required key `url`")
		return
	}
	_, err = url.ParseRequestURI(*data.Url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "invalid value for key `url`")
		return
	}

	var id string
	attempts := 0
	for {
		id = GenerateId()
		query := `SELECT EXISTS(SELECT 1 FROM links WHERE id = $1)`
		var exists bool
		err := ctx.db.QueryRow(query, id).Scan(&exists)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "500 internal server error")
			log.Print(err)
			return
		}

		if !exists {
			break
		} else if attempts >= MaxLookups {
			IdLength += 1
			attempts = 0
		} else {
			attempts += 1
		}
	}

	query := `INSERT INTO links (id, destination) VALUES ($1, $2)`
	_, err = ctx.db.Exec(query, id, data.Url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "500 internal server error")
		log.Print(err)
		return
	}

	response := CreateLinkResponse{Id: &id}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	dbUri := os.Getenv("DB_URI")

	db, err := sql.Open("postgres", dbUri)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	ctx := Context{db: db}

	router.HandleFunc("/a/{id}", ctx.ReadLink).Methods(http.MethodGet)
	router.HandleFunc("/a", ctx.CreateLink).Methods(http.MethodPost)

	addr := fmt.Sprintf("localhost:%v", port)
	err = http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatal(err)
	}
}
