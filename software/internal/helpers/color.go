package helpers

func RGBto565(r, g, b int) uint16 {
	r5 := uint16((r >> 3) & 0x1F)
	g6 := uint16((g >> 2) & 0x3F)
	b5 := uint16((b >> 3) & 0x1F)

	return (r5 << 11) | (g6 << 5) | b5
}
