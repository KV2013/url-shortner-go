package file

import (
	"os"
	"testing"

	"github.com/KV2013/url-shortner-go/internal/logger"
	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/magiconair/properties/assert"
)

func newTempRepo(t *testing.T) (*FileRepository, string) {
	t.Helper()
	f, err := os.CreateTemp("", "file_repo_test_*.json")
	if err != nil {
		t.Fatalf("не удалось создать временный файл: %v", err)
	}
	f.Close()

	logger, err := logger.New("debug")
	if err != nil {
		t.Fatalf("не удалось создать логгер: %v", err)
	}
	repo, err := NewRepository(f.Name(), logger)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	t.Cleanup(func() {
		repo.Close()
		os.Remove(f.Name())
	})
	return repo, f.Name()
}

func TestSave(t *testing.T) {
	repo, _ := newTempRepo(t)

	url := &model.URL{Original: "https://example.com", Short: "abc123"}
	err := repo.Save(url)
	assert.Equal(t, err, nil)
}

func TestGetByID_Found(t *testing.T) {
	repo, _ := newTempRepo(t)

	saved := &model.URL{Original: "https://example.com", Short: "abc123"}
	if err := repo.Save(saved); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// GetByID читает файл с начала — нужно сбросить позицию
	repo.file.Seek(0, 0)

	got, ok := repo.GetByID("abc123")
	assert.Equal(t, ok, true)
	assert.Equal(t, got.Original, saved.Original)
	assert.Equal(t, got.Short, saved.Short)
}

func TestGetByID_NotFound(t *testing.T) {
	repo, _ := newTempRepo(t)

	repo.Save(&model.URL{Original: "https://example.com", Short: "abc123"})
	repo.file.Seek(0, 0)

	_, ok := repo.GetByID("nonexistent")
	assert.Equal(t, ok, false)
}

func TestGetByID_MultipleRecords(t *testing.T) {
	repo, _ := newTempRepo(t)

	urls := []*model.URL{
		{Original: "https://first.com", Short: "aaa"},
		{Original: "https://second.com", Short: "bbb"},
		{Original: "https://third.com", Short: "ccc"},
	}
	for _, u := range urls {
		if err := repo.Save(u); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	repo.file.Seek(0, 0)
	got, ok := repo.GetByID("bbb")
	assert.Equal(t, ok, true)
	assert.Equal(t, got.Original, "https://second.com")
}

func TestSave_PersistsAfterReopen(t *testing.T) {
	f, err := os.CreateTemp("", "file_repo_reopen_*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	logger, err := logger.New("debug")
	if err != nil {
		t.Fatalf("не удалось создать логгер: %v", err)
	}
	repo, err := NewRepository(path, logger)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	repo.Save(&model.URL{Original: "https://example.com", Short: "xyz"})
	repo.Close()

	// открываем заново и проверяем, что данные сохранились
	repo2, err := NewRepository(path, logger)
	if err != nil {
		t.Fatalf("NewRepository (reopen): %v", err)
	}
	defer repo2.Close()

	got, ok := repo2.GetByID("xyz")
	assert.Equal(t, ok, true)
	assert.Equal(t, got.Original, "https://example.com")
}
