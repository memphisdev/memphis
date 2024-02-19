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
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"strconv"
	"strings"

	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type TieredStorageMsg struct {
	Buf         []byte `json:"buf"`
	StationName string `json:"station_name"`
	TenantName  string `json:"tenant_name"`
}

func cacheDetailsS3(keys map[string]interface{}, properties map[string]bool, tenantName string) {
	s3Integration := models.Integration{}
	s3Integration.Keys = make(map[string]interface{})
	s3Integration.Properties = make(map[string]bool)
	if keys == nil {
		deleteIntegrationFromTenant(tenantName, "s3", IntegrationsConcurrentCache)
		return
	}

	s3Integration.Keys["access_key"] = keys["access_key"].(string)
	s3Integration.Keys["secret_key"] = keys["secret_key"].(string)
	s3Integration.Keys["bucket_name"] = keys["bucket_name"].(string)
	s3Integration.Keys["region"] = keys["region"].(string)
	if _, ok := s3Integration.Keys["url"].(string); ok {
		s3Integration.Keys["url"] = keys["url"].(string)
	} else {
		s3Integration.Keys["url"] = _EMPTY_
	}
	if _, ok := s3Integration.Keys["s3_path_style"].(string); ok {
		s3Integration.Keys["s3_path_style"] = keys["s3_path_style"].(string)
	} else {
		s3Integration.Keys["s3_path_style"] = "false"
	}
	s3Integration.Name = "s3"
	if _, ok := IntegrationsConcurrentCache.Load(tenantName); !ok {
		IntegrationsConcurrentCache.Add(tenantName, map[string]interface{}{"s3": s3Integration})
	} else {
		err := addIntegrationToTenant(tenantName, "s3", IntegrationsConcurrentCache, s3Integration)
		if err != nil {
			serv.Errorf("cacheDetailsS3: %s ", err.Error())
			return
		}
	}
}

func (it IntegrationsHandler) handleCreateS3Integration(tenantName string, keys map[string]interface{}) (models.Integration, int, error) {
	statusCode, keys, err := it.handleS3Integration(tenantName, keys)
	if err != nil {
		return models.Integration{}, statusCode, err
	}

	keysMap, properties := createIntegrationsKeysAndProperties("s3", _EMPTY_, _EMPTY_, _EMPTY_, false, false, false, keys["access_key"].(string), keys["secret_key"].(string), keys["bucket_name"].(string), keys["region"].(string), keys["url"].(string), keys["s3_path_style"].(string), map[string]interface{}{}, _EMPTY_, _EMPTY_, _EMPTY_, _EMPTY_)
	s3Integration, err := createS3Integration(tenantName, keysMap, properties)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return models.Integration{}, SHOWABLE_ERROR_STATUS_CODE, err
		} else {
			return models.Integration{}, 500, err
		}
	}
	return s3Integration, statusCode, nil
}

func (it IntegrationsHandler) handleUpdateS3Integration(tenantName string, body models.CreateIntegrationSchema) (models.Integration, int, error) {
	statusCode, keys, err := it.handleS3Integration(tenantName, body.Keys)
	if err != nil {
		return models.Integration{}, statusCode, err
	}
	integrationType := strings.ToLower(body.Name)
	keysMap, properties := createIntegrationsKeysAndProperties(integrationType, _EMPTY_, _EMPTY_, _EMPTY_, false, false, false, keys["access_key"].(string), keys["secret_key"].(string), keys["bucket_name"].(string), keys["region"].(string), keys["url"].(string), keys["s3_path_style"].(string), map[string]interface{}{}, _EMPTY_, _EMPTY_, _EMPTY_, _EMPTY_)
	s3Integration, err := updateS3Integration(tenantName, keysMap, properties)
	if err != nil {
		return s3Integration, 500, err
	}
	return s3Integration, statusCode, nil
}

