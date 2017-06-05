package acronis_gmail

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/kudinovdenis/logger"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type GmailClient struct {
	s             *gmail.Service
	currentUserID string
}

func Init(subject string) (*GmailClient, error) {
	client := GmailClient{}

	ctx := context.Background()

	b, err := ioutil.ReadFile("../Acronis-backup-project-8b80e5be7c37.json")
	if err != nil {
		return nil, err
	}

	data, err := google.JWTConfigFromJSON(b,
		gmail.MailGoogleComScope,
		gmail.GmailComposeScope,
		gmail.GmailInsertScope,
		gmail.GmailLabelsScope,
		gmail.GmailModifyScope,
		gmail.GmailReadonlyScope,
		gmail.GmailSendScope,
		gmail.GmailSettingsBasicScope,
		gmail.GmailSettingsSharingScope,
	)

	if err != nil {
		logger.Logf(logger.LogLevelError, "JWT Config failed, %v", err)
		return nil, err
	}

	data.Subject = subject

	client.s, err = gmail.New(data.Client(ctx))
	if err != nil {
		logger.Logf(logger.LogLevelError, "New Gmail failed, %v", err)
		return nil, err
	}

	return &client, nil
}

func (c *GmailClient) Backup(account string) (err error) {

	threads, err := c.s.Users.Threads.List(account).Do()
	if err != nil {
		logger.Logf(logger.LogLevelError, "Threads List failed , %v", err)
		return
	}

	logger.Logf(logger.LogLevelDefault, "%v", len(threads.Threads))

	for _, thread := range threads.Threads {
		logger.Logf(logger.LogLevelDefault, "Started thread w/ ID : %v", thread.Id)
		logger.Logf(logger.LogLevelDefault, "Thread snippet : %s ", thread.Snippet)

		pathToBackup := "./" + account + "/backup/" + string(thread.Id) + "/"
		err = os.MkdirAll(pathToBackup, 0777)
		if err != nil {
			logger.Logf(logger.LogLevelError, "Directory create failed, %v", err)
			return
		}

		tc := c.s.Users.Threads.Get(account, thread.Id) //.Do()//service.Threads.Get(subject, thread.Id).Do()
		tc = tc.Format("metadata")
		t, err := tc.Do()
		if err != nil {
			logger.Logf(logger.LogLevelError, "Thread Get failed , %v", err)
			return err
		}
		logger.Logf(logger.LogLevelDefault, "Getted Thread Snippet : %s", t.Snippet)
		logger.Logf(logger.LogLevelDefault, "Getted Thread Message Count %v", len(t.Messages))

		for _, mes := range t.Messages {
			logger.Logf(logger.LogLevelDefault, "Started message w/ ID : %v", mes.Id)
			mc := c.s.Users.Messages.Get(account, mes.Id)
			mc = mc.Format("raw")
			m, err := mc.Do()
			if err != nil {
				logger.Logf(logger.LogLevelError, "Message Get failed , %v", err)
				return err
			}
			logger.Logf(logger.LogLevelDefault, "Message snippet : %s", m.Snippet)

			marshalled, err := m.MarshalJSON()
			pb := pathToBackup + m.Id

			err = ioutil.WriteFile(pb, marshalled, 0777)
			if err != nil {
				logger.Logf(logger.LogLevelError, "Write to File failed, %v", err)
				return err
			}

			logger.Logf(logger.LogLevelDefault, "Ended message w/ ID : %v", mes.Id)
		}
		logger.Logf(logger.LogLevelDefault, "Ended thread w/ ID : %v", thread.Id)
	}

	return
}

func (c *GmailClient) Restore(account string, pathToBackup string) (err error) {
	d, err := os.Open(pathToBackup)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Directory open failed, %v", err)
		return
	}
	defer d.Close()
	fi, err := d.Readdir(-1)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Directory open failed, %v", err)
		return
	}
	for _, fi := range fi {

		if fi.IsDir() {
			logger.Logf(logger.LogLevelDefault, "Found dir: %v", fi.Name())
			err = c.restoreThread(account, pathToBackup+fi.Name())
			if err != nil {
				logger.Logf(logger.LogLevelError, "Failed to restore thread if: %v, err: %v", fi.Name(), err.Error())
			}
		}
	}

	return
}

func (c *GmailClient) restoreThread(account string, pathToThread string) (err error) {
	d, err := os.Open(pathToThread)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Directory open failed, %v", err)
		return
	}
	defer d.Close()
	fi, err := d.Readdir(-1)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Directory open failed, %v", err)
		return
	}
	for _, fi := range fi {

		if !fi.IsDir() {
			logger.Logf(logger.LogLevelDefault, "Found file: %v", fi.Name())
			err = c.restoreMessage(account, pathToThread+"/"+fi.Name())
			if err != nil {
				logger.Logf(logger.LogLevelError, "Failed to restore thread id: %v, err: %v", fi.Name(), err.Error())
				return
			}
		}
	}

	return
}

func (c *GmailClient) restoreMessage(account string, pathToMsg string) (err error) {
	raw, err := ioutil.ReadFile(pathToMsg)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to restore message path: %v, err: %v", pathToMsg, err.Error())
		return
	}

	var msg = &gmail.Message{}

	err = json.Unmarshal(raw, msg)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to unmarshal message path: %v, err: %v", pathToMsg, err.Error())
		return err
	}

	var m = &gmail.Message{}
	m.Raw = msg.Raw
	m.Payload = msg.Payload
	m.SizeEstimate = msg.SizeEstimate
	m.LabelIds = msg.LabelIds
	m.Snippet = msg.Snippet
	m.Header = msg.Header
	m.InternalDate = msg.InternalDate

	ic := c.s.Users.Messages.Insert(account, m)

	res, err := ic.Do()
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to restore message path: %v, err: %v", pathToMsg, err.Error())
		return err
	}

	logger.Logf(logger.LogLevelDefault, "Inserted msg: %v", res)

	return
}
