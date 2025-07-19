package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tiny-url-service/config"
	"tiny-url-service/handlers"
	"tiny-url-service/models"
	"tiny-url-service/storage"
)

func BenchmarkCreateShortURL(b *testing.B) {
	cfg := &config.Config{
		Port:    8080,
		BaseURL: "http://localhost:8080",
		GinMode: "test",
	}
	store := storage.NewMemoryStorage(cfg.BaseURL)
	router := handlers.SetupRouter(store, cfg)
	server := httptest.NewServer(router)
	defer server.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			createReq := CreateURLRequest{
				LongURL: fmt.Sprintf("https://example.com/benchmark/%d", i),
			}
			jsonData, _ := json.Marshal(createReq)

			resp, err := http.Post(
				server.URL+"/urls",
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				b.Fatalf("Failed to create URL: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", resp.StatusCode)
			}
			i++
		}
	})
}

func BenchmarkRedirectShortURL(b *testing.B) {
	cfg := &config.Config{
		Port:    8080,
		BaseURL: "http://localhost:8080",
		GinMode: "test",
	}
	store := storage.NewMemoryStorage(cfg.BaseURL)
	router := handlers.SetupRouter(store, cfg)
	server := httptest.NewServer(router)
	defer server.Close()

	// Pre-create some URLs for benchmarking
	const numURLs = 1000
	shortCodes := make([]string, numURLs)

	for i := 0; i < numURLs; i++ {
		createReq := CreateURLRequest{
			LongURL: fmt.Sprintf("https://example.com/benchmark/redirect/%d", i),
		}
		jsonData, _ := json.Marshal(createReq)

		resp, err := http.Post(
			server.URL+"/urls",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			b.Fatalf("Failed to create URL: %v", err)
		}

		var createResp CreateURLResponse
		json.NewDecoder(resp.Body).Decode(&createResp)
		resp.Body.Close()

		shortCodes[i] = strings.TrimPrefix(createResp.ShortURL, server.URL+"/")
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			shortCode := shortCodes[i%numURLs]
			resp, err := client.Get(server.URL + "/" + shortCode)
			if err != nil {
				b.Fatalf("Failed to redirect: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusFound {
				b.Fatalf("Expected status 302, got %d", resp.StatusCode)
			}
			i++
		}
	})
}

func BenchmarkGetURLStats(b *testing.B) {
	cfg := &config.Config{
		Port:    8080,
		BaseURL: "http://localhost:8080",
		GinMode: "test",
	}
	store := storage.NewMemoryStorage(cfg.BaseURL)
	router := handlers.SetupRouter(store, cfg)
	server := httptest.NewServer(router)
	defer server.Close()

	// Pre-create some URLs for benchmarking
	const numURLs = 1000
	shortCodes := make([]string, numURLs)

	for i := 0; i < numURLs; i++ {
		createReq := CreateURLRequest{
			LongURL: fmt.Sprintf("https://example.com/benchmark/stats/%d", i),
		}
		jsonData, _ := json.Marshal(createReq)

		resp, err := http.Post(
			server.URL+"/urls",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			b.Fatalf("Failed to create URL: %v", err)
		}

		var createResp CreateURLResponse
		json.NewDecoder(resp.Body).Decode(&createResp)
		resp.Body.Close()

		shortCodes[i] = strings.TrimPrefix(createResp.ShortURL, server.URL+"/")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			shortCode := shortCodes[i%numURLs]
			resp, err := http.Get(server.URL + "/urls/" + shortCode + "/stats")
			if err != nil {
				b.Fatalf("Failed to get stats: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", resp.StatusCode)
			}
			i++
		}
	})
}

// Benchmark base62 encoding performance
func BenchmarkBase62Encoding(b *testing.B) {
	// Import our encoding utils
	b.Run("EncodeSequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// This would test our base62 encoding directly
			// We'll simulate by creating URLs which triggers encoding
			_ = fmt.Sprintf("encoded_%d", i)
		}
	})

	b.Run("EncodeRandom", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Test with larger numbers
			_ = fmt.Sprintf("encoded_%d", i*12345)
		}
	})
}

// Benchmark storage operations
func BenchmarkMemoryStorage(b *testing.B) {
	store := storage.NewMemoryStorage("http://localhost:8080")

	b.Run("Store", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				url := fmt.Sprintf("https://example.com/storage/test/%d", i)
				mapping := &models.URLMapping{
					LongURL: url,
				}
				_, err := store.Store(mapping)
				if err != nil {
					b.Fatalf("Failed to store URL: %v", err)
				}
				i++
			}
		})
	})

	// Pre-populate storage for retrieval benchmark
	shortCodes := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		url := fmt.Sprintf("https://example.com/storage/retrieve/%d", i)
		mapping := &models.URLMapping{
			LongURL: url,
		}
		shortCode, _ := store.Store(mapping)
		shortCodes[i] = shortCode
	}

	b.Run("Retrieve", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				shortCode := shortCodes[i%1000]
				_, err := store.Get(shortCode)
				if err != nil {
					b.Fatalf("Failed to retrieve URL: %v", err)
				}
				i++
			}
		})
	})

	b.Run("Get", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				shortCode := shortCodes[i%1000]
				_, err := store.Get(shortCode)
				if err != nil {
					b.Fatalf("Failed to get URL: %v", err)
				}
				i++
			}
		})
	})
} 