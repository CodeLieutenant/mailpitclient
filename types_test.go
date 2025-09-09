package mailpitclient

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAttachmentList_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		jsonData    string
		expected    AttachmentList
		expectError bool
	}{
		{
			name: "valid attachment list",
			jsonData: `[
				{
					"PartID": "1",
					"FileName": "document.pdf",
					"ContentType": "application/pdf",
					"Size": 1024
				},
				{
					"PartID": "2",
					"FileName": "image.jpg",
					"ContentType": "image/jpeg",
					"Size": 2048
				}
			]`,
			expected: AttachmentList{
				{
					PartID:      "1",
					FileName:    "document.pdf",
					ContentType: "application/pdf",
					Size:        1024,
				},
				{
					PartID:      "2",
					FileName:    "image.jpg",
					ContentType: "image/jpeg",
					Size:        2048,
				},
			},
			expectError: false,
		},
		{
			name:        "empty attachment list",
			jsonData:    `[]`,
			expected:    AttachmentList{},
			expectError: false,
		},
		{
			name:        "null attachment list",
			jsonData:    `null`,
			expected:    nil,
			expectError: false,
		},
		{
			name:        "invalid JSON",
			jsonData:    `invalid json`,
			expected:    nil,
			expectError: true,
		},
		{
			name: "attachment with missing fields",
			jsonData: `[
				{
					"PartID": "1",
					"Filename": "test.txt"
				}
			]`,
			expected: AttachmentList{
				{
					PartID:   "1",
					FileName: "test.txt",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var attachments AttachmentList
			err := json.Unmarshal([]byte(tt.jsonData), &attachments)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, attachments)
			}
		})
	}
}

func TestListOptions_ToURLValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		opts     *ListOptions
		expected url.Values
		name     string
	}{
		{
			name: "all fields populated",
			opts: &ListOptions{
				Query: "test query",
				Tag:   "important",
				Sort:  "date",
				Start: 10,
				Limit: 25,
			},
			expected: url.Values{
				"query": []string{"test query"},
				"tag":   []string{"important"},
				"sort":  []string{"date"},
				"start": []string{"10"},
				"limit": []string{"25"},
			},
		},
		{
			name: "partial fields populated",
			opts: &ListOptions{
				Query: "search term",
				Limit: 50,
			},
			expected: url.Values{
				"query": []string{"search term"},
				"limit": []string{"50"},
			},
		},
		{
			name: "only start and limit with zero values",
			opts: &ListOptions{
				Start: 0,
				Limit: 0,
			},
			expected: url.Values{},
		},
		{
			name: "start > 0, limit = 0",
			opts: &ListOptions{
				Start: 5,
				Limit: 0,
			},
			expected: url.Values{
				"start": []string{"5"},
			},
		},
		{
			name: "start = 0, limit > 0",
			opts: &ListOptions{
				Start: 0,
				Limit: 10,
			},
			expected: url.Values{
				"limit": []string{"10"},
			},
		},
		{
			name:     "nil options",
			opts:     nil,
			expected: url.Values{},
		},
		{
			name:     "empty options",
			opts:     &ListOptions{},
			expected: url.Values{},
		},
		{
			name: "empty string fields",
			opts: &ListOptions{
				Query: "",
				Tag:   "",
				Sort:  "",
			},
			expected: url.Values{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.opts.ToURLValues()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSearchOptions_ToURLValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		opts     *SearchOptions
		expected url.Values
		name     string
	}{
		{
			name: "all fields populated",
			opts: &SearchOptions{
				Tag:   "work",
				Sort:  "subject",
				Start: 20,
				Limit: 15,
			},
			expected: url.Values{
				"tag":   []string{"work"},
				"sort":  []string{"subject"},
				"start": []string{"20"},
				"limit": []string{"15"},
			},
		},
		{
			name: "partial fields populated",
			opts: &SearchOptions{
				Tag:   "personal",
				Start: 0,
				Limit: 100,
			},
			expected: url.Values{
				"tag":   []string{"personal"},
				"limit": []string{"100"},
			},
		},
		{
			name: "only start and limit with zero values",
			opts: &SearchOptions{
				Start: 0,
				Limit: 0,
			},
			expected: url.Values{},
		},
		{
			name: "start > 0, limit = 0",
			opts: &SearchOptions{
				Start: 3,
				Limit: 0,
			},
			expected: url.Values{
				"start": []string{"3"},
			},
		},
		{
			name: "start = 0, limit > 0",
			opts: &SearchOptions{
				Start: 0,
				Limit: 30,
			},
			expected: url.Values{
				"limit": []string{"30"},
			},
		},
		{
			name:     "nil options",
			opts:     nil,
			expected: url.Values{},
		},
		{
			name:     "empty options",
			opts:     &SearchOptions{},
			expected: url.Values{},
		},
		{
			name: "empty string fields",
			opts: &SearchOptions{
				Tag:  "",
				Sort: "",
			},
			expected: url.Values{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.opts.ToURLValues()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestAttachmentList_JSONMarshaling(t *testing.T) {
	t.Parallel()

	// Test that AttachmentList can be marshaled to JSON and back
	original := AttachmentList{
		{
			PartID:      "1",
			FileName:    "test.pdf",
			ContentType: "application/pdf",
			Size:        1024,
		},
		{
			PartID:      "2",
			FileName:    "image.png",
			ContentType: "image/png",
			Size:        2048,
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled AttachmentList
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Should be equal
	require.Equal(t, original, unmarshaled)
}

func TestURLValues_Integration(t *testing.T) {
	t.Parallel()

	// Test that URL values can be used in actual URL construction
	listOpts := &ListOptions{
		Query: "test search",
		Tag:   "important",
		Sort:  "date",
		Start: 10,
		Limit: 25,
	}

	values := listOpts.ToURLValues()

	// Build URL
	baseURL := "http://example.com/messages"
	u, err := url.Parse(baseURL)
	require.NoError(t, err)

	u.RawQuery = values.Encode()

	// Verify the URL contains all expected parameters
	finalURL := u.String()
	require.Contains(t, finalURL, "query=test+search")
	require.Contains(t, finalURL, "tag=important")
	require.Contains(t, finalURL, "sort=date")
	require.Contains(t, finalURL, "start=10")
	require.Contains(t, finalURL, "limit=25")
}
