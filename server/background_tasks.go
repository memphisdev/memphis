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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/memphis_cache"
	"github.com/memphisdev/memphis/models"

	"strconv"
	"strings"
	"time"
)

const CONN_STATUS_SUBJ = "$memphis_connection_status"
const INTEGRATIONS_UPDATES_SUBJ = "$memphis_integration_updates"
const CONFIGURATIONS_RELOAD_SIGNAL_SUBJ = "$memphis_config_reload_signal"
const NOTIFICATION_EVENTS_SUBJ = "$memphis_notifications"
const PM_RESEND_ACK_SUBJ = "$memphis_pm_acks"
const TIERED_STORAGE_CONSUMER = "$memphis_tiered_storage_consumer"
const DLS_UNACKED_CONSUMER = "$memphis_dls_unacked_consumer"
const NACKED_DLS_SUBJ = "$memphis_nacked_dls"
const NACKED_DLS_INNER_SUBJ = "$memphis_nacked_inner_dls"
const NACKED_DLS_CONSUMER = "$memphis_nacked_dls_consumer"
const SCHEMAVERSE_DLS_SUBJ = "$memphis_schemaverse_dls"
const SCHEMAVERSE_DLS_INNER_SUBJ = "$memphis_schemaverse_inner_dls"
const SCHEMAVERSE_DLS_CONSUMER = "$memphis_schemaverse_dls_consumer"
const FUNCTIONS_DLS_INNER_SUBJ = "$memphis_functions_inner_dls"
const FUNCTIONS_DLS_CONSUMER = "$memphis_functions_dls_consumer"
const CACHE_UDATES_SUBJ = "$memphis_cache_updates"
const NOTIFICATIONS_BUFFER_CONSUMER = "$memphis_notifications_buffer_consumer"
const FUNCTION_TASKS_CONSUMER = "$memphis_function_tasks_consumer"

var LastReadThroughputMap map[string]models.Throughput
var LastWriteThroughputMap map[string]models.Throughput
var tieredStorageMsgsMap *concurrentMap[map[string][]StoredMsg]
var tieredStorageMapLock sync.Mutex

