package cbr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
)

type Client struct {
	url        string
	margin     float64
	httpClient *http.Client
}

func NewClient(url string, margin float64) *Client {
	return &Client{
		url:    url,
		margin: margin,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetCentralBankRate() (float64, error) {
	soapRequest := buildSOAPRequest()

	rawBody, err := c.sendRequest(soapRequest)
	if err != nil {
		return 0, err
	}

	keyRate, err := parseXMLResponse(rawBody)
	if err != nil {
		return 0, err
	}

	return keyRate + c.margin, nil
}

func buildSOAPRequest() string {
	fromDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	toDate := time.Now().Format("2006-01-02")

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
    <soap12:Body>
        <KeyRate xmlns="http://web.cbr.ru/">
            <fromDate>%s</fromDate>
            <ToDate>%s</ToDate>
        </KeyRate>
    </soap12:Body>
</soap12:Envelope>`, fromDate, toDate)
}

func (c *Client) sendRequest(soapRequest string) ([]byte, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		c.url,
		bytes.NewBuffer([]byte(soapRequest)),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cbr request error: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cbr response read error: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cbr bad status: %d", resp.StatusCode)
	}

	return rawBody, nil
}

func parseXMLResponse(rawBody []byte) (float64, error) {
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes(rawBody); err != nil {
		return 0, fmt.Errorf("xml parse error: %w", err)
	}

	krElements := doc.FindElements("//KR")
	if len(krElements) == 0 {
		krElements = doc.FindElements("//diffgram/KeyRate/KR")
	}

	if len(krElements) == 0 {
		return 0, errors.New("key rate data not found")
	}

	latestKR := krElements[len(krElements)-1]

	rateElement := latestKR.FindElement("./Rate")
	if rateElement == nil {
		return 0, errors.New("rate tag not found")
	}

	rateStr := strings.TrimSpace(rateElement.Text())
	rateStr = strings.ReplaceAll(rateStr, ",", ".")

	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return 0, fmt.Errorf("rate conversion error: %w", err)
	}

	return rate, nil
}
