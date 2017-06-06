package acronis_gmail

import (
	"testing"
)

const email = "monica.geller@trueimage.eu"

func TestGmailClient_Backup(t *testing.T) {
	client, err := Init(email)
	if err != nil {
		t.Errorf("Failed to Create service backup %s, err: %v", email, err.Error())
		t.FailNow()
	}

	err = client.Backup(email)
	if err != nil {
		t.Errorf("Failed to backup, err: %v", err.Error())
		t.FailNow()
	}
}

func TestGmailClient_Restore(t *testing.T) {
	client, err := Init(email)
	if err != nil {
		t.Errorf("Failed to create the service: %s", err.Error())
		t.FailNow()
	}

	err = client.Restore(email, "./"+email+"/backup/")
	if err != nil {
		t.Errorf("Failed to restore the messages: %s", err.Error())
		t.FailNow()
	}
}
