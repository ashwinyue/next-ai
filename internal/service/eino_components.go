package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/ashwinyue/next-ai/internal/service/file"
	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/retriever/es8"
	"github.com/cloudwego/eino-ext/components/retriever/es8/search_mode"
	einoembed "github.com/cloudwego/eino/components/embedding"
	ecomodel "github.com/cloudwego/eino/components/model"
	"github.com/elastic/go-elasticsearch/v8"
)

// newChatModel 创建 ChatModel
func newChatModel(ctx context.Context, cfg *config.Config) (ecomodel.ChatModel, error) {
	aiCfg := cfg.AI

	var apiKey, baseURL, modelName string

	switch aiCfg.Provider {
	case "openai":
		apiKey = aiCfg.OpenAI.APIKey
		baseURL = aiCfg.OpenAI.BaseURL
		modelName = aiCfg.OpenAI.Model
	case "alibaba", "qwen", "dashscope":
		apiKey = aiCfg.Alibaba.AccessKeySecret
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		modelName = aiCfg.Alibaba.Model
	case "deepseek":
		apiKey = aiCfg.DeepSeek.APIKey
		baseURL = aiCfg.DeepSeek.BaseURL
		modelName = aiCfg.DeepSeek.Model
	default:
		return nil, fmt.Errorf("unsupported ai provider: %s", aiCfg.Provider)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required for provider: %s", aiCfg.Provider)
	}

	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   modelName,
	})
}

// newToolCallingChatModel 创建支持工具调用的 ChatModel
func newToolCallingChatModel(ctx context.Context, cfg *config.Config) (ecomodel.ToolCallingChatModel, error) {
	aiCfg := cfg.AI

	var apiKey, baseURL, modelName string

	switch aiCfg.Provider {
	case "openai":
		apiKey = aiCfg.OpenAI.APIKey
		baseURL = aiCfg.OpenAI.BaseURL
		modelName = aiCfg.OpenAI.Model
	case "alibaba", "qwen", "dashscope":
		apiKey = aiCfg.Alibaba.AccessKeySecret
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		modelName = aiCfg.Alibaba.Model
	case "deepseek":
		apiKey = aiCfg.DeepSeek.APIKey
		baseURL = aiCfg.DeepSeek.BaseURL
		modelName = aiCfg.DeepSeek.Model
	default:
		return nil, fmt.Errorf("unsupported ai provider: %s", aiCfg.Provider)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required for provider: %s", aiCfg.Provider)
	}

	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	temperature := float32(0.7)

	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		Model:       modelName,
		Temperature: &temperature,
	})
}

// newEmbedder 创建 Embedding 器
func newEmbedder(ctx context.Context, cfg *config.Config) einoembed.Embedder {
	embCfg := cfg.AI.Embedding

	var apiKey, model string
	var timeout int

	switch embCfg.Provider {
	case "alibaba", "qwen", "dashscope", "":
		apiKey = embCfg.APIKey
		model = embCfg.Model
		timeout = embCfg.Timeout
	case "openai":
		apiKey = embCfg.APIKey
		model = embCfg.Model
		timeout = embCfg.Timeout
	default:
		log.Printf("Warning: unsupported embedding provider: %s", embCfg.Provider)
		return nil
	}

	if apiKey == "" {
		log.Printf("Warning: embedding api_key is empty")
		return nil
	}

	if model == "" {
		model = "text-embedding-v3"
	}

	embConfig := &dashscope.EmbeddingConfig{
		APIKey: apiKey,
		Model:  model,
	}

	if timeout > 0 {
		embConfig.Timeout = time.Duration(timeout) * time.Second
	}

	if embCfg.Dimensions > 0 {
		embConfig.Dimensions = &embCfg.Dimensions
	}

	embedder, err := dashscope.NewEmbedder(ctx, embConfig)
	if err != nil {
		log.Printf("Warning: failed to create embedder: %v", err)
		return nil
	}

	return embedder
}

// newES8Retriever 创建 ES8 检索器
// 返回 retriever, esClient, indexName
func newES8Retriever(ctx context.Context, cfg *config.Config, embedder einoembed.Embedder) (*es8.Retriever, *elasticsearch.Client, string) {
	esCfg := cfg.Elastic

	if esCfg.Host == "" {
		log.Printf("Warning: elasticsearch host not configured")
		return nil, nil, ""
	}

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esCfg.Host},
		Username:  esCfg.Username,
		Password:  esCfg.Password,
	})
	if err != nil {
		log.Printf("Warning: failed to create es client: %v", err)
		return nil, nil, ""
	}

	indexName := esCfg.IndexPrefix + "_chunks"

	retriever, err := es8.NewRetriever(ctx, &es8.RetrieverConfig{
		Client:     esClient,
		Index:      indexName,
		TopK:       10,
		SearchMode: search_mode.SearchModeDenseVectorSimilarity(search_mode.DenseVectorSimilarityTypeCosineSimilarity, "content_vector"),
		Embedding:  embedder,
	})
	if err != nil {
		log.Printf("Warning: failed to create retriever: %v", err)
		return nil, esClient, indexName
	}

	return retriever, esClient, indexName
}

// newFileService 创建文件存储服务
func newFileService(repo *repository.Repositories, cfg *config.Config) *file.Service {
	// 默认使用本地存储
	storageType := file.StorageTypeLocal
	fileCfg := make(map[string]string)

	// 从配置中读取文件存储配置
	if cfg.File != nil {
		switch cfg.File.Type {
		case "minio":
			storageType = file.StorageTypeMinIO
			fileCfg = map[string]string{
				"endpoint":   cfg.File.MinIO.Endpoint,
				"access_key": cfg.File.MinIO.AccessKey,
				"secret_key": cfg.File.MinIO.SecretKey,
				"bucket":     cfg.File.MinIO.Bucket,
				"use_ssl":    cfg.File.MinIO.UseSSL,
				"url_prefix": cfg.File.MinIO.URLPrefix,
			}
		case "local":
			storageType = file.StorageTypeLocal
			fileCfg = map[string]string{
				"base_path":  cfg.File.Local.BasePath,
				"url_prefix": cfg.File.Local.URLPrefix,
			}
		}
	}

	// 使用默认本地配置
	if len(fileCfg) == 0 {
		fileCfg = map[string]string{
			"base_path":  "./data/files",
			"url_prefix": "/files",
		}
	}

	fileSvc, err := file.NewServiceFromConfig(repo, storageType, fileCfg)
	if err != nil {
		log.Printf("Warning: failed to create file service: %v, using nil", err)
		return nil
	}

	log.Printf("File service initialized with type: %s", storageType)
	return fileSvc
}
