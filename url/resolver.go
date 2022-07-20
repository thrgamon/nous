package url

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	gurl "net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/thrgamon/nous/links"
	"github.com/thrgamon/nous/logger"
	"mvdan.cc/xurls/v2"
)

func ExtractURLMetadata(body string) {
	ctx := context.Background()
	rxStrict := xurls.Strict()
	urls := rxStrict.FindAllString(body, -1)
	for _, v := range urls {
		go ProcessURL(ctx, v)
	}
}

func ProcessURL(ctx context.Context, url string) error {
	linkRepo := links.NewLinkRepo()
	exists, err := linkRepo.Exists(ctx, url)

	if err != nil {
		logger.Logger.Println(err.Error())
		return err
	}

	if exists {
		logger.Logger.Println("Url already exits: ", url)
		return nil
	}

	linkID, err := linkRepo.AddLink(ctx, url)
	if err != nil {
		logger.Logger.Println(err.Error())
		return err
	}

	resolvedURL, err := Resolve(url)
	if err != nil {
		logger.Logger.Println(err.Error())
		return err
	}
	err = linkRepo.EditLinkURL(ctx, linkID, resolvedURL)
	if err != nil {
		logger.Logger.Println(err.Error())
		return err
	}

	title, err := GetTitle(resolvedURL)
	if err != nil {
		logger.Logger.Println(err.Error())
	} else {
		err = linkRepo.EditLinkTitle(ctx, linkID, title)
		if err != nil {
			logger.Logger.Println(err.Error())
		}
	}

	archiveResponse, err := SubmitToArchive(resolvedURL, "https://web.archive.org/save")
	if err != nil {
		logger.Logger.Println(err.Error())
		return err
	}
	err = linkRepo.EditArchiveStatus(ctx, linkID, links.Pending, archiveResponse.JobID, "")
	if err != nil {
		logger.Logger.Println(err.Error())
		return err
	}

	return nil
}

func Resolve(url string) (string, error) {
	resp, err := http.Head(url)
	return resp.Request.URL.String(), err
}

func GetTitle(url string) (string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, strings.NewReader(""))
	req.Header.Set("User-Agent", "Nous/1.1")
	var title string
	res, err := client.Do(req)
	if err != nil {
		logger.Logger.Println(err.Error())
		return title, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		logger.Logger.Printf("status code error: %d %s\n", res.StatusCode, res.Status)
		return title, err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		logger.Logger.Println(err.Error())
		return title, err
	}

	title = doc.Find("title").Text()

	if title == "" {
		logger.Logger.Println("No title found for url: ", url)
	}

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
		logger.Logger.Println(err.Error())
		return archiveResponse, err
	}

	defer res.Body.Close()
	buf, err := io.ReadAll(res.Body)
	err = json.Unmarshal(buf, &archiveResponse)

	if err != nil {
		logger.Logger.Println(err.Error())
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
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if res.StatusCode != 200 {
		err = errors.New("There was a problem with the request. Status code: " + res.Status)
		return archiveURL, err
	}

	if err != nil {
		logger.Logger.Println(err.Error())
		return archiveURL, err
	}

	defer res.Body.Close()
	buf, err := io.ReadAll(res.Body)
	err = json.Unmarshal(buf, &jobStatusResponse)

	if err != nil {
		logger.Logger.Println(err.Error())
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
