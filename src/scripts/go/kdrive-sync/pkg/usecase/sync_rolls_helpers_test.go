package usecase

import (
	"kdrive-sync/pkg/domain"
	"testing"
	"time"
)

func TestClassifyFiles(t *testing.T) {
	t.Parallel()
	files := []domain.DriveFile{
		{ID: "1", Name: "a.JPG", Type: domain.DriveFileTypeFile},
		{ID: "2", Name: "b.jpeg", Type: domain.DriveFileTypeFile},
		{ID: "3", Name: "poem.md", Type: domain.DriveFileTypeFile},
		{ID: "4", Name: "ambiance.mp3", Type: domain.DriveFileTypeFile},
		{ID: "5", Name: "clip.mp4", Type: domain.DriveFileTypeFile},
		{ID: "6", Name: "subdir", Type: domain.DriveFileTypeDir},
		{ID: "7", Name: "skip.txt", Type: domain.DriveFileTypeFile},
		{ID: "8", Name: "second.md", Type: domain.DriveFileTypeFile},
	}
	got := classifyFiles(files)
	if len(got.images) != 2 {
		t.Fatalf("want 2 images, got %d", len(got.images))
	}
	if got.poetry == nil || got.poetry.ID != "3" {
		t.Fatalf("want poetry=id3, got %+v", got.poetry)
	}
	if got.audio == nil || got.audio.ID != "4" {
		t.Fatalf("want audio=id4, got %+v", got.audio)
	}
	if got.video == nil || got.video.ID != "5" {
		t.Fatalf("want video=id5, got %+v", got.video)
	}
}

func TestClassifyFilesVideoVariants(t *testing.T) {
	t.Parallel()
	cases := []string{"clip.mp4", "clip.webm", "clip.MOV"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			got := classifyFiles([]domain.DriveFile{
				{ID: "1", Name: name, Type: domain.DriveFileTypeFile},
			})
			if got.video == nil {
				t.Fatalf("%s should be classified as video", name)
			}
		})
	}
}

func TestFilterDirs(t *testing.T) {
	t.Parallel()
	in := []domain.DriveFile{
		{ID: "1", Name: "file", Type: domain.DriveFileTypeFile},
		{ID: "2", Name: "dir1", Type: domain.DriveFileTypeDir, CreatedAt: time.Unix(1, 0)},
		{ID: "3", Name: "dir2", Type: domain.DriveFileTypeDir, CreatedAt: time.Unix(2, 0)},
	}
	got := filterDirs(in)
	if len(got) != 2 {
		t.Fatalf("want 2 dirs, got %d", len(got))
	}
	if got[0].ID != "2" || got[1].ID != "3" {
		t.Fatalf("want [dir1,dir2], got %+v", got)
	}
}

func TestBuildAltText(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		roll string
		exif domain.ExifData
		want string
	}{
		{
			name: "empty exif falls back to generic message",
			roll: "Nuit",
			exif: domain.ExifData{},
			want: "Photographie du roll Nuit",
		},
		{
			name: "exif body only",
			roll: "Ignored",
			exif: domain.ExifData{Body: "Leica M11"},
			want: "Photographie prise avec Leica M11",
		},
		{
			name: "full exif",
			roll: "Ignored",
			exif: domain.ExifData{
				Body: "Leica M11", FocalLength: "50mm", Aperture: "f/1.4", Shutter: "1/125",
			},
			want: "Photographie prise avec Leica M11 • 50mm • f/1.4 • 1/125",
		},
		{
			name: "exif body + aperture only",
			roll: "Ignored",
			exif: domain.ExifData{Body: "Leica M11", Aperture: "f/1.4"},
			want: "Photographie prise avec Leica M11 • f/1.4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildAltText(tt.roll, tt.exif)
			if got != tt.want {
				t.Errorf("buildAltText = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSortByOrder(t *testing.T) {
	t.Parallel()
	results := []imageResult{
		{order: 3, image: domain.Image{URL: "c"}},
		{order: 1, image: domain.Image{URL: "a"}},
		{order: 2, image: domain.Image{URL: "b"}},
	}
	sortByOrder(results)
	if results[0].image.URL != "a" || results[1].image.URL != "b" || results[2].image.URL != "c" {
		t.Errorf("sortByOrder failed: got %+v", results)
	}
}
