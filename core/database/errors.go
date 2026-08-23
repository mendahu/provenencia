package database

import "errors"

var (
	ErrAlreadyExists      = errors.New("project already exists")
	ErrAlreadyOpen        = errors.New("project already open")
	ErrNotAProject        = errors.New("not a provenance catalog")
	ErrUnsupportedVersion = errors.New("unsupported catalog version")
	ErrInvalidFolderName  = errors.New("folder name must end in .provenance")
	ErrClosed             = errors.New("catalog closed")
)
