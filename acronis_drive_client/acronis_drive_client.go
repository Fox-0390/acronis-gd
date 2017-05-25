package acronis_drive_client

import (
	"google.golang.org/api/drive/v3"
	"io/ioutil"
	"golang.org/x/oauth2/google"
	"context"
	"github.com/kudinovdenis/logger"
)

type DriveClient struct {
	s *drive.Service
}

func Init(subject string) (*DriveClient, error) {
	client := DriveClient{}

	ctx := context.Background()

	b, err := ioutil.ReadFile("./Acronis-data-backup-db3941030528.json")
	if err != nil {
		return nil, err
	}

	data, err := google.JWTConfigFromJSON(b, drive.DriveScope,drive.DriveAppdataScope,drive.DriveFileScope,drive.DriveMetadataScope,drive.DriveMetadataReadonlyScope,drive.DrivePhotosReadonlyScope,drive.DriveReadonlyScope,drive.DriveScriptsScope)
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

	filesRes, err := list.Do()
	if err != nil {
		logger.Log(logger.LogLevelError, err.Error())
		return nil, err
	}

	return filesRes.Files, nil
}