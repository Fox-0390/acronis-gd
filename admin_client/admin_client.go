package admin_client

import (
	"google.golang.org/api/admin/directory/v1"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"context"
	"fmt"
)

var service *admin.Service

func Init() error {
	ctx := context.Background()

	b, err := ioutil.ReadFile("./Acronis-data-backup-58ecc97b43ae.json")
	if err != nil {
		return err
	}

	data, err := google.JWTConfigFromJSON(b, admin.AdminDirectoryUserScope)
	if err != nil {
		return err
	}

	data.Subject = "admin@dkudinov.com"

	service, err = admin.New(data.Client(ctx))
	if err != nil {
		return err
	}

	return nil
}

func GetListOfUsers() (*admin.Users, error) {
	if service == nil {
		return nil, fmt.Errorf("Service not initialized. Call admin_client.Init() first.")
	}
	usersListCall := service.Users.List()
	usersListCall = usersListCall.Domain("dkudinov.com")
	usersListCall = usersListCall.Projection("full")

	res, err := usersListCall.Do()
	if err != nil {
		return nil, err
	}

	return res, nil
}