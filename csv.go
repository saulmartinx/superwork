package main

import (
	"bytes"
	"encoding/csv"
	"net/http"
)

func exportCSV(w http.ResponseWriter, models []TimeEntry) {
	b := &bytes.Buffer{}
	writer := csv.NewWriter(b)
	writer.Write([]string{
		"Started at",
		"Finished at",
		"Duration",
		"Description",
	})
	const dateFormat = "2006-01-02 15:04:05"
	for i := len(models) - 1; i >= 0; i-- {
		model := models[i]
		if model.FinishedAt == nil {
			continue
		}
		writer.Write([]string{
			model.StartedAt.Format(dateFormat),
			model.FinishedAt.Format(dateFormat),
			model.DurationString(),
			model.Name,
		})
	}
	writer.Flush()

	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("Content-Disposition", "attachment; filename=time_entries.csv")
	w.Header().Set("Content-Type", "text/csv")
	w.Write(b.Bytes())
}
