package unsafemanip

import (
	"bytes"
	"strings"
	"testing"
)

func TestStringToBytes(t *testing.T) {
	var cases = []struct {
		name string
		data string
		size int
	}{
		{"small", "lmao", 4},
		{"big", "This CONTAINS test DATA!! :)", 28},
		{"non ascii", "\x01\x02\x03\x04\r\n", 6},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			converted := StringToBytes(c.data)
			if !bytes.Equal([]byte(c.data), converted) {
				t.Fatal("Failed to convert data")
			}
			if c.size != len(converted) {
				t.Fatal("Invalid len()")
			}
			if c.size != cap(converted) {
				t.Fatal("Invalid cap()")
			}
		})
	}
}

func TestStringToBytesFast(t *testing.T) {
	var cases = []struct {
		name string
		data string
		size int
	}{
		{"small", "lmao", 4},
		{"big", "This CONTAINS test DATA!! :)", 28},
		{"non ascii", "\x01\x02\x03\x04\r\n", 6},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			converted := StringToBytesFast(&c.data)
			if !bytes.Equal([]byte(c.data), converted) {
				t.Fatal("Failed to convert data")
			}
			if c.size != len(converted) {
				t.Fatal("Invalid len()")
			}
			if c.size != cap(converted) {
				t.Fatal("Invalid cap()")
			}
		})
	}
}

func TestStdSetSliceLen(t *testing.T) {
	var cases = []struct {
		name     string
		data     []byte
		size     int
		fullSize int
	}{
		{"short", []byte("data"), 2, 4},
		{"long", []byte("lots of stuff in here"), 4, 21},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			oldlen := len(c.data)
			c.data = c.data[:c.size]
			if len(c.data) != c.size {
				t.Fatal("Failed to set len")
			}

			c.data = c.data[:oldlen]
			if len(c.data) != c.fullSize {
				t.Fatal("Failed to restore len")
			}
		})
	}
}

func TestSetStringLen(t *testing.T) {
	var cases = []struct {
		name     string
		data     string
		size     int
		fullSize int
	}{
		{"short", "data", 2, 4},
		{"long", "lots of stuff in here", 4, 21},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			oldlen := SetStringLen(&c.data, c.size)
			if len(c.data) != c.size {
				t.Fatal("Failed to set len")
			}

			SetStringLen(&c.data, oldlen)
			if len(c.data) != c.fullSize {
				t.Fatal("Failed to restore len")
			}
		})
	}
}

var (
	benchData     = strings.Repeat("Strxx", 1e4)
	benchByteData = bytes.Repeat([]byte("Bytex"), 1e4)
)

func BenchmarkStringToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = StringToBytes(benchData)
	}
}

func BenchmarkStringToBytesFast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = StringToBytesFast(&benchData)
	}
}

func BenchmarkStdSetSliceLen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		oldlen := len(benchByteData)
		benchByteData = benchByteData[:20]
		benchByteData = benchByteData[:oldlen]
	}
}

func BenchmarkSetSliceLen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		oldlen := SetSliceLen(&benchByteData, 20)
		SetSliceLen(&benchByteData, oldlen)
	}
}

func BenchmarkSetStringLen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		oldlen := SetStringLen(&benchData, 20)
		SetStringLen(&benchData, oldlen)
	}
}
