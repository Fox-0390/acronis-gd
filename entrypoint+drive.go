package main

import (
	"github.com/kudinovdenis/logger"
	"path"
	"github.com/kudinovdenis/acronis-gd/utils"
	"github.com/kudinovdenis/acronis-gd/acronis_drive_client"
	"google.golang.org/api/admin/directory/v1"
)

func backupUserGoogleDrive(user *admin.User) {
	logger.Logf(logger.LogLevelDefault, "IN FUNC: %#v", user)
	utils.CreateDirectory(user.Id)
	emails := user.Emails.([]interface{})
	for _, email := range emails {
		email_string := email.(map[string]interface{})["address"].(string)
		utils.CreateDirectory(path.Join(user.Id, email_string))
		logger.Logf(logger.LogLevelDefault, "Backing up User: %+v", email_string)

		drive_client, err := acronis_drive_client.Init(email_string)
		if err != nil {
			processError(err)
			continue
		}

		// Changes
		token, err := drive_client.GetCurrentChangesToken()
		if err != nil {
			processError(err)
			continue
		}

		changes , err := drive_client.GetChangesWithToken(token)
		logger.Logf(logger.LogLevelDefault, "CHANGES: %#v", changes)
		//

		files, err := drive_client.ListAllFiles()
		logger.Logf(logger.LogLevelDefault, "USER: %s FILES: %#v", email_string, files)
		if err != nil {
			processError(err)
			continue
		}

		err = drive_client.SaveChangesToken(path.Join(user.Id, email_string, "token.json"), token)
		if err != nil {
			processError(err)
			continue
		}


		for i, file  := range files {
			logger.Logf(logger.LogLevelDefault, "File %d: %s %s", i, file.Name, file.MimeType)
			// Save metadata
			file_meta, err := drive_client.DownloadMetadata(*file)
			if err != nil {
				processError(err)
				continue
			}
			utils.CreateFileWithReader(path.Join(user.Id, email_string, file.Id + "_meta.json"), file_meta)
			// Save file content
			reader, err := drive_client.GetFileWithReader(*file)
			if err != nil {
				processError(err)
				continue
			}
			utils.CreateFileWithReader(path.Join(user.Id, email_string, file.Id), reader)
		}
	}
}