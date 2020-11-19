package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Severity int

const (
	SeverityInfo = iota
	SeverityWarning
	SeverityError
)

var dateStr = "2006-01-02T15:04:05Z"

var ErrNotJson = fmt.Errorf("not-json")

type Message struct {
	Severity  Severity
	Msg       string
	Timestamp time.Time
	Data      map[string]string
}

func ParseLine(line []byte) (Message, error) {
	type messageContext struct {
		Data json.RawMessage `json:"data"`
	}

	type internalMessage struct {
		Timestamp string         `json:"timestamp"`
		Msg       string         `json:"message"`
		Severity  string         `json:"severity"`
		Context   messageContext `json:"context"`
	}

	var msg internalMessage
	err := json.Unmarshal(line, &msg)
	if err != nil {
		return Message{}, ErrNotJson
	}

	var mf map[string]interface{}
	if len(msg.Context.Data) > 0 {
		err = json.Unmarshal(msg.Context.Data, &mf)
		if err != nil {
			return Message{}, fmt.Errorf("failed to parse message context, %w", err)
		}
	}

	t, err := time.Parse(dateStr, msg.Timestamp)
	if err != nil {
		return Message{}, fmt.Errorf("failed to parse date, %w", err)
	}

	data := make(map[string]string)
	for k, v := range mf {
		data[k] = fmt.Sprintf("%v", v)
	}

	return Message{
		Severity:  getSeverity(msg.Severity),
		Msg:       msg.Msg,
		Timestamp: t,
		Data:      data,
	}, nil
}

func getSeverity(s string) Severity {
	switch s {
	case "DEBUG", "INFO":
		return SeverityInfo
	case "NOTICE", "WARNING":
		return SeverityWarning
	case "ERROR", "CRITICAL", "ALERT", "EMERGENCY":
		return SeverityError
	}
	return SeverityInfo
}
