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
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/memphisdev/memphis/analytics"
	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"
	"github.com/memphisdev/memphis/utils"

	"github.com/gin-gonic/gin"
)

const sendNotificationType = "send_notification"
const INTEGRATIONS_AUDIT_LOGS_CONSUMER = "$memphis_integrations_audit_logs_consumer"

type IntegrationsHandler struct{ S *Server }

var integrationsAuditLogLabelToSubjectMap = map[string]string{
	"slack":   integrationsAuditLogsStream + ".%s.slack",
	"discord": integrationsAuditLogsStream + ".%s.discord",
	"s3":      integrationsAuditLogsStream + ".%s.s3",
	"github":  integrationsAuditLogsStream + ".%s.github",
}

func (it IntegrationsHandler) CreateIntegration(c *gin.Context) {
	var message string
	var body models.CreateIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateIntegration at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	exist, _, err := db.GetTenantByName(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]CreateIntegration at GetTenantByName: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant: %v][user: %v]CreateIntegration : tenant %v does not exist", user.TenantName, user.Username, user.TenantName)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	var integration models.Integration
	integrationType := strings.ToLower(body.Name)
	switch integrationType {
	case "slack":
		if !ValidataAccessToFeature(user.TenantName, "feature-integration-slack") {
			serv.Warnf("[tenant: %v][user: %v]CreateIntegration at ValidataAccessToFeature: %v", user.TenantName, user.Username, "feature-notifications")
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "This feature is not available on your current pricing plan, in order to enjoy it you will have to upgrade your plan"})
			return
		}
		_, _, slackIntegration, errorCode, err := it.handleCreateSlackIntegration(user.TenantName, body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant: %v][user: %v]CreateIntegration at handleCreateSlackIntegration: %v", user.TenantName, user.Username, err.Error())
				message = "Server error"
			} else {
				message = err.Error()
				serv.Warnf("[tenant: %v][user: %v]CreateIntegration at handleCreateSlackIntegration: %v", user.TenantName, user.Username, message)
				auditLog := fmt.Sprintf("Error while trying to connect with Slack: %v", message)
				it.Errorf(integrationType, user.TenantName, auditLog)
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = slackIntegration
	case "discord":
		if !ValidataAccessToFeature(user.TenantName, "feature-integration-discord") {
			serv.Warnf("[tenant: %v][user: %v]CreateIntegration at ValidataAccessToFeature: %v", user.TenantName, user.Username, "feature-notifications")
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "This feature is not available on your current pricing plan, in order to enjoy it you will have to upgrade your plan"})
			return
		}
		_, _, discordIntegration, errorCode, err := it.handleCreateDiscordIntegration(user.TenantName, body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant: %v][user: %v]CreateIntegration at handleCreateDiscordIntegration: %v", user.TenantName, user.Username, err.Error())
				message = "Server error"
			} else {
				message = err.Error()
				serv.Warnf("[tenant: %v][user: %v]CreateIntegration at handleCreateDiscordIntegration: %v", user.TenantName, user.Username, message)
				auditLog := fmt.Sprintf("Error while trying to connect with Discord: %v", message)
				it.Errorf(integrationType, user.TenantName, auditLog)
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = discordIntegration
	case "s3":
		if !ValidataAccessToFeature(user.TenantName, "feature-storage-tiering") {
			serv.Warnf("[tenant: %v][user: %v]CreateIntegration at ValidataAccessToFeature: %v", user.TenantName, user.Username, "feature-storage-tiering")
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "This feature is not available on your current pricing plan, in order to enjoy it you will have to upgrade your plan"})
			return
		}
		s3Integration, errorCode, err := it.handleCreateS3Integration(user.TenantName, body.Keys)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant: %v][user: %v]CreateIntegration at handleCreateS3Integration: %v", user.TenantName, user.Username, err.Error())
				message = "Server error"
			} else {
				message = err.Error()
				serv.Warnf("[tenant: %v][user: %v]CreateIntegration at handleCreateS3Integration: %v", user.TenantName, user.Username, message)
				auditLog := fmt.Sprintf("Error while trying to connect with S3: %v", message)
				it.Errorf(integrationType, user.TenantName, auditLog)
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = s3Integration
	case "github":
		githubIntegration, errorCode, err := it.handleCreateGithubIntegration(user.TenantName, body.Keys)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant: %v][user: %v]CreateIntegration at handleCreateGithubIntegration: %v", user.TenantName, user.Username, err.Error())
				message = "Server error"
			} else {
				message = err.Error()
				serv.Warnf("[tenant: %v][user: %v]CreateIntegration at handleCreateGithubIntegration: %v", user.TenantName, user.Username, message)
				auditLog := fmt.Sprintf("Error while trying to connect with Github: %v", message)
				it.Errorf(integrationType, user.TenantName, auditLog)
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = githubIntegration
	default:
		serv.Warnf("[tenant: %v][user: %v]CreateIntegration: Unsupported integration type - %v", user.TenantName, user.Username, integrationType)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Unsupported integration type - " + integrationType})
		return
	}
	auditLog := fmt.Sprintf("Integration with %v created successfully", integrationType)
	it.Noticef(integrationType, user.TenantName, auditLog)
	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-create-integration-"+integrationType)
	}
	c.IndentedJSON(200, integration)
}

