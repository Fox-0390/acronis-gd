package main

import (
	"github.com/kudinovdenis/acronis-gd/acronis_admin_client"
	"github.com/kudinovdenis/logger"
	"google.golang.org/api/admin/directory/v1"
	"sync"
)

var errors = []error{}

func main() {

	admin_client, err := acronis_admin_client.Init()
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


func processError(err error) {
	logger.Logf(logger.LogLevelError, "Error: %#v", err.Error())
	errors = append(errors, err)
}