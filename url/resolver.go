package url

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	gurl "net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Resolve(url string) (string, error) {
	resp, err := http.Head(url)
	return resp.Request.URL.String(), err
}

func GetTitle(url string) (string, error) {
	var title string
	res, err := http.Get(url)
	if err != nil {
		return title, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		println("status code error: %d %s", res.StatusCode, res.Status)
		return title, err
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		println(err.Error())
		return title, err
	}

	// Find the review items
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		title = s.Text()
	})

	return title, nil
}

type ArchiveResponse struct {
	Url     string `json:"url"`
	JobID   string `json:"job_id"`
	Message string `json:"message"`
}

// "https://web.archive.org/save"
func SubmitToArchive(url string, archiveURL string) (ArchiveResponse, error) {
	var archiveResponse ArchiveResponse
	client := &http.Client{}
	data := gurl.Values{}

	data.Set("url", url)
	req, _ := http.NewRequest("POST", archiveURL, strings.NewReader(data.Encode()))
	req.Header.Add("Authorization", "LOW "+os.Getenv("ARCHIVE_ACCESS_KEY")+":"+os.Getenv("ARCHIVE_SECRET_KEY"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if res.StatusCode != 200 {
		err = errors.New("There was a problem with the request. Status code: " + res.Status)
		return archiveResponse, err
	}
	if err != nil {
		fmt.Println(err.Error())
		return archiveResponse, err
	}

	defer res.Body.Close()
	buf, err := io.ReadAll(res.Body)
	err = json.Unmarshal(buf, &archiveResponse)

	if err != nil {
		fmt.Println(err.Error())
		return archiveResponse, err
	}

	return archiveResponse, err
}

type JobStatusResponse struct {
	OriginalURL string `json:"original_url"`
	JobID       string `json:"job_id"`
	Timestamp   string `json:"timestamp"`
	Status      string `json:"status"`
	Exception   string `json:"exception"`
	StatusExt   string `json:"status-ext"`
}

func CheckArchiveJobStatus(jobID string, checkStatusURL string) (archiveURL string, err error) {
	var jobStatusResponse JobStatusResponse
	client := &http.Client{}

	req, _ := http.NewRequest("GET", checkStatusURL, strings.NewReader(""))
	req.Header.Add("Authorization", "LOW "+os.Getenv("ARCHIVE_ACCESS_KEY")+":"+os.Getenv("ARCHIVE_SECRET_KEY"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if res.StatusCode != 200 {
		err = errors.New("There was a problem with the request. Status code: " + res.Status)
		return archiveURL, err
	}

	if err != nil {
		fmt.Println(err.Error())
		return archiveURL, err
	}

	defer res.Body.Close()
	buf, err := io.ReadAll(res.Body)
	err = json.Unmarshal(buf, &jobStatusResponse)

	if err != nil {
		fmt.Println(err.Error())
		return archiveURL, err
	}

	switch jobStatusResponse.Status {
	case "success":
		archiveURL = fmt.Sprintf("https://web.archive.org/web/%s/%s", jobStatusResponse.Timestamp, jobStatusResponse.OriginalURL)
	case "pending":
		return "", nil
	case "error":
		err = errors.New("There was a problem with the archiving job job. Exception: " + jobStatusResponse.Exception)
		return archiveURL, err
	}

	return archiveURL, nil
}
