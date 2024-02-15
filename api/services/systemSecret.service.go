package services

import (
	"context"
	"encoding/json"
	"fmt"
	"secret-svc/api/dtos"
	"secret-svc/pkg/constants"
	"secret-svc/pkg/utils"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.uber.org/zap"
)

// Get a system secret from the system secret Manager
// /////////////////////////////////////////////////////
// - returns the latest version if version param is not provided
func GetSystemSecret(headers dtos.CustomHeaders, version string) (map[string]interface{}, string, string, string, string, error) {
	REGION := utils.GetEnvVar("REGION")
	secretName := utils.CreatePrefix(headers)
	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(REGION))
	svc := secretsmanager.NewFromConfig(config)

	zap.L().Info("Getting System Secret :: " + secretName)
	input := secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String(utils.SetDefaultIfEmptyValue(version, "AWSCURRENT")),
	}
	result, err := svc.GetSecretValue(context.TODO(), &input)

	if err != nil {
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			zap.L().Error("GetSecretValue Failed :: " + err.Error())
			return nil, "", "", "", "", constants.ErrUnregisteredKey
		}

		zap.L().Error(err.Error())
		return nil, "", "", "", "", err
	}

	secretString := *result.SecretString
	var systemSecret map[string]interface{}
	utils.DeStringifyJson(secretString, &systemSecret)

	// Extract details from the nested map
	storedFlow, _ := systemSecret[constants.FLOW_META_DATA].(string)
	storedARN, _ := systemSecret[constants.ARN_META_DATA].(string)
	storedRegion, _ := systemSecret[constants.REGION_META_DATA].(string)
	storedProvider, _ := systemSecret[constants.PROVIDER_META_DATA].(string)

	return systemSecret, storedARN, storedRegion, storedProvider, storedFlow, nil
}

// Creates a new System Secret in the System secret manager
// ////////////////////////////////////////////////////////////
func CreateSystemSecret(headers dtos.CustomHeaders, requestBody dtos.SystemSecretReq) ([]string, error) {
	REGION := utils.GetEnvVar("REGION")
	ARN := utils.GetEnvVar("SHARED_SECRET_MNGR_ARN")
	secretDescription := fmt.Sprintf("Organization ID: %s", headers.OrgId)
	secretNames := getSecretNames(headers)
	zap.L().Info("Creating System Secrets :: " + strings.Join(secretNames, ","))

	// Construct the desired JSON format
	jsonData := map[string]interface{}{
		constants.ARN_META_DATA:      utils.SetDefaultIfEmptyValue(requestBody.ARN, ARN),
		constants.PROVIDER_META_DATA: utils.SetDefaultIfEmptyValue(requestBody.Provider, "AWS"),
		constants.REGION_META_DATA:   utils.SetDefaultIfEmptyValue(requestBody.Region, REGION),
		constants.FLOW_META_DATA:     requestBody.Flow,
	}

	// Serialize the jsonData to a JSON string
	secretString, err := utils.StringifyJson(jsonData)
	if err != nil {
		zap.L().Error("StringifyJson failed :: " + err.Error())
		return nil, err
	}

	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(REGION))
	svc := secretsmanager.NewFromConfig(config)

	for _, secretName := range secretNames {
		createSecretInput := &secretsmanager.CreateSecretInput{
			Name:         &secretName,
			Description:  &secretDescription,
			SecretString: &secretString,
		}

		_, err = svc.CreateSecret(context.TODO(), createSecretInput)
		if err != nil {
			zap.L().Error(fmt.Sprintf("CreateSecret failed :: %s :: ", secretName) + err.Error())
			return nil, err
		}
	}

	return secretNames, nil
}

