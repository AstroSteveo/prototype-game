package main

import (
	"strings"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name           string
		cellSize       float64
		aoiRadius      float64
		tickHz         int
		snapshotHz     int
		hysteresis     float64
		expectError    bool
		errorSubstring string
	}{
		{
			name:        "valid config",
			cellSize:    256.0,
			aoiRadius:   128.0,
			tickHz:      20,
			snapshotHz:  10,
			hysteresis:  2.0,
			expectError: false,
		},
		{
			name:           "zero cell size",
			cellSize:       0,
			aoiRadius:      128.0,
			tickHz:         20,
			snapshotHz:     10,
			hysteresis:     2.0,
			expectError:    true,
			errorSubstring: "cell size must be > 0",
		},
		{
			name:           "negative cell size",
			cellSize:       -10.0,
			aoiRadius:      128.0,
			tickHz:         20,
			snapshotHz:     10,
			hysteresis:     2.0,
			expectError:    true,
			errorSubstring: "cell size must be > 0",
		},
		{
			name:           "negative AOI radius",
			cellSize:       256.0,
			aoiRadius:      -5.0,
			tickHz:         20,
			snapshotHz:     10,
			hysteresis:     2.0,
			expectError:    true,
			errorSubstring: "AOI radius must be >= 0",
		},
		{
			name:           "zero tick rate",
			cellSize:       256.0,
			aoiRadius:      128.0,
			tickHz:         0,
			snapshotHz:     10,
			hysteresis:     2.0,
			expectError:    true,
			errorSubstring: "tick rate must be >= 1 Hz",
		},
		{
			name:           "zero snapshot rate",
			cellSize:       256.0,
			aoiRadius:      128.0,
			tickHz:         20,
			snapshotHz:     0,
			hysteresis:     2.0,
			expectError:    true,
			errorSubstring: "snapshot rate must be >= 1 Hz",
		},
		{
			name:           "negative hysteresis",
			cellSize:       256.0,
			aoiRadius:      128.0,
			tickHz:         20,
			snapshotHz:     10,
			hysteresis:     -1.0,
			expectError:    true,
			errorSubstring: "handover hysteresis must be >= 0",
		},
		{
			name:        "minimal valid config",
			cellSize:    0.1,
			aoiRadius:   0.0,
			tickHz:      1,
			snapshotHz:  1,
			hysteresis:  0.0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cellSize, tt.aoiRadius, tt.tickHz, tt.snapshotHz, tt.hysteresis)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorSubstring != "" && err.Error() != "" {
					if !strings.Contains(err.Error(), tt.errorSubstring) {
						t.Errorf("expected error to contain %q, got %q", tt.errorSubstring, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}
