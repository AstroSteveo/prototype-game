package sim

import (
	"prototype-game/backend/internal/spatial"
)

type CellInstance struct {
	Key      spatial.CellKey
	Entities map[string]*Entity
}

func NewCellInstance(key spatial.CellKey) *CellInstance {
	return &CellInstance{Key: key, Entities: make(map[string]*Entity)}
}
