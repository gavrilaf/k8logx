package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Severity int

const (
	SeverityInfo = iota
	SeverityWarning
	SeverityError
)

var ErrNotJson = fmt.Errorf("not-json")

type Message struct {
	Severity  Severity
	Msg       string
	Timestamp time.Time
	Data      map[string]string
}

type Parser struct {
	mapping map[string]string
	skip    map[string]struct{}
}

func MakeParser(mapping map[string]string) *Parser {
	skip := map[string]struct{}{}
	for k, v := range skipFields {
		skip[k] = v
	}

	mm := map[string]string{messageKey: messageKey, timestampKey: timestampKey, severityKey: severityKey}

	if m, ok := mapping[messageKey]; ok {
		mm[messageKey] = m
	}
	if m, ok := mapping[timestampKey]; ok {
		mm[timestampKey] = m
	}
	if m, ok := mapping[severityKey]; ok {
		mm[severityKey] = m
	}

	for _, v := range mm {
		skip[v] = struct{}{}
	}

	return &Parser{mapping: mm, skip: skip}
}

const (
	timestampKey = "timestamp"
	messageKey   = "message"
	severityKey  = "severity"
)

var skipFields = map[string]struct{}{
	"stacktrace":                            {},
	"logging.googleapis.com/labels":         {},
	"logging.googleapis.com/sourceLocation": {},
}

var dateLayout = "2006-01-02T15:04:05.99Z"

func (p *Parser) ParseLine(line []byte) (Message, error) {
	var rawMsg map[string]interface{}
	err := json.Unmarshal(line, &rawMsg)
	if err != nil {
		return Message{}, ErrNotJson
	}

	message, _ := rawMsg[p.mapping[messageKey]]
	severity, _ := rawMsg[p.mapping[severityKey]]
	timestamp, _ := rawMsg[p.mapping[timestampKey]]

	t, err := time.Parse(dateLayout, timestamp.(string))
	if err != nil {
		return Message{}, fmt.Errorf("failed to parse date, %w", err)
	}

	data := make(map[string]string)
	for k, v := range rawMsg {
		if _, ok := p.skip[k]; !ok {
			data[k] = fmt.Sprintf("%v", v)
		}
	}

	return Message{
		Severity:  getSeverity(severity.(string)),
		Msg:       message.(string),
		Timestamp: t,
		Data:      data,
	}, nil
}

func getSeverity(s string) Severity {
	switch strings.ToLower(s) {
	case "debug", "info":
		return SeverityInfo
	case "notice", "warning":
		return SeverityWarning
	case "error", "critical", "alert", "emergency", "fatal":
		return SeverityError
	}
	return SeverityInfo
}
