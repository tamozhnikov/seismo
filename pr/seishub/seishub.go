package seishub

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var defClient = http.Client{Timeout: 60 * time.Second}

func GetMsgPages(ctx context.Context, url string) (map[string]string, error) {
	namesPage, err := GetMsgNamesPage(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("GetMsgPages: %v ", err)
	}

	names := parseMsgNames(namesPage)

	msgs := make(map[string]string, len(names))
	for _, n := range names {
		m, err := getMsgPage(ctx, url, n, nil)
		if err != nil {
			log.Printf("GetMsgPages: get message page: %v url: %s, name: %s", err, url, n)
		}
		msgs[n] = m
	}

	return msgs, nil
}

// GetMsgNamesPage retrieves raw html-page containting the link names of messages
func GetMsgNamesPage(ctx context.Context, url string, cl *http.Client) (string, error) {
	if cl == nil {
		cl = &defClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("GetMsgNamesPage: \"%s\": %w", url, err)
	}
	resp, err := cl.Do(req)
	if err != nil {
		return "", fmt.Errorf("GetMsgNamesPage: \"%s\": %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GetMsgNamesPage: \"%s\": %s", url, resp.Status)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", fmt.Errorf("GetMsgNamesPage: copy response body \"%s\": %w", url, err)
	}

	return buf.String(), nil
}

// MonthYearPathSeg returns a string like "2022-April"
func MonthYearPathSeg(m time.Month, y int) string {
	return fmt.Sprintf("%d-%s", y, m.String())
}

func parseMsgNames(s string) []string {
	re := regexp.MustCompile(`\d+\.html`)
	return re.FindAllString(s, -1)
}

func getMsgPage(ctx context.Context, dir string, name string, cl *http.Client) (string, error) {
	if cl == nil {
		cl = &defClient
	}

	url, err := url.JoinPath(dir, name)
	if err != nil {
		return "", fmt.Errorf("getMsgPage: dir arg %q, name erg %q: %w", dir, name, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("getMsgPage: url %q: %w", url, err)
	}

	resp, err := cl.Do(req)
	if err != nil {
		return "", fmt.Errorf("getMsgPage: get %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getMsgPage: get %q: %s", url, resp.Status)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", fmt.Errorf("getMsgPage: copy response body %q: %w", url, err)
	}

	return buf.String(), nil
}
