package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/th1enq/server_management_system/internal/configs"
	"go.uber.org/zap"
)

type IESClient interface {
	Exec(ctx context.Context, indexName string, query map[string]interface{}, dest interface{}) error
	Insert(ctx context.Context, indexName, documentID string, document map[string]interface{}) error
}

type esClient struct {
	es     *elasticsearch.Client
	logger *zap.Logger
}

func (e *esClient) Insert(ctx context.Context, indexName, documentID string, document map[string]interface{}) error {
	newDocumentLine, err := json.Marshal(document)
	if err != nil {
		e.logger.Error("failed to marsal document", zap.Error(err))
		return fmt.Errorf("failed to marsal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       bytes.NewReader(newDocumentLine),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, e.es)
	if err != nil {
		e.logger.Error("Failed to index document", zap.Error(err))
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		e.logger.Error("Error indexing  document", zap.String("index", documentID))
		return fmt.Errorf("error indexing document: %s", res.Status())
	}
	e.logger.Info("Successfully indexed document", zap.Any("document", document))
	return nil
}

func (e *esClient) Exec(ctx context.Context, indexName string, query map[string]interface{}, dest interface{}) error {
	queryBody, err := json.Marshal(query)
	if err != nil {
		e.logger.Error("Failed to marshal query", zap.Error(err))
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{indexName},
		Body:  strings.NewReader(string(queryBody)),
	}
	res, err := req.Do(ctx, e.es)
	if err != nil {
		e.logger.Error("Failed to execute search request", zap.Error(err))
		return fmt.Errorf("failed to execute search request: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		e.logger.Error("Failed to read response body", zap.Error(err))
		return fmt.Errorf("failed to read response body: %w", err)
	}

	e.logger.Info("Search response body", zap.ByteString("body", bodyBytes))

	if err := json.Unmarshal(bodyBytes, &dest); err != nil {
		e.logger.Error("Failed to unmarshal response body", zap.Error(err))
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return nil
}

func LoadElasticSearch(cfg configs.ElasticSearch, logger *zap.Logger) (IESClient, error) {
	esConfig := elasticsearch.Config{
		Addresses: []string{cfg.URL},
	}

	ESClient, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		logger.Error("Error creating Elasticsearch client", zap.Error(err))
	}

	req, err := ESClient.Ping()
	if err != nil {
		return nil, fmt.Errorf("cannot try connect to ES: %w", err)
	}

	logger.Info("Elasticsearch connected successfully",
		zap.String("url", cfg.URL),
		zap.Int("code", req.StatusCode),
	)
	return &esClient{
		es:     ESClient,
		logger: logger,
	}, nil
}
