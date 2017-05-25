package admin_info

import (
	"google.golang.org/api/admin/directory/v1"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"context"
)

func GetListOfUsers() (*admin.Users, error) {
	ctx := context.Background()

	b, err := ioutil.ReadFile("./Acronis-data-backup-58ecc97b43ae.json")
	if err != nil {
		return nil, err
	}

	data, err := google.JWTConfigFromJSON(b, admin.AdminDirectoryUserScope, admin.AdminDirectoryUserReadonlyScope)// admin.AdminDirectoryUserReadonlyScope, "https://www.googleapis.com/auth/drive","https://www.googleapis.com/auth/drive.file","https://www.googleapis.com/auth/drive.readonly","https://www.googleapis.com/auth/drive.metadata.readonly","https://www.googleapis.com/auth/drive.metadata","https://www.googleapis.com/auth/drive.photos.readonly")
	if err != nil {
		return nil, err
	}

	data.Subject = "admin@dkudinov.com"

	adminService, err := admin.New(data.Client(ctx))
	if err != nil {
		return nil, err
	}
	usersListCall := adminService.Users.List()
	usersListCall = usersListCall.Domain("dkudinov.com")
	usersListCall = usersListCall.Projection("full")

	res, err := usersListCall.Do()
	if err != nil {
		return nil, err
	}

	return res, nil
}