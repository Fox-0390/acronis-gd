package main

import (
	"github.com/kudinovdenis/logger"
	"path"
	"github.com/kudinovdenis/acronis-gd/utils"
	"github.com/kudinovdenis/acronis-gd/acronis_drive_client"
	"google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/drive/v2"
)

func backupUserGoogleDrive(user *admin.User) {
	utils.CreateDirectory(user.Id)
	emails := user.Emails.([]interface{})

	for _, email := range emails {
		email_string := email.(map[string]interface{})["address"].(string)
		userEmailDirectoryPath := path.Join(user.Id, email_string)
		userEmailChangesTokenPath := path.Join(user.Id, email_string, "token.json")

		incrementalBackupNeeded, err := utils.IsFileExists(userEmailDirectoryPath)
		if err != nil {
			processError(err)
			incrementalBackupNeeded = false
		}

		drive_client, err := acronis_drive_client.Init(email_string)
		if err != nil {
			processError(err)
			continue
		}

		if incrementalBackupNeeded {
			logger.Logf(logger.LogLevelDefault, "Incremental backup for User: %+v started", email_string)

			changesToken , err := drive_client.LoadChangesToken(userEmailChangesTokenPath)
			if err != nil {
				processError(err)
				continue
			}

			changes, err := drive_client.GetChangesWithToken(changesToken)
			if err != nil {
				processError(err)
				continue
			}

			logger.Logf(logger.LogLevelDefault, "CHANGES: %#v", changes)
			//
		} else {
			logger.Logf(logger.LogLevelDefault, "Full backup for User: %+v started", email_string)

			utils.CreateDirectory(userEmailDirectoryPath)

			files, err := drive_client.ListAllFiles()
			logger.Logf(logger.LogLevelDefault, "USER: %s FILES: %#v", email_string, files)
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

			err = drive_client.SaveChangesToken(userEmailChangesTokenPath, token)
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
				utils.CreateFileWithReader(path.Join(user.Id, email_string, file.Id+"_meta.json"), file_meta)
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
}

func processChanges(list drive.ChangeList, user admin.Users) {
	
}