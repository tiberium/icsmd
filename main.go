package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	ics "github.com/arran4/golang-ical"
	"github.com/mandolyte/mdtopdf"
	"html/template"
)

func safeHTML(s string) template.HTML {
	return template.HTML(s)
}

type Event struct {
	Summary     string
	Start       string
	Description string
	End         string
}

func ConvertEvent(e *ics.VEvent) (*Event, error) {
	summary := e.GetProperty("SUMMARY")
	if summary == nil {
		return nil, errors.New("Event does not have a summary")
	}

	start := e.GetProperty("DTSTART")
	if start == nil {
		return nil, errors.New("Event does not have a start time")
	}

	end := e.GetProperty("DTEND")
	if end == nil {
		return nil, errors.New("Event does not have an end time")
	}

	description := e.GetProperty("DESCRIPTION")
	desc := ""
	if description != nil {
		desc = description.Value
	}

	return &Event{
		Summary:     summary.Value,
		Start:       start.Value,
		End:         end.Value,
		Description: desc,
	}, nil
}

func handleErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	icsFileName := flag.String("ics-file", "ics.ics", "The calendar file to convert")
	outputFileName := flag.String("md-file", "ics.md", "The output file to write the markdown to")
	renderPDF := flag.String("pdf-file", "", "If set, will render the markdown to a PDF file with the given name")
	mkTemplateFileName := flag.String("md-template", "mk.tmpl", "The go http/template file to render the output markdown")
	flag.Parse()

	file, err := os.Open(*icsFileName)
	handleErr(err)

	cal, err := ics.ParseCalendar(file)
	handleErr(err)

	mkTemplate := template.New(*mkTemplateFileName).Funcs(template.FuncMap{"safeHTML": safeHTML})
	mkTemplate.ParseFiles(*mkTemplateFileName)

	events := []*Event{}
	for _, e := range cal.Events() {
		event, err := ConvertEvent(e)
		handleErr(err)

		events = append(events, event)
	}

	outputFile, err := os.Create(*outputFileName)
	handleErr(err)

	err = mkTemplate.Execute(outputFile, events)
	handleErr(err)
	outputFile.Close()

	if *renderPDF != "" {
		markdownFileContents, err := os.ReadFile(*outputFileName)
		handleErr(err)
		mdtopdf.NewPdfRenderer("", "", *renderPDF, "", nil, 1).Process(markdownFileContents)
	}
}