func getS3EndpointResolver(region, url string) aws.EndpointResolverWithOptionsFunc {
	return aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		// Override default endpoint lookup if the url was specified explicitly
		if url != _EMPTY_ {
			return aws.Endpoint{
				PartitionID:   "other",
				URL:           url,
				SigningRegion: region,
			}, nil
		}

		// returning an EndpointNotFoundError will trigger default endpoint lookup
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
}

func (it IntegrationsHandler) handleS3Integration(tenantName string, keys map[string]interface{}) (int, map[string]interface{}, error) {
	accessKey := keys["access_key"]
	secretKey := keys["secret_key"]
	region := keys["region"]
	bucketName := keys["bucket_name"]
	var pathStyle bool
	var url string
	if _, ok := keys["s3_path_style"].(string); ok {
		pathStyle, _ = strconv.ParseBool(keys["s3_path_style"].(string))
	} else {
		pathStyle = false
		keys["s3_path_style"] = "false"
	}
	if _, ok := keys["url"].(string); ok {
		url = keys["url"].(string)
	} else {
		url = _EMPTY_
		keys["url"] = url
	}

	if keys["secret_key"] == _EMPTY_ {
		exist, integrationFromDb, err := db.GetIntegration("s3", tenantName)
		if err != nil {
			return 500, map[string]interface{}{}, err
		}
		if !exist {
			return SHOWABLE_ERROR_STATUS_CODE, map[string]interface{}{}, errors.New("secret key is invalid")
		}
		if value, ok := integrationFromDb.Keys["secret_key"]; ok {
			key := getAESKey()
			decryptedValue, err := DecryptAES(key, value.(string))
			if err != nil {
				return 500, map[string]interface{}{}, err
			}
			integrationFromDb.Keys["secret_key"] = decryptedValue
		}
		secretKey = integrationFromDb.Keys["secret_key"].(string)
		keys["secret_key"] = secretKey
	}

	provider := credentials.NewStaticCredentialsProvider(accessKey.(string), secretKey.(string), _EMPTY_)
	_, err := provider.Retrieve(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "static credentials are empty") {
			return SHOWABLE_ERROR_STATUS_CODE, map[string]interface{}{}, errors.New("credentials are empty")
		} else {
			return 500, map[string]interface{}{}, err
		}
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(provider),
		awsconfig.WithRegion(region.(string)),
		awsconfig.WithEndpointResolverWithOptions(getS3EndpointResolver(region.(string), url)),
	)
	if err != nil {
		return 500, map[string]interface{}{}, err
	}

	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = pathStyle
	})

	statusCode, err := testS3Integration(svc, bucketName.(string), url)
	if err != nil {
		return statusCode, map[string]interface{}{}, err
	}
	return statusCode, keys, nil
}

func createS3Integration(tenantName string, keys map[string]interface{}, properties map[string]bool) (models.Integration, error) {
	exist, s3Integration, err := db.GetIntegration("s3", tenantName)
	if err != nil {
		return models.Integration{}, err
	} else if !exist {
		stringMapKeys := GetKeysAsStringMap(keys)
		cloneKeys := copyMaps(stringMapKeys)
		encryptedValue, err := EncryptAES([]byte(keys["secret_key"].(string)))
		if err != nil {
			return models.Integration{}, err
		}
		cloneKeys["secret_key"] = encryptedValue
		interfaceMapKeys := copyStringMapToInterfaceMap(cloneKeys)
		integrationRes, insertErr := db.InsertNewIntegration(tenantName, "s3", interfaceMapKeys, properties)
		if insertErr != nil {
			return models.Integration{}, insertErr
		}
		s3Integration = integrationRes
		integrationToUpdate := models.CreateIntegration{
			Name:       "s3",
			Keys:       keys,
			Properties: properties,
			TenantName: tenantName,
			IsValid:    integrationRes.IsValid,
		}
		msg, err := json.Marshal(integrationToUpdate)
		if err != nil {
			return models.Integration{}, err
		}
		err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
		if err != nil {
			return models.Integration{}, err
		}
		s3Integration.Keys["secret_key"] = hideIntegrationSecretKey(keys["secret_key"].(string))
		return s3Integration, nil
	}
	return models.Integration{}, errors.New("s3 integration already exists")

}

