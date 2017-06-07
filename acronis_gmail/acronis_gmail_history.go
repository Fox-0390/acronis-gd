package acronis_gmail

import "google.golang.org/api/gmail/v1"

const (
	statusAdded = iota
	statusRemoved
)

type Status int

func (client *GmailClient) createChangeSetFromHistory(account string, lastHistoryId uint64,
	existingMessagesInBackup map[string]struct{}) (map[string]Status, error) {
	statuses := make(map[string]Status)

	response, err := client.service.Users.History.List(account).StartHistoryId(lastHistoryId).Do()
	if err != nil {
		return nil, err
	}

	isNotInTrash := func(message *gmail.Message) bool {
		for _, label := range message.LabelIds {
			if label == "TRASH" || label == "DRAFT" {
				return false
			}
		}
		return true
	}

	calculateStatusFromLabels := func(message *gmail.Message) {
		present := isNotInTrash(message)
		_, existsInBackup := existingMessagesInBackup[message.Id];
		if present && !existsInBackup {
			statuses[message.Id] = statusAdded
		} else if !present && existsInBackup {
			statuses[message.Id] = statusRemoved
		} else {
			delete(statuses, message.Id)
		}
	}

	for _, histRecord := range response.History {
		for _, message := range histRecord.MessagesAdded {
			present := isNotInTrash(message.Message)
			if present {
				statuses[message.Message.Id] = statusAdded
			} else {
				delete(statuses, message.Message.Id)
			}
		}

		for _, message := range histRecord.MessagesDeleted {
			_, existsInBackup := existingMessagesInBackup[message.Message.Id];
			if !existsInBackup {
				delete(statuses, message.Message.Id)
			} else {
				statuses[message.Message.Id] = statusRemoved
			}
		}

		for _, message := range histRecord.LabelsAdded {
			calculateStatusFromLabels(message.Message)
		}

		for _, message := range histRecord.LabelsRemoved {
			calculateStatusFromLabels(message.Message)
		}
	}

	return statuses, err
}
