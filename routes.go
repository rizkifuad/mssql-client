package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type APImsg struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func renderView(w http.ResponseWriter, templatePath string, context interface{}) error {
	w.Header().Set("Content-Type", "text/html")

	pusher, ok := w.(http.Pusher)
	if ok {
		publicDir := "public"
		fileList := []string{}
		err := filepath.Walk(publicDir, func(path string, f os.FileInfo, err error) error {
			if !f.IsDir() {
				fileList = append(fileList, path)
			}
			return nil
		})

		if err == nil {
			for _, file := range fileList {
				staticFile := strings.Replace(file, "public", "", -1)
				pusher.Push(staticFile, nil)
			}
		}
	}

	template := templates.Lookup(templatePath)
	if template != nil {
		err := template.Execute(w, context)
		if err != nil {
			return errors.New("Cannot parse template:" + err.Error())
		} else {
			return nil
		}
	}

	return errors.New("Cannot parse template")

}

func renderError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	context := APImsg{
		Status:  false,
		Message: err.Error(),
		Data:    nil,
	}

	data, _ := json.Marshal(context)

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(data))

}

func renderJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	context := APImsg{
		Status:  true,
		Message: "",
		Data:    data,
	}

	encoded, err := json.Marshal(context)
	if err != nil {
		return errors.New("Cannot parse json: " + err.Error())
	}

	w.Write([]byte(encoded))

	return nil

}

func mainPage(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:  "hash",
		Value: hash,
		Path:  "/",
	}

	http.SetCookie(w, &cookie)
	err := renderView(w, "main.html", nil)
	if err != nil {
		log.Println(err.Error())
	}
}

func apiQuery(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	response, err := connections[id].Query(r.FormValue("query"))

	if err != nil {
		renderError(w, err)
		return
	}

	err = renderJSON(w, response)
	if err != nil {
		log.Println(err.Error())
	}
}

func apiListDatabases(w http.ResponseWriter, r *http.Request) {
	databases, err := connections[0].ListDatabases()
	if err != nil {
		renderError(w, err)
		return
	}

	err = renderJSON(w, databases)
	if err != nil {
		log.Println(err.Error())
	}
}

func apiChangeDatabases(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	connections[id].Disconnect()
	connections[id].Database = r.FormValue("database")
	err := connections[id].Connect()

	if err != nil {
		log.Println(err.Error())
	}

	err = renderJSON(w, true)
	if err != nil {
		log.Println(err.Error())
	}
}

func apiAddConnection(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	connection := new(Connection)

	err := decoder.Decode(connection, r.PostForm)
	if err != nil {
		renderError(w, err)
		return
	}

	err = connection.Connect()
	if err != nil {
		renderError(w, err)
		return
	}

	connections = append(connections, *connection)

	err = renderJSON(w, connection)
	if err != nil {
		log.Println(err.Error())
	}
}

func apiCreateConnection(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	connection := new(Connection)
	err := decoder.Decode(connection, r.PostForm)
	if err != nil {
		renderError(w, err)
		return
	}
	connection.ID = 0
	storage.Create(connection)

	var connections []Connection
	storage.Find(&connections)
	for id, connection := range connections {
		var databases []ActiveDatabase
		storage.Model(&connection).Related(&databases)

		connections[id].Databases = databases
	}

	err = renderJSON(w, connections)
	if err != nil {
		log.Println(err.Error())
	}
}

func apiUpdateConnection(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	connection := new(Connection)

	err := decoder.Decode(connection, r.PostForm)
	if err != nil {
		renderError(w, err)
		return
	}
	storage.Save(&connection)

	err = renderJSON(w, connection)
	if err != nil {
		log.Println(err.Error())
	}
}

func apiGetConnections(w http.ResponseWriter, r *http.Request) {
	var connections []Connection
	storage.Find(&connections)
	for id, connection := range connections {
		var databases []ActiveDatabase
		storage.Model(&connection).Related(&databases)

		connections[id].Databases = databases
	}

	err := renderJSON(w, connections)
	if err != nil {
		log.Println(err.Error())
	}

}

func apiDisconnect(w http.ResponseWriter, r *http.Request) {
	_id := r.FormValue("id")
	id, _ := strconv.Atoi(_id)

	connections[id].Disconnect()
	connections = append(connections[:0], connections[1:]...)

	err := renderJSON(w, true)
	if err != nil {
		log.Println(err.Error())
	}

}
