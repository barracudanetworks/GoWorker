package job

import "testing"

var (
	testJson = []byte(`
		{
			"name":"test",
			"command":"echo",
			"capture_output":true,
			"params": {
				"hello": "hello"
			},
			"type": "cli",
			"retries": 0	
		}
	`)
)

func TestParseConfig(t *testing.T) {
	j, err := ParseConfig(testJson)
	if err != nil {
		t.Error(err)
	}
	// make sure it has all the right fields
	if !j.CaptureOutput {
		t.Error("Capture output is not correct")
	}
	if j.Name != "test" {
		t.Error("Name is not correct")
	}
	if j.Type != "cli" {
		t.Error("Type did not parse correctly")
	}
	if j.Retries != 0 {
		t.Error("Retries did not parse correctly")
	}
}