func updateS3Integration(tenantName string, keys map[string]interface{}, properties map[string]bool) (models.Integration, error) {
	stringMapKeys := GetKeysAsStringMap(keys)
	cloneKeys := copyMaps(stringMapKeys)
	encryptedValue, err := EncryptAES([]byte(keys["secret_key"].(string)))
	if err != nil {
		return models.Integration{}, err
	}
	cloneKeys["secret_key"] = encryptedValue
	interfaceMapKeys := copyStringMapToInterfaceMap(cloneKeys)
	s3Integration, err := db.UpdateIntegration(tenantName, "s3", interfaceMapKeys, properties)
	if err != nil {
		return models.Integration{}, err
	}

	integrationToUpdate := models.CreateIntegration{
		Name:       "s3",
		Keys:       keys,
		Properties: properties,
		TenantName: tenantName,
		IsValid:    s3Integration.IsValid,
	}

	msg, err := json.Marshal(integrationToUpdate)
	if err != nil {
		return s3Integration, err
	}
	err = serv.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), INTEGRATIONS_UPDATES_SUBJ, _EMPTY_, nil, msg, true)
	if err != nil {
		return s3Integration, err
	}

	keys["secret_key"] = hideIntegrationSecretKey(keys["secret_key"].(string))
	s3Integration.Keys = keys
	s3Integration.Properties = properties
	return s3Integration, nil
}

func testS3Integration(svc *s3.Client, bucketName, urlEndpointResolver string) (int, error) {
	_, err := svc.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	var statusCode int
	customUrlEndpointErr := fmt.Sprintf("lookup %s.bucket.: no such host", bucketName)
	if err != nil {
		if strings.Contains(err.Error(), "Unauthorized") {
			err = fmt.Errorf("the provided AWS credentials are incorrect or you do not have the necessary permissions to access the bucket %s", bucketName)
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "Forbidden") {
			err = errors.New("invalid access key or secret key")
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "404") {
			err = errors.New("bucket does not exist")
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "Moved Permanently") {
			err = errors.New("bucket name does not exist in the selected region, it is probably placed in another region")
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "send request failed") {
			err = errors.New("upload failed")
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "could not find region configuration") {
			var oe *smithy.OperationError
			if errors.As(err, &oe) {
				err = fmt.Errorf("%s : region name is empty", oe.Error())
			}
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "validation error(s) found") || strings.Contains(err.Error(), "BadRequest: Bad Request") {
			err = errors.New("invalid bucket name")
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "incorrect region") {
			var oe *smithy.OperationError
			if errors.As(err, &oe) {
				err = fmt.Errorf("%s: incorrect region", oe.Error())
			}
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), customUrlEndpointErr) {
			err = fmt.Errorf("No such host: endpoint url %s is incorrect", urlEndpointResolver)
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "failed to parse endpoint") || strings.Contains(err.Error(), "no such host") {
			err = errors.New("invalid bucket name or region")
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else if strings.Contains(err.Error(), "HeadBucket, exceeded maximum number of attempts") {
			err = fmt.Errorf("The endpoint url %s is incorrect", urlEndpointResolver)
			statusCode = SHOWABLE_ERROR_STATUS_CODE
		} else {
			statusCode = 500
		}
		return statusCode, err
	}

	acl, err := svc.GetBucketAcl(context.Background(), &s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		err = fmt.Errorf("getBucketAcl error: you don’t have enough permissions to get the access to bucket %s ", err.Error())
		return SHOWABLE_ERROR_STATUS_CODE, err
	}

	permission := acl.Grants[0].Permission

	if permission != types.PermissionFullControl {
		err = fmt.Errorf("creds should have full access on this bucket")
		return SHOWABLE_ERROR_STATUS_CODE, err
	}

	uploader := manager.NewUploader(svc)
	reader := strings.NewReader(string("test"))
	// Upload the object to S3.
	_, err = uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("memphis"),
		Body:   reader,
	})
	if err != nil {
		err = fmt.Errorf("could not upload objects -  %s", err.Error())
		return SHOWABLE_ERROR_STATUS_CODE, err
	}
	_, err = svc.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("memphis"),
	})
	if err != nil {
		err = fmt.Errorf("could not delete objects -  %s", err.Error())
		return SHOWABLE_ERROR_STATUS_CODE, err
	}

	return 0, nil
}

