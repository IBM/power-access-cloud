package util

import (
	"strings"
	"testing"
)

func TestFormatErrorForEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string // strings that should be present in output
	}{
		{
			name:  "HTML entity decoding",
			input: `[POST /pcloud/v1/cloud-instances/{cloud_instance_id}/volumes][400] pcloudCloudinstancesVolumesPostBadRequest {&#34;description&#34;:&#34;Bad Request: volume create failed : diskType tier4 is not supported in region lon04&#34;,&#34;error&#34;:&#34;Bad Request&#34;}`,
			expected: []string{
				"API Error Details:",
				"Method:      POST",
				"Status Code: 400",
				"Description: Bad Request: volume create failed : diskType tier4 is not supported in region lon04",
				"Error: Bad Request",
			},
		},
		{
			name:  "Simple error message",
			input: "Connection timeout",
			expected: []string{
				"Connection timeout",
			},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:  "Error with quotes",
			input: `Error: "invalid parameter" provided`,
			expected: []string{
				`Error: "invalid parameter" provided`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatErrorForEmail(tt.input)

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("FormatErrorForEmail() result missing expected string.\nExpected to contain: %q\nGot: %q", expected, result)
				}
			}
		})
	}
}

func TestFormatErrorForLog(t *testing.T) {
	input := `[POST /api][400] {&#34;error&#34;:&#34;Bad Request&#34;}`
	result := FormatErrorForLog(input)

	// Should decode HTML entities but keep structure
	if !strings.Contains(result, `"error"`) {
		t.Errorf("FormatErrorForLog() should decode HTML entities. Got: %q", result)
	}
}

func TestFormatJSONString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "JSON with description and error",
			input: `{"description":"Bad Request: volume create failed","error":"Bad Request"}`,
			expected: []string{
				"Description:",
				"Error:",
			},
		},
		{
			name:  "Simple key-value",
			input: `{"message":"test"}`,
			expected: []string{
				"Message:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatJSONString(tt.input)

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("formatJSONString() result missing expected string.\nExpected to contain: %q\nGot: %q", expected, result)
				}
			}
		})
	}
}
