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
	"errors"
	"fmt"

	"github.com/memphisdev/memphis/db"
)

var IntegrationsConcurrentCache *concurrentMap[map[string]interface{}]
var NotificationFunctionsMap map[string]interface{}
var StorageFunctionsMap map[string]interface{}
var SourceCodeManagementFunctionsMap map[string]map[string]interface{}

const PoisonMAlert = "poison_message_alert"
const SchemaVAlert = "schema_validation_fail_alert"
const DisconEAlert = "disconnection_events_alert"

func InitializeIntegrations() error {
	IntegrationsConcurrentCache = NewConcurrentMap[map[string]interface{}]()
	NotificationFunctionsMap = make(map[string]interface{})
	StorageFunctionsMap = make(map[string]interface{})
	SourceCodeManagementFunctionsMap = make(map[string]map[string]interface{})
	NotificationFunctionsMap["slack"] = sendMessageToSlackChannel
	NotificationFunctionsMap["discord"] = sendMessagesToDiscordChannel
	StorageFunctionsMap["s3"] = serv.uploadToS3Storage
	SourceCodeManagementFunctionsMap["github"] = make(map[string]interface{})
	SourceCodeManagementFunctionsMap["github"]["get_all_repos"] = serv.getGithubRepositories
	SourceCodeManagementFunctionsMap["github"]["get_all_branches"] = serv.getGithubBranches

	err := InitializeConnections()
	if err != nil {
		return err
	}

	return nil
}

func InitializeConnections() error {
	key := getAESKey()
	exist, integrations, err := db.GetAllIntegrations()
	if err != nil {
		return err
	} else if !exist {
		return nil
	}
	for _, integration := range integrations {
		if value, ok := integration.Keys["secret_key"]; ok {
			decryptedValue, err := DecryptAES(key, value.(string))
			if err != nil {
				return err
			}
			integration.Keys["secret_key"] = decryptedValue
		} else if value, ok := integration.Keys["auth_token"]; ok {
			decryptedValue, err := DecryptAES(key, value.(string))
			if err != nil {
				return err
			}
			integration.Keys["auth_token"] = decryptedValue
		} else if value, ok := integration.Keys["webhook_url"]; ok {
			decryptedValue, err := DecryptAES(key, value.(string))
			if err != nil {
				return err
			}
			integration.Keys["webhook_url"] = decryptedValue
		}
		CacheDetails(integration.Name, integration.Keys, integration.Properties, integration.TenantName)
	}
	return nil
}

func CacheDetails(integrationType string, keys map[string]interface{}, properties map[string]bool, tenantName string) {
	switch integrationType {
	case "slack":
		cacheDetailsSlack(keys, properties, tenantName)
	case "discord":
		cacheDetailsDiscord(keys, properties, tenantName)
	case "s3":
		cacheDetailsS3(keys, properties, tenantName)
	case "github":
		cacheDetailsGithub(keys, properties, tenantName)
	}
}

func EncryptOldUnencryptedValues() error {
	tenants, err := db.GetAllTenants()
	if err != nil {
		return err
	}
	for _, tenant := range tenants {
		err := encryptUnencryptedKeysByIntegrationType("s3", "secret_key", tenant.Name)
		if err != nil {
			return err
		}

		err = encryptUnencryptedKeysByIntegrationType("slack", "auth_token", tenant.Name)
		if err != nil {
			return err
		}

		err = encryptUnencryptedKeysByIntegrationType("discord", "webhook_url", tenant.Name)
		if err != nil {
			return err
		}
	}

	err = encryptUnencryptedAppUsersPasswords()
	if err != nil {
		return err
	}
	return nil
}
func encryptUnencryptedKeysByIntegrationType(integrationType, keyTitle string, tenantName string) error {
	exist, integration, err := db.GetIntegration(integrationType, tenantName)
	if err != nil {
		return err
	} else if !exist {
		return nil
	}
	needToEncrypt := false
	key := getAESKey()
	if value, ok := integration.Keys["secret_key"]; ok {
		_, err := DecryptAES(key, value.(string))
		if err != nil {
			needToEncrypt = true
		}
	} else if value, ok := integration.Keys["auth_token"]; ok {
		_, err := DecryptAES(key, value.(string))
		if err != nil {
			needToEncrypt = true
		}
	} else if value, ok := integration.Keys["webhook_url"]; ok {
		_, err := DecryptAES(key, value.(string))
		if err != nil {
			needToEncrypt = true
		}
	}
	if needToEncrypt {
		encryptedValue, err := EncryptAES([]byte(integration.Keys[keyTitle].(string)))
		if err != nil {
			return err
		}
		integration.Keys[keyTitle] = encryptedValue
		_, err = db.UpdateIntegration(integration.TenantName, integrationType, integration.Keys, integration.Properties)
		if err != nil {
			return err
		}
	}
	return nil
}

func encryptUnencryptedAppUsersPasswords() error {
	users, err := db.GetAllUsersByType([]string{"application"})
	if err != nil {
		return err
	}
	key := getAESKey()
	for _, user := range users {
		_, err := DecryptAES(key, user.Password)
		if err != nil {
			password, err := EncryptAES([]byte(user.Password))
			if err != nil {
				return err
			}

			err = db.ChangeUserPassword(user.Username, password, user.TenantName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func addIntegrationToTenant(tenantName string, integrationType string, integrations *concurrentMap[map[string]interface{}], integration interface{}) error {
	integrations.Lock()
	defer integrations.Unlock()
	if i, ok := integrations.m[tenantName]; ok {
		i[integrationType] = integration
		integrations.m[tenantName] = i
	} else {
		return errors.New("AddIntegrationToTenant: tenant not found")
	}
	return nil
}

func deleteIntegrationFromTenant(tenantName string, integrationType string, integrations *concurrentMap[map[string]interface{}]) {
	integrations.Lock()
	defer integrations.Unlock()
	if i, ok := integrations.m[tenantName]; ok {
		delete(i, integrationType)
		integrations.m[tenantName] = i
	}
}

func hideIntegrationSecretKey(secretKey string) string {
	if secretKey != _EMPTY_ {
		lastCharsSecretKey := secretKey[len(secretKey)-4:]
		secretKey = "****" + lastCharsSecretKey
		return secretKey
	}
	return secretKey
}

func copyStringMapToInterfaceMap(srcMap map[string]string) map[string]interface{} {
	destMap := make(map[string]interface{})

	for k, v := range srcMap {
		destMap[k] = v
	}
	return destMap
}

func GetKeysAsStringMap(keys map[string]interface{}) map[string]string {
	stringMap := make(map[string]string)

	for k, v := range keys {
		stringValue := fmt.Sprintf("%v", v)
		stringMap[k] = stringValue
	}
	return stringMap
}
