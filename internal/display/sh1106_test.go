package display

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

// txs splits recorded transactions into command and data transfers.
func txs(bus *recordingBus) (cmds, data [][]byte) {
	for _, w := range bus.writes {
		if len(w) == 0 {
			continue
		}
		switch w[0] {
		case monoCmdPrefix:
			cmds = append(cmds, w[1:])
		case monoDataPrefix:
			data = append(data, w[1:])
		}
	}
	return cmds, data
}

func TestSH1106InitSequence(t *testing.T) {
	rec := &recordingBus{}
	if _, err := newSH1106OnBus(rec, 0x3D, 128, 64, 0); err != nil {
		t.Fatalf("newSH1106OnBus failed: %v", err)
	}

	if len(rec.addrs) == 0 {
		t.Fatal("expected the init sequence to be written to the bus")
	}
	for _, addr := range rec.addrs {
		if addr != 0x3D {
			t.Fatalf("expected address 0x3D, got %#x", addr)
		}
	}

	cmds, _ := txs(rec)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command transaction, got %d", len(cmds))
	}
	init := cmds[0]

	// The DC-DC enable pair is what makes the difference between a dark panel
	// and a lit one on SH1106, so check for it explicitly.
	if i := bytes.Index(init, []byte{0xAD, 0x8B}); i < 0 {
		t.Error("init sequence must enable the internal DC-DC circuit (0xAD 0x8B)")
	}
	// The SSD1306 charge-pump command is meaningless here and its presence in a
	// copied sequence is a common bug.
	if bytes.Contains(init, []byte{0x8D}) {
		t.Error("init sequence must not use the SSD1306 charge-pump command 0x8D")
	}
	// Page addressing only: the SSD1306 addressing-mode and window commands are
	// not implemented by the SH1106.
	for _, unsupported := range []byte{0x20, 0x21, 0x22} {
		if bytes.Contains(init, []byte{unsupported}) {
			t.Errorf("init sequence must not send SSD1306 addressing command %#x", unsupported)
		}
	}
	if init[len(init)-1] != 0xAF {
		t.Errorf("init sequence should end with display-on 0xAF, got %#x", init[len(init)-1])
	}
}

func TestSH1106RotationMirrorCommands(t *testing.T) {
	tests := []struct {
		rotation     int
		segmentRemap byte
		comScan      byte
		otherSegment byte
		otherCOMScan byte
	}{
		{rotation: 0, segmentRemap: 0xA1, comScan: 0xC8, otherSegment: 0xA0, otherCOMScan: 0xC0},
		{rotation: 2, segmentRemap: 0xA0, comScan: 0xC0, otherSegment: 0xA1, otherCOMScan: 0xC8},
	}

	for _, tt := range tests {
		rec := &recordingBus{}
		if _, err := newSH1106OnBus(rec, 0x3C, 128, 64, tt.rotation); err != nil {
			t.Fatalf("rotation %d: newSH1106OnBus failed: %v", tt.rotation, err)
		}
		cmds, _ := txs(rec)
		init := cmds[0]

		if !bytes.Contains(init, []byte{tt.segmentRemap}) {
			t.Errorf("rotation %d: expected segment remap %#x", tt.rotation, tt.segmentRemap)
		}
		if !bytes.Contains(init, []byte{tt.comScan}) {
			t.Errorf("rotation %d: expected COM scan %#x", tt.rotation, tt.comScan)
		}
		if bytes.Contains(init, []byte{tt.otherSegment}) {
			t.Errorf("rotation %d: unexpected segment remap %#x", tt.rotation, tt.otherSegment)
		}
		if bytes.Contains(init, []byte{tt.otherCOMScan}) {
			t.Errorf("rotation %d: unexpected COM scan %#x", tt.rotation, tt.otherCOMScan)
		}
	}
}

