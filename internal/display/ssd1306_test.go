package display

import (
	"testing"

	"periph.io/x/conn/v3/physic"
)

// recordingBus is a fake i2c.Bus that records the addresses of transactions.
type recordingBus struct {
	addrs []uint16
}

func (b *recordingBus) String() string { return "recording" }

func (b *recordingBus) Tx(addr uint16, w, r []byte) error {
	b.addrs = append(b.addrs, addr)
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
