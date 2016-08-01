/*
Copyright 2016 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"errors"
	"fmt"
	"sync"
)

// UpdateServiceStorage represents the storage interface
type UpdateServiceStorage interface {
	// url is the database address or local directory (local://tmp/cache)
	New(url string, km string) (UpdateServiceStorage, error)
	// get the 'url' set by 'New'
	String() string
	Supported(url string) bool
	// key: namespace/repository/appname
	Get(key string) ([]byte, error)
	// key: namespace/repository
	GetMeta(key string) ([]byte, error)
	// key: namespace/repository
	GetMetaSign(key string) ([]byte, error)
	// key: namespace/repository
	GetPublicKey(key string) ([]byte, error)
	// key: namespace/repository/appname
	Put(key string, data []byte) error
	// key: namespace/repository
	List(key string) ([]string, error)
}

var (
	usStoragesLock sync.Mutex
	usStorages     = make(map[string]UpdateServiceStorage)

	// ErrorsUSSNotSupported occurs if a type is not supported
	ErrorsUSSNotSupported = errors.New("storage type is not supported")
)

// RegisterStorage provides a way to dynamically register an implementation of a
// storage type.
//
// If RegisterStorage is called twice with the same name if 'storage type' is nil,
// or if the name is blank, it panics.
func RegisterStorage(name string, f UpdateServiceStorage) {
	if name == "" {
		panic("Could not register a Storage with an empty name")
	}
	if f == nil {
		panic("Could not register a nil Storage")
	}

	usStoragesLock.Lock()
	defer usStoragesLock.Unlock()

	if _, alreadyExists := usStorages[name]; alreadyExists {
		panic(fmt.Sprintf("Storage type '%s' is already registered", name))
	}
	usStorages[name] = f
}

// NewUSStorage creates a storage interface by a url
func NewUSStorage(url string, km string) (UpdateServiceStorage, error) {
	if url == "" {
		url, _ = GetSetting("storage")
	}
	if km == "" {
		km, _ = GetSetting("keymanager")
	}

	for _, f := range usStorages {
		if f.Supported(url) {
			return f.New(url, km)
		}
	}

	return nil, ErrorsUSSNotSupported
}
