import Foundation

enum L10n {
    enum DesignSystem {
        static let requiredMarker = LocalizedStringResource(
            "designSystem.field.requiredMarker",
            defaultValue: "*",
            comment: "Marks a required PVField as required, shown beside its label"
        )

        static let toastDismiss = LocalizedStringResource(
            "designSystem.toast.dismiss",
            defaultValue: "Dismiss",
            comment: "Accessibility label for a PVToast's dismiss button"
        )
    }

    enum Onboarding {
        static let welcomeTitle = LocalizedStringResource(
            "onboarding.chooseFile.welcomeTitle",
            defaultValue: "Welcome to Provenencia",
            comment: "Onboarding choose-file headline"
        )

        static func bodySignedIn(displayName: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "onboarding.chooseFile.bodySignedIn",
                defaultValue: "You're signed in as %@. Open a project you already have, or create a new one.",
                comment: "Onboarding choose-file body when researcher is locked; argument is display name"
            ))
            return String(format: format, locale: .current, displayName)
        }

        static let bodyChoose = LocalizedStringResource(
            "onboarding.chooseFile.bodyChoose",
            defaultValue: "Do you already have a Provenencia project, or do you want to start a new one?",
            comment: "Onboarding choose-file body when researcher is not locked"
        )

        static let createNewTitle = LocalizedStringResource(
            "onboarding.chooseFile.createNewTitle",
            defaultValue: "Create new",
            comment: "Create-new mode card title"
        )

        static let createNewSubtitle = LocalizedStringResource(
            "onboarding.chooseFile.createNewSubtitle",
            defaultValue: "Start a new project folder",
            comment: "Create-new mode card subtitle"
        )

        static let haveFileTitle = LocalizedStringResource(
            "onboarding.chooseFile.haveFileTitle",
            defaultValue: "I have a file",
            comment: "Open-existing mode card title"
        )

        static let haveFileSubtitle = LocalizedStringResource(
            "onboarding.chooseFile.haveFileSubtitle",
            defaultValue: "Open a project you already have",
            comment: "Open-existing mode card subtitle"
        )

        static let signOut = LocalizedStringResource(
            "onboarding.common.signOut",
            defaultValue: "Sign out",
            comment: "Sign out button"
        )

        static let continueAction = LocalizedStringResource(
            "onboarding.common.continue",
            defaultValue: "Continue",
            comment: "Continue button"
        )

        static let back = LocalizedStringResource(
            "onboarding.common.back",
            defaultValue: "Back",
            comment: "Back button"
        )

        static let nameResearchTitle = LocalizedStringResource(
            "onboarding.identify.nameResearchTitle",
            defaultValue: "Name this research",
            comment: "Create-mode identify headline"
        )

        static let nameResearchBody = LocalizedStringResource(
            "onboarding.identify.nameResearchBody",
            defaultValue: "Your researcher name is how work is attributed. The project name is a label; the folder on disk uses a simple slug.",
            comment: "Create-mode identify helper copy"
        )

        static let researcherName = LocalizedStringResource(
            "onboarding.identify.researcherName",
            defaultValue: "Researcher name",
            comment: "Researcher name field label"
        )

        static let researcherNamePrompt = LocalizedStringResource(
            "onboarding.identify.researcherNamePrompt",
            defaultValue: "Jane Smith",
            comment: "Placeholder for researcher name field"
        )

        static let familyName = LocalizedStringResource(
            "onboarding.identify.familyName",
            defaultValue: "Family / project name",
            comment: "Family / project name field label"
        )

        static let familyNamePrompt = LocalizedStringResource(
            "onboarding.identify.familyNamePrompt",
            defaultValue: "Smith Family",
            comment: "Placeholder for family / project name field"
        )

        static func folderNamePreview(folderName: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "onboarding.identify.folderNamePreview",
                defaultValue: "Folder: %@",
                comment: "Live preview of kebab-case project folder name; argument is folder basename"
            ))
            return String(format: format, locale: .current, folderName)
        }

        static let whoAreYouTitle = LocalizedStringResource(
            "onboarding.identify.whoAreYouTitle",
            defaultValue: "Who are you in this file?",
            comment: "Open-mode identify headline"
        )

        static let thisProject = LocalizedStringResource(
            "onboarding.identify.thisProject",
            defaultValue: "this project",
            comment: "Fallback project name when folder basename is unknown"
        )

        static func contributorsBody(projectName: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "onboarding.identify.contributorsBody",
                defaultValue: "These are the contributors already in %@. Choose yourself to keep the same ID, or add a new contributor.",
                comment: "Open-mode identify body; argument is project folder name"
            ))
            return String(format: format, locale: .current, projectName)
        }

        static let contributorPicker = LocalizedStringResource(
            "onboarding.identify.contributorPicker",
            defaultValue: "Contributor",
            comment: "Accessibility label for contributor picker"
        )

        static func contributorOption(displayName: String, ref: String) -> String {
            if ref.isEmpty {
                return displayName
            }
            let format = String(localized: LocalizedStringResource(
                "onboarding.identify.contributorOption",
                defaultValue: "%@ (%@)",
                comment: "Contributor accessibility/combined label; arguments are display name then USR-… ref"
            ))
            return String(format: format, locale: .current, displayName, ref)
        }

        static func contributorRef(ref: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "onboarding.identify.contributorRef",
                defaultValue: "(%@)",
                comment: "Parenthesized short ref beside a display name; argument is USR-… ref"
            ))
            return String(format: format, locale: .current, ref)
        }

        static let notListed = LocalizedStringResource(
            "onboarding.identify.notListed",
            defaultValue: "I’m not listed — add me",
            comment: "Picker option to add a new contributor"
        )

        static let signedInTitle = LocalizedStringResource(
            "onboarding.home.signedInTitle",
            defaultValue: "You're signed in",
            comment: "Home screen headline after successful open/create"
        )

        static func homeFolder(folderName: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "onboarding.home.folder",
                defaultValue: "Folder: %@",
                comment: "Home screen project folder line; argument is folder basename"
            ))
            return String(format: format, locale: .current, folderName)
        }

        static func homeCreated(date: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "onboarding.home.created",
                defaultValue: "Created %@",
                comment: "Home screen created date; argument is localized date"
            ))
            return String(format: format, locale: .current, date)
        }

        static func homeUpdated(date: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "onboarding.home.updated",
                defaultValue: "Updated %@",
                comment: "Home screen updated date; argument is localized date"
            ))
            return String(format: format, locale: .current, date)
        }

        static func homeUpdatedBy(displayName: String, ref: String) -> String {
            if ref.isEmpty {
                let format = String(localized: LocalizedStringResource(
                    "onboarding.home.updatedByName",
                    defaultValue: "Last edited by %@",
                    comment: "Home screen last editor without ref; argument is display name"
                ))
                return String(format: format, locale: .current, displayName)
            }
            let format = String(localized: LocalizedStringResource(
                "onboarding.home.updatedBy",
                defaultValue: "Last edited by %@ (%@)",
                comment: "Home screen last editor; arguments are display name then USR-… ref"
            ))
            return String(format: format, locale: .current, displayName, ref)
        }

        static let projectFolder = LocalizedStringResource(
            "onboarding.open.projectFolder",
            defaultValue: "Project folder",
            comment: "Section label above project folder picker"
        )

        static let noProjectsInDocuments = LocalizedStringResource(
            "onboarding.open.noProjectsInDocuments",
            defaultValue: "No project folders in Documents.",
            comment: "Empty state when Documents has no .provenencia folders"
        )

        static let chooseFolder = LocalizedStringResource(
            "onboarding.open.chooseFolder",
            defaultValue: "Choose…",
            comment: "Button to open a folder picker"
        )

        static let existingProjectPicker = LocalizedStringResource(
            "onboarding.open.existingProjectPicker",
            defaultValue: "Existing project",
            comment: "Accessibility label for existing-project picker"
        )

        static let selectProject = LocalizedStringResource(
            "onboarding.open.selectProject",
            defaultValue: "Select a project",
            comment: "Placeholder row in existing-project picker"
        )

        static let openPanelPrompt = LocalizedStringResource(
            "onboarding.open.openPanelPrompt",
            defaultValue: "Open",
            comment: "NSOpenPanel confirm button"
        )

        static let openPanelMessage = LocalizedStringResource(
            "onboarding.open.openPanelMessage",
            defaultValue: "Choose a Provenencia project folder.",
            comment: "NSOpenPanel message for choosing a project folder"
        )

        static let missingProject = LocalizedStringResource(
            "onboarding.missingProject",
            defaultValue: "The last project could not be found. Create or open a project.",
            comment: "Error when the last active project folder is missing"
        )

        static let createNewFolderNote = LocalizedStringResource(
            "onboarding.chooseFile.createNewFolderNote",
            defaultValue: "The folder is written to ~/Documents when you continue. You name the research on the next screen.",
            comment: "Info note shown when create-new mode is selected, explaining where the project folder is written"
        )

        static let newContributorIDNote = LocalizedStringResource(
            "onboarding.identify.newContributorIDNote",
            defaultValue: "A new ID is minted on continue",
            comment: "Note under the new-contributor name field explaining an ID will be assigned"
        )

        static let loadingFooterNote = LocalizedStringResource(
            "onboarding.loading.footerNote",
            defaultValue: "No project is opened until you choose one",
            comment: "Footer note on the loading screen, reassuring no project is auto-opened"
        )
    }

    /// Maps stable Go/FFI error codes to localized user-facing copy.
    enum Errors {
        static let catalogAlreadyExists = LocalizedStringResource(
            "error.catalog.already_exists",
            defaultValue: "Project already exists.",
            comment: "FFI error catalog.already_exists"
        )
        static let catalogAlreadyOpen = LocalizedStringResource(
            "error.catalog.already_open",
            defaultValue: "Project already open.",
            comment: "FFI error catalog.already_open"
        )
        static let catalogNotAProject = LocalizedStringResource(
            "error.catalog.not_a_project",
            defaultValue: "Not a Provenencia catalog.",
            comment: "FFI error catalog.not_a_project"
        )
        static func catalogUnsupportedVersion(version: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "error.catalog.unsupported_version",
                defaultValue: "Unsupported catalog version (%@).",
                comment: "FFI error catalog.unsupported_version; argument is catalog user_version"
            ))
            return String(format: format, locale: .current, version)
        }
        static let catalogInvalidFolderName = LocalizedStringResource(
            "error.catalog.invalid_folder_name",
            defaultValue: "Folder name must end in .provenencia.",
            comment: "FFI error catalog.invalid_folder_name"
        )
        static let catalogClosed = LocalizedStringResource(
            "error.catalog.closed",
            defaultValue: "Catalog closed.",
            comment: "FFI error catalog.closed"
        )
        static let projectInvalidMetadata = LocalizedStringResource(
            "error.project.invalid_metadata",
            defaultValue: "Invalid project metadata.",
            comment: "FFI error project.invalid_metadata"
        )
        static let projectMissingMetadata = LocalizedStringResource(
            "error.project.missing_metadata",
            defaultValue: "Project metadata missing.",
            comment: "FFI error project.missing_metadata"
        )
        static let usersInvalid = LocalizedStringResource(
            "error.users.invalid",
            defaultValue: "Invalid user ID, display name, or ref.",
            comment: "FFI error users.invalid"
        )
        static let identityNotFound = LocalizedStringResource(
            "error.identity.not_found",
            defaultValue: "Identity file not found.",
            comment: "FFI error identity.not_found"
        )
        static let identityInvalidName = LocalizedStringResource(
            "error.identity.invalid_name",
            defaultValue: "Display name is empty.",
            comment: "FFI error identity.invalid_name"
        )
        static let identityInvalidID = LocalizedStringResource(
            "error.identity.invalid_id",
            defaultValue: "User ID must be UUIDv7.",
            comment: "FFI error identity.invalid_id"
        )
        static let identityInvalidRef = LocalizedStringResource(
            "error.identity.invalid_ref",
            defaultValue: "User ref is invalid.",
            comment: "FFI error identity.invalid_ref"
        )
        static let installNotFound = LocalizedStringResource(
            "error.install.not_found",
            defaultValue: "Active project file not found.",
            comment: "FFI error install.not_found"
        )
        static let installInvalid = LocalizedStringResource(
            "error.install.invalid",
            defaultValue: "Project dir is empty.",
            comment: "FFI error install.invalid"
        )
        static let onboardingBlankName = LocalizedStringResource(
            "error.onboarding.blank_name",
            defaultValue: "Name is empty.",
            comment: "FFI error onboarding.blank_name"
        )
        static let onboardingInvalidFamilyName = LocalizedStringResource(
            "error.onboarding.invalid_family_name",
            defaultValue: "Invalid family name.",
            comment: "FFI error onboarding.invalid_family_name"
        )
        static let onboardingUnknownUser = LocalizedStringResource(
            "error.onboarding.unknown_user",
            defaultValue: "User not in project.",
            comment: "FFI error onboarding.unknown_user"
        )
        static let fileNotFound = LocalizedStringResource(
            "error.file.not_found",
            defaultValue: "File not found.",
            comment: "FFI error file.not_found"
        )
        static let unknown = LocalizedStringResource(
            "error.internal.unknown",
            defaultValue: "Something went wrong. Please try again.",
            comment: "FFI error internal.unknown and other unmapped codes"
        )

        /// Resolves a wire error code (+ params) to localized UI copy.
        static func message(code: String, params: [String] = []) -> String {
            switch code {
            case "catalog.already_exists":
                return String(localized: catalogAlreadyExists)
            case "catalog.already_open":
                return String(localized: catalogAlreadyOpen)
            case "catalog.not_a_project":
                return String(localized: catalogNotAProject)
            case "catalog.unsupported_version":
                return catalogUnsupportedVersion(version: params.first ?? "?")
            case "catalog.invalid_folder_name":
                return String(localized: catalogInvalidFolderName)
            case "catalog.closed":
                return String(localized: catalogClosed)
            case "project.invalid_metadata":
                return String(localized: projectInvalidMetadata)
            case "project.missing_metadata":
                return String(localized: projectMissingMetadata)
            case "users.invalid":
                return String(localized: usersInvalid)
            case "identity.not_found":
                return String(localized: identityNotFound)
            case "identity.invalid_name":
                return String(localized: identityInvalidName)
            case "identity.invalid_id":
                return String(localized: identityInvalidID)
            case "identity.invalid_ref":
                return String(localized: identityInvalidRef)
            case "install.not_found":
                return String(localized: installNotFound)
            case "install.invalid":
                return String(localized: installInvalid)
            case "onboarding.blank_name":
                return String(localized: onboardingBlankName)
            case "onboarding.invalid_family_name":
                return String(localized: onboardingInvalidFamilyName)
            case "onboarding.unknown_user":
                return String(localized: onboardingUnknownUser)
            case "file.not_found":
                return String(localized: fileNotFound)
            default:
                return String(localized: unknown)
            }
        }

        static func message(for error: Error) -> String {
            if let coded = error as? CoreInvokeError {
                switch coded {
                case .coded(_, let code, _, let params):
                    return message(code: code, params: params)
                case .failed:
                    return String(localized: unknown)
                }
            }
            return String(localized: unknown)
        }
    }
}
