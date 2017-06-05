package main

import (
	"github.com/kudinovdenis/acronis-gd/acronis_admin_client"
	"github.com/kudinovdenis/logger"
	"google.golang.org/api/admin/directory/v1"
	"sync"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"net/http"
	"github.com/kudinovdenis/acronis-gd/acronis_gmail"
	"github.com/kudinovdenis/acronis-gd/config"
	"flag"
	"github.com/kudinovdenis/acronis-gd/utils"
	"io/ioutil"
)

var errors = []error{}

func GmailToGmail() {
	
	gb, err := acronis_gmail.Init("ao@dkudinov.com")
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to Create service backup ao, err: %v", err.Error())
	}
	
	err = gb.Backup("ao@dkudinov.com")
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to backup, err: %v", err.Error())
	}
	
	gr, err := acronis_gmail.Init("to@dkudinov.com")
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to Create service backup to, err: %v", err.Error())
	}
	err = gr.Restore("to@dkudinov.com", "./ao@dkudinov.com/backup/")
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to resotre, err: %v", err.Error())
	}
	
}


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
			"<div><a href=\"/backup?domain=" + domain + "&admin_email=" + admin_email + "\">backup now</a></div>" +
			"<div><a href=\"/users?domain=" + domain + "&admin_email=" + admin_email + "\">show users</a></div>" +
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

func googleDomainVerificationHandler(rw http.ResponseWriter, r *http.Request) {
	reader, err := utils.ReadFile("./google7ded6bed08ed3c1b.html")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write(b)
}

func main() {
	mode := flag.String("mode", "prod", "`debug` or `prod` mode")
	flag.Parse()
	err := config.PopulateConfigWithFile(*mode + ".json")
	if err != nil {
		logger.Logf(logger.LogLevelError, "Cant read config file. %s", err.Error())
		return
	}
	logger.Logf(logger.LogLevelDefault, "Mode: %s", *mode)
	logger.Logf(logger.LogLevelDefault, "Config: %#v", config.Cfg)

	n := negroni.Classic()

	r := mux.NewRouter()
	// Main admin panel
	r.HandleFunc("/client", clientHandler).Methods("GET")
	// Methods
	r.HandleFunc("/backup", backupHandler).Methods("GET")
	r.HandleFunc("/users", usersHandler).Methods("GET")
	// Registration / authorization flow
	r.HandleFunc("/authorize", authorizationHandler).Methods("GET")
	r.HandleFunc("/oauth2callback", oauth2CallbackHandler).Methods("GET")
	// Google domain verification
	r.HandleFunc("/google7ded6bed08ed3c1b.html", googleDomainVerificationHandler).Methods("GET")

	n.UseHandler(r)

	if config.Cfg.UseLocalServer {
		logger.Logf(logger.LogLevelError, "%s", http.ListenAndServe(config.Cfg.Port, n).Error())
	} else {
		logger.Logf(logger.LogLevelError, "%s", http.ListenAndServeTLS(config.Cfg.Port, "/etc/letsencrypt/live/dkudinov.com/cert.pem", "/etc/letsencrypt/live/dkudinov.com/privkey.pem", n))
	}
}


func processError(err error) {
	logger.Logf(logger.LogLevelError, "Error: %#v", err.Error())
	errors = append(errors, err)
}