package common

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestGetFilterFields(t *testing.T) {
	fields := GetFilterFields()
	if fields == nil {
		t.Error("GetFilterFields should not return nil")
	}
	if len(fields) != 0 {
		t.Error("GetFilterFields should return an empty slice if FILTER_FIELDS is not set")
	}
	filterFields = nil
	os.Setenv("FILTER_FIELDS", "a,b,c")
	fields = GetFilterFields()
	if len(fields) != 3 {
		t.Error("expected GetFilterFields to return 3 strings")
	}
	if fields[0] != "a" {
		t.Error("wrong value in GetFilterFields result")
	}
	if fields[1] != "b" {
		t.Error("wrong value in GetFilterFields result")
	}
	if fields[2] != "c" {
		t.Error("wrong value in GetFilterFields result")
	}
}

func TestExtractPayload(t *testing.T) {
	line := `{"samplerate":1,"dataset":"travis.serverless-test","data":{"trace.span_id":"28320411-35c9-45dc-b968-b8b9d091f4bf","meta.local_hostname":"ip-10-12-92-115","trace.trace_id":"6614adb4-a74e-47a2-92de-06435f548eb1","name":"sleepytime","service_name":"travis.serverless-test","trace.parent_id":"56c10d68-e8db-46d5-97ad-0d508a3c08bf","duration_ms":500.65700000000004,"meta.beeline_version":"1.0.0"},"user_agent":"libhoney-py/1.4.0","time":"2018-07-23T22:06:58.471593Z"}`
	var data map[string]interface{}
	err := json.Unmarshal([]byte(line), &data)
	if err != nil {
		t.Error("didn't parse json: ", err)
	}
	payload, err := ExtractPayload(data)
	if err != nil {
		t.Error("extractPayload failed: ", err)
	}
	if payload.Dataset != "travis.serverless-test" {
		t.Error("unexpected value for dataset: ", payload.Dataset)
	}
	if payload.SampleRate != 1 {
		t.Error("unexpected value for sampleRate: ", payload.SampleRate)
	}
	expectedTime, err := time.Parse("2006-01-02T15:04:05.000000Z", "2018-07-23T22:06:58.471593Z")
	if err != nil {
		t.Error("error parsing time: ", err)
	}
	if expectedTime != payload.Time {
		t.Error("unexpected value for time: ", payload.Time)
	}
}

func TestGetMatchString(t *testing.T) {
	matchString := GetMatchString()
	if matchString != "" {
		t.Error("GetMatchString should return a blank string")
	}

	os.Setenv("HONEYCOMB_EVENT_MATCH_STRINGS", "libhoney")

	matchString = GetMatchString()
	if matchString != "libhoney" {
		t.Error("GetMatchString should return libhoney")
	}
}