func (s *Server) ListenForZombieConnCheckRequests() error {
	_, err := s.subscribeOnAcc(s.MemphisGlobalAccount(), CONN_STATUS_SUBJ, CONN_STATUS_SUBJ+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			opts := &ConnzOptions{
				Limit:               s.opts.MaxConn,
				Username:            false,
				Subscriptions:       false,
				SubscriptionsDetail: false,
				State:               ConnOpen,
			}
			conns, _ := s.Connz(opts)
			connectionIds := make(map[string]string)
			for _, conn := range conns.Conns {
				connId := strings.Split(conn.Name, "::")[0]
				if connId != _EMPTY_ {
					connectionIds[connId] = _EMPTY_
				}
			}

			if len(connectionIds) > 0 { // in case there are connections
				bytes, err := json.Marshal(connectionIds)
				if err != nil {
					s.Errorf("ListenForZombieConnCheckRequests: %v", err.Error())
				} else {
					s.sendInternalAccountMsgWithReply(s.MemphisGlobalAccount(), reply, _EMPTY_, nil, bytes, true)
				}
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ListenForCacheUpdates() error {
	_, err := s.subscribeOnAcc(s.MemphisGlobalAccount(), CACHE_UDATES_SUBJ, CACHE_UDATES_SUBJ+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var cache_req models.CacheUpdateRequest
			err := json.Unmarshal(msg, &cache_req)
			if err != nil {
				s.Errorf("ListenForUserCacheDeletion at Unmarshal could not delete from cache, error: %v", err)
				return
			}

			switch cache_req.CacheType {
			case "user":
				if cache_req.Operation == "delete" {
					err = memphis_cache.DeleteUser(cache_req.TenantName, cache_req.Usernames)
					if err != nil {
						s.Errorf("ListenForUserCacheDeletion at DeleteUser could not delete from cache, error: %v", err)
						return
					}
				}
			}

		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ListenForIntegrationsUpdateEvents() error {
	_, err := s.subscribeOnAcc(s.MemphisGlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, INTEGRATIONS_UPDATES_SUBJ+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var integrationUpdate models.CreateIntegration
			err := json.Unmarshal(msg, &integrationUpdate)
			if err != nil {
				s.Errorf("[tenant: %v]ListenForIntegrationsUpdateEvents: %v", integrationUpdate.TenantName, err.Error())
				return
			}
			switch strings.ToLower(integrationUpdate.Name) {
			case "slack":
				if s.opts.UiHost == _EMPTY_ {
					EditClusterCompHost("ui_host", integrationUpdate.UIUrl)
				}
				CacheDetails("slack", integrationUpdate.Keys, integrationUpdate.Properties, integrationUpdate.TenantName)
			case "discord":
				if s.opts.UiHost == "" {
					EditClusterCompHost("ui_host", integrationUpdate.UIUrl)
				}
				CacheDetails("discord", integrationUpdate.Keys, integrationUpdate.Properties, integrationUpdate.TenantName)
			case "s3":
				CacheDetails("s3", integrationUpdate.Keys, integrationUpdate.Properties, integrationUpdate.TenantName)
			case "github":
				CacheDetails("github", integrationUpdate.Keys, integrationUpdate.Properties, integrationUpdate.TenantName)
			default:
				s.Warnf("[tenant: %v] ListenForIntegrationsUpdateEvents: %s %s", integrationUpdate.TenantName, strings.ToLower(integrationUpdate.Name), "unknown integration")
				return
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ListenForConfigReloadEvents() error {
	var lock sync.Mutex
	_, err := s.subscribeOnAcc(s.MemphisGlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, CONFIGURATIONS_RELOAD_SIGNAL_SUBJ+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			// reload config
			lock.Lock()
			err := s.Reload()
			if err != nil {
				s.Errorf("Failed reloading: %v", err.Error())
			}
			time.AfterFunc(time.Millisecond*500, func() {
				lock.Unlock()
			})
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ListenForNotificationEvents() error {
	err := s.queueSubscribe(s.MemphisGlobalAccountString(), NOTIFICATION_EVENTS_SUBJ, NOTIFICATION_EVENTS_SUBJ+"_group", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			tenantName, message, err := s.getTenantNameAndMessage(msg)
			if err != nil {
				s.Errorf("ListenForNotificationEvents at getTenantNameAndMessage: %v", err.Error())
				return
			}
			var notification models.Notification
			err = json.Unmarshal([]byte(message), &notification)
			if err != nil {
				s.Errorf("[tenant: %v]ListenForNotificationEvents: %v", tenantName, err.Error())
				return
			}
			notificationMsg := notification.Msg
			if notification.Code != _EMPTY_ {
				notificationMsg = notificationMsg + "\n```" + notification.Code + "```"
			}
			err = s.SendNotification(tenantName, notification.Title, notificationMsg, notification.Type)
			if err != nil {
				s.Errorf("[tenant: %v]ListenForNotificationEvents at SendNotification: %v", tenantName, err.Error())
				return
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ListenForSchemaverseDlsEvents() error {
	err := s.queueSubscribe(s.MemphisGlobalAccountString(), SCHEMAVERSE_DLS_SUBJ, SCHEMAVERSE_DLS_SUBJ+"_group", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			s.sendInternalAccountMsg(s.MemphisGlobalAccount(), SCHEMAVERSE_DLS_INNER_SUBJ, msg)
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) ListenForNackedDlsEvents() error {
	err := s.queueSubscribe(s.MemphisGlobalAccountString(), NACKED_DLS_SUBJ, NACKED_DLS_SUBJ+"_group", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			s.sendInternalAccountMsg(s.MemphisGlobalAccount(), NACKED_DLS_INNER_SUBJ, msg)
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) ListenForPoisonMsgAcks() error {
	err := s.queueSubscribe(s.MemphisGlobalAccountString(), PM_RESEND_ACK_SUBJ, PM_RESEND_ACK_SUBJ+"_group", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			tenantName, message, err := s.getTenantNameAndMessage(msg)
			if err != nil {
				s.Errorf("ListenForPoisonMsgAcks at getTenantNameAndMessage: %v", err.Error())
				return
			}
			var msgToAck models.PmAckMsg
			err = json.Unmarshal([]byte(message), &msgToAck)
			if err != nil {
				s.Errorf("[tenant: %v]ListenForPoisonMsgAcks: %v", tenantName, err.Error())
				return
			}
			err = db.RemoveCgFromDlsMsg(msgToAck.ID, msgToAck.CgName, tenantName)
			if err != nil {
				return
			}

		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func getThroughputSubject(serverName string) string {
	return throughputStreamNameV1 + tsep + serverName
}

func (s *Server) InitializeThroughputSampling() {
	LastReadThroughputMap = map[string]models.Throughput{}
	LastWriteThroughputMap = map[string]models.Throughput{}
	for _, acc := range s.Opts().Accounts {
		LastReadThroughputMap[acc.GetName()] = models.Throughput{
			Bytes:       acc.outBytes,
			BytesPerSec: 0,
		}
		LastWriteThroughputMap[acc.GetName()] = models.Throughput{
			Bytes:       acc.inBytes,
			BytesPerSec: 0,
		}
	}
	go s.CalculateSelfThroughput()
}

func (s *Server) CalculateSelfThroughput() {
	for range time.Tick(time.Second * 1) {
		readMap := map[string]int64{}
		writeMap := map[string]int64{}
		s.accounts.Range(func(_, v interface{}) bool {
			acc := v.(*Account)
			accName := acc.GetName()
			currentRead := acc.outBytes - LastReadThroughputMap[accName].Bytes
			LastReadThroughputMap[accName] = models.Throughput{
				Bytes:       acc.outBytes,
				BytesPerSec: currentRead,
			}
			readMap[accName] = currentRead
			currentWrite := acc.inBytes - LastWriteThroughputMap[accName].Bytes
			LastWriteThroughputMap[accName] = models.Throughput{
				Bytes:       acc.inBytes,
				BytesPerSec: currentWrite,
			}
			writeMap[accName] = currentWrite
			return true
		})
		serverName := s.opts.ServerName
		subj := getThroughputSubject(serverName)
		tpMsg := models.BrokerThroughput{
			Name:     serverName,
			ReadMap:  readMap,
			WriteMap: writeMap,
		}
		s.sendInternalAccountMsg(s.MemphisGlobalAccount(), subj, tpMsg)
	}
}

func (s *Server) StartBackgroundTasks() error {
	err := s.ListenForZombieConnCheckRequests()
	if err != nil {
		return errors.New("Failed subscribing for zombie conns check requests: " + err.Error())
	}

	err = s.ListenForIntegrationsUpdateEvents()
	if err != nil {
		return errors.New("Failed subscribing for integrations updates: " + err.Error())
	}

	err = s.ListenForNotificationEvents()
	if err != nil {
		return errors.New("Failed subscribing for schema validation updates: " + err.Error())
	}

	err = s.ListenForSchemaverseDlsEvents()
	if err != nil {
		return errors.New("Failed to subscribing for schemaverse dls" + err.Error())
	}

	err = s.ListenForNackedDlsEvents()
	if err != nil {
		return errors.New("Failed to subscribing for nacked dls" + err.Error())
	}

	err = s.ListenForPoisonMsgAcks()
	if err != nil {
		return errors.New("Failed subscribing for poison message acks: " + err.Error())
	}

	err = s.ListenForConfigReloadEvents()
	if err != nil {
		return errors.New("Failed subscribing for configurations update: " + err.Error())
	}

	err = s.ListenForCacheUpdates()
	if err != nil {
		return errors.New("Failed to subscribing for cache updates" + err.Error())
	}

	err = s.ListenForCloudCacheUpdates()
	if err != nil {
		return errors.New("Failed to subscribing for cloud cache updates" + err.Error())
	}

	err = s.ListenToFunctionsCounterUpdates()
	if err != nil {
		return errors.New("Failed to subscribing for functions counter updates" + err.Error())
	}

	go s.ConsumeSchemaverseDlsMessages()
	go s.ConsumeNackedDlsMessages()
	go s.ConsumeUnackedMsgs()
	go s.ConsumeFunctionsDlsMessages()
	go s.ConsumeTieredStorageMsgs()
	go s.RemoveOldDlsMsgs()
	go s.uploadMsgsToTier2Storage()
	go s.InitializeThroughputSampling()
	go s.UploadTenantUsageToDB()
	go s.RefreshFirebaseFunctionsKey()
	go s.RemoveOldProducersAndConsumersAndAuditLogs()
	go ScheduledCloudCacheRefresh()
	go s.SendBillingAlertWhenNeeded()
	go s.CheckBrokenConnectedIntegrations()
	go s.ConsumeNotificationsBufferMessages()
	go s.ReleaseStuckLocks()
	go s.ConsumeFunctionTasks()
	go s.ScaleFunctionWorkers()
	go s.ConnectorsDeadPodsRescheduler()
	go s.removeOldAsyncTasks()

	return nil
}

func (s *Server) uploadMsgsToTier2Storage() {
	currentTimeFrame := s.opts.TieredStorageUploadIntervalSec
	ticker := time.NewTicker(time.Duration(currentTimeFrame) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if s.opts.TieredStorageUploadIntervalSec != currentTimeFrame {
			currentTimeFrame = s.opts.TieredStorageUploadIntervalSec
			ticker.Reset(time.Duration(currentTimeFrame) * time.Second)
			// update consumer when TIERED_STORAGE_TIME_FRAME_SEC configuration was changed
			cc := ConsumerConfig{
				DeliverPolicy: DeliverAll,
				AckPolicy:     AckExplicit,
				Durable:       TIERED_STORAGE_CONSUMER,
				FilterSubject: tieredStorageStream + ".>",
				AckWait:       time.Duration(2) * time.Duration(currentTimeFrame) * time.Second,
				MaxAckPending: -1,
				MaxDeliver:    10,
			}
			err := serv.memphisAddConsumer(s.MemphisGlobalAccountString(), tieredStorageStream, &cc)
			if err != nil {
				serv.Errorf("Failed add tiered storage consumer: %v", err.Error())
				return
			}
			TIERED_STORAGE_CONSUMER_CREATED = true
		}
		tieredStorageMapLock.Lock()
		err := flushMapToTier2Storage()
		if err != nil {
			serv.Errorf("Failed upload messages to tiered 2 storage: %v", err.Error())
			tieredStorageMapLock.Unlock()
			continue
		}
		// ack all messages uploaded to tiered 2 storage or when there is no s3 integaration to tenant
		for t, tenant := range tieredStorageMsgsMap.m {
			for i, msgs := range tenant {
				for _, msg := range msgs {
					reply := msg.ReplySubject
					s.sendInternalAccountMsg(s.MemphisGlobalAccount(), reply, []byte(_EMPTY_))
				}
				delete(tenant, i)
			}
			tieredStorageMsgsMap.Delete(t)
		}

		tieredStorageMapLock.Unlock()
	}
}

func (s *Server) ConsumeUnackedMsgs() {
	type unAckedMsg struct {
		Msg          []byte
		ReplySubject string
	}
	amount := 1000
	req := []byte(strconv.FormatUint(uint64(amount), 10))
	for {
		if DLS_UNACKED_CONSUMER_CREATED && DLS_UNACKED_STREAM_CREATED {
			resp := make(chan unAckedMsg)
			replySubj := DLS_UNACKED_CONSUMER + "_reply_" + s.memphis.nuid.Next()

			// subscribe to unacked messages
			sub, err := s.subscribeOnAcc(s.MemphisGlobalAccount(), replySubj, replySubj+"_sid", func(_ *client, subject, reply string, msg []byte) {
				go func(subject, reply string, msg []byte) {
					// Ignore 409 Exceeded MaxWaiting cases
					if reply != _EMPTY_ {
						message := unAckedMsg{
							Msg:          msg,
							ReplySubject: reply,
						}
						resp <- message
					}
				}(subject, reply, copyBytes(msg))
			})
			if err != nil {
				s.Errorf("Failed to subscribe to unacked messages: %v", err.Error())
				continue
			}

			// send JS API request to get more messages
			subject := fmt.Sprintf(JSApiRequestNextT, dlsUnackedStream, DLS_UNACKED_CONSUMER)
			s.sendInternalAccountMsgWithReply(s.MemphisGlobalAccount(), subject, replySubj, nil, req, true)

			timeout := time.NewTimer(5 * time.Second)
			msgs := make([]unAckedMsg, 0)
			stop := false
			for {
				if stop {
					s.unsubscribeOnAcc(s.MemphisGlobalAccount(), sub)
					break
				}
				select {
				case unAckedMsg := <-resp:
					msgs = append(msgs, unAckedMsg)
					if len(msgs) == amount {
						stop = true
					}
				case <-timeout.C:
					stop = true
				}
			}
			for _, msg := range msgs {
				err := s.handleNewUnackedMsg(msg.Msg)
				if err == nil {
					// send ack
					s.sendInternalAccountMsgWithEcho(s.MemphisGlobalAccount(), msg.ReplySubject, []byte(_EMPTY_))
				}
			}
		} else {
			time.Sleep(2 * time.Second)
		}
	}
}

func (s *Server) ConsumeTieredStorageMsgs() {
	type tsMsg struct {
		Msg          []byte
		ReplySubject string
	}

	tieredStorageMsgsMap = NewConcurrentMap[map[string][]StoredMsg]()
	amount := 1000
	req := []byte(strconv.FormatUint(uint64(amount), 10))
	for {
		if TIERED_STORAGE_CONSUMER_CREATED && TIERED_STORAGE_STREAM_CREATED {
			resp := make(chan tsMsg)
			replySubj := TIERED_STORAGE_CONSUMER + "_reply_" + s.memphis.nuid.Next()

			// subscribe to unacked messages
			sub, err := s.subscribeOnAcc(s.MemphisGlobalAccount(), replySubj, replySubj+"_sid", func(_ *client, subject, reply string, msg []byte) {
				go func(subject, reply string, msg []byte) {
					// Ignore 409 Exceeded MaxWaiting cases
					if reply != _EMPTY_ {
						message := tsMsg{
							Msg:          msg,
							ReplySubject: reply,
						}
						resp <- message
					}
				}(subject, reply, copyBytes(msg))
			})
			if err != nil {
				s.Errorf("Failed to subscribe to tiered storage messages: %v", err.Error())
				continue
			}

			// send JS API request to get more messages
			subject := fmt.Sprintf(JSApiRequestNextT, tieredStorageStream, TIERED_STORAGE_CONSUMER)
			s.sendInternalAccountMsgWithReply(s.MemphisGlobalAccount(), subject, replySubj, nil, req, true)

			timeout := time.NewTimer(5 * time.Second)
			msgs := make([]tsMsg, 0)
			stop := false
			for {
				if stop {
					s.unsubscribeOnAcc(s.MemphisGlobalAccount(), sub)
					break
				}
				select {
				case tieredStorageMsg := <-resp:
					msgs = append(msgs, tieredStorageMsg)
					if len(msgs) == amount {
						stop = true
					}
				case <-timeout.C:
					stop = true
				}
			}
			for _, message := range msgs {
				msg := message.Msg
				reply := message.ReplySubject
				s.handleNewTieredStorageMsg(msg, reply)
			}
		} else {
			time.Sleep(2 * time.Second)
		}
	}
}

func (s *Server) ConsumeSchemaverseDlsMessages() {
	type schemaverseDlsMsg struct {
		Msg          []byte
		ReplySubject string
	}
	amount := 1000
	req := []byte(strconv.FormatUint(uint64(amount), 10))
	for {
		if DLS_SCHEMAVERSE_CONSUMER_CREATED && DLS_SCHEMAVERSE_STREAM_CREATED {
			resp := make(chan schemaverseDlsMsg)
			replySubj := SCHEMAVERSE_DLS_CONSUMER + "_reply_" + s.memphis.nuid.Next()

			// subscribe to schemavers dls messages
			sub, err := s.subscribeOnAcc(s.MemphisGlobalAccount(), replySubj, replySubj+"_sid", func(_ *client, subject, reply string, msg []byte) {
				go func(subject, reply string, msg []byte) {
					// Ignore 409 Exceeded MaxWaiting cases
					if reply != _EMPTY_ {
						message := schemaverseDlsMsg{
							Msg:          msg,
							ReplySubject: reply,
						}
						resp <- message
					}
				}(subject, reply, copyBytes(msg))
			})
			if err != nil {
				s.Errorf("Failed to subscribe to schemavers dls messages: %v", err.Error())
				continue
			}

			// send JS API request to get more messages
			subject := fmt.Sprintf(JSApiRequestNextT, dlsSchemaverseStream, SCHEMAVERSE_DLS_CONSUMER)
			s.sendInternalAccountMsgWithReply(s.MemphisGlobalAccount(), subject, replySubj, nil, req, true)

			s.Debugf("ConsumeSchemaverseDlsMessages: sending fetch request")

			timeout := time.NewTimer(5 * time.Second)
			msgs := make([]schemaverseDlsMsg, 0)
			stop := false
			for {
				if stop {
					s.unsubscribeOnAcc(s.MemphisGlobalAccount(), sub)
					break
				}
				select {
				case SchemaDlsMsg := <-resp:
					msgs = append(msgs, SchemaDlsMsg)
					if len(msgs) == amount {
						stop = true
						s.Debugf("ConsumeSchemaverseDlsMessages: finished appending %v messages", len(msgs))
					}
				case <-timeout.C:
					stop = true
					s.Debugf("ConsumeSchemaverseDlsMessages: finished because of timer: %v messages", len(msgs))
				}
			}
			for _, message := range msgs {
				msg := message.Msg
				err := s.handleSchemaverseDlsMsg(msg)
				if err == nil {
					// send ack
					s.sendInternalAccountMsgWithEcho(s.MemphisGlobalAccount(), message.ReplySubject, []byte(_EMPTY_))
				}
			}
		} else {
			s.Warnf("ConsumeSchemaverseDlsMessages: waiting for consumer and stream to be created")
			time.Sleep(2 * time.Second)
		}
	}
}

func (s *Server) ConsumeNackedDlsMessages() {
	type nackedDlsMsg struct {
		Msg          []byte
		ReplySubject string
	}
	amount := 1000
	req := []byte(strconv.FormatUint(uint64(amount), 10))
	for {
		if DLS_NACKED_CONSUMER_CREATED && DLS_NACKED_STREAM_CREATED {
			resp := make(chan nackedDlsMsg)
			replySubj := NACKED_DLS_CONSUMER + "_reply_" + s.memphis.nuid.Next()

			// subscribe to schemavers dls messages
			sub, err := s.subscribeOnAcc(s.MemphisGlobalAccount(), replySubj, replySubj+"_sid", func(_ *client, subject, reply string, msg []byte) {
				go func(subject, reply string, msg []byte) {
					// Ignore 409 Exceeded MaxWaiting cases
					if reply != _EMPTY_ {
						message := nackedDlsMsg{
							Msg:          msg,
							ReplySubject: reply,
						}
						resp <- message
					}
				}(subject, reply, copyBytes(msg))
			})
			if err != nil {
				s.Errorf("Failed to subscribe to nacked dls messages: %v", err.Error())
				continue
			}

			// send JS API request to get more messages
			subject := fmt.Sprintf(JSApiRequestNextT, dlsNackedStream, NACKED_DLS_CONSUMER)
			s.sendInternalAccountMsgWithReply(s.MemphisGlobalAccount(), subject, replySubj, nil, req, true)

			s.Debugf("ConsumeNackedDlsMessages: sending fetch request")

			timeout := time.NewTimer(5 * time.Second)
			msgs := make([]nackedDlsMsg, 0)
			stop := false
			for {
				if stop {
					s.unsubscribeOnAcc(s.MemphisGlobalAccount(), sub)
					break
				}
				select {
				case nackedDlsMsg := <-resp:
					msgs = append(msgs, nackedDlsMsg)
					if len(msgs) == amount {
						stop = true
						s.Debugf("ConsumeNackedDlsMessages: finished appending %v messages", len(msgs))
					}
				case <-timeout.C:
					stop = true
					s.Debugf("ConsumeNackedDlsMessages: finished because of timer: %v messages", len(msgs))
				}
			}
			for _, message := range msgs {
				msg := message.Msg
				err := s.handleNackedDlsMsg(msg)
				if err == nil {
					// send ack
					s.sendInternalAccountMsgWithEcho(s.MemphisGlobalAccount(), message.ReplySubject, []byte(_EMPTY_))
				}
			}
		} else {
			s.Warnf("ConsumeNackedDlsMessages: waiting for consumer and stream to be created")
			time.Sleep(2 * time.Second)
		}
	}
}

func (s *Server) RemoveOldDlsMsgs() {
	ticker := time.NewTicker(2 * time.Minute)
	for range ticker.C {
		for tenantName, rt := range s.opts.DlsRetentionHours {
			configurationTime := time.Now().Add(time.Hour * time.Duration(-rt))
			err := db.DeleteOldDlsMessageByRetention(configurationTime, tenantName)
			if err != nil {
				serv.Errorf("[tenant: %v]RemoveOldDlsMsgs: %v", tenantName, err.Error())
			}
		}
	}
}

func (s *Server) RemoveOldProducersAndConsumersAndAuditLogs() {
	ticker := time.NewTicker(15 * time.Minute)
	for range ticker.C {
		for tenantName, rt := range s.opts.GCProducersConsumersRetentionHours {
			configurationTime := time.Now().Add(time.Hour * time.Duration(-rt))
			err := db.DeleteOldProducersAndConsumers(configurationTime, tenantName)
			if err != nil {
				serv.Errorf("[tenant: %v]RemoveOldProducersAndConsumersAndAuditLogs at DeleteOldProducersAndConsumers : %v", tenantName, err.Error())
			}
			time := time.Now().Add(-time.Hour * 3 * 24)
			err = db.RemoveAuditLogsByTenantAndCreatedAt(tenantName, time)
			if err != nil {
				serv.Errorf("[tenant: %v]RemoveOldProducersAndConsumersAndAuditLogs at RemoveAuditLogsByTenantAndCreatedAt : %v", tenantName, err.Error())
			}
		}
	}
}

func (s *Server) CheckBrokenConnectedIntegrations() error {
	ticker := time.NewTicker(15 * time.Minute)
	for range ticker.C {
		_, integrations, err := db.GetAllIntegrations()
		if err != nil {
			serv.Errorf("CheckBrokenConnectedIntegrations at GetAllIntegrations: %v", err.Error())
		}

		for _, integration := range integrations {
			switch integration.Name {
			case "github":
				if _, ok := integration.Keys["installation_id"].(string); !ok {
					integration.Keys["installation_id"] = _EMPTY_
				}
				err := testGithubIntegration(integration.Keys["installation_id"].(string))
				if err != nil {
					serv.Warnf("[tenant: %s]CheckBrokenConnectedIntegrations at testGithubIntegration: %v", integration.TenantName, err.Error())
					err = db.UpdateIsValidIntegration(integration.TenantName, integration.Name, false)
					if err != nil {
						serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at UpdateIsValidIntegration: %v", integration.TenantName, err.Error())
					}
				} else {
					err = db.UpdateIsValidIntegration(integration.TenantName, integration.Name, true)
					if err != nil {
						serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at UpdateIsValidIntegration: %v", integration.TenantName, err.Error())
					}
				}
			case "slack":
				key := getAESKey()
				if _, ok := integration.Keys["auth_token"].(string); !ok {
					integration.Keys["auth_token"] = _EMPTY_
				}
				if _, ok := integration.Keys["channel_id"].(string); !ok {
					integration.Keys["channel_id"] = _EMPTY_
				}
				authToken, err := DecryptAES(key, integration.Keys["auth_token"].(string))
				if err != nil {
					serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at DecryptAES: %v", integration.TenantName, err.Error())
				}
				err = testSlackIntegration(authToken)
				if err != nil {
					serv.Warnf("[tenant: %s]CheckBrokenConnectedIntegrations at testSlackIntegration: %v", integration.TenantName, err.Error())
					err = db.UpdateIsValidIntegration(integration.TenantName, integration.Name, false)
					if err != nil {
						serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at UpdateIsValidIntegration: %v", integration.TenantName, err.Error())
					}
				} else {
					err = db.UpdateIsValidIntegration(integration.TenantName, integration.Name, true)
					if err != nil {
						serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at UpdateIsValidIntegration: %v", integration.TenantName, err.Error())
					}
				}
			case "discord":
				key := getAESKey()
				if _, ok := integration.Keys["webhook_url"].(string); !ok {
					integration.Keys["webhook_url"] = ""
				}
				webhookUrl, err := DecryptAES(key, integration.Keys["webhook_url"].(string))
				if err != nil {
					serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at DecryptAES: %v", integration.TenantName, err.Error())
				}
				err = testDiscordIntegration(webhookUrl)
				if err != nil {
					serv.Warnf("[tenant: %s]CheckBrokenConnectedIntegrations at testDiscordIntegration: %v", integration.TenantName, err.Error())
					err = db.UpdateIsValidIntegration(integration.TenantName, integration.Name, false)
					if err != nil {
						serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at UpdateIsValidIntegration: %v", integration.TenantName, err.Error())
					}
				} else {
					err = db.UpdateIsValidIntegration(integration.TenantName, integration.Name, true)
					if err != nil {
						serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at UpdateIsValidIntegration: %v", integration.TenantName, err.Error())
					}
				}
			case "s3":
				key := getAESKey()
				if _, ok := integration.Keys["access_key"].(string); !ok {
					integration.Keys["access_key"] = _EMPTY_
				}
				if _, ok := integration.Keys["secret_key"].(string); !ok {
					integration.Keys["secret_key"] = _EMPTY_
				}
				if _, ok := integration.Keys["region"].(string); !ok {
					integration.Keys["region"] = _EMPTY_
				}
				if _, ok := integration.Keys["url"].(string); !ok {
					integration.Keys["url"] = _EMPTY_
				}
				if _, ok := integration.Keys["s3_path_style"].(string); !ok {
					integration.Keys["s3_path_style"] = _EMPTY_
				}
				if _, ok := integration.Keys["bucket_name"].(string); !ok {
					integration.Keys["bucket_name"] = _EMPTY_
				}
				accessKey := integration.Keys["access_key"].(string)
				secretKey, err := DecryptAES(key, integration.Keys["secret_key"].(string))
				if err != nil {
					serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at DecryptAES: %v", integration.TenantName, err.Error())
				}

				provider := credentials.NewStaticCredentialsProvider(accessKey, secretKey, _EMPTY_)
				_, err = provider.Retrieve(context.Background())
				if err != nil {
					if strings.Contains(err.Error(), "static credentials are empty") {
						serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at provider.Retrieve: credentials are empty %v", integration.TenantName, err.Error())
					} else {
						serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at provider.Retrieve: %v", integration.TenantName, err.Error())
					}
				}
				cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
					awsconfig.WithCredentialsProvider(provider),
					awsconfig.WithRegion(integration.Keys["region"].(string)),
					awsconfig.WithEndpointResolverWithOptions(getS3EndpointResolver(integration.Keys["region"].(string), integration.Keys["url"].(string))),
				)
				if err != nil {
					serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at awsconfig.LoadDefaultConfig: %v", integration.TenantName, err.Error())
				}
				var usePathStyle bool
				svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
					switch integration.Keys["s3_path_style"].(string) {
					case "true":
						usePathStyle = true
					case "false":
						usePathStyle = false
					}
					o.UsePathStyle = usePathStyle
				})
				_, err = testS3Integration(svc, integration.Keys["bucket_name"].(string), integration.Keys["url"].(string))
				if err != nil {
					serv.Warnf("[tenant: %s]CheckBrokenConnectedIntegrations at testS3Integration: %v", integration.TenantName, err.Error())
					err = db.UpdateIsValidIntegration(integration.TenantName, integration.Name, false)
					if err != nil {
						serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at UpdateIsValidIntegration: %v", integration.TenantName, err.Error())
					}
				} else {
					err = db.UpdateIsValidIntegration(integration.TenantName, integration.Name, true)
					if err != nil {
						serv.Errorf("[tenant: %s]CheckBrokenConnectedIntegrations at UpdateIsValidIntegration: %v", integration.TenantName, err.Error())
					}
				}
			}
		}
	}
	return nil
}

func (s *Server) ConsumeNotificationsBufferMessages() error {
	const mAmount = 1000
	for {
		if !NOTIFICATIONS_BUFFER_STREAM_CREATED || !NOTIFICATIONS_BUFFER_CONSUMER_CREATED {
			s.Warnf("ConsumeNotificationsBufferMessages: waiting for notifications stream and consumer to be created")
			time.Sleep(2 * time.Second)
			continue
		}

		msgs, err := fetchMessages(s,
			NOTIFICATIONS_BUFFER_CONSUMER,
			notificationsStreamName,
			mAmount,
			3*time.Second)

		if err != nil {
			s.Errorf("Failed to fetch notifications: %v", err.Error())
			continue
		}

		sendNotifications(s, msgs)
	}
}

func createNotificationBufferMsg(msg []byte, reply string) notificationBufferMsg {
	return notificationBufferMsg{
		Msg:          msg,
		ReplySubject: reply,
	}
}

func fetchMessages(s *Server,
	consumer,
	streamName string,
	mAmount int,
	timeToWait time.Duration) ([]notificationBufferMsg, error) {

	req := []byte(strconv.FormatUint(uint64(mAmount), 10))
	resp := make(chan notificationBufferMsg)
	replySubject := consumer + "_reply_" + s.memphis.nuid.Next()

	timeout := time.NewTimer(timeToWait)
	sub, err := s.subscribeOnAcc(s.MemphisGlobalAccount(), replySubject, replySubject+"_sid", func(_ *client, subject string, reply string, msg []byte) {
		go func(subject, reply string, msg []byte) {
			if reply != "" {
				m := createNotificationBufferMsg(msg, reply)
				resp <- m
			}
		}(subject, reply, copyBytes(msg))
	})
	if err != nil {
		return nil, err
	}

	subject := fmt.Sprintf(JSApiRequestNextT, streamName, consumer)
	s.sendInternalAccountMsgWithReply(s.MemphisGlobalAccount(), subject, replySubject, nil, req, true)

	var msgs []notificationBufferMsg
	stop := false
	for {
		if stop {
			s.unsubscribeOnAcc(s.MemphisGlobalAccount(), sub)
			break
		}
		select {
		case m := <-resp:
			msgs = append(msgs, m)
			if len(msgs) == mAmount {
				stop = true
			}
		case <-timeout.C:
			stop = true
		}
	}

	return msgs, nil
}

func (s *Server) ReleaseStuckLocks() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		time := time.Now().Add(-10 * time.Minute)
		err := db.UnlockStuckLocks(time)
		if err != nil {
			serv.Errorf("ReleaseStuckLocks at UnlockStuckLocks: %v", err.Error())
		}

	}
}

func (s *Server) removeOldAsyncTasks() {
	ticker := time.NewTicker(15 * time.Minute)
	for range ticker.C {
		err := db.RemoveOldAsyncTasks()
		if err != nil {
			serv.Errorf("RemoveOldAsyncTasks at db.RemoveOldAsyncTasks : %v", err.Error())
		}
	}
}
