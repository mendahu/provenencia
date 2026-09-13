// Package apperr is the stable error-code registry for the Go core.
// Error.Error() returns the machine code (never user-facing copy).
// Clients map codes to localized strings over FFI.
package apperr

import (
	"errors"
)

// Kind classifies a failure for client UX. Copy still comes from L10n by Code.
type Kind int

const (
	KindUser Kind = iota
	KindConflict
	KindNotFound
	KindInternal
)

// Stable wire codes. Permanent once shipped.
const (
	CodeCatalogAlreadyExists        = "catalog.already_exists"
	CodeCatalogAlreadyOpen          = "catalog.already_open"
	CodeCatalogNotAProject          = "catalog.not_a_project"
	CodeCatalogUnsupportedVersion   = "catalog.unsupported_version"
	CodeCatalogSchemaMismatch       = "catalog.schema_mismatch"
	CodeCatalogInvalidFolderName    = "catalog.invalid_folder_name"
	CodeCatalogClosed               = "catalog.closed"
	CodeProjectInvalidMetadata      = "project.invalid_metadata"
	CodeProjectMissingMetadata      = "project.missing_metadata"
	CodeUsersInvalid                = "users.invalid"
	CodeAuditInvalid                = "audit.invalid"
	CodeDateValuesInvalid           = "datevalues.invalid"
	CodeSourceTypesInvalid          = "sourcetypes.invalid"
	CodeSourceTypesInUse            = "sourcetypes.in_use"
	CodeSourceTypesDuplicateKey     = "sourcetypes.duplicate_key"
	CodeSourceFieldsInvalid         = "sourcefields.invalid"
	CodeSourceFieldsDuplicateKey    = "sourcefields.duplicate_key"
	CodeSourceFieldsInUse           = "sourcefields.in_use"
	CodeSourceVocabInvalid          = "sourcevocab.invalid"
	CodeSourcesInvalid              = "sources.invalid"
	CodeFilesInvalid                = "files.invalid"
	CodeArtifactsInvalid            = "artifacts.invalid"
	CodeArtifactsFileAlreadyAttached = "artifacts.file_already_attached"
	CodeSourceCredibilityInvalid    = "sourcecredibility.invalid"
	CodeSourceMetadataInvalid       = "sourcemetadata.invalid"
	CodeFileDerivativesInvalid      = "filederivatives.invalid"
	CodeIngestInvalid               = "ingest.invalid"
	CodeIngestPermissionDenied      = "ingest.permission_denied"
	CodeIngestUnsupportedOffice     = "ingest.unsupported_office"
	CodeIngestUnsupportedArchive    = "ingest.unsupported_archive"
	CodeIngestUnsupportedExecutable = "ingest.unsupported_executable"
	CodeIngestUnsupportedType       = "ingest.unsupported_type"
	CodeIngestUnidentified          = "ingest.unidentified"
	CodeIngestEmpty                 = "ingest.empty"
	CodeIngestTooLarge              = "ingest.too_large"
	CodeIngestNotAFile              = "ingest.not_a_file"
	CodeIngestSymlink               = "ingest.symlink"
	CodeIngestMissing               = "ingest.missing"
	CodeIdentityNotFound            = "identity.not_found"
	CodeIdentityInvalidName         = "identity.invalid_name"
	CodeIdentityInvalidID           = "identity.invalid_id"
	CodeIdentityInvalidRef          = "identity.invalid_ref"
	CodeInstallNotFound             = "install.not_found"
	CodeInstallInvalid              = "install.invalid"
	CodeOnboardingBlankName         = "onboarding.blank_name"
	CodeOnboardingInvalidFamilyName = "onboarding.invalid_family_name"
	CodeOnboardingUnknownUser       = "onboarding.unknown_user"
	CodeRefInvalidPrefix            = "ref.invalid_prefix"
	CodeRefInvalid                  = "ref.invalid"
	CodeFileNotFound                = "file.not_found"
	CodeInternalUnknown             = "internal.unknown"
	CodeInternalUnknownMethod       = "internal.unknown_method"
	CodeInternalMigrations          = "internal.migrations"

	// Derivative-generation guards: unprocessable = source image too large
	// (bytes or declared pixels) to decode safely; corrupt_object = object
	// bytes do not match the checksum recorded in the catalog.
	CodeFileDerivativesUnprocessable = "filederivatives.unprocessable"
	CodeFileDerivativesCorruptObject = "filederivatives.corrupt_object"
)

// Error is a coded application error. Params are ordered L10n format args.
type Error struct {
	code   string
	kind   Kind
	params []string
}

// New returns a coded error. params are optional interpolation arguments.
func New(code string, kind Kind, params ...string) *Error {
	return &Error{code: code, kind: kind, params: append([]string(nil), params...)}
}

func (e *Error) Error() string {
	if e == nil {
		return CodeInternalUnknown
	}
	return e.code
}

func (e *Error) Code() string {
	if e == nil {
		return CodeInternalUnknown
	}
	return e.code
}

func (e *Error) Kind() Kind {
	if e == nil {
		return KindInternal
	}
	return e.kind
}

func (e *Error) Params() []string {
	if e == nil {
		return nil
	}
	return append([]string(nil), e.params...)
}

// WithParams returns a copy with the given params (same code and kind).
func (e *Error) WithParams(params ...string) *Error {
	if e == nil {
		return New(CodeInternalUnknown, KindInternal, params...)
	}
	return New(e.code, e.kind, params...)
}

// Is reports whether target is an *Error with the same code (params ignored).
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok || e == nil || t == nil {
		return false
	}
	return e.code == t.code
}

// From finds the innermost *Error in err's chain, or returns internal.unknown.
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var ae *Error
	if errors.As(err, &ae) {
		return ae
	}
	return New(CodeInternalUnknown, KindInternal)
}