func TestSH1106Validation(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
		rotation      int
		wantErr       string
	}{
		{name: "128x64", width: 128, height: 64, rotation: 0},
		{name: "128x64 flipped", width: 128, height: 64, rotation: 2},
		{name: "90 degrees", width: 128, height: 64, rotation: 1, wantErr: "rotation"},
		{name: "too tall", width: 128, height: 72, rotation: 0, wantErr: "rows"},
		{name: "page misaligned", width: 128, height: 60, rotation: 0, wantErr: "multiple of 8"},
		{name: "too wide", width: 131, height: 64, rotation: 0, wantErr: "width"},
		{name: "zero height", width: 128, height: 0, rotation: 0, wantErr: "positive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newSH1106OnBus(&recordingBus{}, 0x3C, tt.width, tt.height, tt.rotation)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error mentioning %q", tt.wantErr)
			}
			if !bytes.Contains([]byte(err.Error()), []byte(tt.wantErr)) {
				t.Errorf("error %q does not mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestSH1106ShowPages(t *testing.T) {
	rec := &recordingBus{}
	d, err := newSH1106OnBus(rec, 0x3C, 128, 64, 0)
	if err != nil {
		t.Fatalf("newSH1106OnBus failed: %v", err)
	}
	rec.writes, rec.addrs = nil, nil

	// Light the top-left pixel and the bottom-right pixel of the panel.
	d.img.SetGray(0, 0, color.Gray{Y: 255})
	d.img.SetGray(127, 63, color.Gray{Y: 255})

	if err := d.Show(); err != nil {
		t.Fatalf("Show failed: %v", err)
	}

	cmds, data := txs(rec)
	if len(cmds) != 8 || len(data) != 8 {
		t.Fatalf("expected 8 command and 8 data transactions, got %d and %d", len(cmds), len(data))
	}

	for page := range 8 {
		wantCmd := []byte{0xB0 | byte(page), 0x02, 0x10} // page, column 2 (low, high nibble)
		if !bytes.Equal(cmds[page], wantCmd) {
			t.Errorf("page %d: expected address command %#v, got %#v", page, wantCmd, cmds[page])
		}
		if len(data[page]) != 128 {
			t.Fatalf("page %d: expected 128 data bytes, got %d", page, len(data[page]))
		}
	}

	// First column of the first page carries the top-left pixel (LSB = row 0).
	if data[0][0] != 0x01 {
		t.Errorf("page 0 column 0: expected 0x01, got %#02x", data[0][0])
	}
	// Last column of the last page carries the bottom-right pixel (MSB = row 7).
	if data[7][127] != 0x80 {
		t.Errorf("page 7 column 127: expected 0x80, got %#02x", data[7][127])
	}
	// Everything else stays dark.
	if data[0][1] != 0x00 || data[7][126] != 0x00 {
		t.Error("neighbouring columns should stay unlit")
	}
}

func TestSH1106SetBrightness(t *testing.T) {
	tests := []struct {
		name  string
		level uint8
		want  []byte
	}{
		{name: "off", level: 0, want: []byte{monoCmdPrefix, monoDisplayOff}},
		{name: "dim", level: 64, want: []byte{monoCmdPrefix, monoDisplayOn, monoSetContrast, 64}},
		{name: "max", level: 255, want: []byte{monoCmdPrefix, monoDisplayOn, monoSetContrast, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &recordingBus{}
			d, err := newSH1106OnBus(rec, 0x3D, 128, 64, 0)
			if err != nil {
				t.Fatalf("newSH1106OnBus failed: %v", err)
			}
			rec.writes, rec.addrs = nil, nil

			if err := d.SetBrightness(tt.level); err != nil {
				t.Fatalf("SetBrightness(%d) failed: %v", tt.level, err)
			}
			if len(rec.writes) != 1 {
				t.Fatalf("expected 1 transaction, got %d", len(rec.writes))
			}
			if rec.addrs[0] != 0x3D {
				t.Errorf("expected address 0x3D, got %#x", rec.addrs[0])
			}
			if !bytes.Equal(rec.writes[0], tt.want) {
				t.Errorf("expected %#v, got %#v", tt.want, rec.writes[0])
			}
		})
	}
}

func TestSH1106ImplementsDisplay(t *testing.T) {
	rec := &recordingBus{}
	d, err := newSH1106OnBus(rec, 0x3C, 128, 64, 0)
	if err != nil {
		t.Fatalf("newSH1106OnBus failed: %v", err)
	}

	var disp Display = d
	if err := disp.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if err := disp.DrawImage(0, 0, image.NewUniform(color.White)); err != nil {
		t.Fatalf("DrawImage failed: %v", err)
	}
	if err := disp.Show(); err != nil {
		t.Fatalf("Show failed: %v", err)
	}
	if err := disp.SetBrightness(200); err != nil {
		t.Fatalf("SetBrightness failed: %v", err)
	}
	if b := disp.GetBounds(); b.Dx() != 128 || b.Dy() != 64 {
		t.Errorf("expected 128x64, got %dx%d", b.Dx(), b.Dy())
	}
	if buf := disp.GetBuffer(); len(buf) != 128*64/8 {
		t.Errorf("expected %d buffer bytes, got %d", 128*64/8, len(buf))
	}
	if err := disp.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestSH1106InitClearsPanel(t *testing.T) {
	rec := &recordingBus{}
	d, err := newSH1106OnBus(rec, 0x3C, 128, 64, 0)
	if err != nil {
		t.Fatalf("newSH1106OnBus failed: %v", err)
	}
	d.img.SetGray(10, 10, color.Gray{Y: 255})

	if err := d.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if d.lit(10, 10) {
		t.Error("Init should clear the frame buffer")
	}

	_, data := txs(rec)
	if len(data) < 8 {
		t.Fatalf("Init should flush a blank frame, got %d data transactions", len(data))
	}
	for i, page := range data[len(data)-8:] {
		if !bytes.Equal(page, make([]byte, 128)) {
			t.Errorf("flushed page %d is not blank", i)
		}
	}
}
