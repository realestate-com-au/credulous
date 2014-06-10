package main

import (
	"errors"
	"reflect"
	"strings"

	"github.com/realestate-com-au/goamz/aws"
	"github.com/realestate-com-au/goamz/iam"
)

type Instancer interface {
	GetUser(string) (*iam.GetUserResp, error)
	AccessKeys(string) (*iam.AccessKeysResp, error)
	ListAccountAliases() (*iam.AccountAliasesResp, error)
}

func getAWSUsernameAndAlias(cred Credential) (username, alias string, err error) {
	auth := aws.Auth{
		AccessKey: cred.KeyId,
		SecretKey: cred.SecretKey,
	}
	// Note: the region is irrelevant for IAM
	instance := iam.New(auth, aws.APSoutheast2)
	username, err = getAWSUsername(instance)
	if err != nil {
		return "", "", err
	}

	alias, err = getAWSAccountAlias(instance)
	if err != nil {
		return "", "", err
	}

	return username, alias, nil
}

func getAWSUsername(instance Instancer) (string, error) {
	response, err := instance.GetUser("")
	if err != nil {
		return "", err
	}
	return response.User.Name, nil
}

func getKeyCreateDate(instance Instancer) (string, error) {
	response, err := instance.AccessKeys("")
	panic_the_err(err)
	// This mess is because iam.IAM and TestIamInstance are structs
	elem := reflect.ValueOf(instance).Elem()
	auth := elem.FieldByName("Auth")
	accessKey := auth.FieldByName("AccessKey").String()
	for _, key := range response.AccessKeys {
		if key.Id == accessKey {
			return key.CreateDate, nil
		}
	}
	return "", errors.New("Couldn't find this key")
}

func getAWSAccountAlias(instance Instancer) (string, error) {
	response, err := instance.ListAccountAliases()
	if err != nil {
		return "", err
	}
	// There really is only one alias
	if len(response.Aliases) == 0 {
		// we have to do a getuser instead and parse out the
		// account ID from the ARN
		response, err := instance.GetUser("")
		if err != nil {
			return "", err
		}
		id := strings.Split(response.User.Arn, ":")
		return id[4], nil
	}
	return response.Aliases[0], nil
}

func verify_account(alias string, instance Instancer) error {
	acct_alias, err := getAWSAccountAlias(instance)
	if err != nil {
		return err
	}
	if acct_alias == alias {
		return nil
	}
	err = errors.New("Cannot verify account: does not match alias " + alias)
	return err
}

func verify_user(username string, instance Instancer) error {
	response, err := instance.AccessKeys(username)
	if err != nil {
		return err
	}
	// This mess is because iam.IAM and TestIamInstance are structs
	elem := reflect.ValueOf(instance).Elem()
	auth := elem.FieldByName("Auth")
	accessKey := auth.FieldByName("AccessKey").String()
	for _, key := range response.AccessKeys {
		if key.Id == accessKey {
			return nil
		}
	}
	err = errors.New("Cannot verify user: access keys are not for user " + username)
	return err
}
