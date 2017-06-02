package main

import (
	"github.com/kudinovdenis/acronis-gd/acronis_admin_client"
	"github.com/kudinovdenis/logger"
	"google.golang.org/api/admin/directory/v1"
	"sync"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"net/http"
)

var errors = []error{}

func clientHandler(rw http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	admin_email := r.URL.Query().Get("admin_email")

	if domain == "" || admin_email == "" {
		http.Error(rw, "Must provide domain and admin_email.", http.StatusBadRequest)
		return
	}

	rw.Write(
		[]byte(
			"<h1>Improvised Admin Panel</h1>" +
			"<div><a href=\"/backup?domain=" + domain + "\">backup now</a></div>" +
			"<div><a href=\"/users?domain=" + domain + "\">show users</a></div>" +
			"<br><br><br><br><br><br><br><br><br><br><br><br>" +
			"<div><a href=\"http://cs6.pikabu.ru/post_img/2015/06/09/10/1433867902_2044988577.jpg\">\"Не быть тебе дизайнером\"</a></div>"))
}

func backupHandler(rw http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	admin_email := r.URL.Query().Get("admin_email")

	if domain == "" || admin_email == "" {
		http.Error(rw, "Must provide domain and admin_email.", http.StatusBadRequest)
		return
	}

	admin_client, err := acronis_admin_client.Init(domain, admin_email)
	if err != nil {
		logger.Logf(logger.LogLevelDefault, "Cant initialize admin client. %s", err.Error())
		return
	}

	users, err := admin_client.GetListOfUsers()
	if err != nil {
		logger.Logf(logger.LogLevelDefault, "Error while getting list of users: %s", err.Error())
		return
	}

	group := sync.WaitGroup{}
	group.Add(len(users.Users))

	// Google drive
	for _, user := range users.Users {
		go func(user *admin.User) {
			backupUserGoogleDrive(user)
			group.Done()
		}(user)
	}

	group.Wait()
	logger.Logf(logger.LogLevelDefault, "Errors: %d. %#v", len(errors), errors)
}

func usersHandler(rw http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	admin_email := r.URL.Query().Get("admin_email")

	if domain == "" || admin_email == "" {
		http.Error(rw, "Must provide domain and admin_email.", http.StatusBadRequest)
		return
	}

	admin_client, err := acronis_admin_client.Init(domain, admin_email)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		logger.Logf(logger.LogLevelDefault, "Cant initialize admin client. %s", err.Error())
		return
	}

	users, err := admin_client.GetListOfUsers()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		logger.Logf(logger.LogLevelDefault, "Error while getting list of users: %s", err.Error())
		return
	}

	htmlListOfUsers := "<div><ul>"
	for _, user := range users.Users {
		logger.Logf(logger.LogLevelDefault, "users: %#v", user)
		htmlListOfUsers += "<li>" + "ID: " + user.Id + " Name: " + user.Name.FullName + "</li>"
	}
	htmlListOfUsers += "</ul></div>"

	rw.Write(
		[]byte(
			"<h1>Improvised Admin Panel</h1>" +
			"<h2>Users</h2>" +
			htmlListOfUsers))
}

func main() {
	n := negroni.Classic()

	r := mux.NewRouter()
	r.HandleFunc("/client", clientHandler).Methods("GET")
	r.HandleFunc("/backup", backupHandler).Methods("GET")
	r.HandleFunc("/users", usersHandler).Methods("GET")

	n.UseHandler(r)

	logger.Logf(logger.LogLevelError, "%s", http.ListenAndServe(":8989", n).Error())
}


func processError(err error) {
	logger.Logf(logger.LogLevelError, "Error: %#v", err.Error())
	errors = append(errors, err)
}