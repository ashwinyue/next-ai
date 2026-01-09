package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/ashwinyue/next-ai/internal/service/file"
	"github.com/cloudwego/eino-ext/components/model/openai"
	ecomodel "github.com/cloudwego/eino/components/model"
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
