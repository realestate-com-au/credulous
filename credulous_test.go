package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	Convey("Testing", t, func() {
		So(1, ShouldEqual, 1)
	})
}
