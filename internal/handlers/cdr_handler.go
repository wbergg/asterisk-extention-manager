package handlers

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/wbergg/asterisk-extention-manager/internal/config"
)

type CDRHandler struct {
	Config *config.Config
}

type CDRRecord struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	CallerID    string `json:"callerid"`
	Start       string `json:"start"`
	Answer      string `json:"answer"`
	End         string `json:"end"`
	Duration    int    `json:"duration"`
	BillSec     int    `json:"billsec"`
	Disposition string `json:"disposition"`
	Channel     string `json:"channel"`
	DstChannel  string `json:"dst_channel"`
}

type CDRPage struct {
	Records []CDRRecord `json:"records"`
	Total   int         `json:"total"`
	Offset  int         `json:"offset"`
	HasMore bool        `json:"has_more"`
}

type CDRStats struct {
	TotalCalls    int                `json:"total_calls"`
	Answered      int                `json:"answered"`
	AvgDuration   int                `json:"avg_duration"`
	AnswerRate    int                `json:"answer_rate"`
	CallsPerDay   map[string]int     `json:"calls_per_day"`
	Dispositions  map[string]int     `json:"dispositions"`
}

// ListCDR returns a paginated slice of CDR records (newest first).
// Query params: offset (default 0), limit (default 100).
func (h *CDRHandler) ListCDR(w http.ResponseWriter, r *http.Request) {
	all, err := readCDR(h.Config.CDRLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CDRPage{Records: []CDRRecord{}})
			return
		}
		http.Error(w, `{"error":"failed to read CDR log"}`, http.StatusInternalServerError)
		return
	}

	offset := 0
	limit := 100
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	total := len(all)
	end := offset + limit
	if end > total {
		end = total
	}

	var records []CDRRecord
	if offset < total {
		records = all[offset:end]
	} else {
		records = []CDRRecord{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CDRPage{
		Records: records,
		Total:   total,
		Offset:  offset,
		HasMore: end < total,
	})
}

// Stats returns aggregate statistics computed from all CDR records.
func (h *CDRHandler) Stats(w http.ResponseWriter, r *http.Request) {
	all, err := readCDR(h.Config.CDRLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CDRStats{
				CallsPerDay:  map[string]int{},
				Dispositions: map[string]int{},
			})
			return
		}
		http.Error(w, `{"error":"failed to read CDR log"}`, http.StatusInternalServerError)
		return
	}

	stats := CDRStats{
		CallsPerDay:  make(map[string]int),
		Dispositions: make(map[string]int),
	}
	totalDuration := 0
	for _, rec := range all {
		stats.TotalCalls++
		if rec.Disposition == "ANSWERED" {
			stats.Answered++
		}
		totalDuration += rec.Duration
		stats.Dispositions[rec.Disposition]++
		if len(rec.Start) >= 10 {
			stats.CallsPerDay[rec.Start[:10]]++
		}
	}
	if stats.TotalCalls > 0 {
		stats.AvgDuration = totalDuration / stats.TotalCalls
		stats.AnswerRate = stats.Answered * 100 / stats.TotalCalls
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// readCDR reads the entire CDR CSV file and returns all records in reverse
// chronological order (newest first).
func readCDR(path string) ([]CDRRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var all []CDRRecord
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		rec, err := parseCDRLine(line)
		if err != nil {
			continue
		}
		// Skip AppDial B-leg records (empty destination, inflated durations)
		if rec.Destination == "" {
			continue
		}
		all = append(all, rec)
	}

	// Reverse to newest first
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	return all, nil
}

func parseCDRLine(line string) (CDRRecord, error) {
	r := csv.NewReader(strings.NewReader(line))
	fields, err := r.Read()
	if err != nil && err != io.EOF {
		return CDRRecord{}, err
	}
	if len(fields) < 16 {
		return CDRRecord{}, io.ErrUnexpectedEOF
	}

	duration, _ := strconv.Atoi(fields[12])
	billsec, _ := strconv.Atoi(fields[13])

	rec := CDRRecord{
		Source:      fields[1],
		Destination: fields[2],
		CallerID:    fields[4],
		Channel:     fields[5],
		DstChannel:  fields[6],
		Start:       fields[9],
		Answer:      fields[10],
		End:         fields[11],
		Duration:    duration,
		BillSec:     billsec,
		Disposition: fields[14],
	}

	// Fix CDR for outgoing fax calls: source is "fax", dest is "s",
	// and the real destination is in the Channel field (e.g. PJSIP/2003-xxx).
	if rec.Source == "fax" && rec.Destination == "s" {
		if parts := strings.SplitN(rec.Channel, "/", 2); len(parts) == 2 {
			ext := parts[1]
			if dash := strings.Index(ext, "-"); dash >= 0 {
				ext = ext[:dash]
			}
			rec.Destination = ext
		}
	}

	return rec, nil
}
