package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/th1enq/server_management_system/internal/configs"
	"go.uber.org/zap"
)

func LoadElasticSearch(cfg *configs.ElasticSearch, logger *zap.Logger) (*elasticsearch.Client, func(), error) {
	esConfig := elasticsearch.Config{
		Addresses: []string{cfg.URL},
	}

	ESClient, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		logger.Error("Error creating Elasticsearch client", zap.Error(err))
	}

	req, err := ESClient.Ping()

	cleanup := func() {

	}

	logger.Info("Elasticsearch connected successfully",
		zap.String("url", cfg.URL),
		zap.Int("code", req.StatusCode),
	)
	return ESClient, cleanup, err
}