type Msg struct {
	Payload string            `json:"payload"`
	Headers map[string]string `json:"headers"`
}

func (s *Server) uploadToS3Storage(tenantName string, tenant map[string][]StoredMsg) error {
	for k, msgs := range tenant {
		var credentialsMap models.Integration
		if tenantIntegrations, ok := IntegrationsConcurrentCache.Load(tenantName); !ok {
			continue
		} else {
			if credentialsMap, ok = tenantIntegrations["s3"].(models.Integration); !ok {
				continue
			}
		}
		provider := credentials.NewStaticCredentialsProvider(
			credentialsMap.Keys["access_key"].(string),
			credentialsMap.Keys["secret_key"].(string),
			_EMPTY_,
		)

		region := credentialsMap.Keys["region"]
		url := credentialsMap.Keys["url"]
		pathStyle, _ := strconv.ParseBool(credentialsMap.Keys["s3_path_style"].(string))

		_, err := provider.Retrieve(context.Background())
		if err != nil {
			err = errors.New("uploadToS3Storage: Invalid credentials")
			return err
		}
		cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithCredentialsProvider(provider),
			awsconfig.WithRegion(credentialsMap.Keys["region"].(string)),
			awsconfig.WithEndpointResolverWithOptions(getS3EndpointResolver(region.(string), url.(string))),
		)
		svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = pathStyle
		})
		if err != nil {
			err = errors.New("uploadToS3Storage failure " + err.Error())
			return err
		}
		uploader := manager.NewUploader(svc)
		uid := serv.memphis.nuid.Next()
		var objectName string

		var messages []Msg
		size := int64(0)
		for _, msg := range msgs {
			if tenantName == serv.MemphisGlobalAccountString() {
				tenantName = "global"
			}
			objectName = "memphis/" + tenantName + "/" + k + "/" + uid + "(" + strconv.Itoa(len(msgs)) + ").json"
			var headers string
			hdrs := map[string]string{}
			if len(msg.Header) > 0 {
				headers = strings.ToLower(string(msg.Header))
				headersSplit := strings.Split(headers, CR_LF)
				for _, header := range headersSplit {
					if header != _EMPTY_ && !strings.Contains(header, "nats") {
						keyVal := strings.Split(header, ":")
						key := strings.TrimSpace(keyVal[0])
						value := strings.TrimSpace(keyVal[1])
						hdrs[key] = value
					}
				}
			} else {
				headers = _EMPTY_
			}

			encodedMsg := hex.EncodeToString(msg.Data)
			message := Msg{Payload: encodedMsg, Headers: hdrs}
			messages = append(messages, message)
			size += int64(len(msg.Data)) + int64(len(msg.Header))
		}
		// Upload the object to S3.
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(messages)
		if err != nil {
			return err
		}
		_, err = uploader.Upload(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String(credentialsMap.Keys["bucket_name"].(string)),
			Key:    aws.String(objectName),
			Body:   &buf,
		})
		if err != nil {
			err = errors.New("uploadToS3Storage: failed to upload object to S3: " + err.Error())
			return err
		}
		IncrementEventCounter(tenantName, "tiered", size, int64(len(messages)), _EMPTY_, []byte{}, []byte{})
		serv.Noticef("new file has been uploaded to S3: %s", objectName)
	}

	return nil

}
