package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/liangchenye/update-service/utils"
)

var (
	// ErrorsUCEmptyURL occurs when a repository url is nil
	ErrorsUCEmptyURL = errors.New("empty repository url")
	// ErrorsUCRepoExist occurs when a repository is exist
	ErrorsUCRepoExist = errors.New("repository is already exist")
	// ErrorsUCRepoNotExist occurs when a repository is not exist
	ErrorsUCRepoNotExist = errors.New("repository is not exist")
)

const (
	topDir     = ".update-service"
	configName = "config.json"
	cacheDir   = "cache"
)

// UpdateClientConfig is the local configuation of a update client
type UpdateClientConfig struct {
	DefaultServer string
	CacheDir      string
	Repos         []string
}

func (ucc *UpdateClientConfig) exist() bool {
	configFile := filepath.Join(os.Getenv("HOME"), topDir, configName)
	return utils.IsFileExist(configFile)
}

// Init create directory and setup the cache location
func (ucc *UpdateClientConfig) Init() error {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return errors.New("Cannot get home directory")
	}

	topURL := filepath.Join(homeDir, topDir)
	cacheURL := filepath.Join(topURL, cacheDir)
	if !utils.IsDirExist(cacheURL) {
		if err := os.MkdirAll(cacheURL, os.ModePerm); err != nil {
			return err
		}
	}

	ucc.CacheDir = cacheURL

	if !ucc.exist() {
		return ucc.save()
	}
	return nil
}

func (ucc *UpdateClientConfig) save() error {
	data, err := json.MarshalIndent(ucc, "", "\t")
	if err != nil {
		return err
	}

	configFile := filepath.Join(os.Getenv("HOME"), topDir, configName)
	if err := ioutil.WriteFile(configFile, data, 0666); err != nil {
		return err
	}

	return nil
}

// Load reads the config data
func (ucc *UpdateClientConfig) Load() error {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return errors.New("Cannot get home directory")
	}

	content, err := ioutil.ReadFile(filepath.Join(homeDir, topDir, configName))
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &ucc); err != nil {
		return err
	}

	if ucc.CacheDir == "" {
		ucc.CacheDir = filepath.Join(homeDir, topDir, cacheDir)
	}

	return nil
}

// Add adds a repo url to the config file
func (ucc *UpdateClientConfig) Add(url string) error {
	if url == "" {
		return ErrorsUCEmptyURL
	}

	var err error
	if !ucc.exist() {
		err = ucc.Init()
	} else {
		err = ucc.Load()
	}
	if err != nil {
		return err
	}

	for _, repo := range ucc.Repos {
		if repo == url {
			return ErrorsUCRepoExist
		}
	}
	ucc.Repos = append(ucc.Repos, url)

	return ucc.save()
}

// Remove removes a repo url from the config file
func (ucc *UpdateClientConfig) Remove(url string) error {
	if url == "" {
		return ErrorsUCEmptyURL
	}

	if !ucc.exist() {
		return ErrorsUCRepoNotExist
	}

	if err := ucc.Load(); err != nil {
		return err
	}
	found := false
	for i := range ucc.Repos {
		if ucc.Repos[i] == url {
			found = true
			ucc.Repos = append(ucc.Repos[:i], ucc.Repos[i+1:]...)
			break
		}
	}
	if !found {
		return ErrorsUCRepoNotExist
	}

	return ucc.save()
}
