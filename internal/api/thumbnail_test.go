package api_test

import (
	"github.com/kirillgashkov/assignment-youthumb/internal/api"
	"testing"
)

func TestParseVideoID(t *testing.T) {
	tests := []struct {
		name     string
		videoURL string
		want     string
		wantErr  bool
	}{
		{name: "www.youtube.com", videoURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ", want: "dQw4w9WgXcQ"},
		{name: "youtube.com", videoURL: "https://youtube.com/watch?v=dQw4w9WgXcQ", want: "dQw4w9WgXcQ"},
		{name: "youtu.be", videoURL: "https://youtu.be/dQw4w9WgXcQ", want: "dQw4w9WgXcQ"},
		{name: "invalid", videoURL: "https://example.com/watch?v=dQw4w9WgXcQ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := api.ParseVideoID(tt.videoURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVideoID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseVideoID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
