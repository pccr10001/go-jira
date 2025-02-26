package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Error message from Jira
// See https://docs.atlassian.com/jira/REST/cloud/#error-responses
type Error struct {
	HTTPError     error
	ErrorMessages []string          `json:"errorMessages"`
	Errors        map[string]string `json:"errors"`
}

// NewJiraError creates a new jira Error
func NewJiraError(resp *Response, httpError error) error {
	if resp == nil {
		return fmt.Errorf("no response returned: %w", httpError)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: %w", httpError.Error(), err)
	}
	jerr := Error{HTTPError: httpError}
	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		err = json.Unmarshal(body, &jerr)
		if err != nil {
			return fmt.Errorf("%s: could not parse JSON: %w", httpError.Error(), err)
		}
	} else {
		if httpError == nil {
			return fmt.Errorf("got response status %s:%s", resp.Status, string(body))
		}
		return fmt.Errorf("%s: %s: %w", resp.Status, string(body), httpError)
	}

	return &jerr
}

// Error is a short string representing the error
func (e *Error) Error() string {
	if len(e.ErrorMessages) > 0 {
		// return fmt.Sprintf("%v", e.HTTPError)
		return fmt.Sprintf("%s: %v", e.ErrorMessages[0], e.HTTPError)
	}
	if len(e.Errors) > 0 {
		for key, value := range e.Errors {
			return fmt.Sprintf("%s - %s: %v", key, value, e.HTTPError)
		}
	}
	return e.HTTPError.Error()
}

// LongError is a full representation of the error as a string
func (e *Error) LongError() string {
	var msg bytes.Buffer
	if e.HTTPError != nil {
		msg.WriteString("Original:\n")
		msg.WriteString(e.HTTPError.Error())
		msg.WriteString("\n")
	}
	if len(e.ErrorMessages) > 0 {
		msg.WriteString("Messages:\n")
		for _, v := range e.ErrorMessages {
			msg.WriteString(" - ")
			msg.WriteString(v)
			msg.WriteString("\n")
		}
	}
	if len(e.Errors) > 0 {
		for key, value := range e.Errors {
			msg.WriteString(" - ")
			msg.WriteString(key)
			msg.WriteString(" - ")
			msg.WriteString(value)
			msg.WriteString("\n")
		}
	}
	return msg.String()
}
