package cliod

// This file implements a simple file-system based database for storing
//  and (simple) indexing Clio crates.

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

var (
	// Used when we've created the crate
	STORE_TYPE_SEED = "SEED"

	// Used when we are simply sharing (have no relation)
	STORE_TYPE_SHARE = "SHARE"

	// Used when we are related to the crate and want to hold on too it
	STORE_TYPE_ARCHIVE = "ARCHIVE"
)

type Store struct {
	StorePath string
	Index     *StoreIndex
}

type StoreIndex struct {
	Path string
	Tags map[string][]string
	IDs  []string
	Time map[string]time.Time
	Type map[string][]string
}

func (si *StoreIndex) Sync() {
	f, _ := os.OpenFile(si.Path, os.O_WRONLY, os.ModePerm)
	data, _ := json.Marshal(si)
	f.Write(data)
	f.Sync()
	f.Close()
}

func (si *StoreIndex) Load() {
	f, _ := os.Open(si.Path)
	data, _ := ioutil.ReadAll(f)
	json.Unmarshal(data, si)
	f.Close()
}

func (si *StoreIndex) AddCrate(c Crate) {
	for _, tag := range c.Tags {
		if _, exists := si.Tags[tag]; !exists {
			si.Tags[tag] = []string{}
		}
		si.Tags[tag] = append(si.Tags[tag], c.ID)
	}

	si.IDs = append(si.IDs, c.ID)
	si.Time[c.ID] = c.Time

	si.Type[c.Type] = append(si.Type[c.Type], c.ID)
	si.Sync()
}

func (si *StoreIndex) FindByTags(tags []string) []string {
	results := make([]string, 0)

	for _, tag := range tags {
		if _, exists := si.Tags[tag]; !exists {
			continue
		}

		for _, result := range si.Tags[tag] {
			results = append(results, result)
		}
	}

	return results
}

type Crate struct {
	ID   string
	Time time.Time

	Tags []string
	Raw  []byte

	Expires time.Time
	Type    string
}

func NewCrate(data []byte, tags []string) Crate {
	return Crate{
		ID:   uuid.New(),
		Time: time.Now(),
		Tags: tags,
		Raw:  data,
		Type: STORE_TYPE_SEED,
	}
}

func NewStore(p string) Store {
	s := Store{
		StorePath: p,
		Index: &StoreIndex{
			Path: (p + "/" + "index.json"),
			Tags: make(map[string][]string),
			IDs:  make([]string, 0),
			Time: make(map[string]time.Time),
			Type: make(map[string][]string),
		},
	}

	s.Index.Type[STORE_TYPE_SEED] = make([]string, 0)
	s.Index.Type[STORE_TYPE_ARCHIVE] = make([]string, 0)
	s.Index.Type[STORE_TYPE_SHARE] = make([]string, 0)
	return s
}

func (s *Store) GetPath(id string) string {
	return s.StorePath + "/" + id + ".json"
}

func (s *Store) Init() {
	if !PathExists(s.StorePath) {
		os.MkdirAll(s.StorePath, os.ModePerm)
	}

	if !PathExists(s.Index.Path) {
		f, _ := os.Create(s.Index.Path)
		f.Close()
	}

	s.Index.Load()
}

func (s *Store) PutCrate(c Crate) {
	s.Index.AddCrate(c)
	f, _ := os.Create(s.GetPath(c.ID))
	data, _ := json.Marshal(c)
	f.Write(data)
	f.Close()
}

func (s *Store) HasCrate(id string) bool {
	return PathExists(s.GetPath(id))
}

func (s *Store) GetCrate() *Crate {
	return nil
}

func (s *Store) ListCrates() {

}
