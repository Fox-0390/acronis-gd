package acronis_gmail

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/kudinovdenis/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"net/http"
	"net/http/httputil"
	"strconv"
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

	b, err := ioutil.ReadFile("./Acronis-backup-project-8b80e5be7c37.json")
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

		pathToBackup := "./backups/gmail/" + account + "/backup/" + string(thread.Id) + "/"
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

		for _, mes := range t.Messages {
			_, err := client.saveMessage(account, pathToBackup, mes.Id)
			if err != nil {
				return err
			}
		}
		logger.Logf(logger.LogLevelDefault, "Ended thread w/ ID : %v", thread.Id)
	}

	return
}

func (client *GmailClient) BackupIndividualMessages(account string) (err error) {
	pathToBackup := "./backups/gmail/" + account + "/backup/"

	err = os.RemoveAll(pathToBackup)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Couldn't clear the backup folder, %v", err)
		return
	}

	err = os.MkdirAll(pathToBackup, 0777)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Directory create failed, %v", err)
		return
	}

	nextPageToken := ""
	var latestHistoryId uint64
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

		for _, mes := range messages.Messages {
			historyId, err := client.saveMessage(account, pathToBackup, mes.Id)
			if err != nil {
				return err
			}

			if historyId > latestHistoryId {
				latestHistoryId = historyId
			}
		}

		nextPageToken = messages.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	logger.Logf(logger.LogLevelDefault, "Latest history id: %d", latestHistoryId)
	return client.writeLatestHistoryId(account, latestHistoryId)
}

func (client *GmailClient) saveMessage(account, pathToBackup, messageId string) (uint64, error) {
	logger.Logf(logger.LogLevelDefault, "Started message w/ ID : %v", messageId)
	mc := client.service.Users.Messages.Get(account, messageId)
	mc = mc.Format("raw")
	m, err := mc.Do()
	if err != nil {
		logger.Logf(logger.LogLevelError, "Message Get failed , %v", err)
		return 0, err
	}
	logger.Logf(logger.LogLevelDefault, "Message snippet : %s", m.Snippet)

	marshalled, err := m.MarshalJSON()
	pb := pathToBackup + m.Id

	err = ioutil.WriteFile(pb, marshalled, 0777)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Write to File failed, %v", err)
		return 0, err
	}

	logger.Logf(logger.LogLevelDefault, "Ended message w/ ID : %v", messageId)
	return m.HistoryId, nil
}

func (client *GmailClient) writeLatestHistoryId(account string, latestHistoryId uint64) error {
	data := []byte(strconv.FormatUint(latestHistoryId, 10))
	return ioutil.WriteFile("./backups/gmail/" + account + "/backup.json", data, os.ModePerm)
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
			_, err = client.restoreThread(account, pathToBackup+fi.Name())
			if err != nil {
				logger.Logf(logger.LogLevelError, "Failed to restore thread if: %v, err: %v", fi.Name(), err.Error())
			}
		}
	}

	return
}

func (client *GmailClient) RestoreIndividualMessages(account, pathToBackup string) error {
	latestHistoryId, err := client.restoreThread(account, pathToBackup)
	if err != nil {
		return err
	}

	return client.writeLatestHistoryId(account, latestHistoryId)
}

func (client *GmailClient) restoreThread(account string, pathToThread string) (latestHistoryId uint64, err error) {
	fileList, err := createExistingFileList(pathToThread)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to create file list: %v, err: %v", pathToThread, err.Error())
		return
	}

	var lastMessageId string
	for _, fileName := range fileList {
		lastMessageId, err = client.restoreMessage(account, pathToThread+"/"+fileName)
		if err != nil {
			logger.Logf(logger.LogLevelError, "Failed to restore thread id: %v, err: %v", fileName, err.Error())
			return
		}
	}

	message, err := client.service.Users.Messages.Get(account, lastMessageId).Format("metadata").Do()
	if err != nil {
		logger.Logf(logger.LogLevelError, "Message Get failed , %v", err)
		return 0, err
	}

	latestHistoryId = message.HistoryId
	return
}

func (client *GmailClient) restoreMessage(account string, pathToMsg string) (messageId string, err error) {
	raw, err := ioutil.ReadFile(pathToMsg)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to restore message path: %v, err: %v", pathToMsg, err.Error())
		return
	}

	var msg = &gmail.Message{}

	err = json.Unmarshal(raw, msg)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to unmarshal message path: %v, err: %v", pathToMsg, err.Error())
		return "", err
	}

	var m = &gmail.Message{}
	m.Raw = msg.Raw
	m.LabelIds = msg.LabelIds
	m.ThreadId = msg.ThreadId

	ic := client.service.Users.Messages.Insert(account, m)

	res, err := ic.Do()
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to restore message path: %v, err: %v", pathToMsg, err.Error())
		return "", err
	}

	logger.Logf(logger.LogLevelDefault, "Inserted msg: %v", res)
	messageId = res.Id
	return
}

func (client *GmailClient) BackupIncrementally(account, pathToBackup, pathToBackupDescriptor string) error {
	data, err := ioutil.ReadFile(pathToBackupDescriptor)
	if err != nil {
		return err
	}

	var lastHistoryId uint64
	lastHistoryId, err = strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return err
	}

	return client.backupIncrementally(account, pathToBackup, lastHistoryId)
}

func (client *GmailClient) backupIncrementally(account, pathToBackup string, lastHistoryId uint64) error {
	existingMessages, err := createExistingFileList(pathToBackup)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Failed to create file list: %v, err: %v", pathToBackup, err.Error())
		return err
	}

	existingMessagesSet := make(map[string]struct{})
	for _, message := range existingMessages {
		existingMessagesSet[message] = struct{}{}
	}

	statuses, err := client.createChangeSetFromHistory(account, lastHistoryId, existingMessagesSet)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Couldn't create a status list, err: %v", err.Error())
		return err
	}

	for message, status := range statuses {
		logger.Logf(logger.LogLevelDefault, "Message %s: %d", message, status)
		switch status {
		case statusAdded:
			client.saveMessage(account, pathToBackup, message)
		case statusRemoved:
			os.Remove(pathToBackup + "/" + message)
		}
	}

	return nil
}

func createExistingFileList(pathToFolder string) (fileNames []string, err error) {
	dir, err := os.Open(pathToFolder)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Directory open failed, %v", err)
		return
	}
	defer dir.Close()

	fileList, err := dir.Readdir(-1)
	if err != nil {
		logger.Logf(logger.LogLevelError, "Directory open failed, %v", err)
		return
	}

	for _, file := range fileList {
		if !file.IsDir() {
			logger.Logf(logger.LogLevelDefault, "Found file: %v", file.Name())
			fileNames = append(fileNames, file.Name())
		}
	}

	return
}
