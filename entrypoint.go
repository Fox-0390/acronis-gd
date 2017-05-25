package main

import (
	"github.com/kudinovdenis/acronis-gd/acronis_admin_client"
	"github.com/kudinovdenis/logger"
	"github.com/kudinovdenis/acronis-gd/acronis_drive_client"
)

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

	for i, user := range users.Users {
		emails := user.Emails.([]interface{})
		for _, email := range emails {
			logger.Logf(logger.LogLevelDefault, "User %d: %+v", i, email)

			drive_client, err := acronis_drive_client.Init(email.(map[string]interface{})["address"].(string))
			if err != nil {
				logger.Logf(logger.LogLevelDefault, "Error while getting list of user files: %s", err.Error())
				return
			}

			files, err := drive_client.ListAllFiles()
			if err != nil {
				logger.Logf(logger.LogLevelDefault, "Error while getting list of user files: %s", err.Error())
				return
			}

			for i, file  := range files {
				logger.Logf(logger.LogLevelDefault, "File %d: %+v", i, file.Name)
			}
		}
	}

	// Work with users as you want to
	// see example below
}