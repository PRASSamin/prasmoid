package cmd_tests

import (
	"testing"

	"github.com/PRASSamin/prasmoid/src/cmd"
)

func TestGetNextVersion(t *testing.T) {
	testCases := []struct {
		name          string
		currentVersion string
		bump          string
		expected      string
		expectError   bool
	}{
		{
			name:          "patch bump",
			currentVersion: "1.2.3",
			bump:          "patch",
			expected:      "1.2.4",
			expectError:   false,
		},
		{
			name:          "minor bump",
			currentVersion: "1.2.3",
			bump:          "minor",
			expected:      "1.3.0",
			expectError:   false,
		},
		{
			name:          "major bump",
			currentVersion: "1.2.3",
			bump:          "major",
			expected:      "2.0.0",
			expectError:   false,
		},
		{
			name:          "invalid version format",
			currentVersion: "1.2",
			bump:          "patch",
			expected:      "",
			expectError:   true,
		},
		{
			name:          "invalid bump type",
			currentVersion: "1.2.3",
			bump:          "invalid",
			expected:      "",
			expectError:   true,
		},
        {
			name:          "patch from zero",
			currentVersion: "0.0.0",
			bump:          "patch",
			expected:      "0.0.1",
			expectError:   false,
		},
        {
			name:          "minor with high patch",
			currentVersion: "1.9.15",
			bump:          "minor",
			expected:      "1.10.0",
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nextVersion, err := cmd.GetNextVersion(tc.currentVersion, tc.bump)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
				if nextVersion != tc.expected {
					t.Errorf("Expected version %s, but got %s", tc.expected, nextVersion)
				}
			}
		})
	}
}