// Updates a system secret value in the system secret manager
// //////////////////////////////////////////////////////////////
func UpdateSystemSecret(headers dtos.CustomHeaders, requestBody dtos.SystemSecretReq) ([]string, error) {
	REGION := utils.GetEnvVar("REGION")
	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(REGION))
	svc := secretsmanager.NewFromConfig(config)
	secretNames := getSecretNames(headers)
	var newIDs []string
	zap.L().Info("Updating System Secrets :: " + strings.Join(secretNames, ","))

	for _, secretName := range secretNames {
		getSecretInput := &secretsmanager.GetSecretValueInput{
			SecretId: &secretName,
		}
		getSecretValueOutput, err := svc.GetSecretValue(context.TODO(), getSecretInput)

		if err != nil {
			zap.L().Error(fmt.Sprintf("GetSecretValue %s Failed :: ", secretName) + err.Error())
			return nil, err
		}

		var existingData map[string]interface{}
		if err := json.Unmarshal([]byte(*getSecretValueOutput.SecretString), &existingData); err != nil {
			zap.L().Error("Unmarshalling json Failed :: " + err.Error())
			return nil, err
		}

		switch {
		// SHARED -> PRIVATE Migration
		case requestBody.Flow == constants.PRIVATE_FLOW && existingData[constants.FLOW_META_DATA] == constants.SHARED_FLOW:
			zap.L().Info(fmt.Sprintf("Mirgating %s from SHARED to PRIVATE", secretName))
			ids, err := MigrateSecretsSharedToPvt(headers, secretName, existingData, requestBody.ARN, requestBody.Region)
			if err != nil {
				zap.L().Error(fmt.Sprintf("MigrateSecretsSharedToPvt Failed :: %s :: ", secretName) + err.Error())
				return nil, err
			}
			newIDs = append(newIDs, ids...)

			//Get values for ARN, region, and provider and store
			existingData[constants.ARN_META_DATA] = requestBody.ARN
			existingData[constants.REGION_META_DATA] = requestBody.Region
			existingData[constants.PROVIDER_META_DATA] = requestBody.Provider
			existingData[constants.FLOW_META_DATA] = requestBody.Flow

		default:
			zap.L().Error(constants.ErrInvalidMigration.Error())
		}

		// Serialize the updated data to a JSON string
		updatedSecretString, err := utils.StringifyJson(existingData)

		if err != nil {
			zap.L().Error("StringifyJson Failed :: " + err.Error())
			return nil, err
		}

		// Update the existing secret
		input := &secretsmanager.UpdateSecretInput{
			SecretId:     &secretName,
			SecretString: aws.String(string(updatedSecretString)),
		}

		_, err = svc.UpdateSecret(context.TODO(), input)
		if err != nil {
			zap.L().Error(fmt.Sprintf("UpdateSecret %s Failed :: ", secretName) + err.Error())
			return nil, err
		}
	}

	return newIDs, nil
}

// Returns the different versions of a secret
// ///////////////////////////////////////////////
func GetSystemSecretVersions(headers dtos.CustomHeaders) (secretsmanager.ListSecretVersionIdsOutput, error) {
	REGION := utils.GetEnvVar("REGION")
	secretName := utils.CreatePrefix(headers)
	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(REGION))
	svc := secretsmanager.NewFromConfig(config)
	input := &secretsmanager.ListSecretVersionIdsInput{
		SecretId: &secretName,
	}

	zap.L().Info("Getting System Secret versions :: " + secretName)
	result, err := svc.ListSecretVersionIds(context.TODO(), input)

	if err != nil {
		zap.L().Error(fmt.Sprintf("ListSecretVersionIds Failed :: %s :: ", secretName) + err.Error())
		return secretsmanager.ListSecretVersionIdsOutput{}, err
	}

	return *result, nil
}

// Deletes a key from the System Secret Manager which may be sub projects and Scope(keys/values)
// //////////////////////////////////////////////////////////////////////////////////////////////////
func DeleteSystemSecret(headers dtos.CustomHeaders) ([]string, error) {
	region := utils.GetEnvVar("REGION")
	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	svc := secretsmanager.NewFromConfig(config)
	secretNames := getSecretNames(headers)
	deleteAsap := true
	zap.L().Info("Deleting System Secrets :: " + strings.Join(secretNames, ","))

	for _, secretName := range secretNames {
		getSecretInput := &secretsmanager.GetSecretValueInput{
			SecretId: &secretName,
		}
		getSecretValueOutput, err := svc.GetSecretValue(context.TODO(), getSecretInput)

		if err != nil {
			zap.L().Error(fmt.Sprintf("GetSecretValue Failed :: %s :: ", secretName) + err.Error())
			return nil, err
		}

		var existingData map[string]interface{}
		if err := json.Unmarshal([]byte(*getSecretValueOutput.SecretString), &existingData); err != nil {
			zap.L().Error("Unmarshalling json Failed :: " + err.Error())
			return nil, err
		}

		zap.L().Info("Deleting Sub-sequent Secrets from the shared secret manager :: " + secretName)
		DeleteSecretGroup(headers, existingData[constants.ARN_META_DATA].(string), existingData[constants.REGION_META_DATA].(string))

		input := &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secretName),
			ForceDeleteWithoutRecovery: &deleteAsap,
		}
		_, err = svc.DeleteSecret(context.TODO(), input)

		if err != nil {
			zap.L().Error("DeleteSecret Failed :: " + err.Error())
			return nil, err
		}
	}

	return secretNames, nil
}

// Helper function to get secretNames with scopes
// ///////////////////////////////////////////////////
func getSecretNames(headers dtos.CustomHeaders) []string {
	var secretNames []string
	if headers.ProjectId != "" {
		for _, scope := range constants.ACCEPTED_SCOPES {
			headers.Scope = scope
			secretNames = append(secretNames, utils.CreatePrefix(headers))
		}
	} else {
		secretNames = append(secretNames, utils.CreatePrefix(headers))
	}
	return secretNames
}
