package blockcheck

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data func(t *testing.T) io.ReadCloser

		wantEmpty bool
		wantErr   error
	}{
		{
			name: "256k zeroed",
			data: func(t *testing.T) io.ReadCloser {
				r := bytes.NewBuffer(make([]byte, 256*kilobyte))
				return ioutil.NopCloser(r)
			},
			wantEmpty: true,
			wantErr:   nil,
		},
		{
			name: "256k first 4k populated",
			data: func(t *testing.T) io.ReadCloser {
				buf := make([]byte, 256*kilobyte)
				for i := 0; i < 4*kilobyte; i++ {
					buf[i] = 42
				}
				r := bytes.NewBuffer(buf)
				return ioutil.NopCloser(r)
			},
			wantEmpty: false,
			wantErr:   nil,
		},
		{
			name: "256k last 4k populated",
			data: func(t *testing.T) io.ReadCloser {
				buf := make([]byte, 256*kilobyte)
				start := (256 - 4) * kilobyte
				for i := start; i < len(buf); i++ {
					buf[i] = 42
				}
				r := bytes.NewBuffer(buf)
				return ioutil.NopCloser(r)
			},
			wantEmpty: false,
			wantErr:   nil,
		},
		{
			name: "255k buffer, read up till end of input, last byte populated",
			data: func(t *testing.T) io.ReadCloser {
				buf := make([]byte, 255*kilobyte)
				start := len(buf) - 1
				for i := start; i < len(buf); i++ {
					buf[i] = 42
				}
				r := bytes.NewBuffer(buf)
				return ioutil.NopCloser(r)
			},
			wantEmpty: false,
			wantErr:   nil,
		},
		{
			name: "small input, zeroed",
			data: func(t *testing.T) io.ReadCloser {
				r := bytes.NewBuffer(make([]byte, 32*kilobyte))
				return ioutil.NopCloser(r)
			},
			wantEmpty: true,
			wantErr:   nil,
		},
		{
			name: "small input, first byte populated",
			data: func(t *testing.T) io.ReadCloser {
				x := make([]byte, 32*kilobyte)
				x[0] = 42
				r := bytes.NewBuffer(x)
				return ioutil.NopCloser(r)
			},
			wantEmpty: false,
			wantErr:   nil,
		},
		{
			name: "small input, last byte populated",
			data: func(t *testing.T) io.ReadCloser {
				x := make([]byte, 32*kilobyte)
				x[len(x)-1] = 42
				r := bytes.NewBuffer(x)
				return ioutil.NopCloser(r)
			},
			wantEmpty: false,
			wantErr:   nil,
		},
		{
			name: "btrfs filesystem image",
			data: func(t *testing.T) io.ReadCloser {
				// Open the sample image
				f, err := os.Open(path.Join("testdata", "btrfs.sample.gz"))
				if err != nil {
					t.Fatal(err)
				}
				// Wrap the fd with a gzip stream decoder
				r, err := gzip.NewReader(f)
				if err != nil {
					t.Fatal(err)
				}
				return r
			},
			wantEmpty: false,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := tt.data(t)
			defer buf.Close()

			got, err := isEmpty(buf)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("got %v, want %v", pretty.Sprint(err), pretty.Sprint(tt.wantErr))
			}
			if got != tt.wantEmpty {
				t.Errorf("got %v, want %v", pretty.Sprint(got), pretty.Sprint(tt.wantEmpty))
			}
		})
	}
}