func (it IntegrationsHandler) UpdateIntegration(c *gin.Context) {
	var body models.CreateIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("UpdateIntegration: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	exist, _, err := db.GetTenantByName(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]UpdateIntegration at GetTenantByName: %v", user.TenantName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant: %v]UpdateIntegration at GetTenantByName: tenant %v does not exist", user.TenantName, user.TenantName)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	var integration models.Integration
	var message string
	integrationType := strings.ToLower(body.Name)
	switch integrationType {
	case "slack":
		slackIntegration, errorCode, err := it.handleUpdateSlackIntegration(user.TenantName, "slack", body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant:%v][user: %v]UpdateIntegration at handleUpdateSlackIntegration: %v", user.TenantName, user.Username, err.Error())
				message = "Server error"
			} else {
				message = err.Error()
				serv.Warnf("[tenant:%v][user: %v]UpdateIntegration at handleUpdateSlackIntegration: %v", user.TenantName, user.Username, message)
				auditLog := fmt.Sprintf("Error while trying to connect with Slack: %v", message)
				it.Errorf(integrationType, user.TenantName, auditLog)
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = slackIntegration
	case "discord":
		discordIntegration, errorCode, err := it.handleUpdateDiscordIntegration(user.TenantName, "discord", body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant:%v][user: %v]UpdateIntegration at handleUpdateDiscordIntegration: %v", user.TenantName, user.Username, err.Error())
				message = "Server error"
			} else {
				message = err.Error()
				serv.Warnf("[tenant:%v][user: %v]UpdateIntegration at handleUpdateDiscordIntegration: %v", user.TenantName, user.Username, message)
				auditLog := fmt.Sprintf("Error while trying to connect with Discord: %v", message)
				it.Errorf(integrationType, user.TenantName, auditLog)
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = discordIntegration
	case "s3":
		s3Integration, errorCode, err := it.handleUpdateS3Integration(user.TenantName, body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant: %v][user: %v]UpdateIntegration at handleUpdateS3Integration: %v", user.TenantName, user.Username, err.Error())
				message = "Server error"
			} else {
				message = err.Error()
				serv.Warnf("[tenant: %v][user: %v]UpdateIntegration at handleUpdateS3Integration: %v", user.TenantName, user.Username, message)
				auditLog := fmt.Sprintf("Error while trying to connect with S3: %v", message)
				it.Errorf(integrationType, user.TenantName, auditLog)
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = s3Integration
	case "github":
		_, locked, _, err := db.GetAndLockSharedLock("functions", user.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]UpdateIntegration at GetAndLockSharedLock: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		exist, integrationFromDb, err := db.GetIntegration("github", user.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]UpdateIntegration at GetIntegration: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !exist {
			serv.Warnf("[tenant: %v]UpdateIntegration: Integration does not exist %v", user.TenantName, body.Name)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Unsupported integration type - " + body.Name})
			return
		}
		if !locked {
			serv.Warnf("[tenant: %v]UpdateIntegration: Integration %v: Processing a branch is currently in progress. Please wait for the current processes to complete.", user.TenantName, body.Name)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Processing a branch is currently in progress. Please wait for the current processes to complete."})
			return
		}

		var connectedRepos, connectedReposUpdated []interface{}
		connectedReposVal, oKconnectedRepos := integrationFromDb.Keys["connected_repos"].([]interface{})
		if oKconnectedRepos {
			connectedRepos = connectedReposVal
		}
		connectedReposUpdatedVal, oKconnectedReposUpdated := body.Keys["connected_repos"].([]interface{})
		if oKconnectedReposUpdated {
			connectedReposUpdated = connectedReposUpdatedVal
		}
		for _, connectedRepo := range connectedRepos {
			if !containsRepo(connectedReposUpdated, connectedRepo) {
				connectedRepoName, repoNameOK := connectedRepo.(map[string]interface{})["repo_name"].(string)
				connectedRepoBranch, branchNameOK := connectedRepo.(map[string]interface{})["branch"].(string)
				repoOwner, repoOwnerOK := connectedRepo.(map[string]interface{})["repo_owner"].(string)
				if !repoNameOK || !branchNameOK || !repoOwnerOK {
					serv.Errorf("[tenant: %v][user: %v]UpdateIntegration at handleUpdateGithubIntegration: %v", user.TenantName, user.Username, "Repo name, branch or owner name is missing")
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
				err = sendDeleteAllFunctionsReqToMS(user, user.TenantName, integrationType, connectedRepoName, connectedRepoBranch, "aws_lambda", repoOwner, false)
				if err != nil {
					serv.Errorf("[tenant: %v][user: %v]UpdateIntegration at deleteAllFunctionMs: Repo %v: %v", user.TenantName, user.Username, connectedRepoName, err.Error())
					c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
					return
				}
			}
		}
		for i, connectedRepoUpdated := range connectedReposUpdated {
			if !containsRepo(connectedRepos, connectedRepoUpdated) {
				connectedRepoUpdated.(map[string]interface{})["in_progress"] = true
				connectedRepoUpdated.(map[string]interface{})["time_updated"] = time.Now()
				connectedReposUpdated[i] = connectedRepoUpdated
				go sendCloneFunctionReqToMS(connectedRepoUpdated, user, integrationType, body, i)
			}
		}
		body.Keys["connected_repos"] = connectedReposUpdated
		githubIntegration, errorCode, err := it.handleUpdateGithubIntegration(user, body)
		if err != nil {
			if errorCode == 500 {
				serv.Errorf("[tenant: %v][user: %v]UpdateIntegration at handleUpdateGithubIntegration: %v", user.TenantName, user.Username, err.Error())
				message = "Server error"
			} else {
				message = err.Error()
				serv.Warnf("[tenant: %v][user: %v]UpdateIntegration at handleUpdateGithubIntegration: %v", user.TenantName, user.Username, message)
				auditLog := fmt.Sprintf("Error while trying to connect with Github: %v", message)
				it.Errorf(integrationType, user.TenantName, auditLog)
			}
			c.AbortWithStatusJSON(errorCode, gin.H{"message": message})
			return
		}
		integration = githubIntegration
	default:
		serv.Warnf("[tenant: %v]UpdateIntegration: Unsupported integration type - %v", user.TenantName, body.Name)
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Unsupported integration type - " + body.Name})
		return
	}

	auditLog := fmt.Sprintf("Integration with %v updated successfully", integrationType)
	it.Noticef(integrationType, user.TenantName, auditLog)
	c.IndentedJSON(200, integration)
}

func (it IntegrationsHandler) DisconnectIntegration(c *gin.Context) {
	var body models.DisconnectIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("DisconnectIntegration at getUserDetailsFromMiddleware: Integration %v: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	exist, _, err := db.GetTenantByName(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant:%v]DisconnectIntegration at GetTenantByName: %v", user.TenantName, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant: %v]DisconnectIntegration : tenant %v does not exist", user.TenantName, user.TenantName)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	integrationType := strings.ToLower(body.Name)
	if integrationType == "github" {
		_, locked, _, err := db.GetAndLockSharedLock("functions", user.TenantName)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]DisconnectIntegration at GetAndLockSharedLock: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		if !locked {
			errMsg := "Processing a branch or a function is currently in progress. Please wait for the current process to complete."
			serv.Warnf("[tenant: %v]DisconnectIntegration: %v", user.TenantName, errMsg)
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": errMsg})
			return
		}
		err = deleteInstallationForAuthenticatedGithubApp(user.TenantName)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				serv.Warnf("[tenant:%v]DisconnectIntegration at deleteInstallationForAuthenticatedGithubApp: %v", user.TenantName, err.Error())
				c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
				auditLog := fmt.Sprintf("Failed to disconnect %v integration: %v", integrationType, err.Error())
				it.Errorf(integrationType, user.TenantName, auditLog)
				return
			}
			serv.Errorf("[tenant:%v]DisconnectIntegration at deleteInstallationForAuthenticatedGithubApp: %v", user.TenantName, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		err = sendDeleteAllFunctionsReqToMS(user, user.TenantName, integrationType, _EMPTY_, _EMPTY_, "aws_lambda", _EMPTY_, false)
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]DisconnectIntegration at deleteAllFunctionMs: Repo %v: %v", user.TenantName, user.Username, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
	}

	err = db.DeleteIntegration(integrationType, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]DisconnectIntegration at db.DeleteIntegration: Integration %v: %v", user.TenantName, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	integrationUpdate := models.Integration{
		Name:       strings.ToLower(body.Name),
		Keys:       nil,
		Properties: nil,
		TenantName: user.TenantName,
	}

	msg, err := json.Marshal(integrationUpdate)
	if err != nil {
		serv.Errorf("[tenant: %v]DisconnectIntegration at json.Marshal: Integration %v: %v", user.TenantName, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		serv.Errorf("[tenant: %v]DisconnectIntegration at sendInternalAccountMsgWithReply: Integration %v: %v", user.TenantName, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	switch integrationType {
	case "slack", "discord":
		update := models.SdkClientsUpdates{
			Type:   sendNotificationType,
			Update: false,
		}
		serv.SendUpdateToClients(update)
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-disconnect-integration-"+integrationType)
	}

	auditLog := fmt.Sprintf("Integration with %v has been disconnected by user %v", integrationType, user.Username)
	it.Noticef(integrationType, user.TenantName, auditLog)

	c.IndentedJSON(200, gin.H{})
}

func createIntegrationsKeysAndProperties(integrationType, authToken string, channelID string, webhookUrl string, pmAlert bool, svfAlert bool, disconnectAlert bool, accessKey, secretKey, bucketName, region, url, forceS3PathStyle string, githubIntegrationDetails map[string]interface{}, repo, branch, repoType, repoOwner string) (map[string]interface{}, map[string]bool) {
	keys := make(map[string]interface{})
	properties := make(map[string]bool)
	switch integrationType {
	case "slack":
		keys["auth_token"] = authToken
		keys["channel_id"] = channelID
		properties[PoisonMAlert] = pmAlert
		properties[SchemaVAlert] = svfAlert
		properties[DisconEAlert] = disconnectAlert
	case "discord":
		keys["webhook_url"] = webhookUrl
		properties[PoisonMAlert] = pmAlert
		properties[SchemaVAlert] = svfAlert
		properties[DisconEAlert] = disconnectAlert
	case "s3":
		keys["access_key"] = accessKey
		keys["secret_key"] = secretKey
		keys["bucket_name"] = bucketName
		keys["s3_path_style"] = forceS3PathStyle
		keys["region"] = region
		keys["url"] = url
	case "github":
		keys = getGithubKeys(githubIntegrationDetails, repoOwner, repo, branch, repoType)
	}

	return keys, properties
}

func (it IntegrationsHandler) GetIntegrationDetails(c *gin.Context) {
	var body models.GetIntegrationDetailsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetIntegrationDetails at getUserDetailsFromMiddleware: Integration %v: %v", body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	exist, _, err := db.GetTenantByName(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetIntegrationDetails at GetTenantByName: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("[tenant: %v][user: %v]GetIntegrationDetails : tenant %v does not exist", user.TenantName, user.Username, user.TenantName)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	applicationName := retrieveGithubAppName()
	exist, integration, err := db.GetIntegration(strings.ToLower(body.Name), user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetIntegrationDetails at db.GetIntegration: Integration %v: %v", user.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	} else if !exist {
		if body.Name == "github" {
			c.IndentedJSON(200, gin.H{"application_name": applicationName})
			return
		} else {
			c.IndentedJSON(200, nil)
			return
		}
	}

	if integration.Name == "slack" && integration.Keys["auth_token"] != _EMPTY_ {
		integration.Keys["auth_token"] = "xoxb-****"
	}

	if integration.Name == "discord" && integration.Keys["webhook_url"] != _EMPTY_ {
		integration.Keys["webhook_url"] = hideDiscordWebhookUrl(integration.Keys["webhook_url"].(string))
	}

	if integration.Name == "s3" && integration.Keys["secret_key"] != _EMPTY_ {
		integration.Keys["secret_key"] = hideIntegrationSecretKey(integration.Keys["secret_key"].(string))
	}

	sourceCodeIntegration, branchesMap, err := getSourceCodeDetails(user.TenantName, body, "get_all_repos")
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetIntegrationDetails at getSourceCodeDetails: Integration %v: %v", user.TenantName, user.Username, body.Name, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if integration.Name == "github" {
		githubIntegration := models.Integration{}
		githubIntegration.Keys = map[string]interface{}{}
		githubIntegration.Name = sourceCodeIntegration.Name
		githubIntegration.TenantName = sourceCodeIntegration.TenantName
		githubIntegration.IsValid = integration.IsValid
		githubIntegration.Keys["connected_repos"] = sourceCodeIntegration.Keys["connected_repos"]
		memphisFuncs, err := db.GetMemphisFunctionsByMemphis()
		if err != nil {
			serv.Errorf("[tenant: %v][user: %v]GetIntegrationDetails at GetMemphisFunctionsByMemphis: Integration %v: %v", user.TenantName, user.Username, body.Name, err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		for _, memphisFunc := range memphisFuncs {
			if memphisFunc.Installed {
				memphisFunctions["in_progress"] = memphisFunc.InProgress
				break
			} else {
				memphisFunctions["in_progress"] = memphisFunc.InProgress
			}
		}
		githubIntegration.Keys["memphis_functions"] = memphisFunctions
		githubIntegration.Keys["application_name"] = applicationName
		c.IndentedJSON(200, gin.H{"integration": githubIntegration, "repos": branchesMap})
		return
	}
	c.IndentedJSON(200, integration)
}

func (it IntegrationsHandler) GetAllIntegrations(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		message := fmt.Sprintf("GetAllIntegrations at getUserDetailsFromMiddleware: %v", err.Error())
		serv.Errorf(message)
		c.AbortWithStatusJSON(500, gin.H{"message": message})
		return
	}

	_, integrations, err := db.GetAllIntegrationsByTenant(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetAllIntegrations at db.GetAllIntegrationsByTenant: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	for i := 0; i < len(integrations); i++ {
		if integrations[i].Name == "slack" && integrations[i].Keys["auth_token"] != _EMPTY_ {
			integrations[i].Keys["auth_token"] = "xoxb-****"
		}
		if integrations[i].Name == "discord" && integrations[i].Keys["webhook_url"] != _EMPTY_ {
			integrations[i].Keys["webhook_url"] = hideDiscordWebhookUrl(integrations[i].Keys["webhook_url"].(string))
		}
		if integrations[i].Name == "s3" && integrations[i].Keys["secret_key"] != _EMPTY_ {
			integrations[i].Keys["secret_key"] = hideIntegrationSecretKey(integrations[i].Keys["secret_key"].(string))
		}
		if integrations[i].Name == "github" && integrations[i].Keys["installation_id"] != _EMPTY_ {
			memphisFuncs, err := db.GetMemphisFunctionsByMemphis()
			if err != nil {
				serv.Errorf("[tenant: %v][user: %v]GetAllIntegrations at GetMemphisFunctionsByMemphis: %v", user.TenantName, user.Username, err.Error())
				c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
				return
			}
			for _, memphisFunc := range memphisFuncs {
				if memphisFunc.Installed {
					memphisFunctions["in_progress"] = memphisFunc.InProgress
					break
				} else {
					memphisFunctions["in_progress"] = memphisFunc.InProgress
				}
			}
			integrations[i].Keys["memphis_functions"] = memphisFunctions
			delete(integrations[i].Keys, "installation_id")
		}
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := make(map[string]interface{})
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-enter-integration-page")
	}

	c.IndentedJSON(200, integrations)
}

func (it IntegrationsHandler) RequestIntegration(c *gin.Context) {
	var body models.RequestIntegrationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	shouldSendAnalytics, _ := shouldSendAnalytics()
	if shouldSendAnalytics {
		user, _ := getUserDetailsFromMiddleware(c)
		analyticsParams := map[string]interface{}{"request-content": body.RequestContent}
		analytics.SendEvent(user.TenantName, user.Username, analyticsParams, "user-request-integration")
	}

	c.IndentedJSON(200, gin.H{})
}

func (it IntegrationsHandler) GetIntegrationAuditLogs(c *gin.Context) {
	var body models.GetIntegrationsAuditLogsSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetIntegrationAuditLogs at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	integrationType := strings.ToLower(body.Name)
	auditLogs, err := it.S.getIntegrationAuditLogs(integrationType, user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetIntegrationAuditLogs at getIntegrationAuditLogs: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, auditLogs)
}

func (s *Server) getIntegrationAuditLogs(integrationType, tenantName string) ([]models.IntegrationsAuditLog, error) {
	filterSubject, ok := integrationsAuditLogLabelToSubjectMap[integrationType]
	if !ok {
		return []models.IntegrationsAuditLog{}, fmt.Errorf("Unsupported integration type - %v", integrationType)
	}
	filterSubject = fmt.Sprintf(filterSubject, tenantName)
	streamInfo, err := s.memphisStreamInfo(s.MemphisGlobalAccountString(), integrationsAuditLogsStream)
	if err != nil {
		return []models.IntegrationsAuditLog{}, err
	}

	amount := streamInfo.State.Msgs
	const timeout = 500 * time.Millisecond
	uid := s.memphis.nuid.Next()
	msgs := []StoredMsg{}
	durableName := INTEGRATIONS_AUDIT_LOGS_CONSUMER + "_" + uid
	cc := ConsumerConfig{
		DeliverPolicy: DeliverAll,
		AckPolicy:     AckNone,
		Name:          durableName,
		Replicas:      1,
		FilterSubject: filterSubject,
	}

	err = s.memphisAddConsumer(s.MemphisGlobalAccountString(), integrationsAuditLogsStream, &cc)
	if err != nil {
		return []models.IntegrationsAuditLog{}, err
	}

	responseChan := make(chan StoredMsg)
	subject := fmt.Sprintf(JSApiRequestNextT, integrationsAuditLogsStream, durableName)
	reply := durableName + "_reply"
	req := []byte(strconv.FormatUint(amount, 10))
	sub, err := s.subscribeOnAcc(s.MemphisGlobalAccount(), reply, reply+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(respCh chan StoredMsg, subject, reply string, msg []byte) {
			rawTs := tokenAt(reply, 8)
			seq, _, _ := ackReplyInfo(reply)

			intTs, err := strconv.Atoi(rawTs)
			if err != nil {
				s.Errorf("GetSystemLogs: %v", err.Error())
				return
			}

			respCh <- StoredMsg{
				Subject:  subject,
				Sequence: uint64(seq),
				Data:     msg,
				Time:     time.Unix(0, int64(intTs)),
			}
		}(responseChan, subject, reply, copyBytes(msg))
	})
	if err != nil {
		return []models.IntegrationsAuditLog{}, err
	}

	s.sendInternalAccountMsgWithReply(s.MemphisGlobalAccount(), subject, reply, nil, req, true)

	timer := time.NewTimer(timeout)
	for i := uint64(0); i < amount; i++ {
		select {
		case <-timer.C:
			goto cleanup
		case msg := <-responseChan:
			msgs = append(msgs, msg)
		}
	}

cleanup:
	timer.Stop()
	s.unsubscribeOnAcc(s.MemphisGlobalAccount(), sub)

	resMsgs := []models.IntegrationsAuditLog{}
	for _, msg := range msgs {
		data := string(msg.Data)
		resMsgs = append(resMsgs, models.IntegrationsAuditLog{
			ID:         msg.Sequence,
			Message:    data,
			CreatedAt:  msg.Time,
			TenantName: tenantName,
		})
	}

	sort.Slice(resMsgs, func(i, j int) bool {
		return resMsgs[i].ID < resMsgs[j].ID
	})

	return resMsgs, nil
}

func (s *Server) sendIntegrationAuditLogToSubject(integrationType, tenantName string, log string) {
	filterSubject, ok := integrationsAuditLogLabelToSubjectMap[integrationType]
	if !ok {
		return
	}
	filterSubject = fmt.Sprintf(filterSubject, tenantName)
	s.sendInternalAccountMsg(s.MemphisGlobalAccount(), filterSubject, []byte(log))
}

func (it IntegrationsHandler) Errorf(integrationType, tenantName string, log string) {
	it.S.sendIntegrationAuditLogToSubject(integrationType, tenantName, "[ERR] "+log)
}

func (it IntegrationsHandler) Warnf(integrationType, tenantName string, log string) {
	it.S.sendIntegrationAuditLogToSubject(integrationType, tenantName, "[WRN] "+log)
}

func (it IntegrationsHandler) Noticef(integrationType, tenantName string, log string) {
	it.S.sendIntegrationAuditLogToSubject(integrationType, tenantName, "[INF] "+log)
}

func (s *Server) PurgeIntegrationsAuditLogs(tenantName string) {
	requestSubject := fmt.Sprintf(JSApiStreamPurgeT, integrationsAuditLogsStream)
	subj := fmt.Sprintf("%s.%s.*", integrationsAuditLogsStream, tenantName)
	var resp JSApiStreamPurgeResponse
	req := JSApiStreamPurgeRequest{Subject: subj}
	reqj, _ := json.Marshal(req)
	err := jsApiRequest(MEMPHIS_GLOBAL_ACCOUNT, s, requestSubject, kindDeleteMessage, reqj, &resp)
	if err != nil {
		serv.Errorf("[tenant: %v]PurgeIntegrationsAuditLogs at jsApiRequest: %v", tenantName, err.Error())
	}
	respErr := resp.ToError()
	if respErr != nil {
		serv.Errorf("[tenant: %v]PurgeIntegrationsAuditLogs at respErr: %v", tenantName, respErr.Error())
	}
}

func SharedLockUnlock(name, tenantName string) {
	err := db.SharedLockUnlock(name, tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v] at SharedLockUnlock: %v", tenantName, err.Error())
	}
}

func containsRepo(repos []interface{}, target interface{}) bool {
	for _, repo := range repos {
		var mapBranch, mapRepo, targetBranch, targetRepo string
		repoMap, okRepoMap := repo.(map[string]interface{})
		target, okTarget := target.(map[string]interface{})
		if okRepoMap && okTarget {
			if branchVal, ok := repoMap["branch"].(string); ok {
				mapBranch = branchVal
			}
			if repoVal, ok := repoMap["repo_name"].(string); ok {
				mapRepo = repoVal
			}
			if targetBranchVal, ok := target["branch"].(string); ok {
				targetBranch = targetBranchVal
			}

			if targetRepoVal, ok := target["repo_name"].(string); ok {
				targetRepo = targetRepoVal
			}
			if mapBranch == _EMPTY_ && targetBranch == _EMPTY_ && mapRepo == _EMPTY_ && targetRepo == _EMPTY_ {
				return false
			}
			if mapBranch == targetBranch && mapRepo == targetRepo {
				return true
			}
		}
	}
	return false
}
