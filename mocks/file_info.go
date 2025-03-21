package mocks

import (
	"os"
	"time"

	"github.com/stretchr/testify/mock"
)

type FileInfo struct {
	mock.Mock
}

func (m *FileInfo) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *FileInfo) Size() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *FileInfo) Mode() os.FileMode {
	args := m.Called()
	return args.Get(0).(os.FileMode)
}

func (m *FileInfo) ModTime() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *FileInfo) IsDir() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *FileInfo) Sys() interface{} {
	args := m.Called()
	return args.Get(0)
}