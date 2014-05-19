package main

import (
	"errors"
	"testing"

	"github.com/realestate-com-au/goamz/aws"
	"github.com/realestate-com-au/goamz/iam"
	. "github.com/smartystreets/goconvey/convey"
)

type TestIamInstance struct {
	Auth               aws.Auth
	getUserResp        iam.GetUserResp
	accessKeysResp     iam.AccessKeysResp
	accountAliasesResp iam.AccountAliasesResp
}

func (t *TestIamInstance) GetUser(username string) (*iam.GetUserResp, error) {
	if t.getUserResp.User.Name == "" {
		return &iam.GetUserResp{}, errors.New("No user of that name")
	}
	return &t.getUserResp, nil
}

func (t *TestIamInstance) AccessKeys(username string) (*iam.AccessKeysResp, error) {
	if len(t.accessKeysResp.AccessKeys) == 0 {
		return &iam.AccessKeysResp{}, errors.New("No keys for that user")
	}
	return &t.accessKeysResp, nil
}

func (t *TestIamInstance) ListAccountAliases() (*iam.AccountAliasesResp, error) {
	return &t.accountAliasesResp, nil
}

func TestGetAWSUsername(t *testing.T) {
	Convey("Test getAWSUsername", t, func() {
		tstInst := TestIamInstance{
			getUserResp: iam.GetUserResp{
				RequestId: "abc123",
				User: iam.User{
					Arn:  "some-arn",
					Path: "/",
					Id:   "userid",
					Name: "foonly",
				},
			},
		}
		resp, _ := getAWSUsername(&tstInst)
		So(resp, ShouldEqual, "foonly")
	})
}

func TestGetKeyCreateDate(t *testing.T) {
	Convey("Test getKeyCreateDate", t, func() {
		tstKey := []iam.AccessKey{}
		tstKey = append(tstKey, iam.AccessKey{
			UserName:   "bob",
			Id:         "AKIAtest",
			Secret:     "sooper-seekrit",
			Status:     "happy",
			CreateDate: "1970-01-01T00:00:00:00Z",
		})
		tstInst := TestIamInstance{
			accessKeysResp: iam.AccessKeysResp{
				RequestId:  "abc123",
				AccessKeys: tstKey,
			},
			Auth: aws.Auth{
				AccessKey: "AKIAtest",
				SecretKey: "sooper-seekrit",
			},
		}
		date, err := getKeyCreateDate(&tstInst)
		So(date, ShouldEqual, tstKey[0].CreateDate)
		So(err, ShouldEqual, nil)
	})
}

func TestListAccountAliases(t *testing.T) {
	Convey("Test listAccountAliases", t, func() {
		Convey("when the alias matches", func() {
			tstResp := iam.AccountAliasesResp{
				RequestId: "abc123",
				Aliases:   []string{"test-alias"},
			}
			tstInst := TestIamInstance{accountAliasesResp: tstResp}
			alias, _ := getAWSAccountAlias(&tstInst)
			So(alias, ShouldEqual, "test-alias")
		})
		Convey("when there is no account alias", func() {
			tstResp := iam.AccountAliasesResp{}
			tstInst := TestIamInstance{
				accountAliasesResp: tstResp,
				getUserResp: iam.GetUserResp{
					RequestId: "abc123",
					User: iam.User{
						Arn:  "arn:aws:iam::123456789012:user/foonly",
						Path: "/",
						Id:   "userid",
						Name: "foonly",
					},
				},
			}
			alias, err := getAWSAccountAlias(&tstInst)
			So(err, ShouldEqual, nil)
			So(alias, ShouldNotEqual, "")
		})
		Convey("when there is somehow more than one alias", func() {
			tstResp := iam.AccountAliasesResp{
				RequestId: "abc123",
				Aliases:   []string{"test-alias", "second-alias"},
			}
			tstInst := TestIamInstance{accountAliasesResp: tstResp}
			alias, _ := getAWSAccountAlias(&tstInst)
			So(alias, ShouldEqual, "test-alias")
		})
	})
}

