package acronis_drive_client

import (
	"google.golang.org/api/drive/v3"
	"io/ioutil"
	"golang.org/x/oauth2/google"
	"context"
	"github.com/kudinovdenis/logger"
	"io"
	"encoding/json"
	"bytes"
)

type DriveClient struct {
	s *drive.Service
	currentUserID string
}

func Init(subject string) (*DriveClient, error) {
	client := DriveClient{}

	ctx := context.Background()

	b, err := ioutil.ReadFile("./Acronis-backup-project-8b80e5be7c37.json")
	if err != nil {
		return nil, err
	}

	data, err := google.JWTConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, err
	}

	data.Subject = subject

	client.s, err = drive.New(data.Client(ctx))
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (c *DriveClient) ListAllFiles() ([]*drive.File, error) {
	list := c.s.Files.List()
	list = list.Fields("files(id,name,mimeType,createdTime,modifiedTime,viewedByMeTime,trashed,starred,parents,properties,appProperties)")

	filesRes, err := list.Do()
	if err != nil {
		return nil, err
	}

	return filesRes.Files, nil
}

func (c *DriveClient) GetFileInfo(id string) (*drive.File, error) {
	req := c.s.Files.Get(id)
	req.Fields("id,name,mimeType,createdTime,modifiedTime,viewedByMeTime,trashed,starred,parents,properties,appProperties")

	file, err := req.Do()
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (c *DriveClient) GetFileWithReader(file drive.File) (io.ReadCloser, error) {
	if isExportable(file) {
		return c.exportFileWithReader(file)
	} else {
		return c.downloadFileWithReader(file)
	}
}

func (c *DriveClient) downloadFileWithReader(file drive.File) (io.ReadCloser, error) {
	res, err := c.s.Files.Get(file.Id).Download()
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

// Metadata

func (c* DriveClient) DownloadMetadata(file drive.File) (io.Reader, error) {
	b, err := json.Marshal(file)
	if err != nil {
		return nil, err
	}
	buff := bytes.NewReader(b)
	return buff, nil
}

// Support

func (c *DriveClient) exportFileWithReader(file drive.File) (io.ReadCloser, error) {
	exportedMimeType := exportMimeTypeForMimeType(file.MimeType)
	logger.Logf(logger.LogLevelDefault, "Exporting %s (mime type: %s) with mime-type %s", file.Name, file.MimeType, exportedMimeType)
	res, err := c.s.Files.Export(file.Id, exportedMimeType).Download()
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

/*
application/vnd.google-apps.audio
application/vnd.google-apps.document	Google Docs
application/vnd.google-apps.drawing	Google Drawing
application/vnd.google-apps.file	Google Drive file
application/vnd.google-apps.folder	Google Drive folder
application/vnd.google-apps.form	Google Forms
application/vnd.google-apps.fusiontable	Google Fusion Tables
application/vnd.google-apps.map	Google My Maps
application/vnd.google-apps.photo
application/vnd.google-apps.presentation	Google Slides
application/vnd.google-apps.script	Google Apps Scripts
application/vnd.google-apps.sites	Google Sites
application/vnd.google-apps.spreadsheet	Google Sheets
application/vnd.google-apps.unknown
application/vnd.google-apps.video
application/vnd.google-apps.drive-sdk	3rd party shortcut
*/

var exportableMimeTypes = []string{
	"application/vnd.google-apps.audio",
	"application/vnd.google-apps.document",
	"application/vnd.google-apps.drawing",
	"application/vnd.google-apps.file",
	//"application/vnd.google-apps.folder",
	"application/vnd.google-apps.form",
	"application/vnd.google-apps.fusiontable",
	"application/vnd.google-apps.map",
	"application/vnd.google-apps.photo",
	"application/vnd.google-apps.presentation",
	"application/vnd.google-apps.script",
	"application/vnd.google-apps.sites",
	"application/vnd.google-apps.spreadsheet",
	"application/vnd.google-apps.unknown",
	"application/vnd.google-apps.video",
	"application/vnd.google-apps.drive-sdk",
}

func exportMimeTypeForMimeType(mimeType string) string {
	switch mimeType {
	case "application/vnd.google-apps.document":
		fallthrough
	case "application/vnd.google-apps.drawing":
		return "application/pdf"
	case "application/vnd.google-apps.spreadsheet":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "application/vnd.google-apps.script":
		return "application/vnd.google-apps.script+json"
	case "application/vnd.google-apps.folder":
		return "application/vnd.google-apps.folder"
	}
	logger.Logf(logger.LogLevelError, "No export mime type found for %s", mimeType)
	return ""
}

func isExportable(file drive.File) bool {
	for _, exportableType := range exportableMimeTypes {
		if file.MimeType == exportableType {
			return true
		}
	}
	return false
}