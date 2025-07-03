package database

import (
	"server_management_system/pkg/config"

	"github.com/elastic/go-elasticsearch/v9"
	"go.uber.org/zap"
)

func LoadElasticSearch(cfg config.ElasticSearch, log *zap.Logger) (*elasticsearch.Client, func(), error) {
	esConfig := elasticsearch.Config{
		Addresses: []string{cfg.URL},
	}

	ESClient, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		log.Error("Error creating Elasticsearch client", zap.Error(err))
		return nil, nil, err
	}

	req, err := ESClient.Ping()
	if err != nil {
		log.Error("Error pinging Elasticsearch", zap.Error(err))
		return nil, nil, err
	}

	cleanup := func() {

	}

	log.Info("Elasticsearch connected successfully",
		zap.String("url", cfg.URL),
		zap.Int("code", req.StatusCode),
	)
	return ESClient, cleanup, err
}
