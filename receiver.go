package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

const (
	primaryCl = iota
	secondaryCl
	secondaryClHi
)

var podColors = []*color.Color{
	color.New(color.FgBlue, color.Bold),
	color.New(color.FgMagenta, color.Bold),
	color.New(color.FgYellow, color.Bold),
	color.New(color.FgCyan, color.Bold),
}

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

type Receiver struct {
	pod               string
	container         string
	showPod           bool
	podColor          *color.Color
	significantFields [][]string
}

func MakeReceiver(pod, container string, index int, showPod bool, fields [][]string) *Receiver {
	return &Receiver{
		pod:               pod,
		container:         container,
		showPod:           showPod,
		podColor:          podColors[index%(len(podColors)-1)],
		significantFields: fields,
	}
}

func (r *Receiver) Receive(line []byte) {
	msg, err := ParseLine(line)
	if err != nil {
		if err != ErrNotJson {
			r.internalError(err)
		} else {
			r.printLine(line)
		}
		return
	}

	r.printMsg(msg)
}

func (r *Receiver) Close() {
	r.podColor.Printf("%s:%s ", r.pod, r.container)
	color.Red("closed\n")
	r.termLine()
}

// private

func (r *Receiver) internalError(err error) {
	fmt.Printf("-> %v\n", err)
}

func (r *Receiver) printMsg(msg Message) {
	primary := colors[msg.Severity][primaryCl]
	secondary := colors[msg.Severity][secondaryCl]
	secondaryHi := colors[msg.Severity][secondaryClHi]

	if r.showPod {
		r.podColor.Printf("%s:%s ", r.pod, r.container)
	}

	stime := msg.Timestamp.Local().String()
	primary.Printf("%s %s\n", stime, msg.Msg)

	for _, ll := range r.significantFields {
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
		if v != "" {
			secondaryHi.Printf("%s: ", k)
			secondary.Printf("%s ", v)
			newLine = true
		}
	}
	if newLine {
		fmt.Printf("\n")
	}

	r.termLine()
}

func (r *Receiver) printLine(line []byte) {
	if r.showPod {
		r.podColor.Printf("%s:%s ", r.pod, r.container)
	}
	fmt.Printf("%s\n", string(line))
	r.termLine()
}

func (r *Receiver) termLine() {
	fmt.Println("----------------------------------------")
}
