package file

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/KV2013/url-shortner-go/internal/model"
	"go.uber.org/zap"
)

// type URLRepository interface {
// 	Save(url *model.URL) error
// 	GetByID(id string) (*model.URL, bool)
// }

type FileRepository struct {
	file   *os.File
	writer *bufio.Writer
	logger *zap.Logger
}

func NewRepository(fileStoragePath string, logger *zap.Logger) (*FileRepository, error) {
	file, err := os.OpenFile(fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileRepository{
		file:   file,
		writer: bufio.NewWriter(file),
		logger: logger,
	}, nil
}

func (r *FileRepository) GetByID(id string) (*model.URL, bool) {
	r.file.Seek(0, 0)
	scanner := bufio.NewScanner(r.file)
	var url model.URL
	r.logger.Info("Ищу url по id", zap.String("id", id))
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			r.logger.Error("Ошибка чтения файла", zap.Error(err))
			return nil, false
		}
		if err := json.Unmarshal(scanner.Bytes(), &url); err != nil {
			r.logger.Error("Ошибка при распаковке JSON", zap.Error(err))
			return nil, false
		}
		if url.Short == id {
			r.logger.Debug("URL найден", zap.String("id", id), zap.String("url", url.Original))
			return &url, true
		}
	}

	r.logger.Info("Не найдено", zap.String("id", id))
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
