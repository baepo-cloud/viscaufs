package viscaufs

import (
	"log/slog"
	"sync"

	"github.com/alphadose/haxmap"
	"github.com/baepo-cloud/viscaufs-common/fsindex"
	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
)

type Cache struct {
	indexAttrs *fsindex.Index
	readDirs   *haxmap.Map[string, struct{}]

	m sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		indexAttrs: fsindex.NewFSIndex(),
		readDirs:   haxmap.New[string, struct{}](),
	}
}

func (c *Cache) GetOrFetchAttr(path string, fetch func() (*fspb.GetAttrResponse, error)) (*fspb.File, error) {
	c.m.RLock()
	if file, err := c.indexAttrs.LookupPath(path); err == nil && file != nil {
		c.m.RUnlock()
		slog.Info("cache: hit on file", "path", path, "operation", "getattr")
		return &fspb.File{
			Path:          path,
			Attributes:    file.FileAttributesToProto(),
			SymlinkTarget: nil,
		}, nil
	}
	c.m.RUnlock()

	resp, err := fetch()
	if err == nil && resp.File != nil {

		c.m.Lock()
		c.indexAttrs.AddNode(&fsindex.Node{
			Path:          path,
			Attributes:    fsindex.FSFileAttrFromProto(resp.File.Attributes),
			SymlinkTarget: resp.File.SymlinkTarget,
		})
		c.m.Unlock()

		return resp.File, nil
	}

	return nil, err
}

func (c *Cache) GetOrFetchDir(path string, fetch func() (*fspb.ReadDirResponse, error)) ([]*fspb.File, error) {
	_, ok := c.readDirs.Get(path)
	var files []*fspb.File
	if ok {
		slog.Info("cache: hit on dir", "path", path, "operation", "readdir")
		c.m.RLock()
		search := c.indexAttrs.LookupPrefixSearch(path)
		c.m.RUnlock()
		files = make([]*fspb.File, len(search))
		for i, node := range search {
			files[i] = &fspb.File{
				Path:          node.Path,
				Attributes:    node.FileAttributesToProto(),
				SymlinkTarget: node.SymlinkTarget,
			}
		}
	} else {
		response, err := fetch()
		if err != nil {
			return nil, err
		}

		files = response.Entries

		for _, file := range files {
			c.m.Lock()
			c.indexAttrs.AddNode(&fsindex.Node{
				Path:          path,
				Attributes:    fsindex.FSFileAttrFromProto(file.Attributes),
				SymlinkTarget: file.SymlinkTarget,
			})
			c.m.Unlock()
		}
	}

	c.readDirs.Set(path, struct{}{})
	return files, nil
}
