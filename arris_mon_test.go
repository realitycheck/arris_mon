package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	_ "github.com/realitycheck/arris_mon/statik"
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
		want record
	}{
		{
			name: "Check Upstream 1",
			want: record{
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
			want: record{
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

func Test_fallbackFS_Open(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		ffs     fallbackFS
		args    args
		wantF   bool
		wantErr bool
	}{
		{
			name:    "Check arris_mon.js in  res/static",
			ffs:     fallbackFS{http.Dir("res/static")},
			args:    args{name: "arris_mon.js"},
			wantF:   true,
			wantErr: false,
		},
		{
			name:    "Check arris_mon.js in  res/templates",
			ffs:     fallbackFS{http.Dir("res/templates")},
			args:    args{name: "arris_mon.js"},
			wantF:   false,
			wantErr: true,
		},
		{
			name:    "Check arris_mon.js in  res/templates and res/static",
			ffs:     fallbackFS{http.Dir("res/templates"), http.Dir("res/static")},
			args:    args{name: "arris_mon.js"},
			wantF:   true,
			wantErr: false,
		},
		{
			name:    "Check arris_mon.js in empty",
			ffs:     fallbackFS{},
			args:    args{name: "arris_mon.js"},
			wantF:   false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotF, err := tt.ffs.Open(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("fallbackFS.Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotF != nil) != tt.wantF {
				t.Errorf("fallbackFS.Open() = %v, want %v", gotF, tt.wantF)
			}
		})
	}
}
