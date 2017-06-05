package main

import (
	"github.com/kudinovdenis/logger"
	"path"
	"github.com/kudinovdenis/acronis-gd/utils"
	"github.com/kudinovdenis/acronis-gd/acronis_drive_client"
	"google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/drive/v3"
)

func backupUserGoogleDrive(user *admin.User) {
	emails := user.Emails.([]interface{})

	for _, email := range emails {
		email_string := email.(map[string]interface{})["address"].(string)
		userEmailDirectoryPath := path.Join(user.Id, email_string)

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
			err := processChanges(user, email_string, drive_client)
			if err != nil {
				processError(err)
				continue
			}
		} else {
			logger.Logf(logger.LogLevelDefault, "Full backup for User: %+v started", email)
			err := fullBackup(user, email_string, drive_client)
			if err != nil {
				processError(err)
				continue
			}
		}
	}
}

func processChanges(user *admin.User, email string, drive_client *acronis_drive_client.DriveClient) error {
	userEmailDirectoryPath := path.Join(user.Id, email)
	userEmailChangesTokenPath := path.Join(user.Id, email, "token.json")

	changesToken , err := drive_client.LoadChangesToken(userEmailChangesTokenPath)
	if err != nil {
		return err
	}

	changes, err := drive_client.GetChangesWithToken(changesToken)
	if err != nil {
		return err
	}

	for _, change := range changes.Changes {
		logger.Logf(logger.LogLevelDefault, "Processing change: %#v", change)
		if change.Type == "file" {
			fileID := change.FileId

			filePath := path.Join(userEmailDirectoryPath, fileID)
			fileMetaPath := path.Join(userEmailDirectoryPath, fileID + "_meta.json")

			logger.Logf(logger.LogLevelDefault, "FILE PATH: %s, FILEMETAPATH: %s", filePath, fileMetaPath)

			// Remove and then add again
			err := utils.RemoveFile(filePath)
			if err != nil {
				processError(err)
			}
			utils.RemoveFile(fileMetaPath)
			if err != nil {
				processError(err)
			}

			// If file is not removed
			if change.File != nil && !change.Removed && !change.File.Trashed {
				file_info, err := drive_client.GetFileInfo(fileID)
				if err != nil {
					processError(err)
				}

				err = backupFile(user, email, file_info, drive_client)
				if err != nil {
					processError(err)
				}
			}
		}
	}

	// Save Changes Token
	token, err := drive_client.GetCurrentChangesToken()
	if err != nil {
		return err
	}

	err = drive_client.SaveChangesToken(userEmailChangesTokenPath, token)
	if err != nil {
		return err
	}
	//

	return nil
}

func fullBackup(user *admin.User, email string, drive_client *acronis_drive_client.DriveClient) error {
	userEmailDirectoryPath := path.Join(user.Id, email)
	userEmailChangesTokenPath := path.Join(user.Id, email, "token.json")
	
	utils.CreateDirectory(userEmailDirectoryPath)

	files, err := drive_client.ListAllFiles()
	if err != nil {
		return err
	}

	// Save Changes Token
	token, err := drive_client.GetCurrentChangesToken()
	if err != nil {
		return err
	}

	err = drive_client.SaveChangesToken(userEmailChangesTokenPath, token)
	if err != nil {
		return err
	}
	//

	for _, file := range files {
		err := backupFile(user, email, file, drive_client)
		if err != nil {
			processError(err)
		}
	}

	return nil
}

func backupFile(user *admin.User, email string, file *drive.File, drive_client *acronis_drive_client.DriveClient) error {
	logger.Logf(logger.LogLevelDefault, "Backing up file: %s %s", file.Name, file.MimeType)
	// Save metadata
	file_meta, err := drive_client.DownloadMetadata(*file)
	if err != nil {
		return err
	}
	utils.CreateFileWithReader(path.Join(user.Id, email, file.Id + "_meta.json"), file_meta)
	// Save file content
	reader, err := drive_client.GetFileWithReader(*file)
	if err != nil {
		return err
	}
	utils.CreateFileWithReader(path.Join(user.Id, email, file.Id), reader)

	return nil
}