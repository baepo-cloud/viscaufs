package fsindex

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createMockFileInfo creates a mock os.FileInfo for testing
func createMockFileInfo(isDir bool) os.FileInfo {
	var mode os.FileMode = 0644
	if isDir {
		mode |= os.ModeDir
	}

	// Create a real syscall.Stat_t that can be used directly
	statT := &syscall.Stat_t{
		Ino:     12345,
		Size:    1024,
		Blocks:  8,
		Mode:    uint16(mode),
		Nlink:   1,
		Uid:     1000,
		Gid:     1000,
		Rdev:    0,
		Blksize: 4096,
	}

	// Set timestamps
	now := time.Now()
	statT.Atimespec = syscall.Timespec{Sec: now.Unix(), Nsec: int64(now.Nanosecond())}
	statT.Mtimespec = syscall.Timespec{Sec: now.Unix(), Nsec: int64(now.Nanosecond())}
	statT.Ctimespec = syscall.Timespec{Sec: now.Unix(), Nsec: int64(now.Nanosecond())}

	return mockFileInfo{
		name:    "test",
		size:    1024,
		mode:    mode,
		modTime: now,
		isDir:   isDir,
		sys:     statT, // Use the actual syscall.Stat_t struct
	}
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     interface{}
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return m.sys }

func TestJoinFSIndex(t *testing.T) {
	layer0 := NewFSIndex() // layer 0
	layer0.addPath("file1.txt", createMockFileInfo(false))
	layer0.addPath("file5.txt", createMockFileInfo(false))
	layer0.addPath("file2.txt", createMockFileInfo(false))
	layer0.addPath("dir1", createMockFileInfo(true))
	layer0.addPath("dir1/subfile1.txt", createMockFileInfo(false))
	layer0.addPath("dir1/subfile2.txt", createMockFileInfo(false))
	layer0.addPath("dir2", createMockFileInfo(true))
	layer0.addPath("dir2/subfile1.txt", createMockFileInfo(false))

	layer1 := NewFSIndex() // layer 1
	layer1.addPath("file4.txt", createMockFileInfo(false))
	layer1.addPath("file3.txt", createMockFileInfo(false))
	layer1.addPath("file1.txt", createMockFileInfo(false))
	layer1.addPath("dir1/.wh.subfile1.txt", createMockFileInfo(false))
	layer1.addPath("dir2/.wh..wh.opq", createMockFileInfo(false))
	layer1.addPath("dir2/newfile.txt", createMockFileInfo(false))

	layer2 := NewFSIndex() // layer 2
	layer2.addPath(".wh.file4.txt", createMockFileInfo(false))
	layer2.addPath("file5.txt", createMockFileInfo(false))
	layer2.addPath("file6.txt", createMockFileInfo(false))

	layer3 := NewFSIndex() // layer 3
	layer3.addPath("file7.txt", createMockFileInfo(false))

	// Add a new laye

	// Join the layers
	JoinFSIndex(layer2, layer3, 2, true)
	JoinFSIndex(layer1, layer2, 1, false)
	JoinFSIndex(layer0, layer1, 0, false)
	result := layer0

	fmt.Println(layer0.String())

	// Check that new files are added
	file3, err := result.LookupPath("file3.txt")
	require.NoError(t, err, "Expected file3.txt to be added")
	assert.Equal(t, uint8(1), file3.LayerPosition)

	// Check that updated files are properly updated
	file1, err := result.LookupPath("file1.txt")
	require.NoError(t, err, "Expected file1.txt to exist")
	assert.Equal(t, uint8(1), file1.LayerPosition)

	// Check that whiteout files are removed
	_, err = result.LookupPath("dir1/subfile1.txt")
	assert.Error(t, err, "Expected dir1/subfile1.txt to be removed by whiteout")

	// Check that non-whiteout files in the same directory still exist
	subfile2, err := result.LookupPath("dir1/subfile2.txt")
	require.NoError(t, err, "Expected dir1/subfile2.txt to still exist")
	assert.Equal(t, uint8(0), subfile2.LayerPosition)

	// Check that opaque directory's old content is removed
	_, err = result.LookupPath("dir2/subfile1.txt")
	assert.Error(t, err, "Expected dir2/subfile1.txt to be removed by opaque whiteout")

	// Check that new content in opaque directory is added
	newfile, err := result.LookupPath("dir2/newfile.txt")
	require.NoError(t, err, "Expected dir2/newfile.txt to be added")
	assert.Equal(t, uint8(1), newfile.LayerPosition)

	// Check that the opaque directory itself still exists
	_, err = result.LookupPath("dir2")
	assert.NoError(t, err, "Expected dir2 to still exist")

	// Check that whiteout files aren't included in the final result
	_, err = result.LookupPath("dir1/.wh.subfile1.txt")
	assert.Error(t, err, "Expected whiteout file dir1/.wh.subfile1.txt to not be in the result")

	_, err = result.LookupPath("dir2/.wh..wh.opq")
	assert.Error(t, err, "Expected opaque whiteout file dir2/.wh..wh.opq to not be in the result")

	// Check that the file file4 is not in the result
	_, err = result.LookupPath("file4.txt")
	assert.Error(t, err, "Expected file4.txt to not be in the result")

	// Check that the file file5 is in the result
	file5, err := result.LookupPath("file5.txt")
	require.NoError(t, err, "Expected file5.txt to exist")
	assert.Equal(t, uint8(2), file5.LayerPosition)

}
