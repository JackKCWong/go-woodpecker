package ossindex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/api"
	"io"
	"net/http"
)

type Sonatype struct {
	APIBaseURL   string
	APIBasicAuth string
	HttpClient   http.Client
}

type compoentReportsRequestParam struct {
	Coordinates []string `json:"coordinates"`
}

const requestContentType = "application/vnd.ossindex.component-report-request.v1+json"

func (s Sonatype) GetComponentReports(coordiantes []string) ([]api.ComponentReport, error) {
	reqParam := compoentReportsRequestParam{
		Coordinates: coordiantes,
	}

	reqBody, err := json.Marshal(&reqParam)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ossindex request params: %w", err)
	}

	req, err := http.NewRequest("POST", s.APIBaseURL+"/component-report", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request to ossindex: %v", err)
	}

	req.Header.Add("Content-Type", requestContentType)
	req.Header.Add("Accept", "application/vnd.ossindex.component-report.v1+json")
	req.Header.Add("Authorization", "Basic "+s.APIBasicAuth)

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to ossindex: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var componentReports []api.ComponentReport
	err = json.Unmarshal(body, &componentReports)
	if err != nil {
		return nil, fmt.Errorf("%w: %s - %s", err, resp.Status, string(body))
	}

	return componentReports, nil
}
