package thumbnail_test

import (
	"testing"

	"github.com/kirillgashkov/assignment-youthumb/internal/thumbnail"
)

func TestThumbnailURLFromVideoURL(t *testing.T) {
	tests := []struct {
		name     string
		videoURL string
		want     string
		wantErr  bool
	}{
		{name: "www.youtube.com", videoURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ", want: "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg"},
		{name: "youtube.com", videoURL: "https://youtube.com/watch?v=dQw4w9WgXcQ", want: "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg"},
		{name: "youtu.be", videoURL: "https://youtu.be/dQw4w9WgXcQ", want: "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg"},
		{name: "invalid", videoURL: "https://example.com/watch?v=dQw4w9WgXcQ", wantErr: true},
		{name: "empty", videoURL: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := thumbnail.ThumbnailURLFromVideoURL(tt.videoURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ThumbnailURLFromVideoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ThumbnailURLFromVideoURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
