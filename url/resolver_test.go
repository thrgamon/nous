package url

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolver(t *testing.T) {
	want := "https://www.youtube.com/watch?v=mrkAmmMakMg"

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		h := res.Header()
		h.Set("Location", want)
		res.WriteHeader(303)
		res.Write([]byte("body"))
	}))
	defer testServer.Close()

	url := testServer.URL
	resolvedUrl, err := Resolve(url)
	assert.NoError(t, err)
	assert.Equal(t, resolvedUrl, want)
}

func TestGetTitle(t *testing.T) {
	want := "The Title of the Page"
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		res.Write([]byte("<html><head><title>" + want + "</title></head><body></body></html>"))
	}))
	defer testServer.Close()

	title, err := GetTitle(testServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, title, want)
}

func TestSubmitArchive(t *testing.T) {
	urlToArchive := "http://url-to-archive.com"
	desiredJobID := "ac58789b-f3ca-48d0-9ea6-1d1225e98695"
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		h := res.Header()
		h.Set("Content-Type", "application/json")
		res.WriteHeader(200)
		res.Write([]byte(fmt.Sprintf(`{"url":"%s", "job_id":"%s"}`, urlToArchive, desiredJobID)))
	}))
	defer testServer.Close()
	archiveResponse, err := SubmitToArchive(urlToArchive, testServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, archiveResponse.JobID, desiredJobID)
	assert.Equal(t, archiveResponse.Url, urlToArchive)
}

func TestCheckArchiveJobStatus(t *testing.T) {
	urlToArchive := "http://url-to-archive.com"
	timestamp := "20190102005040"
	archiveURL := fmt.Sprintf("https://web.archive.org/web/%s/%s", timestamp, urlToArchive)
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		h := res.Header()
		h.Set("Content-Type", "application/json")
		res.WriteHeader(200)
		res.Write([]byte(fmt.Sprintf(`{"status":"%s", "timestamp":"%s", "original_url":"%s"}`, "success", timestamp, urlToArchive)))
	}))
	defer testServer.Close()
	result, err := CheckArchiveJobStatus(urlToArchive, testServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, result, archiveURL)
}
