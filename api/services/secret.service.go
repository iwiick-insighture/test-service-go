package services

import (
	"context"
	"encoding/json"
	"fmt"

	"secret-svc/api/dtos"
	"secret-svc/pkg/constants"
	"secret-svc/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.uber.org/zap"
)

// Retrieves the cross account Shared/Private Secret Manager using assume roles
// ////////////////////////////////////////////////////////////////////
func getSecretManager(config aws.Config, arn, region string) (secretsmanager.Client, error) {
	stsClient := sts.NewFromConfig(config)
	assumedRoleObject, err := stsClient.AssumeRole(context.TODO(), &sts.AssumeRoleInput{
		RoleArn:         &arn,
		RoleSessionName: aws.String("ASSUME_ROLE_SESSION_NAME"),
	})

	if err != nil {
		zap.L().Panic("Failed to get Secret Manager Instance :: " + err.Error())
		return secretsmanager.Client{}, err
	}

	credentails := assumedRoleObject.Credentials
	credentialsProvider := credentials.NewStaticCredentialsProvider(
		*credentails.AccessKeyId,
		*credentails.SecretAccessKey,
		*credentails.SessionToken,
	)

	secretsManagerClient := secretsmanager.NewFromConfig(config, func(o *secretsmanager.Options) {
		o.Credentials = credentialsProvider
		o.Region = region
	})

	return *secretsManagerClient, nil
}

// Helper function for getting secret inputs
// //////////////////////////////////////////////
func getSecretInput(secretName string, versionId string) secretsmanager.GetSecretValueInput {
	if versionId == "" {
		return secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(secretName),
			VersionStage: aws.String("AWSCURRENT"),
		}
	}

	return secretsmanager.GetSecretValueInput{
		SecretId:  aws.String(secretName),
		VersionId: &versionId,
	}
}

// Retreives a secret from the Shared/Private Secret Manager by giving uuid
// ////////////////////////////////////////////////////////////////////////
func GetSecret(headers dtos.CustomHeaders, id string, version string) (string, error) {
	secretName := utils.CreatePrefix(headers)
	if headers.Flow == constants.PRIVATE_FLOW {
		secretName = id
	}
	zap.L().Info("Getting Secret :: " + secretName)

	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(headers.Region))
	svc, _ := getSecretManager(config, headers.ARN, headers.Region)
	input := getSecretInput(secretName, version)
	getSecretValueResponse, err := svc.GetSecretValue(context.TODO(), &input)

	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to GetSecretValue :: %s :: ", secretName) + err.Error())
		return "", err
	}

	secretString := *getSecretValueResponse.SecretString
	var secretData map[string]interface{}

	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		zap.L().Error("json unmarshalling failed :: " + err.Error())
		return "", err
	}

	if data, dataExists := secretData[id]; dataExists {
		jsonData, err := json.Marshal(data)
		if err != nil {
			zap.L().Error("Failed to marshal secret data to JSON string: " + err.Error())
			return "", err
		}
		encodedSecret := utils.Base64Encode(string(jsonData))
		return encodedSecret, nil
	}

	return "", constants.ErrKeyNotFound
}

// Retrieves the secret versions for a secret in the Shared/Private Secret Manager
// ////////////////////////////////////////////////////////////////////////////
func GetSecretVersions(headers dtos.CustomHeaders, id string) ([]types.SecretVersionsListEntry, error) {
	secretName := utils.CreatePrefix(headers)
	if headers.Flow == constants.PRIVATE_FLOW {
		secretName = id
	}
	zap.L().Info("Getting Secret Versions :: " + secretName)

	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(headers.Region))
	svc, _ := getSecretManager(config, headers.ARN, headers.Region)
	input := &secretsmanager.ListSecretVersionIdsInput{
		SecretId: &secretName,
	}
	result, err := svc.ListSecretVersionIds(context.TODO(), input)

	if err != nil {
		zap.L().Error("ListSecretVersionIds failed :: " + err.Error())
		return []types.SecretVersionsListEntry{}, err
	}

	return result.Versions, nil
}

