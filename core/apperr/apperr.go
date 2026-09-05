// Package apperr is the stable error-code registry for the Go core.
// Error.Error() returns the machine code (never user-facing copy).
// Clients map codes to localized strings over FFI.
package apperr

import (
	"errors"
	"fmt"
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
	CodeCatalogAlreadyExists       = "catalog.already_exists"
	CodeCatalogAlreadyOpen         = "catalog.already_open"
	CodeCatalogNotAProject         = "catalog.not_a_project"
	CodeCatalogUnsupportedVersion  = "catalog.unsupported_version"
	CodeCatalogInvalidFolderName   = "catalog.invalid_folder_name"
	CodeCatalogClosed              = "catalog.closed"
	CodeProjectInvalidMetadata     = "project.invalid_metadata"
	CodeProjectMissingMetadata     = "project.missing_metadata"
	CodeUsersInvalid               = "users.invalid"
	CodeIdentityNotFound           = "identity.not_found"
	CodeIdentityInvalidName        = "identity.invalid_name"
	CodeIdentityInvalidID          = "identity.invalid_id"
	CodeIdentityInvalidRef         = "identity.invalid_ref"
	CodeInstallNotFound            = "install.not_found"
	CodeInstallInvalid             = "install.invalid"
	CodeOnboardingBlankName        = "onboarding.blank_name"
	CodeOnboardingInvalidFamilyName = "onboarding.invalid_family_name"
	CodeOnboardingUnknownUser      = "onboarding.unknown_user"
	CodeRefInvalidPrefix           = "ref.invalid_prefix"
	CodeRefInvalid                 = "ref.invalid"
	CodeFileNotFound               = "file.not_found"
	CodeInternalUnknown            = "internal.unknown"
	CodeInternalUnknownMethod      = "internal.unknown_method"
	CodeInternalMigrations         = "internal.migrations"
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
	if len(e.params) == 0 {
		return e.code
	}
	return fmt.Sprintf("%s: %v", e.code, e.params)
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

// CodeOf returns the stable code for err, or internal.unknown.
func CodeOf(err error) string {
	return From(err).Code()
}
