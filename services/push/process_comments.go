package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"strings"

	"github.com/TruStory/octopus/services/truapi/db"
	"github.com/gernest/mention"
	stripmd "github.com/writeas/go-strip-markdown"
)

func unique(values []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range values {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func (s *service) parseCosmosMentions(body string) (string, []string) {
	parsedBody := body
	usernameByAddress := map[string]string{}
	terminators := []rune(" \n\r.,():!?'\"")
	addresses := mention.GetTagsAsUniqueStrings('@', body, terminators...)
	for _, address := range addresses {
		user, err := s.db.UserByAddress(address)
		if err != nil || user == nil {
			s.log.WithError(err).Errorf("could not find profile for address %s", address)
			continue
		}
		usernameByAddress[address] = user.Username
	}
	for address, username := range usernameByAddress {
		parsedBody = strings.ReplaceAll(parsedBody, address, username)
	}
	return parsedBody, addresses
}

func (s *service) processCommentsNotifications(cNotifications <-chan *CommentNotificationRequest, notifications chan<- *Notification) {
	for n := range cNotifications {
		c, err := s.db.CommentByID(n.ID)
		if err != nil {
			s.log.WithError(err).Errorf("could not retrieve comment for id [%d]\n", n.ID)
			continue
		}
		var participants []string
		var notificationType db.NotificationType
		var mentionType db.MentionType
		if c.ArgumentID != 0 && c.ElementID != 0 {
			participants, err = s.db.ArgumentLevelCommentsParticipants(c.ArgumentID, c.ElementID)
			notificationType = db.NotificationArgumentCommentAction
			mentionType = db.MentionArgumentComment
		} else {
			participants, err = s.db.ClaimLevelCommentsParticipants(c.ClaimID)
			notificationType = db.NotificationCommentAction
			mentionType = db.MentionComment
		}
		if err != nil {
			s.log.WithError(err).Errorf("could not retrieve participants for comments claim_id[%d] argument_id[%d] element_id[%d]\n", n.ClaimID, n.ArgumentID, n.ElementID)
			continue
		}

		notified := make(map[string]bool)
		// skip comment creator
		notified[n.Creator] = true
		parsedComment, mentions := s.parseCosmosMentions(c.Body)
		parsedComment = stripmd.Strip(parsedComment)
		meta := db.NotificationMeta{
			ClaimID:    &c.ClaimID,
			ArgumentID: &c.ArgumentID,
			ElementID:  &c.ElementID,
			CommentID:  &n.ID,
		}
		typeId := c.ClaimID
		for _, p := range mentions {
			mentionMeta := db.NotificationMeta{
				ClaimID:     &c.ClaimID,
				ArgumentID:  &c.ArgumentID,
				ElementID:   &c.ElementID,
				CommentID:   &n.ID,
				MentionType: &mentionType,
			}
			if _, ok := notified[p]; ok {
				continue
			}
			notified[p] = true
			notifications <- &Notification{
				From:   &c.Creator,
				To:     p,
				TypeID: typeId,
				Type:   db.NotificationMentionAction,
				Msg:    fmt.Sprintf("mentioned you %s: %s", mentionType.String(), parsedComment),
				Meta:   mentionMeta,
				Action: "Mentioned you in a reply",
				Trim:   true,
			}
		}

		for _, p := range participants {
			if _, ok := notified[p]; ok {
				continue
			}
			notified[p] = true
			notifications <- &Notification{
				From:   &c.Creator,
				To:     p,
				TypeID: typeId,
				Type:   notificationType,
				Msg:    fmt.Sprintf("added a Reply: %s", parsedComment),
				Meta:   meta,
				Action: "Added a new reply",
				Trim:   true,
			}
		}

		// notify claim creator if not previously notified
		if _, ok := notified[n.ClaimCreator]; !ok {
			notifications <- &Notification{
				From:   &c.Creator,
				To:     n.ClaimCreator,
				TypeID: typeId,
				Type:   notificationType,
				Msg:    fmt.Sprintf("added a Reply: %s", parsedComment),
				Meta:   meta,
				Action: "Added a new reply",
				Trim:   true,
			}
		}

		// notify argument creator if not previously notified
		if n.ArgumentCreator != "" {
			if _, ok := notified[n.ArgumentCreator]; !ok {
				notifications <- &Notification{
					From:   &c.Creator,
					To:     n.ArgumentCreator,
					TypeID: typeId,
					Type:   notificationType,
					Msg:    fmt.Sprintf("added a Reply: %s", parsedComment),
					Meta:   meta,
					Action: "Added a new reply",
					Trim:   true,
				}
			}
		}

	}
}

func (s *service) startHTTP(stop <-chan struct{}, notifications chan<- *CommentNotificationRequest) {
	mux := http.NewServeMux()
	mux.HandleFunc("/sendCommentNotification", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			fmt.Printf("only POST method allowed received [%s]\n", r.Method)
			return
		}
		n := &CommentNotificationRequest{}
		err := json.NewDecoder(r.Body).Decode(n)
		if err != nil {
			s.log.WithError(err).Error("error decoding request")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.log.WithField("commentId", n.ID).Info("comment notification request received")
		notifications <- n
		w.WriteHeader(http.StatusAccepted)
	})
	server := &http.Server{
		Addr:    ":9001",
		Handler: mux,
	}
	go func() {
		<-stop
		// we are at shutdown
		_ = server.Close()
	}()
	err := server.ListenAndServe()
	if err != nil {
		s.log.WithError(err).Fatal("error starting http service")
	}
}
