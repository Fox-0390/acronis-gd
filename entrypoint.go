package main

import (
	//"google.golang.org/api/drive/v3"
	"google.golang.org/api/admin/directory/v1"
	"golang.org/x/oauth2/google"
	"github.com/kudinovdenis/logger"
	"io/ioutil"
	"context"
)

func main() {
	// Use oauth2.NoContext if there isn't a good context to pass in.
	ctx := context.Background()

	b, err := ioutil.ReadFile("./Acronis-data-backup-58ecc97b43ae.json")
	if err != nil {
		logger.Log(logger.LogLevelError, err.Error())
		return
	}

	data, err := google.JWTConfigFromJSON(b, admin.AdminDirectoryUserScope, admin.AdminDirectoryUserReadonlyScope)// admin.AdminDirectoryUserReadonlyScope, "https://www.googleapis.com/auth/drive","https://www.googleapis.com/auth/drive.file","https://www.googleapis.com/auth/drive.readonly","https://www.googleapis.com/auth/drive.metadata.readonly","https://www.googleapis.com/auth/drive.metadata","https://www.googleapis.com/auth/drive.photos.readonly")
	if err != nil {
		logger.Log(logger.LogLevelError, err.Error())
		return
	}

	data.Subject = "admin@dkudinov.com"

	logger.Logf(logger.LogLevelDefault, "DATA: %+v", data)

	adminService, err := admin.New(data.Client(ctx))
	if err != nil {
		logger.Log(logger.LogLevelError, err.Error())
		return
	}
	usersListCall := adminService.Users.List()
	usersListCall = usersListCall.Domain("dkudinov.com")
	usersListCall = usersListCall.Projection("full")
	usersListCall = usersListCall.MaxResults(5)

	res, err := usersListCall.Do()
	if err != nil {
		logger.Log(logger.LogLevelError, err.Error())
		return
	}

	for i := range res.Users {
		logger.Logf(logger.LogLevelDefault, "User %d: %+v", i, res.Users[i].Emails)
	}

	// Work with users as you want to
	// see example below
}

/*

	driveService, err := drive.New(client)
	if err != nil {
		logger.Log(logger.LogLevelError, err.Error())
		return
	}

	list := driveService.Files.List()

	filesRes, err := list.Do()
	if err != nil {
		logger.Log(logger.LogLevelError, err.Error())
		return
	}

	for i, file  := range filesRes.Files {
		logger.Logf(logger.LogLevelDefault, "File %d: %+v", i, file.Name)
	}

*/