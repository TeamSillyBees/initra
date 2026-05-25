package httpclient

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Factory 按服务名创建并缓存 HTTP Client。
type Factory interface {
	Get(serviceName string) (*Client, error)
	Clear(serviceName string)
	ClearAll()
}

type factory struct {
	cfg     Config
	logger  *zap.Logger
	mu      sync.Mutex
	clients map[string]*Client
}

// NewFactory 根据配置创建 HTTP Client 工厂。
func NewFactory(cfg Config, logger *zap.Logger) (Factory, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	normalized := cfg.withDefaults()
	if err := normalized.Validate(); err != nil {
		return nil, err
	}
	return &factory{
		cfg:     normalized,
		logger:  logger,
		clients: make(map[string]*Client),
	}, nil
}

func (f *factory) Get(serviceName string) (*Client, error) {
	if !f.cfg.Enabled {
		return nil, fmt.Errorf("%w: http_client.enabled=false", ErrDisabled)
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	if client, ok := f.clients[serviceName]; ok {
		return client, nil
	}

	serviceCfg, ok := f.cfg.Services[serviceName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrServiceNotFound, serviceName)
	}
	client, err := newClient(serviceName, f.cfg, serviceCfg.withDefaults(f.cfg), f.logger)
	if err != nil {
		return nil, err
	}
	f.clients[serviceName] = client
	return client, nil
}

func (f *factory) Clear(serviceName string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if client, ok := f.clients[serviceName]; ok {
		client.closeIdleConnections()
		delete(f.clients, serviceName)
	}
}

func (f *factory) ClearAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for name, client := range f.clients {
		client.closeIdleConnections()
		delete(f.clients, name)
	}
}

func newRestyClient(global Config, service ServiceConfig) *resty.Client {
	transport := &http.Transport{
		Proxy:               proxyFunc(service.Proxy),
		MaxIdleConns:        global.MaxIdleConns,
		MaxIdleConnsPerHost: global.MaxIdleConnsPerHost,
		IdleConnTimeout:     global.IdleConnTimeout,
	}
	restyClient := resty.New().
		SetBaseURL(service.BaseURL).
		SetTimeout(service.Timeout).
		SetHeaders(service.Headers).
		SetTransport(transport).
		SetResponseBodyLimit(int(service.MaxResponseBodySize))

	if global.ConnectTimeout > 0 {
		dialer := &net.Dialer{Timeout: global.ConnectTimeout}
		transport.DialContext = dialer.DialContext
	}
	if service.Retry.Enabled && service.Retry.Count > 0 {
		restyClient.
			SetRetryCount(service.Retry.Count).
			SetRetryWaitTime(service.Retry.WaitTime).
			SetRetryMaxWaitTime(service.Retry.MaxWaitTime).
			AddRetryCondition(retryCondition(service.Retry))
	}
	return restyClient
}

func proxyFunc(proxy string) func(*http.Request) (*url.URL, error) {
	proxy = strings.TrimSpace(proxy)
	if proxy == "" {
		return http.ProxyFromEnvironment
	}
	parsed, err := url.Parse(proxy)
	if err != nil {
		return http.ProxyFromEnvironment
	}
	return http.ProxyURL(parsed)
}
