package acronis_gmail

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/kudinovdenis/logger"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"net/http"
	"net/http/httputil"
	"golang.org/x/oauth2"
)

type LoggingTransport struct {
	delegate http.RoundTripper
}

func (transport *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	data, err := httputil.DumpRequest(req, true)
	if err != nil {
		println("Error while reading request")
		return nil, err
	} else {
		println(string(data))
	}

	response, err := transport.delegate.RoundTrip(req)
	if err != nil {
		return response, err
	}

	data, err = httputil.DumpResponse(response, true)
	if err != nil {
		println("Error while reading response")
		return nil, err
	} else {
		println(string(data))
	}

	return response, nil
}

type GmailClient struct {
	service       *gmail.Service
	currentUserID string
}

func Init(subject string) (*GmailClient, error) {
	client := GmailClient{}

	httpClient := &http.Client{}
	httpClient.Transport = &LoggingTransport{
		http.DefaultTransport,
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	b, err := ioutil.ReadFile("../Acronis-backup-project-8b80e5be7c37.json")
	if err != nil {
		return nil, err
	}

	data, err := google.JWTConfigFromJSON(b, gmail.GmailModifyScope)

	if err != nil {
		logger.Logf(logger.LogLevelError, "JWT Config failed, %v", err)
		return nil, err
	}

	data.Subject = subject

	client.service, err = gmail.New(data.Client(ctx))
	if err != nil {
		logger.Logf(logger.LogLevelError, "New Gmail failed, %v", err)
		return nil, err
	}

	return &client, nil
}

func (client *GmailClient) Backup(account string) (err error) {
	threads, err := client.service.Users.Threads.List(account).Do()
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

		tc := client.service.Users.Threads.Get(account, thread.Id) //.Do()//service.Threads.Get(subject, thread.Id).Do()
		tc = tc.Format("metadata")
		t, err := tc.Do()
		if err != nil {
			logger.Logf(logger.LogLevelError, "Thread Get failed , %v", err)
			return err
		}
		logger.Logf(logger.LogLevelDefault, "Getted Thread Snippet : %s", t.Snippet)
		logger.Logf(logger.LogLevelDefault, "Getted Thread Message Count %v", len(t.Messages))

		err = client.saveMessages(pathToBackup, account, t.Messages)
		if err != nil {
			return err
		}
		logger.Logf(logger.LogLevelDefault, "Ended thread w/ ID : %v", thread.Id)
	}

	return
}

func (client *GmailClient) BackupIndividualMessages(account string) (err error) {
	pathToBackup := "./" + account + "/backup/"
	err = os.MkdirAll(pathToBackup, 0777)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Directory create failed, %v", err)
		return
	}

	nextPageToken := ""
	for {
		listCall := client.service.Users.Messages.List(account)
		if nextPageToken != "" {
			listCall.PageToken(nextPageToken)
		}

		messages, err := listCall.Do()
		if err != nil {
			logger.Logf(logger.LogLevelError, "Message list Get failed , %v", err)
			return err
		}
		logger.Logf(logger.LogLevelDefault, "Got Message Count %v", len(messages.Messages))

		err = client.saveMessages(pathToBackup, account, messages.Messages)
		if err != nil {
			return err
		}

		nextPageToken = messages.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return
}

func (client *GmailClient) saveMessages(pathToBackup, account string, messages []*gmail.Message) error {
	for _, mes := range messages {
		logger.Logf(logger.LogLevelDefault, "Started message w/ ID : %v", mes.Id)
		mc := client.service.Users.Messages.Get(account, mes.Id)
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

	return nil
}

func (client *GmailClient) Restore(account string, pathToBackup string) (err error) {
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
			err = client.restoreThread(account, pathToBackup+fi.Name())
			if err != nil {
				logger.Logf(logger.LogLevelError, "Failed to restore thread if: %v, err: %v", fi.Name(), err.Error())
			}
		}
	}

	return
}

func (client *GmailClient) RestoreIndividualMessages(account, pathToBackup string) error {
	return client.restoreThread(account, pathToBackup)
}

func (client *GmailClient) restoreThread(account string, pathToThread string) (err error) {
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
			err = client.restoreMessage(account, pathToThread+"/"+fi.Name())
			if err != nil {
				logger.Logf(logger.LogLevelError, "Failed to restore thread id: %v, err: %v", fi.Name(), err.Error())
				return
			}
		}
	}

	return
}

func (client *GmailClient) restoreMessage(account string, pathToMsg string) (err error) {
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
	m.LabelIds = msg.LabelIds
	m.ThreadId = msg.ThreadId

	ic := client.service.Users.Messages.Insert(account, m)

	res, err := ic.Do()
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to restore message path: %v, err: %v", pathToMsg, err.Error())
		return err
	}

	logger.Logf(logger.LogLevelDefault, "Inserted msg: %v", res)

	return
}
