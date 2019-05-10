package blockcheck

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data func() []byte

		want    bool
		wantErr error
	}{
		{
			name: "256k zeroed",
			data: func() []byte {
				return make([]byte, 256*kilobyte)
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "256k first 4k populated",
			data: func() []byte {
				buf := make([]byte, 256*kilobyte)
				for i := 0; i < 4*kilobyte; i++ {
					buf[i] = 42
				}
				return buf
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "256k last 4k populated",
			data: func() []byte {
				buf := make([]byte, 256*kilobyte)
				start := (256 - 4) * kilobyte
				for i := start; i < len(buf); i++ {
					buf[i] = 42
				}
				return buf
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "255k buffer, read up till end of input, last byte populated",
			data: func() []byte {
				buf := make([]byte, 255*kilobyte)
				start := len(buf) - 1
				for i := start; i < len(buf); i++ {
					buf[i] = 42
				}
				return buf
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "small input, zeroed",
			data: func() []byte {
				return make([]byte, 32*kilobyte)
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "small input, first byte populated",
			data: func() []byte {
				x := make([]byte, 32*kilobyte)
				x[0] = 42
				return x
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "small input, last byte populated",
			data: func() []byte {
				x := make([]byte, 32*kilobyte)
				x[len(x)-1] = 42
				return x
			},
			want:    false,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBuffer(tt.data())

			got, err := isEmpty(buf)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("got %v, want %v", pretty.Sprint(err), pretty.Sprint(tt.wantErr))
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", pretty.Sprint(got), pretty.Sprint(tt.want))
			}
		})
	}
}
