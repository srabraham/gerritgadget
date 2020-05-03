package gerritchange

import (
	"fmt"
	"regexp"
	"strings"
)

type CommitMessage struct {
	Message string
}

func (cm CommitMessage) readFooterRow(key string) (string, bool, error) {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key+":") + `.*$`)
	matches := re.FindAllStringSubmatch(cm.Message, -1)
	if len(matches) > 1 {
		return "", true, fmt.Errorf("found footer key %v twice in commit message %v", key, cm.Message)
	}
	if len(matches) == 1 {
		if len(matches[0]) != 1 {
			// This can't happen.
			return "", false, fmt.Errorf("regexp should have one matching group, but it has %v. How did you pull this off? o_O", matches[0])
		}
		return matches[0][0], true, nil
	}
	// No row found for the key
	return "", false, nil
}

func (cm CommitMessage) WithFooter(key, value string) (CommitMessage, error) {
	footerForKey, hasFooterAlready, err := cm.readFooterRow(key)
	if err != nil {
		return CommitMessage{}, err
	}
	newFooter := fmt.Sprintf("%v: %v", key, value)
	if hasFooterAlready {
		return CommitMessage{Message: strings.ReplaceAll(cm.Message, footerForKey, newFooter)}, nil
	}
	// A footer for this key doesn't exist yet. We'll try to put it before the Change-Id row.
	changeIdFooter, hasChangeId, err := cm.readFooterRow("Change-Id")
	if err != nil {
		return CommitMessage{}, err
	}
	if hasChangeId {
		return CommitMessage{Message: strings.ReplaceAll(cm.Message, changeIdFooter, newFooter+"\n"+changeIdFooter)}, nil
	}
	return CommitMessage{Message: cm.Message + "\n" + newFooter}, nil
}
