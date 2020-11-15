package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type Receiver struct {
}

func (r *Receiver) Receive(line []byte) {
	msg, err := ParseLine(line)
	if err != nil {
		if err != ErrNotJson {
			internalError(err)
		} else {
			printLine(line)
		}
		return
	}

	printMsg(msg)
}

func internalError(err error) {
	fmt.Printf("-> %v\n", err)
}

var significantFields = [][]string{
	{"method", "uri", "status", "latency"},
	{"sql"},
	{"args", "rowCount"},
}

const (
	primaryCl = iota
	secondaryCl
	secondaryClHi
)

var colors = map[Severity]map[int]*color.Color{
	SeverityInfo: {
		primaryCl:     color.New(color.FgHiGreen),
		secondaryCl:   color.New(color.FgGreen),
		secondaryClHi: color.New(color.FgHiGreen)},
	SeverityWarning: {
		primaryCl:     color.New(color.FgHiYellow),
		secondaryCl:   color.New(color.FgGreen),
		secondaryClHi: color.New(color.FgHiGreen)},
	SeverityError: {
		primaryCl:     color.New(color.FgHiRed),
		secondaryCl:   color.New(color.FgGreen),
		secondaryClHi: color.New(color.FgHiGreen)},
}

func printMsg(msg Message) {
	primary := colors[msg.Severity][primaryCl]
	secondary := colors[msg.Severity][secondaryCl]
	secondaryHi := colors[msg.Severity][secondaryClHi]

	primary.Printf("%s %s\n", msg.Timestamp, msg.Msg)

	for _, ll := range significantFields {
		if len(ll) == 1 {
			if v, ok := msg.Data[ll[0]]; ok {
				v = strings.TrimSpace(v)
				secondary.Printf("%s\n", v)

				delete(msg.Data, ll[0])
			}
		} else {
			newLine := false
			for _, k := range ll {
				if v, ok := msg.Data[k]; ok {
					v = strings.TrimSpace(v)
					secondaryHi.Printf("%s: ", k)
					secondary.Printf("%s ", v)
					newLine = true

					delete(msg.Data, k)
				}
			}
			if newLine {
				fmt.Printf("\n")
			}
		}
	}

	newLine := false
	for k, v := range msg.Data {
		v = strings.TrimSpace(v)
		secondaryHi.Printf("%s: ", k)
		secondary.Printf("%s ", v)
		newLine = true
	}
	if newLine {
		fmt.Printf("\n")
	}

	fmt.Println("----------------------------------------")
}

func printLine(line []byte) {
	fmt.Println(string(line))
	fmt.Println("----------------------------------------")
}
