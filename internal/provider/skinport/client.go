package skinport

import (
	"context"
	"encoding/json"
	"fmt"
	"go-test-system/internal/config"
	"go-test-system/internal/domain"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/andybalholm/brotli"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
	logger     *slog.Logger
}

func NewClient(cfg *config.Config, httpClient *http.Client, logger *slog.Logger) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: httpClient,
		logger:     logger,
	}
}

func (client *Client) boolToParam(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (client *Client) FetchItems(ctx context.Context, tradable bool) ([]domain.ItemResponse, error) {
	u, err := url.Parse(client.cfg.SkinportUrl + "/v1/items")
	if err != nil {
		return nil, fmt.Errorf("skinport fetch items url: %w", err)
	}

	q := u.Query()
	q.Set("app_id", client.cfg.SkinportAppID)
	q.Set("currency", client.cfg.SkinportCurrency)
	q.Set("tradable", client.boolToParam(tradable))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "br")

	client.logger.Debug("requesting skinport api",
		slog.String("url", u.String()),
		slog.Bool("tradable", tradable),
	)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("skinport do request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("skinport returned status: %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "br") {
		client.logger.Debug("decompressing brotli response")
		reader = brotli.NewReader(resp.Body)
	}

	var items []domain.ItemResponse
	if err := json.NewDecoder(reader).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode skinport items: %w", err)
	}

	return items, nil
}
