package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

var templates *template.Template
var decoder *schema.Decoder

var connections []Connection

var hash string

func main() {

	rand.Seed(time.Now().UnixNano())
	hash = randSeq(12)

	decoder = schema.NewDecoder()
	fmt.Println("Initializing templates cache...")
	templates, _ = buildTemplates()
	fmt.Println("Initializing storage...")
	InitStorage()

	r := mux.NewRouter()
	r.HandleFunc("/api/query", apiQuery).Methods("POST")
	r.HandleFunc("/api/list_databases", apiListDatabases).Methods("GET")
	r.HandleFunc("/api/change_database", apiChangeDatabases).Methods("POST")
	r.HandleFunc("/api/add_connection", apiAddConnection).Methods("POST")
	r.HandleFunc("/api/save_connection", apiSaveConnection).Methods("POST")
	r.HandleFunc("/api/get_connections", apiGetConnections).Methods("GET")
	r.HandleFunc("/api/disconnect", apiDisconnect).Methods("DELETE")

	r.HandleFunc("/css/{resource}", serveResource).Methods("GET")
	r.HandleFunc("/img/{resource}", serveResource).Methods("GET")
	r.HandleFunc("/js/{resource}", serveResource).Methods("GET")
	r.HandleFunc("/", mainPage).Methods("GET")

	http.Handle("/", r)

	http.HandleFunc("/libs/", serveResource)
	var port string
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	} else {
		port = "4443"
	}

	fmt.Printf("Serving MSSQL-CLIENT from https://localhost:%s\n", port)
	go http.ListenAndServeTLS(":"+port, "tls/cert.pem", "tls/key.pem", nil)
	if os.Getenv("ENV") == "DEV" {
		go func() {
			for range time.Tick(300 * time.Millisecond) {
				tc, needUpdate := buildTemplates()
				if needUpdate {
					fmt.Println("Template change detected, updating..")
					templates = tc
				}
			}
		}()
	}

	if os.Getenv("ENV") != "DEV" {
		if runtime.GOOS == "darwin" {
			cmd := []string{"-a", "/Applications/Google Chrome.app", "https://localhost:4443"}
			out, err := exec.Command("open", cmd...).Output()

			if err != nil {
				log.Printf("Error execution command: %s", err.Error())
			}
			fmt.Printf("%s", out)
		}
	}

	fmt.Scanln()
}

func serveResource(w http.ResponseWriter, r *http.Request) {
	path := "public" + r.URL.Path
	var contentType string
	if strings.HasSuffix(path, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(path, "png") {
		contentType = "image/png"
	} else {
		contentType = "text/plain"
	}

	f, err := os.Open(path)
	if err != nil {
		log.Println(err.Error())
	}

	defer f.Close()
	w.Header().Set("Content-Type", contentType)
	br := bufio.NewReader(f)
	br.WriteTo(w)

}

var lastMod time.Time = time.Unix(0, 0)

func buildTemplates() (*template.Template, bool) {
	needUpdate := false
	templates := template.New("templates")
	funcMap := template.FuncMap{
		"ParseValue": ParseValue,
	}
	templates = templates.Funcs(funcMap)
	basePath := "views"
	templateDir, _ := os.Open(basePath)
	defer templateDir.Close()

	fileInfos, _ := templateDir.Readdir(-1)
	fileNames := make([]string, len(fileInfos))
	for i, fi := range fileInfos {
		if !fi.IsDir() {
			if fi.ModTime().After(lastMod) {
				lastMod = fi.ModTime()
				needUpdate = true
			}
			fileNames[i] = basePath + "/" + fi.Name()
		}
	}
	if needUpdate {
		templates.ParseFiles(fileNames...)
	}
	return templates, needUpdate
}
