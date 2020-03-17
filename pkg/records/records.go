package records

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// LogEntry matches the cloudwatch log entry structure
type LogEntry struct {
	ID        string `json:"id,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Message   string `json:"message,omitempty"`
}

// LogBatch matches the cloudwatch logs batch structure
type LogBatch struct {
	MessageType         string      `json:"messageType,omitempty"`
	Owner               string      `json:"owner,omitempty"`
	LogGroup            string      `json:"logGroup,omitempty"`
	LogStream           string      `json:"logStream,omitempty"`
	SubscriptionFilters []string    `json:"subscriptionFilters,omitempty"`
	LogEvents           []*LogEntry `json:"logEvents,omitempty"`
}

// LogMessage log message after decompression and parsing
type LogMessage struct {
	LogGroup  string // optional log group
	Message   string
	Timestamp string
}

// UncompressLogs takes gziped CloudWatch Log batch data and returns LogMessage(s)
func UncompressLogs(ts *time.Time, data []byte) ([]*LogMessage, error) {

	dataReader := bytes.NewReader(data)

	gzipReader, err := gzip.NewReader(dataReader)
	if err != nil {
		return nil, errors.Wrap(err, "ungzip data failed")
	}

	var batch LogBatch

	err = json.NewDecoder(gzipReader).Decode(&batch)
	if err != nil {
		return nil, errors.Wrap(err, "json decode failed")
	}

	LogEvents := make([]*LogMessage, len(batch.LogEvents))

	for i, entry := range batch.LogEvents {
		LogEvents[i] = &LogMessage{
			LogGroup:  batch.LogGroup,
			Timestamp: ts.Format(time.RFC3339),
			Message:   strings.TrimSuffix(entry.Message, "\n"),
		}
	}

	return LogEvents, nil

}
