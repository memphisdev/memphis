package server

import (
	"context"
	"errors"
	"fmt"
	"memphis-broker/models"
	"memphis-broker/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tallstoat/pbparser"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SchemasHandler struct{ S *Server }

const (
	schemaObjectName = "Schema"
)

func validateProtobufContent(schemaContent string) error {
	r := strings.NewReader(schemaContent)
	_, err := pbparser.Parse(r, nil)
	if err != nil {
		return errors.New("Your Proto file is invalid: " + err.Error())
	}

	return nil
}

func validateSchemaName(schemaName string) error {
	return validateName(schemaName, schemaObjectName)
}

func validateSchemaType(schemaType string) error {
	invalidTypeErrStr := fmt.Sprintf("unsupported schema type")
	invalidTypeErr := errors.New(invalidTypeErrStr)
	invalidSupportTypeErrStr := fmt.Sprintf("Json/Avro types are not supported at this time")
	invalidSupportTypeErr := errors.New(invalidSupportTypeErrStr)

	if schemaType == "protobuf" {
		return nil
	} else if schemaType == "avro" || schemaType == "json" {
		return invalidSupportTypeErr
	} else {
		return invalidTypeErr
	}
}

func validateSchemaContent(schemaContent, schemaType string) error {
	switch schemaType {
	case "protobuf":
		err := validateProtobufContent(schemaContent)
		if err != nil {
			return err
		}
	case "json":
		break
	case "avro":
		break
	}
	return nil
}

func (sh SchemasHandler) updateActiveVersion(schemaName string, versionNumber int) error {
	var schema models.Schema
	err := schemasCollection.FindOne(context.TODO(), bson.M{"name": schemaName}).Decode(&schema)
	if err != nil {
		return err
	}

	_, err = schemaVersionCollection.UpdateMany(context.TODO(),
		bson.M{"schema_id": schema.ID},
		bson.M{"$set": bson.M{"active": false}},
	)

	if err != nil {
		return err
	}

	_, err = schemaVersionCollection.UpdateOne(context.TODO(), bson.M{"schema_id": schema.ID, "version_number": versionNumber}, bson.M{"$set": bson.M{"active": true}})
	if err != nil {
		return err
	}
	return nil
}

func (sh SchemasHandler) GetVersionsCount(schemaName string) (models.Schema, int, error) {
	filter := bson.M{"name": schemaName}
	var schema models.Schema
	err := schemasCollection.FindOne(context.TODO(), filter).Decode(&schema)
	if err != nil {
		return models.Schema{}, 0, err
	}
	countVersions, err := schemaVersionCollection.CountDocuments(context.TODO(), bson.M{"schema_id": schema.ID})

	if err != nil {
		return schema, 0, err
	}

	return schema, int(countVersions), err
}

func (sh SchemasHandler) GetSchemaDetailsBySchemaName(schemaName string) (models.ExtendedSchemaDetails, error) {
	var schema models.Schema
	err := schemasCollection.FindOne(context.TODO(), bson.M{"name": schemaName}).Decode(&schema)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}
	var schemaVersions []models.SchemaVersion
	filter := bson.M{"schema_id": schema.ID}
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"creation_date": -1})

	cursor, err := schemaVersionCollection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return models.ExtendedSchemaDetails{}, err
	}
	if err = cursor.All(context.TODO(), &schemaVersions); err != nil {
		return models.ExtendedSchemaDetails{}, err
	}

	extedndedSchemaDetails := models.ExtendedSchemaDetails{
		ID:         schema.ID,
		SchemaName: schema.Name,
		Type:       schema.Type,
		Versions:   schemaVersions,
	}
	return extedndedSchemaDetails, nil
}

