package database

import (
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/th1enq/server_management_system/internal/config"
	"github.com/th1enq/server_management_system/pkg/logger"
	"go.uber.org/zap"
)

var ESClient *elasticsearch.Client

func InitElasticsearchClient(cfg *config.Config) error {

	esConfig := elasticsearch.Config{
		Addresses: []string{cfg.Elasticsearch.URL},
	}

	var err error
	ESClient, err = elasticsearch.NewClient(esConfig)
	if err != nil {
		logger.Error("Error creating Elasticsearch client", err)
	}

	req, err := ESClient.Ping()

	logger.Info("Elasticsearch connected successfully",
		zap.String("url", cfg.Elasticsearch.URL),
		zap.Int("code", req.StatusCode),
	)
	return nil
}
