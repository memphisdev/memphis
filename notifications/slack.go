package notifications

import (
	"context"
	"errors"

	"memphis-broker/db"
	"memphis-broker/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/slack-go/slack"
)

var SlackAuthToken string
var SlackChannelID string
var SlackPoisonMessageAlert bool
var SlackSchemaValidationFailAlert bool
var UIUrl string
var SlackClient *slack.Client
var integrationsCollection *mongo.Collection

func InitializeSlackConnection(c *mongo.Client) error {
	integrationsCollection = db.GetCollection("integrations", c)
	filter := bson.M{"name": "slack"}
	var slackIntegration models.Integration
	err := integrationsCollection.FindOne(context.TODO(),
		filter).Decode(&slackIntegration)
	if err == mongo.ErrNoDocuments {
		UpdateEmptySlackDetails()
		return nil
	} else if err != nil {
		return err
	}
	UpdateSlackDetails(slackIntegration.Keys, slackIntegration.Properties, slackIntegration.UIUrl)
	return nil
}

func UpdateEmptySlackDetails() {
	UIUrl = ""
	SlackAuthToken = ""
	SlackChannelID = ""
	SlackPoisonMessageAlert = false
	SlackSchemaValidationFailAlert = false
	SlackClient = nil
}

func UpdateSlackDetails(keys map[string]string, properties map[string]bool, url string) {
	var authToken, channelID string
	var poisonMessageAlert, schemaValidationFailAlert bool
	UIUrl = url
	if keys == nil {
		authToken = ""
		channelID = ""
		SlackClient = nil
	}
	if properties == nil {
		poisonMessageAlert = false
		schemaValidationFailAlert = false
	}
	authToken, ok := keys["auth_token"]
	if !ok {
		authToken = ""
		SlackClient = nil
	}
	channelID, ok = keys["channel_id"]
	if !ok {
		channelID = ""
	}
	poisonMessageAlert, ok = properties["poison_message_alert"]
	if !ok {
		poisonMessageAlert = false
	}
	schemaValidationFailAlert, ok = properties["schema_validation_fail_alert"]
	if !ok {
		schemaValidationFailAlert = false
	}
	if SlackAuthToken != authToken {
		SlackAuthToken = authToken
		if authToken != "" {
			SlackClient = slack.New(authToken)
		}
	}
	SlackChannelID = channelID
	SlackPoisonMessageAlert = poisonMessageAlert
	SlackSchemaValidationFailAlert = schemaValidationFailAlert
}

func SendMessageToSlackChannel(message string) error {
	if SlackChannelID != "" || SlackClient != nil {
		attachment := slack.Attachment{
			Pretext: "Memphis Notification",
			Text:    message,
			Color:   "#6557FF",
		}

		_, _, err := SlackClient.PostMessage(
			SlackChannelID,
			slack.MsgOptionAttachments(attachment),
		)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Invalid slack credentials")
}
