package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

var (
	fileSampleHTML = "sample_source.html"
	fileSampleZIP  = "test/sample_source.zip"
	tblUpstream    = table{
		{
			"",
			"UCID",
			"Freq",
			"Power",
			"Channel Type",
			"Symbol Rate",
			"Modulation",
		},
		{
			"Upstream 1",
			"5",
			"36.00 MHz",
			"46.50 dBmV",
			"DOCSIS2.0 (ATDMA)",
			"5120 kSym/s",
			"32QAM",
		},
		{
			"Upstream 2",
			"6",
			"44.00 MHz",
			"46.50 dBmV",
			"DOCSIS2.0 (ATDMA)",
			"5120 kSym/s",
			"32QAM",
		},
	}
	tblDownstream = table{
		{
			"",
			"DCID",
			"Freq",
			"Power",
			"SNR",
			"Modulation",
			"Octets",
			"Correcteds",
			"Uncorrectables",
		},
		{
			"Downstream 1",
			"73",
			"114.00 MHz",
			"0.82 dBmV",
			"32.77 dB",
			"256QAM",
			"1144704283",
			"760100388",
			"26454645",
		},
		{
			"Downstream 2",
			"74",
			"122.00 MHz",
			"2.70 dBmV",
			"36.39 dB",
			"256QAM",
			"991440664",
			"9866185",
			"12556",
		},
		{
			"Downstream 3",
			"75",
			"130.00 MHz",
			"2.36 dBmV",
			"37.09 dB",
			"256QAM",
			"990710609",
			"2502231",
			"8385",
		},
		{
			"Downstream 4",
			"76",
			"138.00 MHz",
			"2.33 dBmV",
			"37.94 dB",
			"256QAM",
			"991393690",
			"29883",
			"11059",
		},
	}
)

func unzip(zipName, fileName string) ([]byte, error) {
	r, err := zip.OpenReader(zipName)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == fileName {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			return ioutil.ReadAll(rc)
		}
	}

	return nil, err
}

func Test_parseTable(t *testing.T) {
	bytes, err := unzip(fileSampleZIP, fileSampleHTML)
	if err != nil {
		t.Fatalf("Error open %s: %v %v", fileSampleZIP, fileSampleHTML, err)
	}

	doc, err := html.Parse(strings.NewReader(string(bytes)))
	if err != nil {
		t.Fatalf("Error parse %s: %v %v", fileSampleZIP, fileSampleHTML, err)
	}

	tests := []struct {
		name string
		expr string
		want table
	}{
		{
			name: "Check downstream",
			expr: "//table[2]/tbody",
			want: tblDownstream,
		},
		{
			name: "Check upstream",
			expr: "//table[4]/tbody",
			want: tblUpstream,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTable(doc, tt.expr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_table_iterator(t *testing.T) {
	tests := []struct {
		name string
		want channel
	}{
		{
			name: "Check Upstream 1",
			want: channel{
				"":             "Upstream 1",
				"UCID":         "5",
				"Freq":         "36.00 MHz",
				"Power":        "46.50 dBmV",
				"Channel Type": "DOCSIS2.0 (ATDMA)",
				"Symbol Rate":  "5120 kSym/s",
				"Modulation":   "32QAM",
			},
		},
		{
			name: "Check Upstream 2",
			want: channel{
				"":             "Upstream 2",
				"UCID":         "6",
				"Freq":         "44.00 MHz",
				"Power":        "46.50 dBmV",
				"Channel Type": "DOCSIS2.0 (ATDMA)",
				"Symbol Rate":  "5120 kSym/s",
				"Modulation":   "32QAM",
			},
		},
		{
			name: "Check nil",
			want: nil,
		},
	}

	it := tblUpstream.iterator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := it(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("it() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_table_iterator_for_loop(t *testing.T) {
	tbl := tblUpstream
	it := tblUpstream.iterator()

	n := 0
	for c := it(); c != nil; c = it() {
		n++
	}
	fmt.Print(n)
	want := len(tbl) - 1
	if n != want {
		t.Errorf("for cycles count = %d, want %d", n, want)
	}
}