// Create a new secret in the Shared/Private Secret Manager
// ///////////////////////////////////////////////////////////
func CreateSecret(headers dtos.CustomHeaders, secret string) (string, error) {
	uuid := utils.GetPrefixedUuid()
	secretName := utils.CreatePrefix(headers)
	secretDescription := fmt.Sprintf("Organization ID: %s", headers.OrgId)
	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(headers.Region))
	svc, _ := getSecretManager(config, headers.ARN, headers.Region)
	input := getSecretInput(secretName, "")

	// creating secrets for the PRIVATE flow
	//--------------------------------------------------------------------------------------------
	if headers.Flow == constants.PRIVATE_FLOW {
		secretName = uuid
		zap.L().Info("Creating Secret :: " + secretName)
		updatedSecretString, err := json.Marshal(secret)
		if err != nil {
			zap.L().Error("Marshalling secret: " + err.Error())
			return "", err
		}
		input := &secretsmanager.CreateSecretInput{
			Name:         &secretName,
			Description:  &secretDescription,
			SecretString: aws.String(string(updatedSecretString)),
		}

		_, err = svc.CreateSecret(context.TODO(), input)

		if err != nil {
			zap.L().Error("Creating Secret Failed :: " + err.Error())
			return "", err
		}
		return uuid, nil
	}

	// creating secrets for the SHARED flow
	//--------------------------------------------------------------------------------------------
	zap.L().Info("Creating Secret :: " + secretName)
	_, err := svc.GetSecretValue(context.TODO(), &input)
	if err != nil {
		input := &secretsmanager.CreateSecretInput{
			Name:         &secretName,
			Description:  &secretDescription,
			SecretString: aws.String(fmt.Sprintf(`{"%s": {}}`, uuid)),
		}

		_, err := svc.CreateSecret(context.TODO(), input)

		if err != nil {
			zap.L().Error("Creating Secret Failed :: " + err.Error())
			return "", err
		}
	}

	getSecretInput := getSecretInput(secretName, "")
	getSecretValueResponse, err := svc.GetSecretValue(context.TODO(), &getSecretInput)

	if err != nil {
		zap.L().Error("Failed to GetSecretValue :: " + err.Error())
		return "", err
	}

	secretString := *getSecretValueResponse.SecretString
	var secretData map[string]interface{}

	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		zap.L().Error("Unmarshalling json: " + err.Error())
		return "", err
	}

	secretData[uuid] = secret

	// Marshalling the updated secretData
	updatedSecretString, err := json.Marshal(secretData)
	if err != nil {
		zap.L().Error("Marshalling json: " + err.Error())
		return "", err
	}

	putSecretInput := &secretsmanager.PutSecretValueInput{
		SecretId:     &secretName,
		SecretString: aws.String(string(updatedSecretString)),
	}

	_, err = svc.PutSecretValue(context.TODO(), putSecretInput)

	if err != nil {
		zap.L().Error("PutSecretValue failed :: " + err.Error())
		return "", err
	}

	return uuid, nil
}

