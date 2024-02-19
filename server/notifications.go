// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/memphisdev/memphis/models"
	"time"
)

type NotificationMsg struct {
	TenantName      string    `json:"tenantName"`
	Title           string    `json:"title"`
	Message         string    `json:"message"`
	MsgType         string    `json:"msgType"`
	IntegrationName string    `json:"integrationName"`
	Time            time.Time `json:"time"`
}

type NotificationMsgWithReply struct {
	NotificationMsg *NotificationMsg
	ReplySubject    string
}

type notificationBufferMsg struct {
	Msg          []byte
	ReplySubject string
}

func (s *Server) SendNotification(tenantName string, title string, message string, msgType string) error {
	for k := range NotificationFunctionsMap {
		switch k {
		case "slack":
			if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); ok {
				if slackIntegration, ok := tenantInetgrations["slack"].(models.SlackIntegration); ok {
					if slackIntegration.Properties[msgType] {
						// TODO: if the stream doesn't exist save the messages in buffer
						if !NOTIFICATIONS_BUFFER_STREAM_CREATED {
							return nil
						}

						// TODO: do we need msg-id here? if yes - what's the best way to generate it? hash title?
						if tenantName == "" {
							tenantName = serv.MemphisGlobalAccountString()
						}
						notificationMsg := NotificationMsg{
							TenantName:      tenantName,
							Title:           title,
							Message:         message,
							MsgType:         msgType,
							IntegrationName: "slack",
							Time:            time.Now(),
						}

						err := saveNotificationToQueue(s, notificationsStreamName+".user_notifications", tenantName, &notificationMsg)
						if err != nil {
							return err
						}
					}
				}
			}
		case "discord":
			if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); ok {
				if slackIntegration, ok := tenantInetgrations["slack"].(models.SlackIntegration); ok {
					if slackIntegration.Properties[msgType] {
						// TODO: if the stream doesn't exist save the messages in buffer
						if !NOTIFICATIONS_BUFFER_STREAM_CREATED {
							return nil
						}

						// TODO: do we need msg-id here? if yes - what's the best way to generate it? hash title?
						if tenantName == "" {
							tenantName = serv.MemphisGlobalAccountString()
						}
						notificationMsg := NotificationMsg{
							TenantName:      tenantName,
							Title:           title,
							Message:         message,
							MsgType:         msgType,
							IntegrationName: "discord",
							Time:            time.Now(),
						}

						err := saveNotificationToQueue(s, notificationsStreamName+".user_notifications", tenantName, &notificationMsg)
						if err != nil {
							return err
						}
					}
				}
			}
		default:
			return errors.New("failed sending notification: unsupported integration")
		}
	}
	return nil
}

func saveNotificationToQueue(s *Server, subject, tenantName string, notificationMsg *NotificationMsg) error {
	msg, err := json.Marshal(notificationMsg)
	if err != nil {
		return err
	}
	err = s.sendInternalAccountMsgWithEcho(s.MemphisGlobalAccount(), subject, msg)
	if err != nil {
		return fmt.Errorf("SendNotification (tenant %s): %w", tenantName, err)
	}

	return nil
}

func shouldSendNotification(tenantName string, alertType string) bool {
	if tenantInetgrations, ok := IntegrationsConcurrentCache.Load(tenantName); ok {
		if slackIntegration, ok := tenantInetgrations["slack"].(models.SlackIntegration); ok {
			if slackIntegration.Properties[alertType] {
				return true
			}
		}

		if discordIntegration, ok := tenantInetgrations["discord"].(models.DiscordIntegration); ok {
			if discordIntegration.Properties[alertType] {
				return true
			}
		}
	}
	return false
}

func sendNotifications(s *Server, msgs []notificationBufferMsg) {
	groupedMsgs := groupMessagesByTenantAndIntegration(msgs, s)
	for integrationName, tenantMsgs := range groupedMsgs {
		for tenantName, tMsgs := range tenantMsgs {
			switch integrationName {
			case "slack":
				sendSlackTenantNotifications(s, tenantName, tMsgs)
			case "discord":
				sendDiscordTenantNotifications(s, tenantName, tMsgs)
			}
		}
	}
}

func nackMsgs(s *Server, msgs []NotificationMsgWithReply, nackDuration time.Duration) error {
	nakPayload := []byte(fmt.Sprintf("%s {\"delay\": %d}", AckNak, nackDuration.Nanoseconds()))
	for i := 0; i < len(msgs); i++ {
		m := msgs[i]
		err := s.sendInternalAccountMsg(s.MemphisGlobalAccount(), m.ReplySubject, nakPayload)
		if err != nil {
			return err
		}
	}

	return nil
}

func ackMsgs(s *Server, msgs []NotificationMsgWithReply) {
	for i := 0; i < len(msgs); i++ {
		m := msgs[i]
		s.sendInternalAccountMsg(s.MemphisGlobalAccount(), m.ReplySubject, []byte(_EMPTY_))
	}
}

func groupMessagesByTenantAndIntegration(msgs []notificationBufferMsg, l Logger) map[string]map[string][]NotificationMsgWithReply {
	groupMsgs := make(map[string]map[string][]NotificationMsgWithReply)
	for _, message := range msgs {
		msg := message.Msg
		reply := message.ReplySubject
		var nm NotificationMsg
		err := json.Unmarshal(msg, &nm)
		if err != nil {
			// TODO: does it make sense to send ack for this message?
			// TODO: it's malformed and won't be unmarshalled next time as well
			l.Errorf("failed to unmarshal notification message: %v", err)
			continue
		}
		nmr := NotificationMsgWithReply{
			NotificationMsg: &nm,
			ReplySubject:    reply,
		}
		if _, ok := groupMsgs[nm.IntegrationName]; !ok {
			groupMsgs[nm.IntegrationName] = make(map[string][]NotificationMsgWithReply)
		}
		if _, ok := groupMsgs[nm.IntegrationName][nm.TenantName]; !ok {
			groupMsgs[nm.IntegrationName][nm.TenantName] = []NotificationMsgWithReply{}
		}
		groupMsgs[nm.IntegrationName][nm.TenantName] = append(groupMsgs[nm.IntegrationName][nm.TenantName], nmr)
	}

	return groupMsgs
}
