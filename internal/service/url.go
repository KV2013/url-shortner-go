package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/KV2013/url-shortner-go/internal/model"
	"github.com/KV2013/url-shortner-go/internal/service/random"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type URLRepository interface {
	Save(ctx context.Context, url *model.URL) error
	SaveMany(ctx context.Context, urls []*model.URL) error
	GetByID(ctx context.Context, id string) (*model.URL, bool)
	GetAllByUserID(ctx context.Context, userID string) ([]model.URL, error)
	DeleteUserURLs(ctx context.Context, userID string, urls []string) error
}

type DeleteJob struct {
	id     string
	urls   []string
	userID string
}

type URLService struct {
	urlRepository URLRepository
	delCh         chan DeleteJob
	logger        *zap.Logger
}

func NewURLService(urlRepository URLRepository, logger *zap.Logger) *URLService {
	service := &URLService{
		urlRepository: urlRepository,
		delCh:         make(chan DeleteJob, 1024),
		logger:        logger,
	}

	go service.startDeleteQueue()

	return service
}

func (s *URLService) SaveURL(ctx context.Context, url string, userID string) (*model.URL, error) {
	newID := random.NewRandomString(10)
	newURL := &model.URL{
		Original: url,
		Short:    newID,
		UserID:   userID,
	}
	err := s.urlRepository.Save(ctx, newURL)
	if err != nil {
		return nil, err
	}

	return newURL, nil
}

func (s *URLService) SaveManyURL(ctx context.Context, urls []string, userID string) ([]model.URL, error) {
	newURLs := make([]*model.URL, 0, len(urls))
	for _, url := range urls {
		newURLs = append(newURLs, &model.URL{
			Original: url,
			Short:    random.NewRandomString(10),
			UserID:   userID,
		})
	}

	if err := s.urlRepository.SaveMany(ctx, newURLs); err != nil {
		return nil, err
	}

	result := make([]model.URL, 0, len(newURLs))
	for _, u := range newURLs {
		result = append(result, *u)
	}
	return result, nil
}

func (s *URLService) GetAllByUserID(ctx context.Context, userID string) ([]model.URL, error) {
	return s.urlRepository.GetAllByUserID(ctx, userID)
}

func (s *URLService) GetByID(ctx context.Context, id string) (*model.URL, error) {
	url, exists := s.urlRepository.GetByID(ctx, id)
	if !exists {
		return nil, &model.ErrURLNotFound{Short: id}
	}

	if url.DeletedAt != nil {
		return nil, &model.ErrURLDeleted{Short: id}
	}

	return url, nil
}

func (s *URLService) DeleteURLs(ctx context.Context, shortUrls []string, userID string) error {

	if len(shortUrls) == 0 {
		return errors.New("получен пустой массив ссылок")
	}
	job := DeleteJob{
		userID: userID,
		urls:   shortUrls,
		id:     uuid.New().String(),
	}

	s.delCh <- job

	return nil
}

func (s *URLService) startDeleteQueue() {
	ticker := time.NewTicker(3 * time.Second)
	var jobs []DeleteJob

	s.logger.Info("URLService: запуск очереди на удаление URL")
	for {
		select {
		case job := <-s.delCh:
			jobs = append(jobs, job)
			s.logger.Debug(
				"URLService: добавлена задача на удаление",
				zap.String("UserID", job.userID),
				zap.String("JobID", job.id),
				zap.String("urls", strings.Join(job.urls, ",")),
				zap.Int("jobsCount", len(jobs)),
			)
		case <-ticker.C:
			if len(jobs) == 0 {
				continue
			}
			s.logger.Debug("URLService: запуск удаления URL", zap.Int("count", len(jobs)))
			var failedJobs []DeleteJob
			for _, job := range jobs {
				err := s.urlRepository.DeleteUserURLs(context.TODO(), job.userID, job.urls)
				if err != nil {
					s.logger.Error(
						"URLService: ошибка при удалении URL",
						zap.Error(err),
						zap.String("UserID", job.userID),
						zap.String("JobID", job.id),
						zap.String("urls", strings.Join(job.urls, ",")),
					)
					failedJobs = append(failedJobs, job)
				}
			}
			jobs = failedJobs
			s.logger.Debug("URLService: процедура удаления завершена", zap.Int("failedJobsCount", len(failedJobs)))
		}
	}
}
