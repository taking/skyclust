package sse

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
)

const (
	// CompressionThreshold is the minimum size in bytes to compress SSE messages
	CompressionThreshold = 1024 // 1KB

	// CompressionEnabled indicates if compression is enabled
	CompressionEnabled = true
)

// CompressMessage compresses a JSON message if it exceeds the threshold
func CompressMessage(jsonData []byte) ([]byte, bool, error) {
	if !CompressionEnabled || len(jsonData) < CompressionThreshold {
		return jsonData, false, nil
	}

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	if _, err := writer.Write(jsonData); err != nil {
		writer.Close()
		return nil, false, fmt.Errorf("failed to write to gzip writer: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, false, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	compressed := buf.Bytes()

	// 압축이 효과적이지 않은 경우 원본 반환
	if len(compressed) >= len(jsonData) {
		return jsonData, false, nil
	}

	// Base64 인코딩하여 문자열로 변환
	encoded := base64.StdEncoding.EncodeToString(compressed)
	return []byte(encoded), true, nil
}

// DecompressMessage decompresses a base64-encoded gzip message
func DecompressMessage(compressedData []byte) ([]byte, error) {
	// Base64 디코딩
	decoded, err := base64.StdEncoding.DecodeString(string(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Gzip 압축 해제
	reader, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from gzip reader: %w", err)
	}

	return decompressed, nil
}
