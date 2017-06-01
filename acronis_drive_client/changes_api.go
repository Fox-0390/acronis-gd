package acronis_drive_client

import (
	"encoding/json"
	"github.com/kudinovdenis/acronis-gd/utils"
	"bytes"
	"io/ioutil"
	"google.golang.org/api/drive/v3"
)

type userChangesToken struct {
	Token string `json:"token"`
}

func (c *DriveClient) GetCurrentChangesToken() (string, error) {
	changes, err := c.s.Changes.GetStartPageToken().Do()
	if err != nil {
		return "", err
	}

	return changes.StartPageToken, nil
}

// Not saving just string because in future it might be not just a string, but a user info
func (c *DriveClient) SaveChangesToken(path string, token string) error {
	tokenJson := userChangesToken{token}

	b, err := json.Marshal(tokenJson)
	if err != nil {
		return err
	}

	return utils.CreateFileWithReader(path, bytes.NewReader(b))
}

func (c *DriveClient) LoadChangesToken(path string) (string, error) {
	token := userChangesToken{}

	reader, err := utils.ReadFile(path)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(b, &token)
	if err != nil {
		return "", err
	}

	return token.Token, nil
}

func (c *DriveClient) GetChangesWithToken(token string) (*drive.ChangeList, error) {
	changes, err := c.s.Changes.List(token).Do()
	return changes, err
}
