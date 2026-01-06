package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Elastic  ElasticConfig
	AI       AIConfig
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string
	Environment string
	Version     string
	Debug       bool
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host        string
	Port        int
	Mode        string
	ReadTimeout int
	WriteTimeout int
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  int
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// ElasticConfig Elasticsearch配置
type ElasticConfig struct {
	Host        string
	Username    string
	Password    string
	IndexPrefix string
}

// AIConfig AI配置
type AIConfig struct {
	Provider  string
	OpenAI    OpenAIConfig
	Alibaba   AlibabaConfig
	DeepSeek  DeepSeekConfig
	Embedding EmbeddingConfig
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout int
}

// AlibabaConfig 阿里云配置
type AlibabaConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	Model           string
	Timeout         int
}

// DeepSeekConfig DeepSeek配置
type DeepSeekConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout int
}

// EmbeddingConfig Embedding配置
type EmbeddingConfig struct {
	Provider   string
	Model      string
	APIKey     string
	BaseURL    string
	Timeout    int
	Dimensions int
}

var globalConfig *Config

// Load 加载配置
func Load(path string) (*Config, error) {
	v := viper.New()

	if path != "" {
		v.SetConfigFile(path)
		v.SetConfigType("yaml")
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	} else {
		// 默认配置
		setDefaults(v)
	}

	// 环境变量
	v.SetEnvPrefix("NEXT_AI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("config not loaded")
	}
	return globalConfig
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetAddr 获取服务器地址
func (c *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetAddr 获取 Redis 地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func setDefaults(v *viper.Viper) {
	// App
	v.SetDefault("app.name", "next-ai")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.debug", true)

	// Server
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.readTimeout", 30)
	v.SetDefault("server.writeTimeout", 30)

	// Database
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.dbname", "next_ai")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.maxOpenConns", 25)
	v.SetDefault("database.maxIdleConns", 5)
	v.SetDefault("database.maxLifetime", 300)

	// Redis
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)

	// Elastic
	v.SetDefault("elastic.host", "http://localhost:9200")
	v.SetDefault("elastic.indexPrefix", "next_ai")

	// AI
	v.SetDefault("ai.provider", "openai")
	v.SetDefault("ai.openai.baseUrl", "https://api.openai.com/v1")
	v.SetDefault("ai.openai.model", "gpt-4o-mini")
}
