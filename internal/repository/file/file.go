package file

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/KV2013/url-shortner-go/internal/model"
)

// type URLRepository interface {
// 	Save(url *model.URL) error
// 	GetByID(id string) (*model.URL, bool)
// }

type FileRepository struct {
	file   *os.File
	writer *bufio.Writer
}

func NewRepository(fileStoragePath string) (*FileRepository, error) {
	file, err := os.OpenFile(fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileRepository{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (r *FileRepository) GetByID(id string) (*model.URL, bool) {
	scanner := bufio.NewScanner(r.file)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, false
		}
		var url model.URL
		if err := json.Unmarshal(scanner.Bytes(), &url); err != nil {
			return nil, false
		}
		if url.Short == id {
			return &url, true
		}
	}

	return nil, false
}

func (r *FileRepository) Save(url *model.URL) error {
	jsonurl, err := json.Marshal(url)
	if err != nil {
		return err
	}
	if _, err = r.writer.Write(jsonurl); err != nil {
		return err
	}
	if err = r.writer.WriteByte('\n'); err != nil {
		return err
	}

	return r.writer.Flush()
}

func (r *FileRepository) Close() error {
	return r.file.Close()
}
