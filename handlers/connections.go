// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"memphis-broker/models"

	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConnectionsHandler struct{}

func (ch ConnectionsHandler) CreateConnection(username string, clientAddress string) (primitive.ObjectID, error) {
	connectionId := primitive.NewObjectID()

	username = strings.ToLower(username)
	exist, _, err := IsUserExist(username)
	if err != nil {
		serv.Errorf("CreateConnection error: " + err.Error())
		return connectionId, err
	}
	if !exist {
		return connectionId, errors.New("User was not found")
	}

	newConnection := models.Connection{
		ID:            connectionId,
		CreatedByUser: username,
		IsActive:      true,
		CreationDate:  time.Now(),
		LastPing:      time.Now(),
		ClientAddress: clientAddress,
	}

	_, err = connectionsCollection.InsertOne(context.TODO(), newConnection)
	if err != nil {
		serv.Errorf("CreateConnection error: " + err.Error())
		return connectionId, err
	}
	return connectionId, nil
}

func (ch ConnectionsHandler) KillConnection(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		serv.Errorf("KillConnection error: " + err.Error())
		return err
	}

	return nil
}

func (ch ConnectionsHandler) ReliveConnection(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"is_active": true}},
	)
	if err != nil {
		serv.Errorf("ReliveConnection error: " + err.Error())
		return err
	}

	return nil
}

func (ch ConnectionsHandler) UpdatePingTime(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.UpdateOne(context.TODO(),
		bson.M{"_id": connectionId},
		bson.M{"$set": bson.M{"last_ping": time.Now()}},
	)
	if err != nil {
		serv.Errorf("UpdatePingTime error: " + err.Error())
		return err
	}

	return nil
}
