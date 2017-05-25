package main

import (
	"github.com/kudinovdenis/acronis-gd/admin_client"
	"github.com/kudinovdenis/logger"
)

func main() {
	err := admin_client.Init()
	if err != nil {
		logger.Logf(logger.LogLevelDefault, "Cant initialize admin client. %s", err.Error())
		return
	}

	users, err := admin_client.GetListOfUsers()
	if err != nil {
		logger.Logf(logger.LogLevelDefault, "Error while getting list of users: %s", err.Error())
		return
	}

	for i, user := range users.Users {
		logger.Logf(logger.LogLevelDefault, "User %d: %+v", i, user.Emails)
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