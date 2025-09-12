package spatial

import "math"

type Vec2 struct {
	X float64
	Z float64
}

// CellKey uniquely identifies a grid cell.
type CellKey struct {
	Cx int
	Cz int
}

// WorldToCell converts world coordinates into a cell coordinate.
func WorldToCell(x, z, cellSize float64) (int, int) {
	return int(math.Floor(x / cellSize)), int(math.Floor(z / cellSize))
}

// CellBounds returns [minX,maxX) and [minZ,maxZ) for the cell.
func CellBounds(ck CellKey, cellSize float64) (minX, maxX, minZ, maxZ float64) {
	minX = float64(ck.Cx) * cellSize
	maxX = float64(ck.Cx+1) * cellSize
	minZ = float64(ck.Cz) * cellSize
	maxZ = float64(ck.Cz+1) * cellSize
	return
}

// Neighbors3x3 returns the 3x3 neighborhood centered at key (including center).
func Neighbors3x3(center CellKey) []CellKey {
	out := make([]CellKey, 0, 9)
	for dz := -1; dz <= 1; dz++ {
		for dx := -1; dx <= 1; dx++ {
			out = append(out, CellKey{Cx: center.Cx + dx, Cz: center.Cz + dz})
		}
	}
	return out
}

// Dist2 returns squared distance between two points.
func Dist2(a, b Vec2) float64 {
	dx := a.X - b.X
	dz := a.Z - b.Z
	return dx*dx + dz*dz
}

// InsideCell returns true if the point is within the inclusive min, exclusive max bounds of the cell.
func InsideCell(p Vec2, key CellKey, cellSize float64) bool {
	minX, maxX, minZ, maxZ := CellBounds(key, cellSize)
	return p.X >= minX && p.X < maxX && p.Z >= minZ && p.Z < maxZ
}