func (sh SchemasHandler) GetAllSchemasDetails() ([]models.ExtendedSchema, error) {
	var schemas []models.ExtendedSchema
	cursor, err := schemasCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$lookup", bson.D{{"from", "schema_versions"}, {"localField", "_id"}, {"foreignField", "schema_id"}, {"as", "extendedSchema"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$extendedSchema"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$match", bson.D{{"extendedSchema.version_number", 1}}}},
		bson.D{{"$project", bson.D{{"_id", 1}, {"name", 1}, {"type", 1}, {"created_by_user", "$extendedSchema.created_by_user"}, {"creation_date", "$extendedSchema.creation_date"}}}},
		bson.D{{"$sort", bson.D{{"creation_date", -1}}}},
	})

	if err != nil {
		return []models.ExtendedSchema{}, err
	}

	if err = cursor.All(context.TODO(), &schemas); err != nil {
		return []models.ExtendedSchema{}, err
	}
	if len(schemas) == 0 {
		return []models.ExtendedSchema{}, nil
	} else {
		return schemas, nil
	}
}

func (sh SchemasHandler) GetAllActiveVersions() ([]models.ActiveVersions, error) {
	var schemas []models.ActiveVersions
	cursor, err := schemasCollection.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$lookup", bson.D{{"from", "schema_versions"}, {"localField", "_id"}, {"foreignField", "schema_id"}, {"as", "extendedSchema"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$extendedSchema"}, {"preserveNullAndEmptyArrays", true}}}},
		bson.D{{"$match", bson.D{{"extendedSchema.active", true}}}},
		bson.D{{"$project", bson.D{{"name", 1}, {"version_number", "$extendedSchema.version_number"}}}},
	})

	if err != nil {
		return []models.ActiveVersions{}, err
	}

	if err = cursor.All(context.TODO(), &schemas); err != nil {
		return []models.ActiveVersions{}, err
	}

	if len(schemas) == 0 {
		return []models.ActiveVersions{}, nil

	} else {
		return schemas, nil
	}
}

func (sh SchemasHandler) findAndDeleteSchema(schemaName string) error {
	var schema models.Schema
	filter := bson.M{"name": schemaName}
	err := schemasCollection.FindOneAndDelete(context.TODO(), filter).Decode(&schema)
	if err != nil {
		return err
	}
	filter = bson.M{"schema_id": schema.ID}
	_, err = schemaVersionCollection.DeleteMany(context.TODO(), filter)

	if err != nil {
		return err
	}
	return nil
}

func (sh SchemasHandler) CreateNewSchema(c *gin.Context) {
	var body models.CreateNewSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	schemaName := strings.ToLower(body.Name)
	err := validateSchemaName(schemaName)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	exist, _, err := IsSchemaExist(schemaName)
	if err != nil {
		serv.Errorf("CreateNewSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if exist {
		serv.Warnf("Schema with that name already exists")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema with that name already exists"})
		return
	}
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateNewSchema error: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
	}
	schemaType := strings.ToLower(body.Type)
	err = validateSchemaType(schemaType)

	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schemaType)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	newSchema := models.Schema{
		ID:   primitive.NewObjectID(),
		Name: schemaName,
		Type: schemaType,
	}

	filter := bson.M{"name": newSchema.Name}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":  newSchema.ID,
			"type": newSchema.Type,
		},
	}

	newSchemaVersion := models.SchemaVersion{
		ID:            primitive.NewObjectID(),
		VersionNumber: 1,
		Active:        true,
		CreatedByUser: user.Username,
		CreationDate:  time.Now(),
		SchemaContent: schemaContent,
		SchemaId:      newSchema.ID,
	}
	opts := options.Update().SetUpsert(true)
	updateResults, err := schemasCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		serv.Errorf("CreateSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if updateResults.MatchedCount == 0 {
		_, err = schemaVersionCollection.InsertOne(context.TODO(), newSchemaVersion)
		if err != nil {
			serv.Errorf("CreateSchema error: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
			return
		}
		message := "Schema " + schemaName + " has been created"
		serv.Noticef(message)
	} else {
		serv.Warnf("Schema with that name already exists")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema with that name already exists"})
		return
	}

	c.IndentedJSON(200, newSchema)
}

