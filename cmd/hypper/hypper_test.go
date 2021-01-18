package main

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/time"
)

func testTimestamper() time.Time { return time.Unix(242085845, 0).UTC() }

func init() {
	action.Timestamper = testTimestamper
}
