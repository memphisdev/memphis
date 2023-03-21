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
package conf

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tkanos/gonfig"
)

type Configuration struct {
	DEV_ENV                        string
	LOCAL_CLUSTER_ENV              bool
	JWT_SECRET                     string
	REFRESH_JWT_SECRET             string
	DOCKER_ENV                     string
	ANALYTICS                      string
	K8S_NAMESPACE                  string
	LOGS_RETENTION_IN_DAYS         string
	// GOOGLE_CLIENT_ID               string
	// GOOGLE_CLIENT_SECRET           string
	// SANDBOX_ENV                    string
	// GITHUB_CLIENT_ID               string
	// GITHUB_CLIENT_SECRET           string
	// SANDBOX_REDIRECT_URI           string
	POISON_MSGS_RETENTION_IN_HOURS int
	// SANDBOX_SLACK_BOT_TOKEN        string
	// SANDBOX_SLACK_CHANNEL_ID       string
	// SANDBOX_UI_URL                 string
	TIERED_STORAGE_TIME_FRAME_SEC  int
	EXPORTER                       bool
	POSTGRESQL_USER                string
	POSTGRESQL_PASS                string
	POSTGRESQL_DBNAME              string
	POSTGRESQL_HOST                string
	POSTGRESQL_PORT                string
	POSTGRESQL_TLS_ENABLED         bool
	POSTGRESQL_TLS_KEY             string
	POSTGRESQL_TLS_CRT             string
	POSTGRESQL_TLS_CA              string
}

func GetConfig() Configuration {
	configuration := Configuration{}
	if os.Getenv("DOCKER_ENV") != "" || os.Getenv("LOCAL_CLUSTER_ENV") != "" {
		gonfig.GetConf("./conf/docker-config.json", &configuration)
	} else {
		gonfig.GetConf("./conf/config.json", &configuration)
	}

	gin.SetMode(gin.ReleaseMode)
	return configuration
}
