package display

import (
	"bytes"
	"testing"

	"periph.io/x/conn/v3/physic"
)

// recordingBus is a fake i2c.Bus that records transactions.
type recordingBus struct {
	addrs  []uint16
	writes [][]byte
}

func (b *recordingBus) String() string { return "recording" }

func (b *recordingBus) Tx(addr uint16, w, r []byte) error {
	b.addrs = append(b.addrs, addr)
	b.writes = append(b.writes, append([]byte(nil), w...))
	return nil
}

func (b *recordingBus) SetSpeed(f physic.Frequency) error { return nil }

func TestRemappedI2CBus(t *testing.T) {
	rec := &recordingBus{}
	bus := &remappedI2CBus{Bus: rec, from: ssd1306DefaultAddr, to: 0x3D}

	if err := bus.Tx(ssd1306DefaultAddr, []byte{0x00}, nil); err != nil {
		t.Fatalf("Tx failed: %v", err)
	}
	if err := bus.Tx(0x18, []byte{0x00}, nil); err != nil {
		t.Fatalf("Tx failed: %v", err)
	}

	want := []uint16{0x3D, 0x18}
	if len(rec.addrs) != len(want) {
		t.Fatalf("expected %d transactions, got %d", len(want), len(rec.addrs))
	}
	for i, addr := range want {
		if rec.addrs[i] != addr {
			t.Errorf("transaction %d: expected address %#x, got %#x", i, addr, rec.addrs[i])
		}
	}
}

func TestSSD1306SetBrightness(t *testing.T) {
	rec := &recordingBus{}
	d := &SSD1306Display{bus: rec, addr: 0x3D}

	if err := d.SetBrightness(0); err != nil {
		t.Fatalf("SetBrightness(0) failed: %v", err)
	}
	if err := d.SetBrightness(128); err != nil {
		t.Fatalf("SetBrightness(128) failed: %v", err)
	}

	want := [][]byte{
		{ssd1306CmdPrefix, ssd1306DisplayOff},
		{ssd1306CmdPrefix, ssd1306DisplayOn, ssd1306SetContrast, 128},
	}
	if len(rec.writes) != len(want) {
		t.Fatalf("expected %d transactions, got %d", len(want), len(rec.writes))
	}
	for i, w := range want {
		if rec.addrs[i] != 0x3D {
			t.Errorf("transaction %d: expected address 0x3D, got %#x", i, rec.addrs[i])
		}
		if !bytes.Equal(rec.writes[i], w) {
			t.Errorf("transaction %d: expected bytes %#v, got %#v", i, w, rec.writes[i])
		}
	}
}

func TestParseI2CAddr(t *testing.T) {
	tests := []struct {
		in      string
		want    uint16
		wantErr bool
	}{
		{"0x3C", 0x3C, false},
		{"0x3d", 0x3D, false},
		{"0X18", 0x18, false},
		{"3c", 0x3C, false},
		{"", 0, true},
		{"0xZZ", 0, true},
	}

	for _, tt := range tests {
		got, err := parseI2CAddr(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseI2CAddr(%q): error = %v, wantErr %v", tt.in, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("parseI2CAddr(%q) = %#x, want %#x", tt.in, got, tt.want)
		}
	}
}
