// Package cfs fetches the Center for Financial Stability Divisia monetary
// aggregates. CFS publishes only an Excel workbook with no API, so the file
// is downloaded and parsed at runtime (the XLSX container is a zip of XML).
// The client is stateless; callers apply their own caching.
package cfs

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Point is a monthly Divisia M4 index value (1967 = 100).
type Point struct {
	Date  time.Time
	Index float64
}

const (
	defaultBaseURL = "https://centerforfinancialstability.org"
	userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
)

// Client fetches Divisia data from the CFS site.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the site. Used by tests.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient overrides the HTTP client, including its timeout.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) { c.httpClient = client }
}

// New returns a Client with sensible defaults.
func New(opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  userAgent,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// FetchM4Index downloads the Divisia workbook and returns the monthly Divisia
// M4 index (1967 = 100), oldest first.
func (c *Client) FetchM4Index() ([]Point, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/amfm/Divisia.xlsx", nil)
	if err != nil {
		return nil, fmt.Errorf("build CFS request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch CFS workbook: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch CFS workbook: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read CFS workbook: %w", err)
	}

	return ParseM4Index(body)
}

// ParseM4Index parses the monthly Divisia M4 "Including" index column from a
// CFS workbook (the Broad sheet, 1967 = 100).
func ParseM4Index(xlsx []byte) ([]Point, error) {
	zr, err := zip.NewReader(bytes.NewReader(xlsx), int64(len(xlsx)))
	if err != nil {
		return nil, fmt.Errorf("open CFS workbook: %w", err)
	}

	sheet, err := findSheet(zr, "Broad")
	if err != nil {
		return nil, err
	}
	if sheet == nil {
		return nil, fmt.Errorf("CFS workbook has no Broad sheet")
	}

	dec := xml.NewDecoder(bytes.NewReader(sheet))
	var doc struct {
		Rows []struct {
			Cells []struct {
				Ref string `xml:"r,attr"`
				T   string `xml:"t,attr"`
				V   string `xml:"v"`
			} `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("parse CFS Broad sheet: %w", err)
	}

	points := make([]Point, 0, len(doc.Rows))
	for i, row := range doc.Rows {
		if i < 2 { // title and header rows
			continue
		}

		var dateVal, indexVal string
		for _, cell := range row.Cells {
			col := columnLetter(cell.Ref)
			switch col {
			case 0: // A: the observation date
				dateVal = cell.V
			case 1: // B: Divisia M4 index level
				indexVal = cell.V
			}
		}
		if dateVal == "" || indexVal == "" {
			continue
		}

		date, err := parseDateCell(dateVal)
		if err != nil {
			continue
		}
		index, err := strconv.ParseFloat(indexVal, 64)
		if err != nil {
			continue
		}

		points = append(points, Point{Date: date, Index: index})
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("no Divisia M4 data in CFS workbook")
	}
	return points, nil
}

// excelEpoch is the zero date of the Excel serial calendar.
var excelEpoch = time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)

// parseDateCell converts an Excel serial date (days since 1899-12-30) or an
// ISO string to a time.
func parseDateCell(v string) (time.Time, error) {
	if day, err := strconv.ParseFloat(v, 64); err == nil {
		return excelEpoch.AddDate(0, 0, int(day)), nil
	}
	return time.Parse("2006-01-02", strings.TrimSpace(v))
}

// findSheet returns the decoded bytes of the worksheet with the given name.
func findSheet(zr *zip.Reader, name string) ([]byte, error) {
	var sheetID string
	for _, entry := range zr.File {
		if entry.Name == "xl/workbook.xml" {
			rc, err := entry.Open()
			if err != nil {
				return nil, err
			}
			var wb struct {
				Sheets []struct {
					Name string `xml:"name,attr"`
					RID  string `xml:"id,attr"`
				} `xml:"sheets>sheet"`
			}
			err = xml.NewDecoder(rc).Decode(&wb)
			_ = rc.Close()
			if err != nil {
				return nil, fmt.Errorf("parse CFS workbook.xml: %w", err)
			}
			for _, s := range wb.Sheets {
				if s.Name == name {
					sheetID = s.RID
				}
			}
			break
		}
	}

	target := ""
	for _, entry := range zr.File {
		if entry.Name == "xl/_rels/workbook.xml.rels" {
			rc, err := entry.Open()
			if err != nil {
				return nil, err
			}
			var rels struct {
				Relationships []struct {
					ID     string `xml:"Id,attr"`
					Target string `xml:"Target,attr"`
				} `xml:"Relationship"`
			}
			err = xml.NewDecoder(rc).Decode(&rels)
			_ = rc.Close()
			if err != nil {
				return nil, fmt.Errorf("parse CFS workbook rels: %w", err)
			}
			for _, r := range rels.Relationships {
				if r.ID == sheetID {
					// Worksheet targets in the workbook rels are relative to
					// the xl/ directory.
					target = "xl/" + strings.TrimPrefix(r.Target, "/")
					break
				}
			}
			break
		}
	}

	if sheetID == "" || target == "" {
		return nil, nil // sheet not referenced by this workbook
	}
	for _, entry := range zr.File {
		if entry.Name == target {
			rc, err := entry.Open()
			if err != nil {
				return nil, err
			}
			defer func() { _ = rc.Close() }()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("CFS workbook has no worksheet %s", name)
}

// columnLetter returns the zero-based column index for a cell reference like
// "B3", and -1 if it cannot be parsed.
func columnLetter(ref string) int {
	col := 0
	found := false
	for i := 0; i < len(ref); i++ {
		ch := ref[i]
		if ch >= 'A' && ch <= 'Z' {
			col = col*26 + int(ch-'A'+1)
			found = true
		} else {
			break
		}
	}
	if !found {
		return -1
	}
	return col - 1
}
