package ci

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

var (
	testPort = ":4045"
	testURL  = "http://localhost" + testPort + "/gh_hook"
)

func TestMain(m *testing.M) {
	go ListenAndServe(testPort)
	os.Exit(m.Run())
}

func TestPushEvent(t *testing.T) {
	err := sendEvent(eventPush)
	if err != nil {
		t.Error(err)
	}
}

func sendEvent(eventType string) error {
	filename := fmt.Sprintf("test/%s_event.json", eventType)
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cant open file '%s': %s", filename, err)
	}

	req, err := http.NewRequest("post", testURL, f)
	if err != nil {
		return fmt.Errorf("error while preparing request %s: %s", filename, err)
	}

	req.Header = map[string][]string{
		"Content-Type":   {"application/json"},
		"X-Github-Event": {"push"},
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error after sending %s: %s", filename, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wrong status after sending %s: got %d, expected %d",
			filename, resp.StatusCode, http.StatusOK)
	}

	return nil
}
