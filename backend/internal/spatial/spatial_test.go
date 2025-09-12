package spatial

import "testing"

func TestWorldToCell(t *testing.T) {
	cell := 256.0
	cases := []struct {
		x, z   float64
		cx, cz int
	}{
		{0, 0, 0, 0},
		{255.9, 0, 0, 0},
		{256.0, 0, 1, 0},
		{-0.1, 0, -1, 0},
		{0, -0.1, 0, -1},
		{-256.0, -256.0, -1, -1},
		{512.0, 512.0, 2, 2},
	}
	for i, c := range cases {
		cx, cz := WorldToCell(c.x, c.z, cell)
		if cx != c.cx || cz != c.cz {
			t.Fatalf("case %d: got (%d,%d), want (%d,%d)", i, cx, cz, c.cx, c.cz)
		}
	}
}

func TestNeighbors3x3(t *testing.T) {
	ns := Neighbors3x3(CellKey{Cx: 5, Cz: -3})
	if len(ns) != 9 {
		t.Fatalf("expected 9 neighbors, got %d", len(ns))
	}
	// Ensure center included
	found := false
	for _, k := range ns {
		if k.Cx == 5 && k.Cz == -3 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("center cell missing in neighbors")
	}
}
