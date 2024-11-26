package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"jason-short-server/tools"
	"log"
	"net/http"
	"time"
)

type Config struct {
	Base struct {
		Website   string `mapstructure:"website"`   // 服务启动的端口号
		Port      string `mapstructure:"port"`      // 短连接服务的 对外域名，需要加上 http:// 或 https:// 末尾不需要加 /
		Length    int    `mapstructure:"length"`    // 短连接后边字符的长度
		CacheTime int    `mapstructure:"cacheTime"` // 缓存时间，分钟
	} `mapstructure:"base"`
	Redis struct {
		Addr string `mapstructure:"addr"` // redis 缓存地址
		Port string `mapstructure:"port"` // redis 端口号
		Pwd  string `mapstructure:"pwd"`  // redis 密码
	} `mapstructure:"redis"`
}

var (
	conf     *Config
	cacheMap *tools.AutoDeleteMap
)

type ShortLinkService struct {
	redisClient *redis.Client
	baseURL     string
}

func NewShortLinkService(baseURL string, redisAddr, redisPwd string) *ShortLinkService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPwd,
	})
	return &ShortLinkService{
		redisClient: rdb,
		baseURL:     baseURL,
	}
}

// Shorten handles long URL to short URL conversion
func (s *ShortLinkService) Shorten(c *gin.Context) {
	var req struct {
		LongURL   string `json:"long_url" binding:"required"`
		DiyDomain string `json:"diy_domain"` // 自定义域名
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortCode := generateShortCode(conf.Base.Length)
	ctx := context.Background()

	// Store the mapping in Redis
	err := s.redisClient.Set(ctx, fmt.Sprintf("short:short:%s", shortCode), req.LongURL, redis.KeepTTL).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save short URL"})
		return
	}
	// 将 redis 数据 添加到内存
	cacheMap.Set(shortCode, req.LongURL, time.Duration(conf.Base.CacheTime)*time.Minute)

	resUrl := ""
	if req.DiyDomain == "" {
		resUrl = s.baseURL + "/" + shortCode
	} else {
		resUrl = req.DiyDomain + "/" + shortCode
	}

	c.JSON(http.StatusOK, gin.H{
		"short_url": resUrl,
	})
}

// Redirect handles short URL redirection to long URL
func (s *ShortLinkService) Redirect(c *gin.Context) {
	shortCode := c.Param("shortCode")
	ctx := context.Background()
	longURL := fmt.Sprintf("%v/%v", conf.Base.Website, shortCode) // 默认值
	// 先查询当前内存，没有查询 redis
	if v, ok := cacheMap.Load(shortCode); ok {
		longURL = v.(string)
	} else {
		// Get the long URL from Redis
		longURL, _ = s.redisClient.Get(ctx, fmt.Sprintf("short:short:%s", shortCode)).Result()

		// 将 redis 数据 添加到内存
		cacheMap.Set(shortCode, longURL, time.Duration(conf.Base.CacheTime)*time.Minute)
	}
	// Increment visit count
	_ = s.redisClient.Incr(ctx, fmt.Sprintf("short:stats:%s", shortCode)).Err()

	c.Redirect(http.StatusMovedPermanently, longURL)
}

// Stats provides statistics for a given short URL
func (s *ShortLinkService) Stats(c *gin.Context) {
	shortCode := c.Param("shortCode")
	ctx := context.Background()

	// Get the long URL from Redis
	longURL, err := s.redisClient.Get(ctx, fmt.Sprintf("short:short:%s", shortCode)).Result()
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "short URL not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve short URL"})
		return
	}

	// Get visit count
	visitCount, err := s.redisClient.Get(ctx, fmt.Sprintf("stats:%s", shortCode)).Result()
	if err == redis.Nil {
		visitCount = "0"
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_url":   s.baseURL + "/" + shortCode,
		"long_url":    longURL,
		"visit_count": visitCount,
	})
}

// creates a unique short code
func generateShortCode(count int) string {
	return uuid.New().String()[:count] // Use the first 8 characters of a UUID
}

// 初始化配置信息
func initConf() {
	// 从当前目录读取 config.yaml 文件，并且放入全局变量中
	conf = new(Config)
	// 初始化 Viper
	viper.SetConfigName("config") // 配置文件名称（不带扩展名）
	viper.SetConfigType("yaml")   // 配置文件类型
	viper.AddConfigPath(".")      // 配置文件路径（当前目录）

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(conf); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
}

func main() {
	initConf()
	fmt.Println(conf)
	cacheMap = new(tools.AutoDeleteMap)
	r := gin.Default()
	service := NewShortLinkService(conf.Base.Website, fmt.Sprintf("%v:%v", conf.Redis.Addr, conf.Redis.Port), conf.Redis.Pwd)

	r.POST("/shorten", service.Shorten)
	r.GET("/:shortCode", service.Redirect)
	r.GET("/stats/:shortCode", service.Stats)

	r.Run(fmt.Sprintf(":%v", conf.Base.Port))
}
