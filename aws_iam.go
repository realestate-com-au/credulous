package main

import (
	"errors"
	"log"

	"github.com/realestate-com-au/goamz/iam"
	"launchpad.net/goamz/aws"
)

type AwsCaller interface {
	GetAwsUsername(keyId, secret string) (string, error)
	GetKeyCreateDate(keyId, secret string) (string, error)
	GetAwsAccountAlias(KeyId, secret string) (string, error)
}

func getAWSUsername(key_id, secret string) (string, error) {
	auth := aws.Auth{key_id, secret}
	instance := iam.New(auth, aws.APSoutheast2)
	response, err := instance.GetUser("")
	panic_the_err(err)
	return response.User.Name, nil
}

func getKeyCreateDate(key_id, secret string) (string, error) {
	auth := aws.Auth{key_id, secret}
	instance := iam.New(auth, aws.APSoutheast2)
	response, err := instance.AccessKeys("")
	panic_the_err(err)
	for _, key := range response.AccessKeys {
		if key.Id == key_id {
			return key.CreateDate, nil
		}
	}
	return "", errors.New("Couldn't find this key")
}

func getAWSAccountAlias(key_id, secret string) (string, error) {
	auth := aws.Auth{key_id, secret}
	instance := iam.New(auth, aws.APSoutheast2)
	response, err := instance.ListAccountAliases()
	panic_the_err(err)
	// There really is only one alias
	return response.Aliases[0], nil
}

func verify_account(alias string, iam_instance *iam.IAM) (bool, error) {
	// TODO: the GetAccountAlias function needs to be implemented in goamz/iam
	response, err := iam_instance.ListAccountAliases()
	if err != nil {
		log.Fatal("Could not ListAccountAliases")
	}
	for _, acct_alias := range response.Aliases {
		if acct_alias == alias {
			return true, nil
		}
	}
	log.Print("Cannot verify account: does not match alias " + alias)
	return false, nil
}

func verify_user(username string, iam_instance *iam.IAM) (bool, error) {
	response, err := iam_instance.AccessKeys(username)
	if err != nil {
		log.Fatal("Cannot get AccessKeys")
	}
	for _, key := range response.AccessKeys {
		if key.Id == iam_instance.AccessKey {
			return true, nil
		}
	}
	log.Print("Cannot verify user: access keys are not for user " + username)
	return false, nil
}

func verifyUserAndAccount(creds Credential) (bool, error) {
	// need to check both the username and the account alias for the
	// supplied creds match the passed-in username and account alias
	auth := aws.Auth{creds.KeyId, creds.SecretKey}
	// Note: the region is irrelevant for IAM
	instance := iam.New(auth, aws.APSoutheast2)

	// Make sure the account is who we expect
	result, err := verify_account(creds.AccountAliasOrId, instance)
	if err != nil {
		log.Fatal(err)
	}

	if !result {
		log.Fatal("Could not verify account")
	}

	// Make sure the user is who we expect
	result, err = verify_user(creds.IamUsername, instance)
	if err != nil {
		log.Fatal(err)
	}

	if !result {
		log.Fatal("Could not verify user")
	}

	return true, nil
}