func TestVerifyAccount(t *testing.T) {
	Convey("Test verifyAccount", t, func() {
		Convey("when there is no account alias but we think there is", func() {
			tstResp := iam.AccountAliasesResp{}
			tstInst := TestIamInstance{
				accountAliasesResp: tstResp,
				getUserResp: iam.GetUserResp{
					RequestId: "abc123",
					User: iam.User{
						Arn:  "arn:aws:iam::123456789012:user/foonly",
						Path: "/",
						Id:   "userid",
						Name: "foonly",
					},
				},
			}
			err := verify_account("test-alias", &tstInst)
			So(err, ShouldNotEqual, nil)
			So(err.Error(), ShouldEqual, "Cannot verify account: does not match alias test-alias")
		})
		Convey("when returned alias matches expected alias", func() {
			tstResp := iam.AccountAliasesResp{
				RequestId: "abc123",
				Aliases:   []string{"test-alias"},
			}
			tstInst := TestIamInstance{accountAliasesResp: tstResp}
			err := verify_account("test-alias", &tstInst)
			So(err, ShouldEqual, nil)
		})
		Convey("when returned alias does not match expected alias", func() {
			tstResp := iam.AccountAliasesResp{
				RequestId: "abc123",
				Aliases:   []string{"test-alias"},
			}
			tstInst := TestIamInstance{accountAliasesResp: tstResp}
			err := verify_account("nomatch-alias", &tstInst)
			So(err.Error(), ShouldEqual, "Cannot verify account: does not match alias nomatch-alias")
		})
		Convey("when there is no alias and we expect that", func() {
			tstResp := iam.AccountAliasesResp{}
			tstInst := TestIamInstance{
				accountAliasesResp: tstResp,
				getUserResp: iam.GetUserResp{
					RequestId: "abc123",
					User: iam.User{
						Arn:  "arn:aws:iam::123456789012:user/foonly",
						Path: "/",
						Id:   "userid",
						Name: "foonly",
					},
				},
			}
			err := verify_account("123456789012", &tstInst)
			So(err, ShouldEqual, nil)
		})

	})
}

func TestVerifyUser(t *testing.T) {
	Convey("Test verify_user", t, func() {
		Convey("when username not found", func() {
			tstInst := TestIamInstance{
				Auth: aws.Auth{
					AccessKey: "AKIAtest",
					SecretKey: "sooper-seekrit",
				},
			}
			err := verify_user("bob", &tstInst)
			So(err, ShouldNotEqual, nil)
		})
		Convey("when returned user matches expected user", func() {
			tstKey := []iam.AccessKey{}
			// this is the set of responses
			tstKey = append(tstKey, iam.AccessKey{
				UserName:   "bob",
				Id:         "AKIAtest",
				Secret:     "sooper-seekrit",
				Status:     "happy",
				CreateDate: "1970-01-01T00:00:00:00Z",
			})
			tstInst := TestIamInstance{
				accessKeysResp: iam.AccessKeysResp{
					RequestId:  "abc123",
					AccessKeys: tstKey,
				},
				Auth: aws.Auth{
					AccessKey: "AKIAtest",
					SecretKey: "sooper-seekrit",
				},
			}
			err := verify_user("bob", &tstInst)
			So(err, ShouldEqual, nil)
		})
		Convey("when returned user does not match expected user", func() {
			tstKey := []iam.AccessKey{}
			// this is the set of responses
			tstKey = append(tstKey, iam.AccessKey{
				UserName:   "fred",
				Id:         "AKIAcheese",
				Secret:     "notso-seekrit",
				Status:     "indifferent",
				CreateDate: "1970-01-01T00:00:00:00Z",
			})
			tstInst := TestIamInstance{
				accessKeysResp: iam.AccessKeysResp{
					RequestId:  "abc123",
					AccessKeys: tstKey,
				},
				Auth: aws.Auth{
					AccessKey: "AKIAtest",
					SecretKey: "sooper-seekrit",
				},
			}
			err := verify_user("bob", &tstInst)
			So(err, ShouldNotEqual, nil)
			So(err.Error(), ShouldEqual, "Cannot verify user: access keys are not for user bob")
		})

	})
}
