package main

import (
	"encoding/json"
	"fmt"
)

type Severity int

const (
	SeverityDebug = iota
	SeverityInfo
)

type Message struct {
	Pod       string
	Severity  Severity
	Msg       string
	Timestamp string
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
		fmt.Printf("\n***\n%s\n", string(line))
		return Message{}, fmt.Errorf("failed to parse line, %w", err)
	}

	var mf map[string]interface{}
	if len(msg.Context.Data) > 0 {
		err = json.Unmarshal(msg.Context.Data, &mf)
		if err != nil {
			fmt.Printf("\n***\n%s\n", string(line))
			return Message{}, fmt.Errorf("failed to parse message context, %w", err)
		}
	}

	data := make(map[string]string)
	for k, v := range mf {
		data[k] = fmt.Sprintf("%v", v)
	}

	return Message{
		Pod:       "",
		Severity:  0,
		Msg:       msg.Msg,
		Timestamp: msg.Timestamp,
		Data:      data,
	}, nil
}
