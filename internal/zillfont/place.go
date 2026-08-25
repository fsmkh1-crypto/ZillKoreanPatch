// SPDX-License-Identifier: GPL-3.0-or-later

package zillfont

import "fmt"

// PlaceRaster creates an exact cell-sized raster by clearing the target cell
// and pasting source at the requested offset. This separates font rendering
// from atlas writing: a renderer may consistently produce (for example) the
// proven 10x10 Korean glyph raster, while existing retail cells may be 10x10,
// 11x12, 12x11, and so on.
func PlaceRaster(cellWidth, cellHeight, pasteX, pasteY int, source Raster) (Raster, error) {
	if err := source.validate(); err != nil {
		return Raster{}, fmt.Errorf("place raster: %w", err)
	}
	if cellWidth <= 0 || cellHeight <= 0 {
		return Raster{}, fmt.Errorf("place raster: invalid cell dimensions %dx%d", cellWidth, cellHeight)
	}
	if pasteX < 0 || pasteY < 0 || pasteX+source.Width > cellWidth || pasteY+source.Height > cellHeight {
		return Raster{}, fmt.Errorf("place raster: source %dx%d at (%d,%d) does not fit cell %dx%d", source.Width, source.Height, pasteX, pasteY, cellWidth, cellHeight)
	}
	pixels := make([]uint8, cellWidth*cellHeight)
	for row := 0; row < source.Height; row++ {
		copy(
			pixels[(pasteY+row)*cellWidth+pasteX:(pasteY+row)*cellWidth+pasteX+source.Width],
			source.Pixels[row*source.Width:(row+1)*source.Width],
		)
	}
	return Raster{Width: cellWidth, Height: cellHeight, Pixels: pixels}, nil
}

// PlaceCenteredRaster centers source inside a retail cell using deterministic
// integer floor division. Explicit placement should be preferred when bearing
// calibration calls for a known offset; this helper is for cells whose metrics
// already match the desired visual box.
func PlaceCenteredRaster(cellWidth, cellHeight int, source Raster) (Raster, error) {
	if source.Width > cellWidth || source.Height > cellHeight {
		return Raster{}, fmt.Errorf("center raster: source %dx%d does not fit cell %dx%d", source.Width, source.Height, cellWidth, cellHeight)
	}
	return PlaceRaster(cellWidth, cellHeight, (cellWidth-source.Width)/2, (cellHeight-source.Height)/2, source)
}
