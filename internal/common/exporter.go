package common

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/harness/harness-migrate/internal/codeerror"
	"github.com/harness/harness-migrate/internal/types"
	"log"
	"os"
	"path/filepath"
)

const (
	maxChunkSize = 25 * 1024 * 1024 // 25 MB
	infoFileName = "info.json"
)

type Exporter struct {
	Exporter    Interface
	ZipLocation string
	Checkpoint  map[string]interface{}
}

// Export calls exporter methods in order and serialize an object for import.
func (e *Exporter) Export(ctx context.Context) {
	path := filepath.Join(".", e.ZipLocation)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		panic("cannot create folder")
	}
	data, _ := e.getData(ctx)
	for _, repo := range data {
		err = e.writeJsonForRepo(repo)
		if err != nil {
			panic("error writing data")
		}
	}
}

// Calculate the size of the struct in bytes
func calculateSize(s *types.PullRequestData) int {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
	if err != nil {
		panic(err)
	}
	return buf.Len()
}

// Split the array into smaller chunks if the size exceeds the maxChunkSize
func splitArray(arr []*types.PullRequestData) [][]*types.PullRequestData {
	var chunks [][]*types.PullRequestData
	var currentChunk []*types.PullRequestData
	currentSize := 0

	for _, item := range arr {
		itemSize := calculateSize(item)
		if currentSize+itemSize > maxChunkSize {
			chunks = append(chunks, currentChunk)
			currentChunk = []*types.PullRequestData{}
			currentSize = 0
		}
		currentChunk = append(currentChunk, item)
		currentSize += itemSize
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}
	return chunks
}

func (e *Exporter) writeJsonForRepo(repo *types.RepoData) error {
	repoJson, err := getJsonContent(repo.Repository)
	if err != nil {
		// todo: fix this
		log.Printf("cannot serialize into json: %v", err)
	}
	pathRepo := filepath.Join(".", e.ZipLocation, repo.Repository.RepoSlug)
	err = os.MkdirAll(pathRepo, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot create folder")
	}

	err = os.WriteFile(filepath.Join(pathRepo, infoFileName), repoJson, os.ModePerm)
	if err != nil {
		return err
	}

	if len(repo.PullRequestData) == 0 {
		return nil
	}

	prDataInSize := splitArray(repo.PullRequestData)
	pathPR := filepath.Join(pathRepo, "pr")
	err = createFolder(pathPR)
	if err != nil {
		return err
	}

	for i, data := range prDataInSize {
		prJson, err := getJsonContent(data)
		if err != nil {
			// todo: fix this
			log.Printf("cannot serialize into json: %v", err)
		}
		prFilePath := fmt.Sprintf("pr%d.json", i)

		err = writeFile(filepath.Join(pathPR, prFilePath), prJson)
		if err != nil {
			return err
		}
	}

	return nil
}

func createFolder(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func writeFile(path string, prJson []byte) error {
	err := os.WriteFile(path, prJson, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func getJsonContent(data interface{}) ([]byte, error) {
	jsonString, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("cannot serialize json string for data: %w", err)
	}
	return jsonString, nil
}

func (e *Exporter) getData(ctx context.Context) ([]*types.RepoData, error) {
	repoData := make([]*types.RepoData, 0)
	// 1. list all the repos for the given org
	repositories, err := e.Exporter.ListRepositories(ctx, types.ListRepoOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot list repositories: %w", err)
	}

	for _, repository := range repositories {
		data := &types.RepoData{Repository: repository}
		repoData = append(repoData, data)
	}

	// 2. list pr per repo
	for i, repo := range repositories {
		prs, err := e.Exporter.ListPullRequest(ctx, repo.RepoSlug, types.PullRequestListOptions{})
		var notSupportedErr *codeerror.ErrorOpNotSupported
		if errors.As(err, &notSupportedErr) {
			return repoData, nil
		}
		if err != nil {
			log.Default().Printf("encountered error in getting pr: %q", err)
		}

		// 3. get all data for each pr
		prData := make([]*types.PullRequestData, len(prs))

		for j, pr := range prs {
			// todo: implement comment
			prData[j] = mapPrData(pr, nil)
		}

		repoData[i].PullRequestData = prData
	}

	return repoData, nil
}

func mapPrData(pr types.PRResponse, comments []types.PRComments) *types.PullRequestData {
	return &types.PullRequestData{
		PullRequest: pr,
		Comments:    comments,
	}
}
