package database

import "errors"

var (
	ErrAlreadyExists      = errors.New("Project already exists.")
	ErrAlreadyOpen        = errors.New("Project already open.")
	ErrNotAProject        = errors.New("Not a Provenencia catalog.")
	ErrUnsupportedVersion = errors.New("Unsupported catalog version.")
	ErrInvalidFolderName  = errors.New("Folder name must end in .provenencia.")
	ErrClosed             = errors.New("Catalog closed.")
)