func (sh SchemasHandler) GetAllSchemas(c *gin.Context) {
	schemas, err := sh.GetAllSchemasDetails()
	if err != nil {
		serv.Errorf("GetAllSchemas error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, schemas)
}

func (sh SchemasHandler) GetSchemaDetails(c *gin.Context) {
	var body models.GetSchemaDetails
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	schemaName := strings.ToLower(body.SchemaName)
	exist, _, err := IsSchemaExist(schemaName)
	if err != nil {
		serv.Errorf("GetSchemaDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Schema does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema does not exist"})
		return
	}

	schemaDetails, err := sh.GetSchemaDetailsBySchemaName(schemaName)

	if err != nil {
		serv.Errorf("GetSchemaDetails error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, schemaDetails)
}

func (sh SchemasHandler) RemoveSchema(c *gin.Context) {
	var body models.RemoveSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}
	schemaName := strings.ToLower(body.SchemaName)
	exist, _, err := IsSchemaExist(schemaName)
	if err != nil {
		serv.Errorf("RemoveSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Schema does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema does not exist"})
		return
	}
	err = sh.findAndDeleteSchema(schemaName)

	if err != nil {
		serv.Errorf("RemoveSchema error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return

	}
	serv.Noticef("Schema " + schemaName + " has been deleted")
	c.IndentedJSON(200, gin.H{})
}

func (sh SchemasHandler) CreateNewVersion(c *gin.Context) {
	var body models.CreateNewVersion
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	schemaName := strings.ToLower(body.SchemaName)
	err := validateSchemaName(schemaName)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	exist, schema, err := IsSchemaExist(schemaName)
	if err != nil {
		serv.Errorf("CreateNewVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if !exist {
		serv.Warnf("Schema does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema does not exist"})
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("CreateNewVersion error: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	schemaContent := body.SchemaContent
	err = validateSchemaContent(schemaContent, schema.Type)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}

	schema, countVersions, err := sh.GetVersionsCount(schemaName)
	if err != nil {
		serv.Errorf("CreateNewVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	versionNumber := countVersions + 1
	newSchema := models.Schema{
		ID:   schema.ID,
		Name: schema.Name,
		Type: schema.Type,
	}

	newSchemaVersion := models.SchemaVersion{
		ID:            primitive.NewObjectID(),
		VersionNumber: versionNumber,
		Active:        false,
		CreatedByUser: user.Username,
		CreationDate:  time.Now(),
		SchemaContent: schemaContent,
		SchemaId:      newSchema.ID,
	}

	filter := bson.M{"schema_id": schema.ID, "version_number": newSchemaVersion.VersionNumber}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":             newSchemaVersion.ID,
			"active":          newSchemaVersion.Active,
			"created_by_user": newSchemaVersion.CreatedByUser,
			"creation_date":   newSchemaVersion.CreationDate,
			"schema_content":  newSchemaVersion.SchemaContent,
		},
	}

	opts := options.Update().SetUpsert(true)
	updateResults, err := schemaVersionCollection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		serv.Errorf("CreateNewVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if updateResults.MatchedCount == 0 {
		message := "Version " + strconv.Itoa(newSchemaVersion.VersionNumber) + " has been created"
		serv.Noticef(message)
	} else {
		serv.Warnf("Version already exists")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Version already exists"})
		return
	}
	c.IndentedJSON(200, newSchemaVersion)

}

func (sh SchemasHandler) RollBackVersion(c *gin.Context) {
	var body models.RollBackVersion
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	schemaName := strings.ToLower(body.SchemaName)
	err := validateSchemaName(schemaName)
	if err != nil {
		serv.Warnf(err.Error())
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": err.Error()})
		return
	}
	exist, _, err := IsSchemaExist(schemaName)
	if err != nil {
		serv.Errorf("RollBackVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server Error"})
		return
	}
	if !exist {
		serv.Warnf("Schema does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Schema does not exist"})
		return
	}

	schemaVersion := body.VersionNumber
	exist, _, err = isSchemaVersionExists(schemaVersion, schemaName)

	if err != nil {
		serv.Errorf("RollBackVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("Version does not exist")
		c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "version does not exist"})
		return
	}

	err = sh.updateActiveVersion(schemaName, body.VersionNumber)
	if err != nil {
		serv.Errorf("RollBackVersion error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(200, gin.H{})
}

func (sh SchemasHandler) GetActiveVersions(c *gin.Context) {

	schemas, err := sh.GetAllActiveVersions()
	if err != nil {
		serv.Errorf("GetActiveVersions error: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	c.IndentedJSON(200, schemas)

}
