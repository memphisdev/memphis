// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"memphis-control-plane/broker"
	"memphis-control-plane/db"
	"memphis-control-plane/handlers"
	"memphis-control-plane/http_server"
	"memphis-control-plane/logger"
	"memphis-control-plane/tcp_server"
	"memphis-control-plane/utils"
	"os"
	"sync"
)

func main() {
	err := handlers.CreateRootUserOnFirstSystemLoad()
	if err != nil {
		logger.Error("Failed to create root user: " + err.Error())
		panic("Failed to create root user: " + err.Error())
	}

	defer db.Close()
	defer broker.Close()

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go tcp_server.InitializeTcpServer(wg)
	go http_server.InitializeHttpServer(wg)
	go utils.KillZombieConnections(wg)

	env := os.Getenv("ENVIRONMENT")
	if env == "" && os.Getenv("DOCKER_ENV") != "" {
		env = "Docker"
		logger.Info("\n**********\n\nDashboard: http://localhost:9000\nMemphis broker: localhost, Open ports: 80 (For CLI), 6666 (For SDK connections), 7766 (brokerPort - For SDK connections)\n\n**********")
	} else if env == "" && os.Getenv("DOCKER_ENV") == "" {
		env = "K8S"
	}

	logger.Info("Memphis control plane is up and running, ENV: " + env)
	wg.Wait()
}
