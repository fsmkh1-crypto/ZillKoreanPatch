// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import "testing"

func TestPlaceRasterPadsAndCopiesExactly(t *testing.T) {
	source := Raster{Width: 2, Height: 2, Pixels: []uint8{1, 2, 3, 4}}
	got, err := PlaceRaster(4, 3, 1, 1, source)
	if err != nil {
		t.Fatal(err)
	}
	want := []uint8{
		0, 0, 0, 0,
		0, 1, 2, 0,
		0, 3, 4, 0,
	}
	if len(got.Pixels) != len(want) {
		t.Fatalf("pixels = %d, want %d", len(got.Pixels), len(want))
	}
	for i := range want {
		if got.Pixels[i] != want[i] {
			t.Fatalf("pixel %d = %d, want %d", i, got.Pixels[i], want[i])
		}
	}
}

func TestPlaceCenteredRasterIsDeterministic(t *testing.T) {
	source := Raster{Width: 2, Height: 1, Pixels: []uint8{5, 6}}
	got, err := PlaceCenteredRaster(5, 3, source)
	if err != nil {
		t.Fatal(err)
	}
	// Floor centering yields paste=(1,1) for a 2x1 raster in a 5x3 cell.
	if got.Pixels[1*5+1] != 5 || got.Pixels[1*5+2] != 6 {
		t.Fatalf("unexpected centered raster: %#v", got.Pixels)
	}
}

func TestPlaceRasterRejectsOverflow(t *testing.T) {
	source := Raster{Width: 3, Height: 3, Pixels: make([]uint8, 9)}
	if _, err := PlaceRaster(4, 4, 2, 2, source); err == nil {
		t.Fatal("expected placement overflow error")
	}
}
