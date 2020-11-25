package main

import (
	"fmt"
	"strings"

	au "github.com/logrusorgru/aurora"
)

const (
	primaryCl = iota
	secondaryCl
)

var podColors = []au.Color{
	au.BlueFg,
	au.MagentaFg,
	au.CyanFg,
	au.YellowFg,
}

var colors = map[Severity]map[int]au.Color{
	SeverityInfo: {
		primaryCl:   au.GreenFg,
		secondaryCl: au.GreenFg},
	SeverityWarning: {
		primaryCl:   au.YellowFg,
		secondaryCl: au.GreenFg},
	SeverityError: {
		primaryCl:   au.RedFg,
		secondaryCl: au.GreenFg},
}

type Receiver struct {
	pod               string
	container         string
	podColor          au.Color
	significantFields [][]string
	parser            *Parser
}

func MakeReceiver(pod, container string, index int, fields [][]string, parser *Parser) *Receiver {
	return &Receiver{
		pod:               pod,
		container:         container,
		podColor:          podColors[index%len(podColors)],
		significantFields: fields,
		parser:            parser,
	}
}

func (r *Receiver) Receive(line []byte) {
	msg, err := r.parser.ParseLine(line)
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
	fmt.Printf("%s %s\n", au.Colorize(fmt.Sprintf("%s:%s ", r.pod, r.container), r.podColor), au.Red("closed"))
	r.termLine()
}

// private

func (r *Receiver) internalError(err error) {
	fmt.Printf("-> %v\n", err)
}

const layout = "2006-01-02T15:04:05"

func (r *Receiver) printMsg(msg Message) {
	primary := colors[msg.Severity][primaryCl]
	secondary := colors[msg.Severity][secondaryCl]

	var sb strings.Builder

	sb.WriteString(au.Colorize(fmt.Sprintf("%s:%s ", r.pod, r.container), r.podColor).String())

	stime := msg.Timestamp.Local().Format(layout)
	sb.WriteString(au.Colorize(fmt.Sprintf("%s %s\n", stime, msg.Msg), primary).String())

	for _, ll := range r.significantFields {
		if len(ll) == 1 {
			if v, ok := msg.Data[ll[0]]; ok {
				v = strings.TrimSpace(v)
				if v != "" {
					sb.WriteString(v)
					sb.WriteByte('\n')
				}
				delete(msg.Data, ll[0])
			}
		} else {
			newLine := false
			for _, k := range ll {
				if v, ok := msg.Data[k]; ok {
					v = strings.TrimSpace(v)
					sb.WriteString(au.Colorize(k, secondary).String())
					sb.WriteByte(' ')
					sb.WriteString(v)
					sb.WriteByte(' ')

					newLine = true
					delete(msg.Data, k)
				}
			}
			if newLine {
				sb.WriteByte('\n')
			}
		}
	}

	newLine := false
	for k, v := range msg.Data {
		v = strings.TrimSpace(v)
		sb.WriteString(au.Colorize(k, secondary).String())
		sb.WriteByte(' ')
		sb.WriteString(v)
		sb.WriteByte(' ')

		newLine = true
	}

	if newLine {
		sb.WriteByte('\n')
	}

	fmt.Print(sb.String())

	r.termLine()
}

func (r *Receiver) printLine(line []byte) {
	fmt.Printf("%s ", au.Colorize(fmt.Sprintf("%s:%s ", r.pod, r.container), r.podColor))
	fmt.Printf("%s\n", string(line))
	r.termLine()
}

func (r *Receiver) termLine() {
	fmt.Println("----------------------------------------")
}
