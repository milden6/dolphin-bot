package whisper

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	httpClientGlobalTimeout = 60 * time.Second

	// timeout for the dial, the TLS handshake,
	// and reading the response header
	httpTimeout = 60 * time.Second
)

type Client struct {
	host       string
	httpClient *http.Client
}

func NewAPIClient(host string) *Client {
	c := &Client{
		host: host,
		httpClient: &http.Client{
			Timeout: httpClientGlobalTimeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: httpTimeout,
				}).DialContext,
				TLSHandshakeTimeout:   httpTimeout,
				ResponseHeaderTimeout: httpTimeout,
			},
		},
	}

	return c
}

type respFromAPI struct {
	Text string `json:"text"`
}

func (c *Client) DoInference(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			slog.Error("Failed to close output file", "error", err.Error())
		}
	}()

	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	part, err := multipartWriter.CreateFormFile("file", filePath)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, f)
	if err != nil {
		return "", err
	}

	err = multipartWriter.WriteField("temperature", "0.0")
	if err != nil {
		return "", err
	}

	err = multipartWriter.WriteField("temperature_inc", "0.2")
	if err != nil {
		return "", err
	}

	err = multipartWriter.WriteField("response_format", "json")
	if err != nil {
		return "", err
	}

	err = multipartWriter.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.host+"/inference", &requestBody)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())

	// TODO: add retry
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.Error("Failed to close body from whisper server", "error", err.Error())
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var r respFromAPI
	err = json.Unmarshal(body, &r)
	if err != nil {
		return "", err
	}

	return r.Text, nil
}