// Updates a secret in the Shared/Private Secret Manager
// ///////////////////////////////////////////////////////
func UpdateSecret(headers dtos.CustomHeaders, id string, secret string) (string, error) {
	secretName := utils.CreatePrefix(headers)
	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(headers.Region))
	svc, _ := getSecretManager(config, headers.ARN, headers.Region)

	// Updating secrets in the PRIVATE Flow
	//----------------------------------------------------------------------------------------------
	if headers.Flow == constants.PRIVATE_FLOW {
		secretName = id
		zap.L().Info("Updating Secret :: " + secretName)
		updatedSecretString, err := json.Marshal(secret)
		if err != nil {
			zap.L().Error("Marshalling secret failed :: " + err.Error())
			return "", err
		}

		updateInput := &secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(secretName),
			SecretString: aws.String(string(updatedSecretString)),
		}
		_, err = svc.UpdateSecret(context.TODO(), updateInput)

		if err != nil {
			zap.L().Error("UpdateSecret failed :: " + err.Error())
			return "", err
		}
		return id, nil

	}

	// Updating secrets in the SHARED Flow
	//----------------------------------------------------------------------------------------------
	zap.L().Info("Updating Secret :: " + secretName)
	input := getSecretInput(secretName, "")
	getSecretValueResponse, err := svc.GetSecretValue(context.TODO(), &input)

	if err != nil {
		zap.L().Error("Failed to GetSecretValue :: " + err.Error())
		return "", err
	}

	secretString := *getSecretValueResponse.SecretString

	// Parse the JSON content of the secret
	var secretData map[string]interface{}
	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		zap.L().Error("Unmarshalling json failed :: " + err.Error())
		return "", err
	}

	// Check if the specified ID (UUID) exists under the organization key
	if _, idExists := secretData[id]; idExists {
		// Update the secret value associated with the specified ID
		secretData[id] = secret
		// Serialize the updated secret data back to JSON
		updatedSecretString, err := json.Marshal(secretData)
		if err != nil {
			zap.L().Error("Marshalling json failed :: " + err.Error())
			return "", err
		}

		// Update the secret in AWS Secrets Manager with the modified data
		updateInput := &secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(secretName),
			SecretString: aws.String(string(updatedSecretString)),
		}
		_, err = svc.UpdateSecret(context.TODO(), updateInput)

		if err != nil {
			zap.L().Error("UpdateSecret failed :: " + err.Error())
			return "", err
		}
		return id, nil
	}

	return "", constants.ErrKeyNotFound
}

// Deletes a secret in the Shared/Private Secret Manager by giving UUID
// /////////////////////////////////////////////////////////////////////
func DeleteSecret(headers dtos.CustomHeaders, id string) (string, error) {
	secretName := utils.CreatePrefix(headers)
	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(headers.Region))
	svc, _ := getSecretManager(config, headers.ARN, headers.Region)
	deleteAsap := true // Bypasses the recovery window

	// Deleting secrets in the PRIVATE Flow
	//----------------------------------------------------------------------------------------------
	if headers.Flow == constants.PRIVATE_FLOW {
		secretName = id
		zap.L().Info("Deleting Secret :: " + secretName)
		// Deleting the Secret
		input := &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secretName),
			ForceDeleteWithoutRecovery: &deleteAsap,
		}

		_, err := svc.DeleteSecret(context.TODO(), input)
		if err != nil {
			zap.L().Error("DeleteSecret failed :: " + err.Error())
			return "", err
		}
		return id, nil
	}

	// Deleting secrets in the SHARED Flow
	//----------------------------------------------------------------------------------------------
	zap.L().Info("Deleting Secret :: " + secretName)
	input := getSecretInput(secretName, "")
	getSecretValueResponse, err := svc.GetSecretValue(context.TODO(), &input)

	if err != nil {
		zap.L().Error("Failed to GetSecretValue :: " + err.Error())
		return "", err
	}

	secretString := *getSecretValueResponse.SecretString
	var secretData map[string]interface{}

	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		zap.L().Error("Unmarshalling json failed :: " + err.Error())
		return "", err
	}

	// Check if the specified ID exists under the organization
	if _, idExists := secretData[id]; idExists {
		// Delete the ID data
		delete(secretData, id)
		// Encode the updated data to JSON
		updatedSecretString, err := json.Marshal(secretData)

		if err != nil {
			zap.L().Error("Marshalling json failed :: " + err.Error())
			return "", err
		}

		// Update the secret with the updated data
		putSecretInput := &secretsmanager.PutSecretValueInput{
			SecretId:     aws.String(secretName),
			SecretString: aws.String(string(updatedSecretString)),
		}
		_, err = svc.PutSecretValue(context.TODO(), putSecretInput)

		if err != nil {
			zap.L().Error("PutSecretValue failed :: " + err.Error())
			return "", err
		}
		return id, nil
	}
	return "", constants.ErrKeyNotFound
}

