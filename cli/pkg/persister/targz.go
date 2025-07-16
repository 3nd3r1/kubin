package persister

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/3nd3r1/kubin/cli/pkg/collector"
	"github.com/3nd3r1/kubin/cli/pkg/log"
)

type TarGzPersister struct {
    basePath string
}

func NewTarGzPersister() (*TarGzPersister, error) {
    basePath, err := os.MkdirTemp("", "kubin-persister-*")
    if err != nil {
        return nil, err
    }

	return &TarGzPersister{
        basePath: basePath,
    }, nil
}

func (p *TarGzPersister) Persist(resource collector.ClusterResource) error {
	path := filepath.Join(p.basePath, resource.Kind)
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fileName := fmt.Sprintf("%s.json", resource.Name)
	filePath := filepath.Join(path, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(resource.Data)
}

func (p *TarGzPersister) Finalize(outputPath string) error {
    defer p.cleanup()

	// Create the output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk through the source directory and add files to the archive
	return filepath.Walk(p.basePath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Create tar header
		relPath, err := filepath.Rel(p.basePath, filePath)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		return err
	})
}

func (p *TarGzPersister) cleanup() {
    if err := os.RemoveAll(p.basePath); err != nil {
        log.WithError(err).Errorf("Failed to cleanup tmp dir %s", p.basePath)
    }
}
