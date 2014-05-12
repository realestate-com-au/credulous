package main

import (
	"bytes"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFunctions(t *testing.T) {
	Convey("Retrieve public key", t, func() {

		Convey("returns a valid public key", func() {
			reader := bytes.NewReader([]byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDXg9Vmhy9YSB8BcN3yHgQjdX9lN3j2KRpv7kVDXSiIana2WbKP7IiTS0uJcJWUM3vlHjdL9KOO0jCWWzVFIcmLhiVVG+Fy2tothBp/NhjR8WWG/6Jg/6tXvVkLG6bDgfbDaLWdE5xzjL0YG8TrIluqnu0J5GHKrQcXF650PlqkGo+whpXrS8wOG+eUmsHX9L1w/Z3TkQlMjQNJEoRbqqSrp7yGj4JqzbtLpsglPRlobD7LHp+5ZDxzpk9i+6hoMxp2muDFxnEtZyED6IMQlNNEGkc3sdmGPOo26oW2+ePkBcjpOpdVif/Iya/jDLuLFHAOol6G34Tr4IdTgaL0qCCr foo@example.com"))
			key, err := RetrievePublicKey(reader)
			So(err, ShouldEqual, nil)
			So(key, ShouldNotEqual, nil)
		})

		Convey("returns error for garbage key", func() {
			reader := bytes.NewReader([]byte("not a key!"))
			_, err := RetrievePublicKey(reader)
			So(err, ShouldNotEqual, nil)
		})
	})
}
