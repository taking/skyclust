package messaging

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/golang/snappy"
)

// CompressionType represents the compression algorithm to use
type CompressionType string

const (
	CompressionNone   CompressionType = "none"
	CompressionGzip   CompressionType = "gzip"
	CompressionSnappy CompressionType = "snappy"
)

// Compress compresses data using the specified compression type
func Compress(data []byte, compressionType CompressionType) ([]byte, error) {
	switch compressionType {
	case CompressionNone:
		return data, nil
	case CompressionGzip:
		return compressGzip(data)
	case CompressionSnappy:
		return compressSnappy(data)
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", compressionType)
	}
}

// Decompress decompresses data using the specified compression type
func Decompress(data []byte, compressionType CompressionType) ([]byte, error) {
	switch compressionType {
	case CompressionNone:
		return data, nil
	case CompressionGzip:
		return decompressGzip(data)
	case CompressionSnappy:
		return decompressSnappy(data)
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", compressionType)
	}
}

// compressGzip compresses data using gzip
func compressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to write to gzip writer: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// decompressGzip decompresses gzip-compressed data
func decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
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

// compressSnappy compresses data using snappy
func compressSnappy(data []byte) ([]byte, error) {
	return snappy.Encode(nil, data), nil
}

// decompressSnappy decompresses snappy-compressed data
func decompressSnappy(data []byte) ([]byte, error) {
	decompressed, err := snappy.Decode(nil, data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode snappy data: %w", err)
	}

	return decompressed, nil
}

// ShouldCompress determines if data should be compressed based on size threshold
func ShouldCompress(data []byte, threshold int, compressionType CompressionType) bool {
	if compressionType == CompressionNone {
		return false
	}
	return len(data) >= threshold
}
