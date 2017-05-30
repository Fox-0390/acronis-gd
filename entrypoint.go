package main

import (
	"github.com/kudinovdenis/acronis-gd/acronis_admin_client"
	"github.com/kudinovdenis/logger"
	"github.com/kudinovdenis/acronis-gd/acronis_drive_client"
	"github.com/kudinovdenis/acronis-gd/utils"
	"path"
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

	// Google drive
	for i, user := range users.Users {
		utils.CreateDirectory(user.Id)
		emails := user.Emails.([]interface{})
		for _, email := range emails {
			email_string := email.(map[string]interface{})["address"].(string)
			logger.Logf(logger.LogLevelDefault, "\nUser %d: %+v", i, email_string)

			drive_client, err := acronis_drive_client.Init(email_string)
			if err != nil {
				logger.Logf(logger.LogLevelDefault, "Error while getting list of user files: %s", err.Error())
				errors = append(errors, err)
				continue
			}

			files, err := drive_client.ListAllFiles()
			if err != nil {
				logger.Logf(logger.LogLevelDefault, "Error while getting list of user files: %s", err.Error())
				errors = append(errors, err)
				continue
			}

			for i, file  := range files {
				logger.Logf(logger.LogLevelDefault, "File %d: %#v", i, file)
				reader, err := drive_client.GetFileWithReader(*file)
				if err != nil {
					logger.Logf(logger.LogLevelDefault, "Error while exporting file: %s", err.Error())
					errors = append(errors, err)
					continue
				}
				utils.CreateFileWithReader(path.Join(user.Id, file.Name), reader)
			}
		}
	}

	logger.Logf(logger.LogLevelDefault, "Errors: %d. %#v", len(errors), errors)
}