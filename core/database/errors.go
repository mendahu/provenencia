package database

import "errors"

var (
	ErrAlreadyExists      = errors.New("project already exists")
	ErrAlreadyOpen        = errors.New("project already open")
	ErrNotAProject        = errors.New("not a provenencia catalog")
	ErrUnsupportedVersion = errors.New("unsupported catalog version")
	ErrInvalidFolderName  = errors.New("folder name must end in .provenencia")
	ErrClosed             = errors.New("catalog closed")
)
