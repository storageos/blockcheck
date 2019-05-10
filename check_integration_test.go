// +build linux

package blockcheck

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

func TestIsBlockDeviceEmpty(t *testing.T) {
	t.Parallel()

	runCmd := func(t *testing.T, cmd string, args ...string) {
		t.Helper()

		buf := &bytes.Buffer{}

		c := exec.Command(cmd, args...)
		c.Stdout = buf
		c.Stderr = buf
		if err := c.Run(); err != nil {
			t.Log(buf.String())
			t.Fatalf("failed to exec %q (args: %v): %v", cmd, args, err)
		}
	}

	tests := []struct {
		name string
		size int
		init func(t *testing.T, path string, i int)

		wantEmpty bool
	}{
		{
			name: "ext4",
			size: 256 * kilobyte,
			init: func(t *testing.T, path string, i int) {
				runCmd(t, "mkfs.ext4", "-F", path)
			},
			wantEmpty: false,
		},
		{
			name: "ext3",
			size: 256 * kilobyte,
			init: func(t *testing.T, path string, i int) {
				runCmd(t, "mkfs.ext3", "-F", path)
			},
			wantEmpty: false,
		},
		{
			name: "xfs",
			size: 100 * 1024 * kilobyte,
			init: func(t *testing.T, path string, i int) {
				runCmd(t, "mkfs.xfs", path)
			},
			wantEmpty: false,
		},
		{
			name: "btrfs",
			size: 1024 * 1024 * kilobyte,
			init: func(t *testing.T, path string, i int) {
				runCmd(t, "mkfs.btrfs", path)
			},
			wantEmpty: false,
		},
		{
			name:      "no fs",
			size:      256 * kilobyte,
			init:      func(t *testing.T, path string, i int) {},
			wantEmpty: true,
		},
	}

	for i, tt := range tests {
		var tt = tt
		var i = i
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := path.Join(".", fmt.Sprintf("test-%d", i))
			t.Logf("using file %q as block device image", path)

			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(path)
			defer f.Close()

			// Write tt.size number of zeros
			if _, err := f.Write(make([]byte, tt.size)); err != nil {
				t.Fatal(err)
			}

			// Close the file and run init, which populates the file
			f.Close()
			tt.init(t, path, i+100)

			// Mount the block device
			dev := fmt.Sprintf("/dev/loop%d", 100+i)
			t.Logf("using loopback dev %q", dev)

			// Try to create the device - this is going to fail when the user
			// running the test does not have root permissions, so instead of
			// failing, this test skips if the device cannot be created.
			buf := &bytes.Buffer{}
			c := exec.Command("losetup", dev, path)
			c.Stdout = buf
			c.Stderr = buf
			if err := c.Run(); err != nil {
				t.Log(buf.String())
				t.Skipf("failed to create loopback device: %v", err)
			}

			// Always free the device
			defer func() {
				runCmd(t, "losetup", "-d", dev)
			}()

			got, err := IsBlockDeviceEmpty(dev)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.wantEmpty {
				t.Errorf("got %v, want %v", pretty.Sprint(got), pretty.Sprint(tt.wantEmpty))
			}
		})
	}

	t.Run("normal file", func(t *testing.T) {
		path := path.Join(".", "test-file")
		t.Logf("using file %q", path)

		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(path)
		defer f.Close()

		// Write tt.size number of zeros
		if _, err := f.Write(make([]byte, 256*kilobyte)); err != nil {
			t.Fatal(err)
		}

		got, err := IsBlockDeviceEmpty(path)
		if !reflect.DeepEqual(err, ErrNotBlockDevice) {
			t.Errorf("got %v, want %v", pretty.Sprint(err), pretty.Sprint(ErrNotBlockDevice))
		}
		if got != false {
			t.Errorf("got %v, want %v", pretty.Sprint(got), pretty.Sprint(false))
		}
	})
}