// Deletes a Shared/Private secret Group along with it's individual secrets
// /////////////////////////////////////////////////////////////////////////////
func DeleteSecretGroup(headers dtos.CustomHeaders, arn, region string) (string, error) {
	secretName := utils.CreatePrefix(headers)
	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	svc, _ := getSecretManager(config, arn, region)
	deleteAsap := true // Bypasses the recovery window
	zap.L().Info("Deleting Secret Group :: " + secretName)

	// Deleting the Group
	input := &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretName),
		ForceDeleteWithoutRecovery: &deleteAsap,
	}

	_, err := svc.DeleteSecret(context.TODO(), input)

	if err != nil {
		zap.L().Error("DeleteSecret failed :: " + err.Error())
		return "", err
	}

	return secretName, nil
}

// Migrating secrets from Shared acc to Pvt acc
// /////////////////////////////////////////////////
func MigrateSecretsSharedToPvt(headers dtos.CustomHeaders, secretName string, prevMetaData map[string]interface{}, newArn string, newRegion string) ([]string, error) {
	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(newRegion))
	secretDescription := fmt.Sprintf("Organization ID: %s", headers.OrgId)
	svc, _ := getSecretManager(config, newArn, newRegion)
	var keys []string

	// Get Prev Account Data
	secretData, err := getPrevAccountSecrets(secretName, prevMetaData)
	if err != nil {
		zap.L().Info(fmt.Sprintf("%s :: Secrets Doesn't Exist", secretName) + err.Error())
		return keys, nil
	}

	// Insert them to the PRIVATE account
	for key, value := range secretData {
		valueStringyfied, _ := utils.StringifyJson(value)
		input := &secretsmanager.CreateSecretInput{
			Name:         &key,
			Description:  &secretDescription,
			SecretString: aws.String(valueStringyfied),
		}

		_, err = svc.CreateSecret(context.TODO(), input)
		if err != nil {
			zap.L().Error("CreateSecret Failed" + err.Error())
			return nil, err
		}
		keys = append(keys, key)
	}

	// Delete Prev Account Data
	_, err = deletePrevAccountSecrets(secretName, prevMetaData)
	if err != nil {
		zap.L().Error("DeleteSecretGroup Failed" + err.Error())
		return keys, err
	}

	return keys, nil
}

func getPrevAccountSecrets(secretName string, prevMetaData map[string]interface{}) (map[string]interface{}, error) {
	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(prevMetaData[constants.REGION_META_DATA].(string)))
	svc, _ := getSecretManager(config, prevMetaData[constants.ARN_META_DATA].(string), prevMetaData[constants.REGION_META_DATA].(string))
	input := getSecretInput(secretName, "")
	getSecretValueResponse, err := svc.GetSecretValue(context.TODO(), &input)
	zap.L().Info("Getting Previous Account Secrets :: " + secretName)

	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to GetSecretValue :: %s :: ", secretName) + err.Error())
		return nil, err
	}

	secretString := *getSecretValueResponse.SecretString
	var secretData map[string]interface{}

	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		zap.L().Error("json unmarshalling failed :: " + err.Error())
		return nil, err
	}

	return secretData, err
}

func deletePrevAccountSecrets(secretName string, prevMetaData map[string]interface{}) (string, error) {
	config, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(prevMetaData[constants.REGION_META_DATA].(string)))
	svc, _ := getSecretManager(config, prevMetaData[constants.ARN_META_DATA].(string), prevMetaData[constants.REGION_META_DATA].(string))
	deleteAsap := true
	zap.L().Info("Deleting Previous Account Secrets :: " + secretName)

	input := &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretName),
		ForceDeleteWithoutRecovery: &deleteAsap,
	}

	_, err := svc.DeleteSecret(context.TODO(), input)
	if err != nil {
		zap.L().Error(fmt.Sprintf("DeleteSecret failed :: %s :: ", secretName) + err.Error())
		return "", err
	}

	return secretName, nil
}
