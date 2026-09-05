package database

import "github.com/mendahu/provenencia/core/apperr"

var (
	ErrAlreadyExists      = apperr.New(apperr.CodeCatalogAlreadyExists, apperr.KindConflict)
	ErrAlreadyOpen        = apperr.New(apperr.CodeCatalogAlreadyOpen, apperr.KindConflict)
	ErrNotAProject        = apperr.New(apperr.CodeCatalogNotAProject, apperr.KindUser)
	ErrUnsupportedVersion = apperr.New(apperr.CodeCatalogUnsupportedVersion, apperr.KindUser)
	ErrInvalidFolderName  = apperr.New(apperr.CodeCatalogInvalidFolderName, apperr.KindUser)
	ErrClosed             = apperr.New(apperr.CodeCatalogClosed, apperr.KindConflict)
)
