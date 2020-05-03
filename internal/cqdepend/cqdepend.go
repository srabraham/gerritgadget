package cqdepend

import (
	"fmt"
	aggerrit "github.com/andygrunwald/go-gerrit"
	"github.com/srabraham/gerritgadget/internal/gerritchange"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	gerritRateLimit = time.Second / 10 // 10 calls per second
)

func getCommitMessage(authedClient *http.Client, cl gerritchange.ChangeList) (gerritchange.CommitMessage, error) {
	var cm gerritchange.CommitMessage
	agClient, err := aggerrit.NewClient(cl.LongHost(), authedClient)
	if err != nil {
		return cm, fmt.Errorf("failed to create Gerrit client: %v", err)
	}
	ci, _, err := agClient.Changes.GetCommit(strconv.FormatInt(cl.ChangeNum, 10), "current", &aggerrit.CommitOptions{})
	if err != nil {
		return cm, fmt.Errorf("failed GetCommit: %v", err)
	}
	return gerritchange.CommitMessage{Message: ci.Message}, nil
}

func updateCommitMessage(authedClient *http.Client, cl gerritchange.ChangeList, cm gerritchange.CommitMessage) error {
	agClient, err := aggerrit.NewClient(cl.LongHost(), authedClient)
	if err != nil {
		return fmt.Errorf("failed to create Gerrit client: %v", err)
	}
	_, err = agClient.Changes.SetCommitMessage(strconv.FormatInt(cl.ChangeNum, 10), &aggerrit.CommitMessageInput{Message: cm.Message})
	if err != nil {
		return fmt.Errorf("SetCommitMessage: %v", err)
	}
	return nil
}

type clCm struct {
	cl gerritchange.ChangeList
	cm gerritchange.CommitMessage
}

func getCommitMessages(authedClient *http.Client, cls []gerritchange.ChangeList) (map[gerritchange.ChangeList]gerritchange.CommitMessage, error) {
	var g errgroup.Group
	ch := make(chan clCm, len(cls))
	throttle := time.Tick(gerritRateLimit)
	for _, cl := range cls {
		<-throttle
		cl := cl
		g.Go(func() error {
			log.Printf("fetching commit for %v", cl)
			cm, err := getCommitMessage(authedClient, cl)
			if err != nil {
				return fmt.Errorf("getCommitMessage for cl %v: %v", cl, err)
			}
			ch <- clCm{cl: cl, cm: cm}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	// safe to close, since all the sender goroutines are done
	close(ch)
	clToCm := make(map[gerritchange.ChangeList]gerritchange.CommitMessage)
	for result := range ch {
		clToCm[result.cl] = result.cm
	}
	return clToCm, nil
}

func updateCommitMessages(authedClient *http.Client, cls []gerritchange.ChangeList, clToCm map[gerritchange.ChangeList]gerritchange.CommitMessage) error {
	var g errgroup.Group
	throttle := time.Tick(gerritRateLimit)
	for _, cl := range cls {
		<-throttle
		cl := cl
		g.Go(func() error {
			cm := clToCm[cl]
			cqDependValue := gerritchange.ChangeListsString(cls)
			newCm, err := cm.WithFooter("Cq-Depend", cqDependValue)
			if err != nil {
				return fmt.Errorf("WithFooter for cl %v: %v", cl, err)
			}
			if cm.Message == newCm.Message {
				log.Printf("footer is already correct for CL %v", cl)
			} else {
				updateCommitMessage(authedClient, cl, newCm)
				log.Printf("updated commit message for CL %v", cl)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func UpdateCqDepend(authedClient *http.Client, clString string) error {
	cls, err := gerritchange.ParseChangeListsSorted(clString)
	if err != nil {
		return fmt.Errorf("ParseChangeListsSorted: %v", err)
	}
	clToCm, err := getCommitMessages(authedClient, cls)
	if err != nil {
		return fmt.Errorf("getCommitMessages: %v", err)
	}
	if err := updateCommitMessages(authedClient, cls, clToCm); err != nil {
		return fmt.Errorf("updateCommitMessages: %v", err)
	}
	return nil
}
