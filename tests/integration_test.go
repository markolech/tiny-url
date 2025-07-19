package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"tiny-url-service/config"
	"tiny-url-service/handlers"
	"tiny-url-service/storage"
)

// Test data structures
type CreateURLRequest struct {
	LongURL        string `json:"long_url"`
	ExpirationDate string `json:"expiration_date,omitempty"`
}

type CreateURLResponse struct {
	ShortURL string `json:"short_url"`
}

type URLStats struct {
	ShortCode   string    `json:"short_code"`
	LongURL     string    `json:"long_url"`
	AccessCount int       `json:"access_count"`
	CreatedAt   time.Time `json:"created_at"`
}

func setupTestServer() *httptest.Server {
	server := httptest.NewServer(nil)
	
	cfg := &config.Config{
		Port:    8080,
		BaseURL: server.URL,
		GinMode: "test",
	}
	
	store := storage.NewMemoryStorage(cfg.BaseURL)
	router := handlers.SetupRouter(store, cfg)
	server.Config.Handler = router
	
	return server
}

func TestCreateShortURL(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	tests := []struct {
		name           string
		request        CreateURLRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid HTTP URL",
			request:        CreateURLRequest{LongURL: "http://example.com"},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Valid HTTPS URL",
			request:        CreateURLRequest{LongURL: "https://www.google.com"},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Valid URL with path",
			request:        CreateURLRequest{LongURL: "https://example.com/path/to/page?query=value"},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid URL - no protocol",
			request:        CreateURLRequest{LongURL: "example.com"},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Invalid URL - wrong protocol",
			request:        CreateURLRequest{LongURL: "ftp://example.com"},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Empty URL",
			request:        CreateURLRequest{LongURL: ""},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.request)
			resp, err := http.Post(
				server.URL+"/urls",
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if !tt.expectError {
				var response CreateURLResponse
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.ShortURL == "" {
					t.Error("Expected short_url in response")
				}

				if !strings.HasPrefix(response.ShortURL, server.URL) {
					t.Errorf("Short URL should start with %s, got %s", server.URL, response.ShortURL)
				}
			}
		})
	}
}

func TestRedirectToLongURL(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// First create a short URL
	createReq := CreateURLRequest{LongURL: "https://www.example.com/test"}
	jsonData, _ := json.Marshal(createReq)
	resp, err := http.Post(
		server.URL+"/urls",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Failed to create short URL: %v", err)
	}
	defer resp.Body.Close()

	var createResp CreateURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	// Extract short code from URL
	shortCode := strings.TrimPrefix(createResp.ShortURL, server.URL+"/")

	// Test redirect
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	resp, err = client.Get(server.URL + "/" + shortCode)
	if err != nil {
		t.Fatalf("Failed to make redirect request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location != createReq.LongURL {
		t.Errorf("Expected location %s, got %s", createReq.LongURL, location)
	}
}

func TestGetURLStats(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create a short URL
	createReq := CreateURLRequest{LongURL: "https://www.example.com/stats-test"}
	jsonData, _ := json.Marshal(createReq)
	resp, err := http.Post(
		server.URL+"/urls",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Failed to create short URL: %v", err)
	}
	defer resp.Body.Close()

	var createResp CreateURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	shortCode := strings.TrimPrefix(createResp.ShortURL, server.URL+"/")

	// Get initial stats
	resp, err = http.Get(server.URL + "/urls/" + shortCode + "/stats")
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var stats URLStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode stats response: %v", err)
	}

	if stats.ShortCode != shortCode {
		t.Errorf("Expected short_code %s, got %s", shortCode, stats.ShortCode)
	}

	if stats.LongURL != createReq.LongURL {
		t.Errorf("Expected long_url %s, got %s", createReq.LongURL, stats.LongURL)
	}

	if stats.AccessCount != 0 {
		t.Errorf("Expected access_count 0, got %d", stats.AccessCount)
	}
}

func TestHealthCheck(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to get health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestErrorCases(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	tests := []struct {
		name           string
		method         string
		url            string
		contentType    string
		body           string
		expectedStatus int
	}{
		{
			name:           "Missing Content-Type",
			method:         "POST",
			url:            "/urls",
			contentType:    "",
			body:           `{"long_url": "https://example.com"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Wrong Content-Type",
			method:         "POST",
			url:            "/urls",
			contentType:    "text/plain",
			body:           `{"long_url": "https://example.com"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			method:         "POST",
			url:            "/urls",
			contentType:    "application/json",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent short code",
			method:         "GET",
			url:            "/nonexistent",
			contentType:    "",
			body:           "",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Non-existent stats",
			method:         "GET",
			url:            "/urls/nonexistent/stats",
			contentType:    "",
			body:           "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			if tt.method == "POST" {
				req, _ := http.NewRequest("POST", server.URL+tt.url, strings.NewReader(tt.body))
				if tt.contentType != "" {
					req.Header.Set("Content-Type", tt.contentType)
				}
				resp, err = http.DefaultClient.Do(req)
			} else {
				resp, err = http.Get(server.URL + tt.url)
			}

			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	const numRequests = 50
	var wg sync.WaitGroup
	results := make(chan CreateURLResponse, numRequests)
	errors := make(chan error, numRequests)

	// Test concurrent URL creation
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			createReq := CreateURLRequest{
				LongURL: fmt.Sprintf("https://example.com/concurrent/%d", id),
			}
			jsonData, _ := json.Marshal(createReq)

			resp, err := http.Post(
				server.URL+"/urls",
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
				return
			}

			var response CreateURLResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				errors <- err
				return
			}

			results <- response
		}(i)
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	var errorCount int
	for err := range errors {
		t.Errorf("Concurrent request error: %v", err)
		errorCount++
	}

	// Check results
	var resultCount int
	shortCodes := make(map[string]bool)
	for result := range results {
		resultCount++
		shortCode := strings.TrimPrefix(result.ShortURL, server.URL+"/")
		if shortCodes[shortCode] {
			t.Errorf("Duplicate short code detected: %s", shortCode)
		}
		shortCodes[shortCode] = true
	}

	if resultCount != numRequests {
		t.Errorf("Expected %d results, got %d", numRequests, resultCount)
	}

	if len(shortCodes) != numRequests {
		t.Errorf("Expected %d unique short codes, got %d", numRequests, len(shortCodes))
	}

	t.Logf("Successfully created %d unique short URLs concurrently", len(shortCodes))
}

func TestConcurrentRedirects(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create a URL first
	createReq := CreateURLRequest{LongURL: "https://example.com/concurrent-redirect"}
	jsonData, _ := json.Marshal(createReq)
	resp, err := http.Post(
		server.URL+"/urls",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Failed to create short URL: %v", err)
	}
	defer resp.Body.Close()

	var createResp CreateURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	shortCode := strings.TrimPrefix(createResp.ShortURL, server.URL+"/")

	// Test concurrent access to the same URL
	const numRequests = 30
	var wg sync.WaitGroup
	successCount := make(chan bool, numRequests)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, err := client.Get(server.URL + "/" + shortCode)
			if err != nil {
				t.Errorf("Failed to make redirect request: %v", err)
				successCount <- false
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusFound {
				successCount <- true
			} else {
				successCount <- false
			}
		}()
	}

	wg.Wait()
	close(successCount)

	var successful int
	for success := range successCount {
		if success {
			successful++
		}
	}

	if successful != numRequests {
		t.Errorf("Expected %d successful redirects, got %d", numRequests, successful)
	}

	t.Logf("Successfully handled %d concurrent redirects", successful)
} 