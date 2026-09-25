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

        static let reorderHandle = LocalizedStringResource(
            "designSystem.reorder.handle",
            defaultValue: "Drag to reorder",
            comment: "Accessibility label for a PVReorderHandle drag affordance"
        )

        static let tableSortNone = LocalizedStringResource(
            "designSystem.table.sortNone",
            defaultValue: "Not sorted. Activate to sort ascending",
            comment: "Spoken sort state of an unsorted PVTable column header"
        )

        static let tableSortAscending = LocalizedStringResource(
            "designSystem.table.sortAscending",
            defaultValue: "Sorted ascending. Activate to sort descending",
            comment: "Spoken sort state of a PVTable column header sorted ascending"
        )

        static let tableSortDescending = LocalizedStringResource(
            "designSystem.table.sortDescending",
            defaultValue: "Sorted descending. Activate to sort ascending",
            comment: "Spoken sort state of a PVTable column header sorted descending"
        )

        static func tableFilterColumn(column: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "designSystem.table.filterColumn",
                defaultValue: "Filter %@",
                comment: "Accessibility label for a PVTable column's filter menu; argument is the column title"
            ))
            return String(format: format, locale: .current, column)
        }

        static let selectState = LocalizedStringResource(
            "designSystem.select.state",
            defaultValue: "State",
            comment: "VoiceOver custom-content key for whether a PVSelect menu is open"
        )

        static let selectExpanded = LocalizedStringResource(
            "designSystem.select.expanded",
            defaultValue: "Expanded",
            comment: "VoiceOver value when a PVSelect menu is open"
        )

        static let selectCollapsed = LocalizedStringResource(
            "designSystem.select.collapsed",
            defaultValue: "Collapsed",
            comment: "VoiceOver value when a PVSelect menu is closed"
        )

        static let selectPosition = LocalizedStringResource(
            "designSystem.select.position",
            defaultValue: "Position",
            comment: "VoiceOver custom-content key for the highlighted PVSelect option index"
        )

        static func selectOptionPosition(current: Int, count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "designSystem.select.optionPosition",
                defaultValue: "%1$lld of %2$lld",
                comment: "VoiceOver position of the highlighted PVSelect option; arguments are 1-based index and count"
            ))
            return String(format: format, locale: .current, current, count)
        }

        static func tableFilterOptionCount(label: String, count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "designSystem.table.filterOptionCount",
                defaultValue: "%1$@ (%2$lld)",
                comment: "A PVTable filter menu option with its row count; arguments are the option label and the count"
            ))
            return String(format: format, locale: .current, label, count)
        }

        static let thumbnailEmpty = LocalizedStringResource(
            "designSystem.thumbnail.empty",
            defaultValue: "No preview",
            comment: "Accessibility label for an empty PVThumbnail placeholder"
        )

        static let thumbnailLoading = LocalizedStringResource(
            "designSystem.thumbnail.loading",
            defaultValue: "Loading preview",
            comment: "Accessibility label for a PVThumbnail in the loading state"
        )

        static let markFilePDF = LocalizedStringResource(
            "designSystem.mark.filePdf",
            defaultValue: "PDF file",
            comment: "Accessibility name for the file_pdf evidence glyph"
        )
        static let markFileDoc = LocalizedStringResource(
            "designSystem.mark.fileDoc",
            defaultValue: "Word-processed file",
            comment: "Accessibility name for the file_doc evidence glyph"
        )
        static let markFileTxt = LocalizedStringResource(
            "designSystem.mark.fileTxt",
            defaultValue: "Plain text file",
            comment: "Accessibility name for the file_txt evidence glyph"
        )
        static let markFileSheet = LocalizedStringResource(
            "designSystem.mark.fileSheet",
            defaultValue: "Spreadsheet file",
            comment: "Accessibility name for the file_sheet evidence glyph"
        )
        static let markFileSlides = LocalizedStringResource(
            "designSystem.mark.fileSlides",
            defaultValue: "Presentation file",
            comment: "Accessibility name for the file_slides evidence glyph"
        )
        static let markFileVideo = LocalizedStringResource(
            "designSystem.mark.fileVideo",
            defaultValue: "Video file",
            comment: "Accessibility name for the file_video evidence glyph"
        )
        static let markFileAudio = LocalizedStringResource(
            "designSystem.mark.fileAudio",
            defaultValue: "Audio file",
            comment: "Accessibility name for the file_audio evidence glyph"
        )
        static let markFileImageMissing = LocalizedStringResource(
            "designSystem.mark.fileImageMissing",
            defaultValue: "Image file, preview unavailable",
            comment: "Accessibility name for the file_image_missing evidence glyph"
        )
        static let markFileGeneric = LocalizedStringResource(
            "designSystem.mark.fileGeneric",
            defaultValue: "File of unknown type",
            comment: "Accessibility name for the file_generic evidence glyph"
        )
        static let markTypeCertificate = LocalizedStringResource(
            "designSystem.mark.typeCertificate",
            defaultValue: "Certificate source type",
            comment: "Accessibility name for the type_certificate evidence glyph"
        )
        static let markTypeBook = LocalizedStringResource(
            "designSystem.mark.typeBook",
            defaultValue: "Book source type",
            comment: "Accessibility name for the type_book evidence glyph"
        )
        static let markTypeDocument = LocalizedStringResource(
            "designSystem.mark.typeDocument",
            defaultValue: "Document source type",
            comment: "Accessibility name for the type_document evidence glyph"
        )
        static let markTypeScroll = LocalizedStringResource(
            "designSystem.mark.typeScroll",
            defaultValue: "Register source type",
            comment: "Accessibility name for the type_scroll evidence glyph"
        )
        static let markTypePhotograph = LocalizedStringResource(
            "designSystem.mark.typePhotograph",
            defaultValue: "Photograph source type",
            comment: "Accessibility name for the type_photograph evidence glyph"
        )
        static let markTypeNewspaper = LocalizedStringResource(
            "designSystem.mark.typeNewspaper",
            defaultValue: "Newspaper source type",
            comment: "Accessibility name for the type_newspaper evidence glyph"
        )
        static let markTypeMap = LocalizedStringResource(
            "designSystem.mark.typeMap",
            defaultValue: "Map source type",
            comment: "Accessibility name for the type_map evidence glyph"
        )
        static let markTypeMicrofilm = LocalizedStringResource(
            "designSystem.mark.typeMicrofilm",
            defaultValue: "Microfilm source type",
            comment: "Accessibility name for the type_microfilm evidence glyph"
        )
        static let markTypeCassette = LocalizedStringResource(
            "designSystem.mark.typeCassette",
            defaultValue: "Magnetic tape source type",
            comment: "Accessibility name for the type_cassette evidence glyph"
        )
        static let markTypeOralHistory = LocalizedStringResource(
            "designSystem.mark.typeOralHistory",
            defaultValue: "Oral history source type",
            comment: "Accessibility name for the type_oral_history evidence glyph"
        )
        static let markTypeVideo = LocalizedStringResource(
            "designSystem.mark.typeVideo",
            defaultValue: "Moving image source type",
            comment: "Accessibility name for the type_video evidence glyph"
        )
        static let markTypeWebsite = LocalizedStringResource(
            "designSystem.mark.typeWebsite",
            defaultValue: "Website source type",
            comment: "Accessibility name for the type_website evidence glyph"
        )
        static let markTypeCensus = LocalizedStringResource(
            "designSystem.mark.typeCensus",
            defaultValue: "Census source type",
            comment: "Accessibility name for the type_census evidence glyph"
        )
        static let markTypeDNA = LocalizedStringResource(
            "designSystem.mark.typeDna",
            defaultValue: "DNA match source type",
            comment: "Accessibility name for the type_dna evidence glyph"
        )
        static let markTypeGEDCOM = LocalizedStringResource(
            "designSystem.mark.typeGedcom",
            defaultValue: "GEDCOM source type",
            comment: "Accessibility name for the type_gedcom evidence glyph"
        )
        static let markTypeGrave = LocalizedStringResource(
            "designSystem.mark.typeGrave",
            defaultValue: "Memorial source type",
            comment: "Accessibility name for the type_grave evidence glyph"
        )
        static let markTypeScrapbook = LocalizedStringResource(
            "designSystem.mark.typeScrapbook",
            defaultValue: "Scrapbook source type",
            comment: "Accessibility name for the type_scrapbook evidence glyph"
        )
        static let markTypeEvidence = LocalizedStringResource(
            "designSystem.mark.typeEvidence",
            defaultValue: "Evidence source type",
            comment: "Accessibility name for the type_evidence evidence glyph"
        )
        static let markTypeFolderArchive = LocalizedStringResource(
            "designSystem.mark.typeFolderArchive",
            defaultValue: "Archival container source type",
            comment: "Accessibility name for the type_folder_archive evidence glyph"
        )
        static let markTypeEmail = LocalizedStringResource(
            "designSystem.mark.typeEmail",
            defaultValue: "Correspondence source type",
            comment: "Accessibility name for the type_email evidence glyph"
        )
        static let markTypePostcard = LocalizedStringResource(
            "designSystem.mark.typePostcard",
            defaultValue: "Postcard source type",
            comment: "Accessibility name for the type_postcard evidence glyph"
        )
        static let markTypePassport = LocalizedStringResource(
            "designSystem.mark.typePassport",
            defaultValue: "Passport source type",
            comment: "Accessibility name for the type_passport evidence glyph"
        )

        // Short titles for the Source-type icon picker (design pack names).
        static let markTypeCertificateTitle = LocalizedStringResource(
            "designSystem.mark.typeCertificate.title", defaultValue: "Certificate",
            comment: "Short title for type_certificate in the icon picker"
        )
        static let markTypeBookTitle = LocalizedStringResource(
            "designSystem.mark.typeBook.title", defaultValue: "Book",
            comment: "Short title for type_book in the icon picker"
        )
        static let markTypeDocumentTitle = LocalizedStringResource(
            "designSystem.mark.typeDocument.title", defaultValue: "Document",
            comment: "Short title for type_document in the icon picker"
        )
        static let markTypeScrollTitle = LocalizedStringResource(
            "designSystem.mark.typeScroll.title", defaultValue: "Register",
            comment: "Short title for type_scroll in the icon picker"
        )
        static let markTypePhotographTitle = LocalizedStringResource(
            "designSystem.mark.typePhotograph.title", defaultValue: "Photograph",
            comment: "Short title for type_photograph in the icon picker"
        )
        static let markTypeNewspaperTitle = LocalizedStringResource(
            "designSystem.mark.typeNewspaper.title", defaultValue: "Newspaper",
            comment: "Short title for type_newspaper in the icon picker"
        )
        static let markTypeMapTitle = LocalizedStringResource(
            "designSystem.mark.typeMap.title", defaultValue: "Map",
            comment: "Short title for type_map in the icon picker"
        )
        static let markTypeMicrofilmTitle = LocalizedStringResource(
            "designSystem.mark.typeMicrofilm.title", defaultValue: "Microfilm",
            comment: "Short title for type_microfilm in the icon picker"
        )
        static let markTypeCassetteTitle = LocalizedStringResource(
            "designSystem.mark.typeCassette.title", defaultValue: "Magnetic tape",
            comment: "Short title for type_cassette in the icon picker"
        )
        static let markTypeOralHistoryTitle = LocalizedStringResource(
            "designSystem.mark.typeOralHistory.title", defaultValue: "Oral history",
            comment: "Short title for type_oral_history in the icon picker"
        )
        static let markTypeVideoTitle = LocalizedStringResource(
            "designSystem.mark.typeVideo.title", defaultValue: "Moving image",
            comment: "Short title for type_video in the icon picker"
        )
        static let markTypeWebsiteTitle = LocalizedStringResource(
            "designSystem.mark.typeWebsite.title", defaultValue: "Website",
            comment: "Short title for type_website in the icon picker"
        )
        static let markTypeCensusTitle = LocalizedStringResource(
            "designSystem.mark.typeCensus.title", defaultValue: "Census",
            comment: "Short title for type_census in the icon picker"
        )
        static let markTypeDNATitle = LocalizedStringResource(
            "designSystem.mark.typeDNA.title", defaultValue: "DNA match",
            comment: "Short title for type_dna in the icon picker"
        )
        static let markTypeGEDCOMTitle = LocalizedStringResource(
            "designSystem.mark.typeGEDCOM.title", defaultValue: "GEDCOM",
            comment: "Short title for type_gedcom in the icon picker"
        )
        static let markTypeGraveTitle = LocalizedStringResource(
            "designSystem.mark.typeGrave.title", defaultValue: "Memorial",
            comment: "Short title for type_grave in the icon picker"
        )
        static let markTypeScrapbookTitle = LocalizedStringResource(
            "designSystem.mark.typeScrapbook.title", defaultValue: "Scrapbook",
            comment: "Short title for type_scrapbook in the icon picker"
        )
        static let markTypeEvidenceTitle = LocalizedStringResource(
            "designSystem.mark.typeEvidence.title", defaultValue: "Evidence",
            comment: "Short title for type_evidence in the icon picker"
        )
        static let markTypeFolderArchiveTitle = LocalizedStringResource(
            "designSystem.mark.typeFolderArchive.title", defaultValue: "Archival container",
            comment: "Short title for type_folder_archive in the icon picker"
        )
        static let markTypeEmailTitle = LocalizedStringResource(
            "designSystem.mark.typeEmail.title", defaultValue: "Correspondence",
            comment: "Short title for type_email in the icon picker"
        )
        static let markTypePostcardTitle = LocalizedStringResource(
            "designSystem.mark.typePostcard.title", defaultValue: "Postcard",
            comment: "Short title for type_postcard in the icon picker"
        )
        static let markTypePassportTitle = LocalizedStringResource(
            "designSystem.mark.typePassport.title", defaultValue: "Passport",
            comment: "Short title for type_passport in the icon picker"
        )

        static let markTypeCertificateMetaphor = LocalizedStringResource(
            "designSystem.mark.typeCertificate.metaphor",
            defaultValue: "Ruled formal record under an impressed seal",
            comment: "Metaphor for type_certificate in the icon picker footer"
        )
        static let markTypeBookMetaphor = LocalizedStringResource(
            "designSystem.mark.typeBook.metaphor",
            defaultValue: "Bound volume seen spine-on",
            comment: "Metaphor for type_book in the icon picker footer"
        )
        static let markTypeDocumentMetaphor = LocalizedStringResource(
            "designSystem.mark.typeDocument.metaphor",
            defaultValue: "Loose sheet with a turned corner — a letter, a note, a form",
            comment: "Metaphor for type_document in the icon picker footer"
        )
        static let markTypeScrollMetaphor = LocalizedStringResource(
            "designSystem.mark.typeScroll.metaphor",
            defaultValue: "Parchment rolled at both ends — parish register, roll, cartulary",
            comment: "Metaphor for type_scroll in the icon picker footer"
        )
        static let markTypePhotographMetaphor = LocalizedStringResource(
            "designSystem.mark.typePhotograph.metaphor",
            defaultValue: "A print with its white margin below the image",
            comment: "Metaphor for type_photograph in the icon picker footer"
        )
        static let markTypeNewspaperMetaphor = LocalizedStringResource(
            "designSystem.mark.typeNewspaper.metaphor",
            defaultValue: "Masthead over a photo block and columns",
            comment: "Metaphor for type_newspaper in the icon picker footer"
        )
        static let markTypeMapMetaphor = LocalizedStringResource(
            "designSystem.mark.typeMap.metaphor",
            defaultValue: "Sheet folded into panels",
            comment: "Metaphor for type_map in the icon picker footer"
        )
        static let markTypeMicrofilmMetaphor = LocalizedStringResource(
            "designSystem.mark.typeMicrofilm.metaphor",
            defaultValue: "Reel on its hub — film or fiche as delivered by an archive",
            comment: "Metaphor for type_microfilm in the icon picker footer"
        )
        static let markTypeCassetteMetaphor = LocalizedStringResource(
            "designSystem.mark.typeCassette.metaphor",
            defaultValue: "Shell with two hubs — the physical medium, not the recording",
            comment: "Metaphor for type_cassette in the icon picker footer"
        )
        static let markTypeOralHistoryMetaphor = LocalizedStringResource(
            "designSystem.mark.typeOralHistory.metaphor",
            defaultValue: "A person speaking outward — testimony, not equipment",
            comment: "Metaphor for type_oral_history in the icon picker footer"
        )
        static let markTypeVideoMetaphor = LocalizedStringResource(
            "designSystem.mark.typeVideo.metaphor",
            defaultValue: "Sprocketed strip with a frame to play",
            comment: "Metaphor for type_video in the icon picker footer"
        )
        static let markTypeWebsiteMetaphor = LocalizedStringResource(
            "designSystem.mark.typeWebsite.metaphor",
            defaultValue: "Captured page — window chrome around a globe",
            comment: "Metaphor for type_website in the icon picker footer"
        )
        static let markTypeCensusMetaphor = LocalizedStringResource(
            "designSystem.mark.typeCensus.metaphor",
            defaultValue: "Enumeration schedule — ruled both ways, no prose",
            comment: "Metaphor for type_census in the icon picker footer"
        )
        static let markTypeDNAMetaphor = LocalizedStringResource(
            "designSystem.mark.typeDNA.metaphor",
            defaultValue: "Two strands crossing on three rungs — scientific, not decorative",
            comment: "Metaphor for type_dna in the icon picker footer"
        )
        static let markTypeGEDCOMMetaphor = LocalizedStringResource(
            "designSystem.mark.typeGEDCOM.metaphor",
            defaultValue: "Structured genealogy data — a pedigree bracket inside a file",
            comment: "Metaphor for type_gedcom in the icon picker footer"
        )
        static let markTypeGraveMetaphor = LocalizedStringResource(
            "designSystem.mark.typeGrave.metaphor",
            defaultValue: "Inscribed stone on its plinth — a cemetery record",
            comment: "Metaphor for type_grave in the icon picker footer"
        )
        static let markTypeScrapbookMetaphor = LocalizedStringResource(
            "designSystem.mark.typeScrapbook.metaphor",
            defaultValue: "Album leaf with clippings pasted at an angle",
            comment: "Metaphor for type_scrapbook in the icon picker footer"
        )
        static let markTypeEvidenceMetaphor = LocalizedStringResource(
            "designSystem.mark.typeEvidence.metaphor",
            defaultValue: "A catalogued tag — the fallback for any custom type",
            comment: "Metaphor for type_evidence in the icon picker footer"
        )
        static let markTypeFolderArchiveMetaphor = LocalizedStringResource(
            "designSystem.mark.typeFolderArchive.metaphor",
            defaultValue: "Lidded box with a written label — a box, folder or bundle",
            comment: "Metaphor for type_folder_archive in the icon picker footer"
        )
        static let markTypeEmailMetaphor = LocalizedStringResource(
            "designSystem.mark.typeEmail.metaphor",
            defaultValue: "Digital letter — correspondence that was never on paper",
            comment: "Metaphor for type_email in the icon picker footer"
        )
        static let markTypePostcardMetaphor = LocalizedStringResource(
            "designSystem.mark.typePostcard.metaphor",
            defaultValue: "Divided back — message, stamp, address",
            comment: "Metaphor for type_postcard in the icon picker footer"
        )
        static let markTypePassportMetaphor = LocalizedStringResource(
            "designSystem.mark.typePassport.metaphor",
            defaultValue: "Travel booklet — cover emblem over the title line",
            comment: "Metaphor for type_passport in the icon picker footer"
        )

        static let markSubjectPerson = LocalizedStringResource(
            "designSystem.mark.subjectPerson",
            defaultValue: "Person subject"
        )
        static let markSubjectEvent = LocalizedStringResource(
            "designSystem.mark.subjectEvent",
            defaultValue: "Event subject"
        )
        static let markSubjectPlace = LocalizedStringResource(
            "designSystem.mark.subjectPlace",
            defaultValue: "Place subject"
        )
        static let markSubjectRelationship = LocalizedStringResource(
            "designSystem.mark.subjectRelationship",
            defaultValue: "Relationship bridge"
        )
        static let markSubjectParticipation = LocalizedStringResource(
            "designSystem.mark.subjectParticipation",
            defaultValue: "Participation bridge"
        )
        static let markSubjectLocation = LocalizedStringResource(
            "designSystem.mark.subjectLocation",
            defaultValue: "Location bridge"
        )
        static let markSubjectSource = LocalizedStringResource(
            "designSystem.mark.subjectSource",
            defaultValue: "Source subject"
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

        static let workspaceMissingContext = LocalizedStringResource(
            "onboarding.workspaceMissingContext",
            defaultValue: "Provenencia could not open the workspace because the project or your account is missing. Create or open a project to continue.",
            comment: "Error when entering the workspace without a project directory or user id"
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

    enum Workspace {
        static let navGroupLabel = LocalizedStringResource(
            "workspace.sidebar.navGroupLabel",
            defaultValue: "Source layer",
            comment: "Eyebrow label above the workspace sidebar's nav destinations"
        )

        static let sourcesTitle = LocalizedStringResource(
            "workspace.section.sources.title",
            defaultValue: "Sources",
            comment: "Workspace sidebar destination and page title: Sources"
        )

        static let sourceTypesTitle = LocalizedStringResource(
            "workspace.section.sourceTypes.title",
            defaultValue: "Source types",
            comment: "Workspace sidebar destination and page title: Source types"
        )

        static let sourceFieldsTitle = LocalizedStringResource(
            "workspace.section.sourceFields.title",
            defaultValue: "Source fields",
            comment: "Workspace sidebar destination and page title: Source fields"
        )

        static let subjectFieldsTitle = LocalizedStringResource(
            "workspace.section.subjectFields.title",
            defaultValue: "Subject fields",
            comment: "Workspace sidebar destination and page title: Subject fields"
        )

        static let evidenceGraphTitle = LocalizedStringResource(
            "workspace.section.evidenceGraph.title",
            defaultValue: "Evidence graph",
            comment: "Evidence graph deep place title (toolbar breadcrumb and stub)"
        )

        static func evidenceGraphFor(sourceTitle: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "workspace.section.evidenceGraph.forSource",
                defaultValue: "Evidence graph for %@",
                comment: "Composer breadcrumb segment; argument is the Source title"
            ))
            return String(format: format, locale: .current, sourceTitle)
        }

        static let evidenceGraphStubBody = LocalizedStringResource(
            "workspace.stub.evidenceGraph.body",
            defaultValue: "Interpret subjects on this source — canvas coming soon.",
            comment: "Evidence graph coming-soon stub body"
        )

        static let countsRefreshFailedTitle = LocalizedStringResource(
            "workspace.counts.refreshFailedTitle",
            defaultValue: "Couldn’t refresh catalog counts",
            comment: "Toast title when GetWorkspaceNavCounts fails at workspace appear"
        )

        static let navigationHistoryLoadFailedTitle = LocalizedStringResource(
            "workspace.navigation.historyLoadFailedTitle",
            defaultValue: "Couldn’t restore navigation history",
            comment: "Toast title when navigation history JSON is missing or corrupt on project open"
        )

        static let navigationHistoryLoadFailedBody = LocalizedStringResource(
            "workspace.navigation.historyLoadFailedBody",
            defaultValue: "Starting from Sources. Back and Forward may not match your last session.",
            comment: "Toast body when navigation history could not be loaded; workspace still opens"
        )

        static let navigationHistoryPersistFailedTitle = LocalizedStringResource(
            "workspace.navigation.historyPersistFailedTitle",
            defaultValue: "Couldn’t save navigation history",
            comment: "Toast title when writing navigation history JSON fails"
        )

        static let navigationHistoryPersistFailedBody = LocalizedStringResource(
            "workspace.navigation.historyPersistFailedBody",
            defaultValue: "Back and Forward still work now, but may not survive quitting the app.",
            comment: "Toast body when navigation history could not be persisted"
        )

        static let collapseSidebar = LocalizedStringResource(
            "workspace.sidebar.collapse",
            defaultValue: "Collapse labels",
            comment: "Tooltip/accessibility label for the sidebar toggle when expanded"
        )

        static let expandSidebar = LocalizedStringResource(
            "workspace.sidebar.expand",
            defaultValue: "Show labels",
            comment: "Tooltip/accessibility label for the sidebar toggle when collapsed"
        )

        static let goBack = LocalizedStringResource(
            "workspace.navigation.goBack",
            defaultValue: "Back",
            comment: "Menu / keyboard command to go back in workspace navigation history (⌘[)"
        )

        static let goForward = LocalizedStringResource(
            "workspace.navigation.goForward",
            defaultValue: "Forward",
            comment: "Menu / keyboard command to go forward in workspace navigation history (⌘])"
        )

        static let focusOmnibar = LocalizedStringResource(
            "workspace.navigation.focusOmnibar",
            defaultValue: "Focus Search",
            comment: "Menu / keyboard command to focus the workspace omnibar field (⌘K)"
        )

        static let omnibarPlaceholder = LocalizedStringResource(
            "workspace.omnibar.placeholder",
            defaultValue: "Search people, sources, places and files",
            comment: "Placeholder in the main-column toolbar omnibar field shell"
        )

        static let omnibarAccessibilityLabel = LocalizedStringResource(
            "workspace.omnibar.accessibilityLabel",
            defaultValue: "Search everything",
            comment: "Accessibility label for the toolbar omnibar field"
        )

        static let omnibarKindSource = LocalizedStringResource(
            "workspace.omnibar.kind.source",
            defaultValue: "Source",
            comment: "Kind chip on an omnibar hit for a Source"
        )

        static let omnibarKindType = LocalizedStringResource(
            "workspace.omnibar.kind.type",
            defaultValue: "Type",
            comment: "Kind chip on an omnibar hit for a source type"
        )

        static let omnibarKindField = LocalizedStringResource(
            "workspace.omnibar.kind.field",
            defaultValue: "Field",
            comment: "Kind chip on an omnibar hit for a source field"
        )

        static func omnibarNoMatchesTitle(query: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "workspace.omnibar.noMatchesTitle",
                defaultValue: "No matches for “%@”.",
                comment: "Omnibar empty-state title; argument is the typed query"
            ))
            return String(format: format, locale: .current, query)
        }

        static let omnibarNoMatchesHint = LocalizedStringResource(
            "workspace.omnibar.noMatchesHint",
            defaultValue: "Check the spelling, or paste a catalog reference such as SRC-3K9M2",
            comment: "Omnibar empty-state hint under no matches"
        )

        static let omnibarResultsAccessibilityLabel = LocalizedStringResource(
            "workspace.omnibar.resultsAccessibilityLabel",
            defaultValue: "Search results",
            comment: "Accessibility label for the omnibar results panel"
        )

        static let omnibarSearchFailed = LocalizedStringResource(
            "workspace.omnibar.searchFailed",
            defaultValue: "Search couldn’t finish. Try again.",
            comment: "Omnibar panel message when SearchCatalog fails"
        )

        static let omnibarSearching = LocalizedStringResource(
            "workspace.omnibar.searching",
            defaultValue: "Searching…",
            comment: "Omnibar panel loading label while SearchCatalog is in flight"
        )

        /// Omnibar match-context line for a note body hit; argument is the raw snippet.
        static func omnibarMatchNote(snippet: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "workspace.omnibar.match.note",
                defaultValue: "Note: %@",
                comment: "Omnibar match context when the query hit a Source note; argument is a short snippet"
            ))
            return String(format: format, locale: .current, snippet)
        }

        /// Omnibar match-context line for metadata text; argument is the raw snippet.
        static func omnibarMatchMetadata(snippet: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "workspace.omnibar.match.metadata",
                defaultValue: "Metadata: %@",
                comment: "Omnibar match context when the query hit Source metadata; argument is a short snippet"
            ))
            return String(format: format, locale: .current, snippet)
        }

        /// Omnibar match-context line for a filename/artifact label; argument is the raw snippet.
        static func omnibarMatchFilename(snippet: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "workspace.omnibar.match.filename",
                defaultValue: "Filename: %@",
                comment: "Omnibar match context when the query hit a filename or artifact label; argument is a short snippet"
            ))
            return String(format: format, locale: .current, snippet)
        }

        static let omnibarMatchDescription = LocalizedStringResource(
            "workspace.omnibar.match.description",
            defaultValue: "Description",
            comment: "Omnibar match context when the query hit a description field"
        )

        static let jumpMenuAccessibilityLabel = LocalizedStringResource(
            "workspace.navigation.jumpMenuAccessibilityLabel",
            defaultValue: "History jump menu",
            comment: "Accessibility label for the Back/Forward history jump menu panel"
        )
    }

    enum EvidenceGraph {
        static let emptyTitle = LocalizedStringResource(
            "evidenceGraph.empty.title",
            defaultValue: "No subjects yet",
            comment: "Empty Evidence graph title before any Person/Event/Place cards exist"
        )

        static let emptyMessage = LocalizedStringResource(
            "evidenceGraph.empty.message",
            defaultValue: "Pick a tool above, then click the grid to place the first person, event or place from this source.",
            comment: "Empty Evidence graph body guiding the researcher to arm a palette tool"
        )

        static let uncitedAccessibility = LocalizedStringResource(
            "evidenceGraph.subject.uncitedAccessibility",
            defaultValue: "Uncited — no citations attached yet",
            comment: "VoiceOver fragment when a primary subject card has no Observations"
        )

        static let citedAccessibility = LocalizedStringResource(
            "evidenceGraph.subject.citedAccessibility",
            defaultValue: "Cited",
            comment: "VoiceOver fragment when a primary subject card has at least one Observation"
        )

        static let subjectsRotor = LocalizedStringResource(
            "evidenceGraph.rotor.subjects",
            defaultValue: "Subjects",
            comment: "VoiceOver rotor name for jumping between Evidence graph subject cards"
        )

        static let toolPerson = LocalizedStringResource(
            "evidenceGraph.palette.person",
            defaultValue: "Add person",
            comment: "Evidence graph palette tool to place a Person subject"
        )

        static let toolEvent = LocalizedStringResource(
            "evidenceGraph.palette.event",
            defaultValue: "Add event",
            comment: "Evidence graph palette tool to place an Event subject"
        )

        static let toolPlace = LocalizedStringResource(
            "evidenceGraph.palette.place",
            defaultValue: "Add place",
            comment: "Evidence graph palette tool to place a Place subject"
        )

        static let toolRole = LocalizedStringResource(
            "evidenceGraph.palette.toolRole",
            defaultValue: "tool",
            comment: "VoiceOver middle token for a palette toggle (Add person, tool, off)"
        )

        static let toolOn = LocalizedStringResource(
            "evidenceGraph.palette.toolOn",
            defaultValue: "on",
            comment: "VoiceOver state when a palette tool is armed"
        )

        static let toolOff = LocalizedStringResource(
            "evidenceGraph.palette.toolOff",
            defaultValue: "off",
            comment: "VoiceOver state when a palette tool is idle"
        )

        static let armedHintPerson = LocalizedStringResource(
            "evidenceGraph.armed.hintPerson",
            defaultValue: "Click the grid to place a person",
            comment: "Banner when Add person is armed"
        )

        static let armedHintEvent = LocalizedStringResource(
            "evidenceGraph.armed.hintEvent",
            defaultValue: "Click the grid to place an event",
            comment: "Banner when Add event is armed"
        )

        static let armedHintPlace = LocalizedStringResource(
            "evidenceGraph.armed.hintPlace",
            defaultValue: "Click the grid to place a place",
            comment: "Banner when Add place is armed"
        )

        static let armedEscHint = LocalizedStringResource(
            "evidenceGraph.armed.escHint",
            defaultValue: "esc to cancel",
            comment: "Secondary hint beside the armed placement banner"
        )

        static let addPersonTitle = LocalizedStringResource(
            "evidenceGraph.create.personTitle",
            defaultValue: "Add person",
            comment: "Create-subject dialog title when placing a Person"
        )

        static let addEventTitle = LocalizedStringResource(
            "evidenceGraph.create.eventTitle",
            defaultValue: "Add event",
            comment: "Create-subject dialog title when placing an Event"
        )

        static let addPlaceTitle = LocalizedStringResource(
            "evidenceGraph.create.placeTitle",
            defaultValue: "Add place",
            comment: "Create-subject dialog title when placing a Place"
        )

        static let createConfirm = LocalizedStringResource(
            "evidenceGraph.create.confirm",
            defaultValue: "Add",
            comment: "Confirm button on the Evidence graph create-subject dialog"
        )

        static let createCancel = LocalizedStringResource(
            "evidenceGraph.create.cancel",
            defaultValue: "Cancel",
            comment: "Cancel button on the Evidence graph create-subject dialog"
        )

        static let editPersonTitle = LocalizedStringResource(
            "evidenceGraph.edit.personTitle",
            defaultValue: "Edit person",
            comment: "Edit-subject dialog title for a Person"
        )

        static let editEventTitle = LocalizedStringResource(
            "evidenceGraph.edit.eventTitle",
            defaultValue: "Edit event",
            comment: "Edit-subject dialog title for an Event"
        )

        static let editPlaceTitle = LocalizedStringResource(
            "evidenceGraph.edit.placeTitle",
            defaultValue: "Edit place",
            comment: "Edit-subject dialog title for a Place"
        )

        static let editRelationshipTitle = LocalizedStringResource(
            "evidenceGraph.edit.relationshipTitle",
            defaultValue: "Edit relationship",
            comment: "Edit-subject dialog title for a relationship bridge"
        )

        static let editParticipationTitle = LocalizedStringResource(
            "evidenceGraph.edit.participationTitle",
            defaultValue: "Edit participation",
            comment: "Edit-subject dialog title for a participation bridge"
        )

        static let editLocationTitle = LocalizedStringResource(
            "evidenceGraph.edit.locationTitle",
            defaultValue: "Edit location",
            comment: "Edit-subject dialog title for a location bridge"
        )

        static let editConfirm = LocalizedStringResource(
            "evidenceGraph.edit.confirm",
            defaultValue: "Save",
            comment: "Confirm button on the Evidence graph edit-subject dialog"
        )

        static let editAccessibility = LocalizedStringResource(
            "evidenceGraph.subject.editAccessibility",
            defaultValue: "Edit label and description",
            comment: "VoiceOver action / tooltip for the subject card edit pencil"
        )

        static let editPropertyAccessibility = LocalizedStringResource(
            "evidenceGraph.subject.editPropertyAccessibility",
            defaultValue: "Edit citation",
            comment: "VoiceOver action for a cited property row pencil that opens the citation composer"
        )

        static let editCitationAccessibility = LocalizedStringResource(
            "evidenceGraph.bridge.editCitationAccessibility",
            defaultValue: "Edit citation",
            comment: "VoiceOver action for the pencil beside a bridge card’s relationship sentence"
        )

        static let deleteAccessibility = LocalizedStringResource(
            "evidenceGraph.subject.deleteAccessibility",
            defaultValue: "Delete subject",
            comment: "VoiceOver action / tooltip for trash on an uncited subject card"
        )

        /// Board `PVConfirm` title for uncited subject delete.
        static let deleteConfirmTitle = LocalizedStringResource(
            "evidenceGraph.subject.deleteConfirmTitle",
            defaultValue: "Do you want to delete this?",
            comment: "Title on the PVConfirm sheet when deleting an uncited Evidence graph subject"
        )

        /// Board message: name (+ ref) and description removed; no citations so no evidence lost.
        static func deleteConfirmMessage(label: String, ref: String) -> String {
            let trimmedRef = ref.trimmingCharacters(in: .whitespacesAndNewlines)
            if trimmedRef.isEmpty {
                let format = String(localized: LocalizedStringResource(
                    "evidenceGraph.subject.deleteConfirmMessageNoRef",
                    defaultValue: "%@ and its description are removed from this graph. It has no citations, so no evidence is lost.",
                    comment: "Delete-confirm body when the subject has no ref; argument is the label"
                ))
                return String(format: format, locale: .current, label)
            }
            let format = String(localized: LocalizedStringResource(
                "evidenceGraph.subject.deleteConfirmMessage",
                defaultValue: "%@ (%@) and its description are removed from this graph. It has no citations, so no evidence is lost.",
                comment: "Delete-confirm body; arguments are subject label then catalog ref"
            ))
            return String(format: format, locale: .current, label, trimmedRef)
        }

        static let deleteConfirm = LocalizedStringResource(
            "evidenceGraph.subject.deleteConfirm",
            defaultValue: "Delete",
            comment: "Confirm button on the Evidence graph delete-subject sheet"
        )

        static let deleteCancel = LocalizedStringResource(
            "evidenceGraph.subject.deleteCancel",
            defaultValue: "Keep",
            comment: "Cancel button on the Evidence graph delete-subject sheet"
        )

        static let addProperty = LocalizedStringResource(
            "evidenceGraph.subject.addProperty",
            defaultValue: "Add property",
            comment: "Quiet footer control on primary cards that opens the citation composer"
        )

        static let citedValueUnavailable = LocalizedStringResource(
            "evidenceGraph.subject.citedValueUnavailable",
            defaultValue: "—",
            comment: "Placeholder when a cited Observation has no displayable value summary yet"
        )

        static let noArtifactTitle = LocalizedStringResource(
            "evidenceGraph.noArtifact.title",
            defaultValue: "No Artifacts on this source",
            comment: "Title on the graph-wide No-Artifact warning callout"
        )

        static let noArtifactMessage = LocalizedStringResource(
            "evidenceGraph.noArtifact.message",
            defaultValue: "Citing needs an Artifact. Place tools, Connect, and Add property stay disabled until you attach one.",
            comment: "Body on the graph-wide No-Artifact warning callout"
        )

        static let noArtifactAction = LocalizedStringResource(
            "evidenceGraph.noArtifact.action",
            defaultValue: "Add an Artifact",
            comment: "Callout action that navigates to the Source page to attach an Artifact"
        )

        static let labelField = LocalizedStringResource(
            "evidenceGraph.create.labelField",
            defaultValue: "Label",
            comment: "Label field title on the create-subject dialog"
        )

        static let workingLabelHint = LocalizedStringResource(
            "evidenceGraph.create.workingLabelHint",
            defaultValue: "Working label. Cards show the cited name once one is recorded.",
            comment: "Caption under Label on the graph primary create/edit dialog"
        )

        static let descriptionField = LocalizedStringResource(
            "evidenceGraph.create.descriptionField",
            defaultValue: "Description",
            comment: "Description field title on the create-subject dialog"
        )

        static let labelRequired = LocalizedStringResource(
            "evidenceGraph.create.labelRequired",
            defaultValue: "Enter a working label",
            comment: "Validation when create-subject label is blank"
        )

        static let typesUnavailable = LocalizedStringResource(
            "evidenceGraph.create.typesUnavailable",
            defaultValue: "Subject types are not available for this project",
            comment: "Error when seeded Person/Event/Place types could not be loaded"
        )

        static let defaultLabelPerson = LocalizedStringResource(
            "evidenceGraph.create.defaultLabelPerson",
            defaultValue: "New person",
            comment: "Prefill label when creating a Person from the palette"
        )

        static let defaultLabelEvent = LocalizedStringResource(
            "evidenceGraph.create.defaultLabelEvent",
            defaultValue: "New event",
            comment: "Prefill label when creating an Event from the palette"
        )

        static let defaultLabelPlace = LocalizedStringResource(
            "evidenceGraph.create.defaultLabelPlace",
            defaultValue: "New place",
            comment: "Prefill label when creating a Place from the palette"
        )

        static let positionPersistFailedTitle = LocalizedStringResource(
            "evidenceGraph.position.persistFailedTitle",
            defaultValue: "Couldn't save position",
            comment: "Toast title when setSubjectPosition fails after drag or arrow move"
        )

        static let toolConnect = LocalizedStringResource(
            "evidenceGraph.palette.connect",
            defaultValue: "Connect",
            comment: "Evidence graph palette tool to link two primary subjects with a bridge"
        )

        static let armedHintConnect = LocalizedStringResource(
            "evidenceGraph.armed.hintConnect",
            defaultValue: "Select the first subject",
            comment: "Banner when Connect is armed and no origin is chosen yet"
        )

        static let armedHintConnectPickB = LocalizedStringResource(
            "evidenceGraph.armed.hintConnectPickB",
            defaultValue: "Select the second subject",
            comment: "Banner when Connect has origin A and is waiting for B"
        )

        static let connectingFrom = LocalizedStringResource(
            "evidenceGraph.connect.connectingFrom",
            defaultValue: "connecting from",
            comment: "Mono status line on primary card A while Connect waits for B"
        )

        static let addRelationshipTitle = LocalizedStringResource(
            "evidenceGraph.create.relationshipTitle",
            defaultValue: "Add relationship",
            comment: "Create-bridge dialog title for a relationship mid-card"
        )

        static let addParticipationTitle = LocalizedStringResource(
            "evidenceGraph.create.participationTitle",
            defaultValue: "Add participation",
            comment: "Create-bridge dialog title for a participation mid-card"
        )

        static let addLocationTitle = LocalizedStringResource(
            "evidenceGraph.create.locationTitle",
            defaultValue: "Add location",
            comment: "Create-bridge dialog title for a location mid-card"
        )

        static let defaultLabelRelationship = LocalizedStringResource(
            "evidenceGraph.create.defaultLabelRelationship",
            defaultValue: "New relationship",
            comment: "Prefill label when creating a relationship bridge"
        )

        static let defaultLabelParticipation = LocalizedStringResource(
            "evidenceGraph.create.defaultLabelParticipation",
            defaultValue: "New participation",
            comment: "Prefill label when creating a participation bridge"
        )

        static let defaultLabelLocation = LocalizedStringResource(
            "evidenceGraph.create.defaultLabelLocation",
            defaultValue: "New location",
            comment: "Prefill label when creating a location bridge"
        )

        static let bridgeHonestyBody = LocalizedStringResource(
            "evidenceGraph.bridge.honestyBody",
            defaultValue: "Prototype link — not cited evidence",
            comment: "Permanent honesty copy on S6-04 bridge cards"
        )

        static let bridgeHonestyAccessibility = LocalizedStringResource(
            "evidenceGraph.bridge.honestyAccessibility",
            defaultValue: "Prototype link, not cited evidence",
            comment: "VoiceOver fragment for bridge honesty"
        )

        /// Location: "{event} took place in {place}".
        static func bridgeSummaryLocation(event: String, place: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "evidenceGraph.bridge.summary.location",
                defaultValue: "%@ took place in %@",
                comment: "Location edge summary; arguments are event label then place label"
            ))
            return String(format: format, locale: .current, event, place)
        }

        /// Location mid-phrase when endpoint labels are incomplete.
        static let bridgeSummaryLocationBare = LocalizedStringResource(
            "evidenceGraph.bridge.summary.locationBare",
            defaultValue: "Took place in",
            comment: "Location edge summary when event/place endpoint labels are missing"
        )

        /// Relationship: "{person} is the {type} of {related_to}".
        static func bridgeSummaryRelationship(person: String, type: String, related: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "evidenceGraph.bridge.summary.relationship",
                defaultValue: "%@ is the %@ of %@",
                comment: "Relationship edge summary; arguments are person, relationship_type, related_to"
            ))
            return String(format: format, locale: .current, person, type, related)
        }

        /// Relationship without type: "{person} is related to {related_to}".
        static func bridgeSummaryRelationshipFallback(person: String, related: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "evidenceGraph.bridge.summary.relationshipFallback",
                defaultValue: "%@ is related to %@",
                comment: "Relationship edge summary without relationship_type; person then related_to"
            ))
            return String(format: format, locale: .current, person, related)
        }

        /// Relationship mid-phrase when only the type term is known.
        static func bridgeSummaryRelationshipTypeOnly(type: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "evidenceGraph.bridge.summary.relationshipTypeOnly",
                defaultValue: "Is the %@ of",
                comment: "Relationship edge summary when endpoint labels are missing; argument is type"
            ))
            return String(format: format, locale: .current, type)
        }

        static let bridgeSummaryRelationshipBare = LocalizedStringResource(
            "evidenceGraph.bridge.summary.relationshipBare",
            defaultValue: "Is related to",
            comment: "Relationship edge summary when type and endpoint labels are missing"
        )

        /// Participation with role: "{person} participated as {role} at {event}".
        static func bridgeSummaryParticipation(person: String, role: String, event: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "evidenceGraph.bridge.summary.participation",
                defaultValue: "%@ participated as %@ at %@",
                comment: "Participation edge summary; arguments are person, role, event"
            ))
            return String(format: format, locale: .current, person, role, event)
        }

        /// Participation without role: "{person} participated in {event}".
        static func bridgeSummaryParticipationFallback(person: String, event: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "evidenceGraph.bridge.summary.participationFallback",
                defaultValue: "%@ participated in %@",
                comment: "Participation edge summary without role; arguments are person then event"
            ))
            return String(format: format, locale: .current, person, event)
        }

        /// Participation mid-phrase when only the role term is known.
        static func bridgeSummaryParticipationRoleOnly(role: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "evidenceGraph.bridge.summary.participationRoleOnly",
                defaultValue: "Participated as %@",
                comment: "Participation edge summary when endpoint labels are missing; argument is role"
            ))
            return String(format: format, locale: .current, role)
        }

        static let bridgeSummaryParticipationBare = LocalizedStringResource(
            "evidenceGraph.bridge.summary.participationBare",
            defaultValue: "Participated in",
            comment: "Participation edge summary when role and endpoint labels are missing"
        )

        static let linksRotor = LocalizedStringResource(
            "evidenceGraph.rotor.links",
            defaultValue: "Links",
            comment: "VoiceOver rotor name for jumping between Evidence graph bridge cards"
        )

        static let connectInvalidPairTitle = LocalizedStringResource(
            "evidenceGraph.connect.invalidPairTitle",
            defaultValue: "Can't connect those",
            comment: "Toast title when Connect picks an unsupported primary pair"
        )

        static let connectRulesUnavailableTitle = LocalizedStringResource(
            "evidenceGraph.connect.rulesUnavailableTitle",
            defaultValue: "Connect is unavailable",
            comment: "Toast title when the connect-rules catalog read fails"
        )

        static let connectInvalidPairBody = LocalizedStringResource(
            "evidenceGraph.connect.invalidPairBody",
            defaultValue: "Try person↔event, event↔place, or two people. Person and place cannot be linked directly.",
            comment: "Toast body explaining which primary pairs Connect accepts; person↔place is refused"
        )

        /// Endpoint noun fallback: "{type} {ref}".
        static func bridgeNounTypeAndRef(type: String, ref: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "evidenceGraph.bridge.nounTypeAndRef",
                defaultValue: "%@ %@",
                comment: "Bridge endpoint noun when the working label is blank; type label then ref"
            ))
            return String(format: format, locale: .current, type, ref)
        }

        /// Whole-name fallback: "{kind phrase} · {ref}".
        static func bridgeNameKindAndRef(phrase: String, ref: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "evidenceGraph.bridge.nameKindAndRef",
                defaultValue: "%@ · %@",
                comment: "Bridge name when edges cannot be read; kind phrase then bridge ref"
            ))
            return String(format: format, locale: .current, phrase, ref)
        }

        static func subjectCount(count: Int) -> LocalizedStringResource {
            count == 1 ? subjectCountOne : subjectCountOther(count: count)
        }

        private static let subjectCountOne = LocalizedStringResource(
            "evidenceGraph.header.subjectCountOne",
            defaultValue: "1 subject",
            comment: "Evidence graph header count when exactly one placed primary is on the canvas"
        )

        private static func subjectCountOther(count: Int) -> LocalizedStringResource {
            LocalizedStringResource(
                "evidenceGraph.header.subjectCountOther",
                defaultValue: "\(count) subjects",
                comment: "Evidence graph header count; argument is how many placed primaries are on the canvas"
            )
        }
    }

    /// Product Property-term display names (`propertyTerm.<propertyKey>.<termKey>`).
    enum PropertyTerm {
        static func resource(propertyKey: String, termKey: String) -> LocalizedStringResource? {
            switch (propertyKey, termKey) {
            case ("role", "subject"): return roleSubject
            case ("role", "father"): return roleFather
            case ("role", "mother"): return roleMother
            case ("role", "spouse"): return roleSpouse
            case ("role", "child"): return roleChild
            case ("role", "witness"): return roleWitness
            case ("role", "informant"): return roleInformant
            case ("relationship_type", "spouse"): return relationshipSpouse
            case ("relationship_type", "sibling"): return relationshipSibling
            case ("relationship_type", "cousin"): return relationshipCousin
            case ("relationship_type", "parent"): return relationshipParent
            case ("relationship_type", "child"): return relationshipChild
            case ("relationship_type", "grandparent"): return relationshipGrandparent
            case ("relationship_type", "grandchild"): return relationshipGrandchild
            case ("relationship_type", "pibling"): return relationshipPibling
            case ("relationship_type", "nibling"): return relationshipNibling
            case ("relationship_type", "guardian"): return relationshipGuardian
            case ("relationship_type", "ward"): return relationshipWard
            default: return nil
            }
        }

        static let roleSubject = LocalizedStringResource(
            "propertyTerm.role.subject",
            defaultValue: "Subject",
            comment: "Product role term: subject"
        )
        static let roleFather = LocalizedStringResource(
            "propertyTerm.role.father",
            defaultValue: "Father",
            comment: "Product role term: father"
        )
        static let roleMother = LocalizedStringResource(
            "propertyTerm.role.mother",
            defaultValue: "Mother",
            comment: "Product role term: mother"
        )
        static let roleSpouse = LocalizedStringResource(
            "propertyTerm.role.spouse",
            defaultValue: "Spouse",
            comment: "Product role term: spouse"
        )
        static let roleChild = LocalizedStringResource(
            "propertyTerm.role.child",
            defaultValue: "Child",
            comment: "Product role term: child"
        )
        static let roleWitness = LocalizedStringResource(
            "propertyTerm.role.witness",
            defaultValue: "Witness",
            comment: "Product role term: witness"
        )
        static let roleInformant = LocalizedStringResource(
            "propertyTerm.role.informant",
            defaultValue: "Informant",
            comment: "Product role term: informant"
        )
        static let relationshipSpouse = LocalizedStringResource(
            "propertyTerm.relationship_type.spouse",
            defaultValue: "Spouse",
            comment: "Product relationship_type term: spouse"
        )
        static let relationshipSibling = LocalizedStringResource(
            "propertyTerm.relationship_type.sibling",
            defaultValue: "Sibling",
            comment: "Product relationship_type term: sibling"
        )
        static let relationshipCousin = LocalizedStringResource(
            "propertyTerm.relationship_type.cousin",
            defaultValue: "Cousin",
            comment: "Product relationship_type term: cousin"
        )
        static let relationshipParent = LocalizedStringResource(
            "propertyTerm.relationship_type.parent",
            defaultValue: "Parent",
            comment: "Product relationship_type term: parent"
        )
        static let relationshipChild = LocalizedStringResource(
            "propertyTerm.relationship_type.child",
            defaultValue: "Child",
            comment: "Product relationship_type term: child"
        )
        static let relationshipGrandparent = LocalizedStringResource(
            "propertyTerm.relationship_type.grandparent",
            defaultValue: "Grandparent",
            comment: "Product relationship_type term: grandparent"
        )
        static let relationshipGrandchild = LocalizedStringResource(
            "propertyTerm.relationship_type.grandchild",
            defaultValue: "Grandchild",
            comment: "Product relationship_type term: grandchild"
        )
        static let relationshipPibling = LocalizedStringResource(
            "propertyTerm.relationship_type.pibling",
            defaultValue: "Aunt / uncle",
            comment: "Product relationship_type term: pibling"
        )
        static let relationshipNibling = LocalizedStringResource(
            "propertyTerm.relationship_type.nibling",
            defaultValue: "Niece / nephew",
            comment: "Product relationship_type term: nibling"
        )
        static let relationshipGuardian = LocalizedStringResource(
            "propertyTerm.relationship_type.guardian",
            defaultValue: "Guardian",
            comment: "Product relationship_type term: guardian"
        )
        static let relationshipWard = LocalizedStringResource(
            "propertyTerm.relationship_type.ward",
            defaultValue: "Ward",
            comment: "Product relationship_type term: ward"
        )
    }

    /// Shared genealogical DateValue display (list rows, previews). Not Source-page-owned.
    enum Dates {
        static func displayAbout(_ date: String, locale: Locale = .autoupdatingCurrent) -> String {
            var resource = LocalizedStringResource(
                "dates.display.about",
                defaultValue: "About %@",
                comment: "DateValue summary prefix for ABT; argument is the formatted point date"
            )
            resource.locale = locale
            return String(format: String(localized: resource), locale: locale, date)
        }

        static func displayBefore(_ date: String, locale: Locale = .autoupdatingCurrent) -> String {
            var resource = LocalizedStringResource(
                "dates.display.before",
                defaultValue: "Before %@",
                comment: "DateValue summary prefix for BEF; argument is the formatted point date"
            )
            resource.locale = locale
            return String(format: String(localized: resource), locale: locale, date)
        }

        static func displayAfter(_ date: String, locale: Locale = .autoupdatingCurrent) -> String {
            var resource = LocalizedStringResource(
                "dates.display.after",
                defaultValue: "After %@",
                comment: "DateValue summary prefix for AFT; argument is the formatted point date"
            )
            resource.locale = locale
            return String(format: String(localized: resource), locale: locale, date)
        }

        static func displayBetween(start: String, end: String, locale: Locale = .autoupdatingCurrent) -> String {
            var resource = LocalizedStringResource(
                "dates.display.between",
                defaultValue: "Between %@ and %@",
                comment: "DateValue summary for a range; arguments are formatted start then end"
            )
            resource.locale = locale
            return String(format: String(localized: resource), locale: locale, start, end)
        }
    }

    /// Shared genealogical NameValue editor (S7-02b / S7-D5). Not composer-owned.
    enum NameValue {
        static let titleAdd = LocalizedStringResource(
            "nameValue.title.add",
            defaultValue: "Add name",
            comment: "NameValue editor dialog title when creating"
        )
        static let titleEdit = LocalizedStringResource(
            "nameValue.title.edit",
            defaultValue: "Edit name",
            comment: "NameValue editor dialog title when editing"
        )
        static let confirmAdd = LocalizedStringResource(
            "nameValue.confirm.add",
            defaultValue: "Add name",
            comment: "NameValue editor confirm verb when creating"
        )
        static let confirmSave = LocalizedStringResource(
            "nameValue.confirm.save",
            defaultValue: "Save name",
            comment: "NameValue editor confirm verb when editing"
        )
        static let cancel = LocalizedStringResource(
            "nameValue.cancel",
            defaultValue: "Cancel",
            comment: "NameValue editor cancel button"
        )
        static let formLabel = LocalizedStringResource(
            "nameValue.form.label",
            defaultValue: "Full form",
            comment: "Required NameValue form field label"
        )
        static let formRequired = LocalizedStringResource(
            "nameValue.form.required",
            defaultValue: "required",
            comment: "Required badge next to the NameValue form label"
        )
        static let formHint = LocalizedStringResource(
            "nameValue.form.hint",
            defaultValue: "The name as you read it, in normalized spelling — the source's own wording stays on the citation",
            comment: "Hint under the NameValue form field"
        )
        static let formPlaceholder = LocalizedStringResource(
            "nameValue.form.placeholder",
            defaultValue: "John William Alderwick",
            comment: "Placeholder in the NameValue full-form field"
        )
        static let formErrorMissing = LocalizedStringResource(
            "nameValue.form.error.missing",
            defaultValue: "Enter the name as one full form. Parts are optional; the form is not.",
            comment: "Validation error when NameValue form is empty"
        )
        static let partsHeading = LocalizedStringResource(
            "nameValue.parts.heading",
            defaultValue: "Parts · optional",
            comment: "Heading above the optional NameValue parts list"
        )
        static let partsEmpty = LocalizedStringResource(
            "nameValue.parts.empty",
            defaultValue: "No parts — the form stands alone.",
            comment: "Empty-state title when NameValue has no parts"
        )
        static let partsEmptyHint = LocalizedStringResource(
            "nameValue.parts.empty.hint",
            defaultValue: "Add parts only when the name segments cleanly and the segments matter to your research",
            comment: "Empty-state hint when NameValue has no parts"
        )
        static let partsAdd = LocalizedStringResource(
            "nameValue.parts.add",
            defaultValue: "Add part",
            comment: "Button that appends a NameValue part row"
        )

        static func partsCount(_ count: Int) -> String {
            if count == 0 {
                return String(localized: LocalizedStringResource(
                    "nameValue.parts.count.none",
                    defaultValue: "none",
                    comment: "Parts heading count when the NameValue has no parts"
                ))
            }
            if count == 1 {
                return String(localized: LocalizedStringResource(
                    "nameValue.parts.count.one",
                    defaultValue: "1 part",
                    comment: "Parts heading count for a single NameValue part"
                ))
            }
            let format = String(localized: LocalizedStringResource(
                "nameValue.parts.count.other",
                defaultValue: "%lld parts",
                comment: "Parts heading count; argument is the part count"
            ))
            return String(format: format, locale: .current, count)
        }
        static let partsHint = LocalizedStringResource(
            "nameValue.parts.hint",
            defaultValue: "Order is the order you enter. Leave a part untyped when no type fits",
            comment: "Hint under the NameValue parts list"
        )
        static let partValueLabel = LocalizedStringResource(
            "nameValue.part.value.label",
            defaultValue: "Value",
            comment: "Label for a NameValue part value field"
        )
        static let partTypeLabel = LocalizedStringResource(
            "nameValue.part.type.label",
            defaultValue: "Type",
            comment: "Label for a NameValue part type picker"
        )
        static let partTypeNone = LocalizedStringResource(
            "nameValue.part.type.none",
            defaultValue: "No type",
            comment: "Picker option for an untyped NameValue part"
        )
        static let partTypePrefix = LocalizedStringResource(
            "nameValue.part.type.prefix",
            defaultValue: "Prefix",
            comment: "NameValue part type: prefix"
        )
        static let partTypeGiven = LocalizedStringResource(
            "nameValue.part.type.given",
            defaultValue: "Given name",
            comment: "NameValue part type: given"
        )
        static let partTypeInitial = LocalizedStringResource(
            "nameValue.part.type.initial",
            defaultValue: "Initial",
            comment: "NameValue part type: initial"
        )
        static let partTypeNick = LocalizedStringResource(
            "nameValue.part.type.nick",
            defaultValue: "Nickname",
            comment: "NameValue part type: nick"
        )
        static let partTypeSurnamePrefix = LocalizedStringResource(
            "nameValue.part.type.surname_prefix",
            defaultValue: "Surname prefix",
            comment: "NameValue part type: surname_prefix"
        )
        static let partTypeSurname = LocalizedStringResource(
            "nameValue.part.type.surname",
            defaultValue: "Surname",
            comment: "NameValue part type: surname"
        )
        static let partTypeSuffix = LocalizedStringResource(
            "nameValue.part.type.suffix",
            defaultValue: "Suffix",
            comment: "NameValue part type: suffix"
        )
        static let partTypeUndetermined = LocalizedStringResource(
            "nameValue.part.type.undetermined",
            defaultValue: "Undetermined",
            comment: "NameValue part type: undetermined"
        )
        static let storedAs = LocalizedStringResource(
            "nameValue.storedAs",
            defaultValue: "Stored as",
            comment: "Heading for the NameValue draft readout"
        )
        static let summaryAdd = LocalizedStringResource(
            "nameValue.summary.add",
            defaultValue: "Add name…",
            comment: "Host control that opens the NameValue editor when empty"
        )
        static let summaryEdit = LocalizedStringResource(
            "nameValue.summary.edit",
            defaultValue: "Edit name…",
            comment: "Host control that reopens the NameValue editor when set"
        )
        static let summaryNotNormalized = LocalizedStringResource(
            "nameValue.summary.notNormalized",
            defaultValue: "Not normalized yet",
            comment: "Host preview when no NameValue form has been saved"
        )
        static let partMoveUpLabel = LocalizedStringResource(
            "nameValue.part.moveUp",
            defaultValue: "Move up",
            comment: "Tooltip for moving a NameValue part up"
        )
        static let partMoveDownLabel = LocalizedStringResource(
            "nameValue.part.moveDown",
            defaultValue: "Move down",
            comment: "Tooltip for moving a NameValue part down"
        )
        static let partRemoveLabel = LocalizedStringResource(
            "nameValue.part.remove",
            defaultValue: "Remove part",
            comment: "Tooltip for removing a NameValue part"
        )

        static func partEmptyValue(position: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "nameValue.part.error.emptyValue",
                defaultValue: "Part %lld cannot be empty — remove it instead.",
                comment: "Validation when a NameValue part value is blank; argument is 1-based index"
            ))
            return String(format: format, locale: .current, position)
        }

        static func partAccessibility(position: Int, of count: Int, typeLabel: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "a11y.nameValue.part",
                defaultValue: "Part %1$lld of %2$lld, %3$@",
                comment: "VoiceOver name for a NameValue part row; arguments are index, count, type label"
            ))
            return String(format: format, locale: .current, position, count, typeLabel)
        }

        static func partMoveUp(position: Int) -> LocalizedStringResource {
            LocalizedStringResource(
                "a11y.nameValue.part.moveUp",
                defaultValue: "Move part \(position) up",
                comment: "VoiceOver for move-up; argument is 1-based part index"
            )
        }

        static func partMoveDown(position: Int) -> LocalizedStringResource {
            LocalizedStringResource(
                "a11y.nameValue.part.moveDown",
                defaultValue: "Move part \(position) down",
                comment: "VoiceOver for move-down; argument is 1-based part index"
            )
        }

        static func partRemove(position: Int) -> LocalizedStringResource {
            LocalizedStringResource(
                "a11y.nameValue.part.remove",
                defaultValue: "Remove part \(position)",
                comment: "VoiceOver for remove; argument is 1-based part index"
            )
        }

        static func partMoved(position: Int, of count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "a11y.nameValue.part.moved",
                defaultValue: "Part moved to position %1$lld of %2$lld",
                comment: "Announcement after Option-arrow reorder; arguments are new index and count"
            ))
            return String(format: format, locale: .current, position, count)
        }
    }

    /// Artifact document viewer (S7-06) — image/PDF now; audio/video kinds reserved.
    enum ArtifactViewer {
        static let previousPage = LocalizedStringResource(
            "artifactViewer.previousPage",
            defaultValue: "Previous page",
            comment: "Artifact viewer: go to previous PDF page"
        )

        static let nextPage = LocalizedStringResource(
            "artifactViewer.nextPage",
            defaultValue: "Next page",
            comment: "Artifact viewer: go to next PDF page"
        )

        static let pageField = LocalizedStringResource(
            "artifactViewer.pageField",
            defaultValue: "Page number",
            comment: "Accessibility label for the editable PDF page field"
        )

        static func pageOf(total: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "artifactViewer.pageOf",
                defaultValue: "of %lld",
                comment: "Artifact viewer page count suffix; argument is total pages"
            ))
            return String(format: format, locale: .current, total)
        }

        static let zoomOut = LocalizedStringResource(
            "artifactViewer.zoomOut",
            defaultValue: "Zoom out",
            comment: "Artifact viewer: decrease magnification"
        )

        static let zoomIn = LocalizedStringResource(
            "artifactViewer.zoomIn",
            defaultValue: "Zoom in",
            comment: "Artifact viewer: increase magnification"
        )

        static func zoomPercent(percent: Int) -> LocalizedStringResource {
            LocalizedStringResource(
                "artifactViewer.zoomPercent",
                defaultValue: "Zoom \(percent) percent",
                comment: "VoiceOver label for the live zoom percentage"
            )
        }

        static let canvasAccessibility = LocalizedStringResource(
            "artifactViewer.canvasAccessibility",
            defaultValue: "Artifact document",
            comment: "VoiceOver name for the pan/zoom document canvas"
        )

        static let emptyIdleTitle = LocalizedStringResource(
            "artifactViewer.emptyIdleTitle",
            defaultValue: "No Artifact selected",
            comment: "Empty canvas title before a document is loaded"
        )

        static let emptyIdleMessage = LocalizedStringResource(
            "artifactViewer.emptyIdleMessage",
            defaultValue: "Choose an Artifact to view it here.",
            comment: "Empty canvas body before a document is loaded"
        )

        static let unsupportedTitle = LocalizedStringResource(
            "artifactViewer.unsupportedTitle",
            defaultValue: "Can't preview this file",
            comment: "Title when mediaType is not a supported viewer kind"
        )

        static let unsupportedMessage = LocalizedStringResource(
            "artifactViewer.unsupportedMessage",
            defaultValue: "This Artifact type isn't supported in the viewer yet.",
            comment: "Body when mediaType is unsupported"
        )

        static let mediaComingSoonTitle = LocalizedStringResource(
            "artifactViewer.mediaComingSoonTitle",
            defaultValue: "Audio and video coming later",
            comment: "Title when kind is audio/video (players not shipped)"
        )

        static let mediaComingSoonMessage = LocalizedStringResource(
            "artifactViewer.mediaComingSoonMessage",
            defaultValue: "You can still cite this Artifact; in-app playback isn't available yet.",
            comment: "Body when kind is audio/video"
        )

        static let missingFileTitle = LocalizedStringResource(
            "artifactViewer.missingFileTitle",
            defaultValue: "File missing",
            comment: "Title when ProjectFiles cannot resolve or find the object"
        )

        static let missingFileMessage = LocalizedStringResource(
            "artifactViewer.missingFileMessage",
            defaultValue: "The stored file could not be found in this project.",
            comment: "Body when the object file is missing"
        )

        static let loadFailedTitle = LocalizedStringResource(
            "artifactViewer.loadFailedTitle",
            defaultValue: "Couldn't open file",
            comment: "Title when image/PDF decode fails"
        )

        static let loadFailedMessage = LocalizedStringResource(
            "artifactViewer.loadFailedMessage",
            defaultValue: "The file exists but couldn't be read as an image or PDF.",
            comment: "Body when decode fails"
        )

        static let setPage = LocalizedStringResource(
            "artifactViewer.setPage",
            defaultValue: "Set page",
            comment: "Commits the current viewer page into the citation locator"
        )

        static func pageSet(page: Int) -> LocalizedStringResource {
            LocalizedStringResource(
                "artifactViewer.pageSet",
                defaultValue: "Page \(page) set",
                comment: "Set page button after the locator page matches the viewer; argument is page number"
            )
        }

        static let regionRectangle = LocalizedStringResource(
            "artifactViewer.regionRectangle",
            defaultValue: "Rectangle",
            comment: "Region tool: axis-aligned rectangle"
        )

        static let regionLOpenTopRight = LocalizedStringResource(
            "artifactViewer.regionLOpenTopRight",
            defaultValue: "L open top-right",
            comment: "Region tool: L-shape with the missing quarter at top-right"
        )

        static let regionLOpenTopLeft = LocalizedStringResource(
            "artifactViewer.regionLOpenTopLeft",
            defaultValue: "L open top-left",
            comment: "Region tool: L-shape with the missing quarter at top-left"
        )

        static let regionLOpenBottomRight = LocalizedStringResource(
            "artifactViewer.regionLOpenBottomRight",
            defaultValue: "L open bottom-right",
            comment: "Region tool: L-shape with the missing quarter at bottom-right"
        )

        static let regionLOpenBottomLeft = LocalizedStringResource(
            "artifactViewer.regionLOpenBottomLeft",
            defaultValue: "L open bottom-left",
            comment: "Region tool: L-shape with the missing quarter at bottom-left"
        )

        static let regionCircle = LocalizedStringResource(
            "artifactViewer.regionCircle",
            defaultValue: "Circle",
            comment: "Region tool: circle stored as a 32-point ring"
        )

        static let regionFreeform = LocalizedStringResource(
            "artifactViewer.regionFreeform",
            defaultValue: "Freeform",
            comment: "Region tool: click-to-place polygon vertices"
        )
    }

    /// Citation composer place (S7-08 thin submit path; board-aligned shell).
    enum CitationComposer {
        static func breadcrumbCitationFor(scope: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.breadcrumb.citationFor",
                defaultValue: "Citation for %@",
                comment: "Composer breadcrumb leaf; argument is subject label or bridge edge sentence"
            ))
            return String(format: format, locale: .current, scope)
        }

        static let accessibilityTitle = LocalizedStringResource(
            "citationComposer.accessibilityTitle",
            defaultValue: "Citation composer",
            comment: "VoiceOver name for the citation composer workspace place"
        )

        static let save = LocalizedStringResource(
            "citationComposer.save",
            defaultValue: "Save citation",
            comment: "Citation fields button that writes only the Citation row"
        )

        static let done = LocalizedStringResource(
            "citationComposer.done",
            defaultValue: "Done",
            comment: "Footer button that returns to the Evidence graph through the leave guard"
        )

        static let citationStatusNew = LocalizedStringResource(
            "citationComposer.citationStatusNew",
            defaultValue: "First save creates the citation",
            comment: "Status under citation fields when the Citation does not exist yet"
        )

        static let citationStatusDirty = LocalizedStringResource(
            "citationComposer.citationStatusDirty",
            defaultValue: "Citation has unsaved changes",
            comment: "Status under citation fields when the Citation row is dirty"
        )

        static let citationStatusSaved = LocalizedStringResource(
            "citationComposer.citationStatusSaved",
            defaultValue: "Citation saved",
            comment: "Status under citation fields when the Citation row matches the catalog"
        )

        static let identityMenusDisabledHint = LocalizedStringResource(
            "citationComposer.identityMenusDisabledHint",
            defaultValue: "Save or discard changes to switch citation",
            comment: "Hint under Artifact / Citation menus when unsaved document work disables them"
        )

        static let rowStateNew = LocalizedStringResource(
            "citationComposer.rowStateNew",
            defaultValue: "New",
            comment: "Badge on a draft Observation or pending connection row"
        )

        static let rowStateEdited = LocalizedStringResource(
            "citationComposer.rowStateEdited",
            defaultValue: "Edited",
            comment: "Badge on an Observation or connection role that differs from its baseline"
        )

        static let rowStateSaving = LocalizedStringResource(
            "citationComposer.rowStateSaving",
            defaultValue: "Saving",
            comment: "Badge while a row-level Save is in flight"
        )

        static let rowStateError = LocalizedStringResource(
            "citationComposer.rowStateError",
            defaultValue: "Not saved",
            comment: "Badge when a row-level Save failed"
        )

        static let rowStateSaved = LocalizedStringResource(
            "citationComposer.rowStateSaved",
            defaultValue: "Saved",
            comment: "Spoken status for a saved connection without a ref yet"
        )

        static let negateObservation = LocalizedStringResource(
            "citationComposer.negateObservation",
            defaultValue: "Negate observation",
            comment: "Row menu item that flips a positive Observation to negative"
        )

        static let affirmObservation = LocalizedStringResource(
            "citationComposer.affirmObservation",
            defaultValue: "Affirm observation",
            comment: "Row menu item that flips a negative Observation to positive"
        )

        static let applyValue = LocalizedStringResource(
            "citationComposer.applyValue",
            defaultValue: "Apply",
            comment: "Confirm on the name / date dialog; writes into the row only"
        )

        static let valueDialogSubtitle = LocalizedStringResource(
            "citationComposer.valueDialogSubtitle",
            defaultValue: "Updates the row only. The row's Save commits it to CIT-…",
            comment: "Subtitle on the name / date dialog"
        )

        static let connectionRelationship = LocalizedStringResource(
            "citationComposer.connectionRelationship",
            defaultValue: "Relationship",
            comment: "Term picker label on a relationship connection row"
        )

        static let connectionRole = LocalizedStringResource(
            "citationComposer.connectionRole",
            defaultValue: "Role",
            comment: "Term picker label on a participation connection row"
        )

        static let connectionNoRole = LocalizedStringResource(
            "citationComposer.connectionNoRole",
            defaultValue: "No role or type",
            comment: "Muted slot on a location connection that has no term"
        )

        static let connectionTermNotChosen = LocalizedStringResource(
            "citationComposer.connectionTermNotChosen",
            defaultValue: "not chosen",
            comment: "VoiceOver fallback when a connection role or type is empty"
        )

        static let saveRow = LocalizedStringResource(
            "citationComposer.saveRow",
            defaultValue: "Save",
            comment: "Row-level Save on an Observation"
        )

        static let revertRow = LocalizedStringResource(
            "citationComposer.revertRow",
            defaultValue: "Revert",
            comment: "Row-level Revert on an Observation"
        )

        static let saveConnection = LocalizedStringResource(
            "citationComposer.saveConnection",
            defaultValue: "Save connection",
            comment: "Writes a pending Connect row as a cited bridge"
        )

        static let discardConnection = LocalizedStringResource(
            "citationComposer.discardConnection",
            defaultValue: "Discard",
            comment: "Removes a pending connection without writing"
        )

        static let saveRole = LocalizedStringResource(
            "citationComposer.saveRole",
            defaultValue: "Save role",
            comment: "Commits an edited role or type on a saved connection"
        )

        static let revertRole = LocalizedStringResource(
            "citationComposer.revertRole",
            defaultValue: "Revert",
            comment: "Restores a saved connection role to its baseline"
        )

        static let subjectRequiredError = LocalizedStringResource(
            "citationComposer.subjectRequiredError",
            defaultValue: "Choose a subject.",
            comment: "Row error when Save is pressed with no subject"
        )

        static let propertyRequiredAfterSubject = LocalizedStringResource(
            "citationComposer.propertyRequiredAfterSubject",
            defaultValue: "This subject does not allow that property.",
            comment: "Property field error after an incompatible subject change"
        )

        static let newSubjectTitle = LocalizedStringResource(
            "citationComposer.newSubjectTitle",
            defaultValue: "New subject",
            comment: "Dialog title when creating a subject from a row picker"
        )

        static let newSubjectConfirm = LocalizedStringResource(
            "citationComposer.newSubjectConfirm",
            defaultValue: "Create",
            comment: "Confirm creating a subject from the composer"
        )

        static let newSubjectLabel = LocalizedStringResource(
            "citationComposer.newSubjectLabel",
            defaultValue: "Label",
            comment: "Label field on the composer new-subject dialog"
        )

        static let deleteObservationTitle = LocalizedStringResource(
            "citationComposer.deleteObservationTitle",
            defaultValue: "Delete observation?",
            comment: "Confirm title before deleting one Observation"
        )

        static let deleteObservationMessage = LocalizedStringResource(
            "citationComposer.deleteObservationMessage",
            defaultValue: "This observation will be removed from the citation. The citation stays.",
            comment: "Confirm body before deleting one Observation"
        )

        static let deleteObservationConfirm = LocalizedStringResource(
            "citationComposer.deleteObservationConfirm",
            defaultValue: "Delete",
            comment: "Confirm button that deletes one Observation"
        )

        static let leaveTitle = LocalizedStringResource(
            "citationComposer.leaveTitle",
            defaultValue: "Leave without saving?",
            comment: "Leave-guard confirm title"
        )

        static let leaveDiscard = LocalizedStringResource(
            "citationComposer.leaveDiscard",
            defaultValue: "Leave",
            comment: "Leave-guard confirm that discards unsaved composer work"
        )

        static let leaveKeepEditing = LocalizedStringResource(
            "citationComposer.leaveKeepEditing",
            defaultValue: "Keep editing",
            comment: "Leave-guard cancel that stays on the composer"
        )

        static func newSubject(typeKey: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.newSubjectOption",
                defaultValue: "New %@…",
                comment: "Subject picker option; argument is person, event, or place"
            ))
            return String(format: format, locale: .current, typeKey)
        }

        static func unsavedSummary(
            citationDirty: Bool,
            observationCount: Int,
            connectionTouched: Bool
        ) -> String {
            var parts: [String] = []
            if citationDirty {
                parts.append(String(localized: LocalizedStringResource(
                    "citationComposer.unsavedCitationPart",
                    defaultValue: "the citation",
                    comment: "Unsaved-summary clause for dirty citation fields"
                )))
            }
            if observationCount == 1 {
                parts.append(String(localized: LocalizedStringResource(
                    "citationComposer.unsavedObservationOne",
                    defaultValue: "1 observation",
                    comment: "Unsaved-summary clause for one dirty observation row"
                )))
            } else if observationCount > 1 {
                let format = String(localized: LocalizedStringResource(
                    "citationComposer.unsavedObservationMany",
                    defaultValue: "%lld observations",
                    comment: "Unsaved-summary clause for several dirty observation rows"
                ))
                parts.append(String(format: format, locale: .current, observationCount))
            }
            if connectionTouched {
                parts.append(String(localized: LocalizedStringResource(
                    "citationComposer.unsavedConnectionPart",
                    defaultValue: "a new connection",
                    comment: "Unsaved-summary clause for a touched pending connection"
                )))
            }
            let joined = parts.joined(separator: ", ")
            let format = String(localized: LocalizedStringResource(
                "citationComposer.unsavedSummary",
                defaultValue: "Unsaved: %@",
                comment: "Footer and leave-guard body; argument is the joined unsaved parts"
            ))
            return String(format: format, locale: .current, joined)
        }

        static func connectionAccessibility(
            sentence: String,
            termLabel: String,
            term: String,
            status: String
        ) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.connectionAccessibility",
                defaultValue: "Connection, %@. %@ %@. %@",
                comment: "VoiceOver for a relationship or participation connection"
            ))
            return String(format: format, locale: .current, sentence, termLabel, term, status)
        }

        static func connectionAccessibilityLocation(status: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.connectionAccessibilityLocation",
                defaultValue: "Connection, Location. %@",
                comment: "VoiceOver for a location connection; argument is New/Edited/Saved ref"
            ))
            return String(format: format, locale: .current, status)
        }

        static func connectionSavedStatus(ref: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.connectionSavedStatus",
                defaultValue: "Saved, %@",
                comment: "VoiceOver status for a saved connection; argument is the bridge ref"
            ))
            return String(format: format, locale: .current, ref)
        }

        static let cancel = LocalizedStringResource(
            "citationComposer.cancel",
            defaultValue: "Cancel",
            comment: "Cancel button that returns to the Evidence graph without writing"
        )

        static let transcriptionLabel = LocalizedStringResource(
            "citationComposer.transcriptionLabel",
            defaultValue: "Transcription",
            comment: "Label for the Citation transcription field"
        )

        static let uncertainLabel = LocalizedStringResource(
            "citationComposer.uncertainLabel",
            defaultValue: "Uncertain reading",
            comment: "Checkbox when the researcher is unsure of the transcription"
        )

        static let uncertainNoteLabel = LocalizedStringResource(
            "citationComposer.uncertainNoteLabel",
            defaultValue: "Why the transcription is uncertain",
            comment: "Note field shown when transcription uncertain is checked"
        )

        static let descriptionLabel = LocalizedStringResource(
            "citationComposer.descriptionLabel",
            defaultValue: "Description",
            comment: "Optional Citation description field"
        )

        static let observationsSection = LocalizedStringResource(
            "citationComposer.observationsSection",
            defaultValue: "Observations",
            comment: "Section header for the Observations list"
        )

        static let addObservation = LocalizedStringResource(
            "citationComposer.addObservation",
            defaultValue: "Add observation",
            comment: "Button / dialog confirm that commits an Observation"
        )

        static let editObservation = LocalizedStringResource(
            "citationComposer.editObservation",
            defaultValue: "Edit observation",
            comment: "Icon button that reopens the observation dialog for a row"
        )

        static let connectSystemBadge = LocalizedStringResource(
            "citationComposer.connectSystemBadge",
            defaultValue: "System · Connect",
            comment: "Badge on fixed connect-edge Observation rows prefilled from Connect (Frame 10)"
        )

        static let removeObservation = LocalizedStringResource(
            "citationComposer.removeObservation",
            defaultValue: "Remove observation",
            comment: "Icon button that deletes one Observation row"
        )

        static let propertyLabel = LocalizedStringResource(
            "citationComposer.propertyLabel",
            defaultValue: "Property",
            comment: "ComboBox label for choosing which Property an Observation asserts"
        )

        static let propertyPlaceholder = LocalizedStringResource(
            "citationComposer.propertyPlaceholder",
            defaultValue: "Search properties",
            comment: "Placeholder in the Property picker ComboBox"
        )

        static let propertyEmpty = LocalizedStringResource(
            "citationComposer.propertyEmpty",
            defaultValue: "No matching properties",
            comment: "Empty state when Property search has no hits"
        )

        static let polarityLabel = LocalizedStringResource(
            "citationComposer.polarityLabel",
            defaultValue: "Polarity",
            comment: "Label above Asserts / Negates chips"
        )

        static let polarityAsserts = LocalizedStringResource(
            "citationComposer.polarityAsserts",
            defaultValue: "Asserts",
            comment: "Positive Observation polarity chip"
        )

        static let polarityNegates = LocalizedStringResource(
            "citationComposer.polarityNegates",
            defaultValue: "Negates",
            comment: "Negative Observation polarity chip"
        )

        static let valueLabel = LocalizedStringResource(
            "citationComposer.valueLabel",
            defaultValue: "Value",
            comment: "Generic value label before a Property type is chosen"
        )

        static let valueLabelText = LocalizedStringResource(
            "citationComposer.valueLabelText",
            defaultValue: "Value — text",
            comment: "Value slot label for a text Property"
        )

        static let valueLabelTerm = LocalizedStringResource(
            "citationComposer.valueLabelTerm",
            defaultValue: "Value — term",
            comment: "Value slot label for a term Property"
        )

        static let valueLabelDate = LocalizedStringResource(
            "citationComposer.valueLabelDate",
            defaultValue: "Value — date",
            comment: "Value slot label for a date Property"
        )

        static let valueLabelInteger = LocalizedStringResource(
            "citationComposer.valueLabelInteger",
            defaultValue: "Value — whole number",
            comment: "Value slot label for an integer Property"
        )

        static let valueLabelName = LocalizedStringResource(
            "citationComposer.valueLabelName",
            defaultValue: "Value — name",
            comment: "Value slot label for a name Property"
        )

        static let termLabel = LocalizedStringResource(
            "citationComposer.termLabel",
            defaultValue: "Term",
            comment: "ComboBox label for a term Observation value"
        )

        static let termPlaceholder = LocalizedStringResource(
            "citationComposer.termPlaceholder",
            defaultValue: "Search terms",
            comment: "Placeholder in the term picker"
        )

        static let termEmpty = LocalizedStringResource(
            "citationComposer.termEmpty",
            defaultValue: "No matching terms",
            comment: "Empty state when term search has no hits"
        )

        static let addCustomTerm = LocalizedStringResource(
            "citationComposer.addCustomTerm",
            defaultValue: "Add custom…",
            comment: "Opens dialog to mint a user Property term"
        )

        static let addTermTitle = LocalizedStringResource(
            "citationComposer.addTermTitle",
            defaultValue: "Add custom term",
            comment: "Dialog title for creating a user Property term"
        )

        static let addTermConfirm = LocalizedStringResource(
            "citationComposer.addTermConfirm",
            defaultValue: "Add",
            comment: "Confirm creating a user Property term"
        )

        static let addTermLabel = LocalizedStringResource(
            "citationComposer.addTermLabel",
            defaultValue: "Label",
            comment: "Label field when minting a custom Property term"
        )

        static let dateUnset = LocalizedStringResource(
            "citationComposer.dateUnset",
            defaultValue: "No date set",
            comment: "Summary when a date Observation has no valid DateValue yet"
        )

        static func artifactIndexOf(index: Int, total: Int, kind: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.artifactIndexOf",
                defaultValue: "Artifact %lld of %lld · %@",
                comment: "Viewer header index; arguments are 1-based index, total, and media kind"
            ))
            return String(format: format, locale: .current, index, total, kind)
        }

        static let artifactKindPDF = LocalizedStringResource(
            "citationComposer.artifactKindPDF",
            defaultValue: "PDF",
            comment: "Media kind label for a PDF Artifact"
        )

        static let artifactKindImage = LocalizedStringResource(
            "citationComposer.artifactKindImage",
            defaultValue: "Image",
            comment: "Media kind label for an image Artifact"
        )

        static let artifactKindUnknown = LocalizedStringResource(
            "citationComposer.artifactKindUnknown",
            defaultValue: "File",
            comment: "Media kind fallback when type is neither image nor PDF"
        )

        static let mediaCaptionPDF = LocalizedStringResource(
            "citationComposer.mediaCaptionPDF",
            defaultValue: "PDF",
            comment: "Caption under an Artifact tile for PDF media"
        )

        static let mediaCaptionImage = LocalizedStringResource(
            "citationComposer.mediaCaptionImage",
            defaultValue: "Image · 1 page",
            comment: "Caption under an Artifact tile for image media"
        )

        static let mediaCaptionNoFile = LocalizedStringResource(
            "citationComposer.mediaCaptionNoFile",
            defaultValue: "No file",
            comment: "Caption under a fileless Artifact tile in the picker"
        )

        static func mediaCaptionNoFileDetail(detail: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.mediaCaptionNoFileDetail",
                defaultValue: "No file · %@",
                comment: "Caption under a fileless Artifact tile; argument is the Artifact description"
            ))
            return String(format: format, locale: .current, detail)
        }

        static func locatorPage(_ page: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.locatorPage",
                defaultValue: "Page %lld",
                comment: "Locator list row for a committed page; argument is page number"
            ))
            return String(format: format, locale: .current, page)
        }

        static let clearMenu = LocalizedStringResource(
            "citationComposer.clearMenu",
            defaultValue: "Clear",
            comment: "Opens the Clear region / Reset to entire artifact menu"
        )

        static let clearRegion = LocalizedStringResource(
            "citationComposer.clearRegion",
            defaultValue: "Clear region",
            comment: "Menu item that drops the region locator layer only"
        )

        static let resetToEntireArtifact = LocalizedStringResource(
            "citationComposer.resetToEntireArtifact",
            defaultValue: "Reset to entire artifact",
            comment: "Menu item that drops page and region, leaving the artifact floor"
        )

        static let locatorSection = LocalizedStringResource(
            "citationComposer.locatorSection",
            defaultValue: "Locator",
            comment: "Header above the layered locator summary list"
        )

        static let locatorOuterToInner = LocalizedStringResource(
            "citationComposer.locatorOuterToInner",
            defaultValue: "Outer to inner",
            comment: "Caption explaining locator list order"
        )

        static let locatorEntireArtifact = LocalizedStringResource(
            "citationComposer.locatorEntireArtifact",
            defaultValue: "Entire artifact",
            comment: "Floor locator row — always present, not removable"
        )

        static let locatorAlwaysIncluded = LocalizedStringResource(
            "citationComposer.locatorAlwaysIncluded",
            defaultValue: "Always included",
            comment: "Trailing lock label on the Entire artifact locator row"
        )

        static func locatorPoints(_ count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.locatorPoints",
                defaultValue: "%lld points",
                comment: "Locator list helper for a region; argument is vertex count"
            ))
            return String(format: format, locale: .current, count)
        }

        static let locatorRectangle = LocalizedStringResource(
            "citationComposer.locatorRectangle",
            defaultValue: "Rectangle",
            comment: "Locator list noun for a rectangle region"
        )

        static let locatorLShape = LocalizedStringResource(
            "citationComposer.locatorLShape",
            defaultValue: "L-shape",
            comment: "Locator list noun for an L-shaped region"
        )

        static let locatorCircle = LocalizedStringResource(
            "citationComposer.locatorCircle",
            defaultValue: "Circle",
            comment: "Locator list noun for a circle region"
        )

        static let locatorPolygon = LocalizedStringResource(
            "citationComposer.locatorPolygon",
            defaultValue: "Polygon",
            comment: "Locator list noun for a freeform region"
        )

        static func removePage(_ page: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.removePage",
                defaultValue: "Remove page %lld",
                comment: "Accessibility name to remove a page locator; argument is page number"
            ))
            return String(format: format, locale: .current, page)
        }

        static let removeRegion = LocalizedStringResource(
            "citationComposer.removeRegion",
            defaultValue: "Remove region",
            comment: "Accessibility name to remove the region locator layer"
        )

        static let removeLocatorLayer = LocalizedStringResource(
            "citationComposer.removeLocatorLayer",
            defaultValue: "Remove",
            comment: "Tooltip on the locator list remove control"
        )

        static let freeformDeleteVertex = LocalizedStringResource(
            "citationComposer.freeformDeleteVertex",
            defaultValue: "Right-click a vertex to delete it. At least three points remain.",
            comment: "Tooltip while a freeform region is committed"
        )

        static let noArtifactsCallout = LocalizedStringResource(
            "citationComposer.noArtifactsCallout",
            defaultValue: "Every citation needs an artifact to point at. Attach one from the Source page, then return here.",
            comment: "Left-pane callout when the Source has no Artifacts"
        )

        static let goToSourcePage = LocalizedStringResource(
            "citationComposer.goToSourcePage",
            defaultValue: "Go to Source page",
            comment: "Recovery action from the no-Artifact cite gate"
        )

        static let loadFailedBack = LocalizedStringResource(
            "citationComposer.loadFailedBack",
            defaultValue: "Back to evidence graph",
            comment: "Leaves the citation composer after catalog data failed to load"
        )

        static let fieldsDisabledHint = LocalizedStringResource(
            "citationComposer.fieldsDisabledHint",
            defaultValue: "Fields disabled until an artifact is attached",
            comment: "Hint on the inert form pane when no Artifact exists"
        )

        static let needArtifact = LocalizedStringResource(
            "citationComposer.needArtifact",
            defaultValue: "Choose an Artifact before saving.",
            comment: "Validation when Save is pressed without an Artifact"
        )

        static let dialogPropertyRequired = LocalizedStringResource(
            "citationComposer.dialogPropertyRequired",
            defaultValue: "Pick the Property this observation is about.",
            comment: "Inline error under Property in the Add observation dialog"
        )

        static let dialogValueRequired = LocalizedStringResource(
            "citationComposer.dialogValueRequired",
            defaultValue: "A value is required.",
            comment: "Inline error under Value in the Add observation dialog"
        )

        static let dialogValuePickPropertyFirst = LocalizedStringResource(
            "citationComposer.dialogValuePickPropertyFirst",
            defaultValue: "Pick a Property first — the editor follows its type",
            comment: "Placeholder in the value slot before a Property is chosen"
        )

        static let missingPropertyError = LocalizedStringResource(
            "citationComposer.missingPropertyError",
            defaultValue: "Each observation needs a Property.",
            comment: "Validation when a committed Observation has no Property"
        )

        static let invalidIntegerError = LocalizedStringResource(
            "citationComposer.invalidIntegerError",
            defaultValue: "Enter a whole number for the integer value.",
            comment: "Validation when integer Observation text is not Int64"
        )

        static let invalidDateError = LocalizedStringResource(
            "citationComposer.invalidDateError",
            defaultValue: "Enter a valid date for each date observation.",
            comment: "Validation when a date Observation DateValue draft is incomplete"
        )

        static let unsupportedValueTypeError = LocalizedStringResource(
            "citationComposer.unsupportedValueTypeError",
            defaultValue: "This Property type is not editable here yet.",
            comment: "Shown for name/subject rows deferred past the thin composer"
        )

        static let newCitation = LocalizedStringResource(
            "citationComposer.newCitation",
            defaultValue: "New citation",
            comment: "Citation identity control when composing a new reading"
        )

        static let subjectLabel = LocalizedStringResource(
            "citationComposer.subjectLabel",
            defaultValue: "Subject",
            comment: "ComboBox label for the Observation subject"
        )

        static let subjectPlaceholder = LocalizedStringResource(
            "citationComposer.subjectPlaceholder",
            defaultValue: "Search subjects",
            comment: "Placeholder in the Observation subject picker"
        )

        static let subjectEmpty = LocalizedStringResource(
            "citationComposer.subjectEmpty",
            defaultValue: "No matching subjects",
            comment: "Empty state when subject search has no hits"
        )

        static let observationActions = LocalizedStringResource(
            "citationComposer.observationActions",
            defaultValue: "Observation actions",
            comment: "VoiceOver name for the Observation row overflow menu"
        )

        static let editValueTitle = LocalizedStringResource(
            "citationComposer.editValueTitle",
            defaultValue: "Edit value",
            comment: "Title of the name or date Observation editor dialog"
        )

        static let editValuePlaceholder = LocalizedStringResource(
            "citationComposer.editValuePlaceholder",
            defaultValue: "Set value",
            comment: "Placeholder on the name/date button before a value is set"
        )

        static let viewerGroup = LocalizedStringResource(
            "citationComposer.viewerGroup",
            defaultValue: "Artifact viewer",
            comment: "VoiceOver group name for the composer viewer pane"
        )

        static let formGroup = LocalizedStringResource(
            "citationComposer.formGroup",
            defaultValue: "Citation form",
            comment: "VoiceOver group name for the composer form pane"
        )

        static let citationMenuNew = LocalizedStringResource(
            "citationComposer.citationMenuNew",
            defaultValue: "Citation, new",
            comment: "VoiceOver name for the Citation identity control when New is selected"
        )

        static let noArtifactsTitle = LocalizedStringResource(
            "citationComposer.noArtifactsTitle",
            defaultValue: "This source has no artifacts",
            comment: "Callout title when the Source has no Artifacts to cite"
        )

        static let identityChangedNew = LocalizedStringResource(
            "citationComposer.identityChangedNew",
            defaultValue: "Now composing a new citation",
            comment: "VoiceOver announcement when Citation identity resets to New"
        )

        static func artifactMenuLabel(title: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.artifactMenuLabel",
                defaultValue: "Artifact, %@",
                comment: "VoiceOver name for the Artifact identity control; argument is Artifact title"
            ))
            return String(format: format, locale: .current, title)
        }

        static func artifactMenuMeta(kind: String, count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.artifactMenuMeta",
                defaultValue: "%@ · %lld citations",
                comment: "Artifact menu row meta; arguments are media kind and citation count"
            ))
            return String(format: format, locale: .current, kind, count)
        }

        static func citationMenuCount(count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.citationMenuCount",
                defaultValue: "%lld obs",
                comment: "Observation count on a Citation menu row; argument is count"
            ))
            return String(format: format, locale: .current, count)
        }

        static func citationMenuRef(ref: String, count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.citationMenuRef",
                defaultValue: "Citation, %@, %lld observations",
                comment: "VoiceOver name for the Citation identity control; arguments are ref and count"
            ))
            return String(format: format, locale: .current, ref, count)
        }

        static func identityChangedCitation(ref: String, count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.identityChangedCitation",
                defaultValue: "Now citing %@, %lld observations",
                comment: "VoiceOver announcement after switching Citation; arguments are ref and count"
            ))
            return String(format: format, locale: .current, ref, count)
        }

        static func identityChangedArtifact(title: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "citationComposer.identityChangedArtifact",
                defaultValue: "Now citing %@",
                comment: "VoiceOver announcement after switching Artifact; argument is Artifact title"
            ))
            return String(format: format, locale: .current, title)
        }

    }

    /// Origin markers shared by every catalog vocabulary destination —
    /// `OriginBadge` on a detail panel, `OriginPill` inline in a list.
    enum Origin {
        static let provenencia = LocalizedStringResource(
            "origin.badge.provenencia",
            defaultValue: "provenencia",
            comment: "Origin badge for a vocabulary row seeded by Provenencia"
        )

        static let user = LocalizedStringResource(
            "origin.badge.user",
            defaultValue: "you",
            comment: "Origin badge for a vocabulary row the researcher added"
        )

        static let seededPill = LocalizedStringResource(
            "origin.pill.seeded",
            defaultValue: "Seeded by Provenencia",
            comment: "Accessibility label and tooltip for the pill marking a Provenencia-seeded row in a vocabulary list"
        )

        static func pluginPill(pluginID: String) -> LocalizedStringResource {
            LocalizedStringResource(
                "origin.pill.plugin",
                defaultValue: "Supplied by the \(pluginID) plugin",
                comment: "Accessibility label and tooltip for the pill marking a plugin-owned row in a vocabulary list; argument is the plugin id"
            )
        }
    }

    /// The **Sources** workspace destination (S2-17–18): browse Sources as an
    /// evidence list, create via a thin dialog, open a separate Source page.
    enum Sources {
        static let description = LocalizedStringResource(
            "sources.list.description",
            defaultValue: "Every record behind this project. Open a source to read its artifacts and citation.",
            comment: "Explanatory copy under the Sources page title"
        )

        static func countLine(total: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "sources.list.countLine",
                defaultValue: "%lld sources",
                comment: "Sources list count when unfiltered; argument is total"
            ))
            return String(format: format, locale: .current, total)
        }

        static func countLineFiltered(visible: Int, total: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "sources.list.countLineFiltered",
                defaultValue: "%lld of %lld sources",
                comment: "Sources list count when search/filter narrows the list; arguments are visible then total"
            ))
            return String(format: format, locale: .current, visible, total)
        }

        static let addSource = LocalizedStringResource(
            "sources.list.addSource",
            defaultValue: "Add source",
            comment: "Button: open the Add Source dialog"
        )

        static let noFilterMatchesTitle = LocalizedStringResource(
            "sources.list.noFilterMatchesTitle",
            defaultValue: "No sources for this type",
            comment: "Empty-state title when the Sources type filter matches nothing"
        )

        static let noFilterMatchesMessage = LocalizedStringResource(
            "sources.list.noFilterMatchesMessage",
            defaultValue: "Try another type, or choose All types.",
            comment: "Empty-state body when the Sources type filter matches nothing"
        )

        static let filterAllTypes = LocalizedStringResource(
            "sources.list.filterAllTypes",
            defaultValue: "All types",
            comment: "Sources type filter menu: show every type"
        )

        static let filterMenu = LocalizedStringResource(
            "sources.list.filterMenu",
            defaultValue: "Filter by type",
            comment: "Accessibility label for the Sources type filter control"
        )

        static let sortMenu = LocalizedStringResource(
            "sources.list.sortMenu",
            defaultValue: "Sort sources",
            comment: "Accessibility label for the Sources sort control"
        )

        static let sortAdded = LocalizedStringResource(
            "sources.list.sort.added",
            defaultValue: "Date added",
            comment: "Sources sort option: catalog list order (date added)"
        )

        static let sortUpdated = LocalizedStringResource(
            "sources.list.sort.updated",
            defaultValue: "Date updated",
            comment: "Sources sort option: reverse catalog order (stand-in until audit timestamps surface)"
        )

        static let sortAZ = LocalizedStringResource(
            "sources.list.sort.az",
            defaultValue: "Alphabetical",
            comment: "Sources sort option: title A to Z"
        )

        static let sortZA = LocalizedStringResource(
            "sources.list.sort.za",
            defaultValue: "Reverse alphabetical",
            comment: "Sources sort option: title Z to A"
        )

        static func sortedBy(_ label: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sources.list.sortedBy",
                defaultValue: "Sorted by %@",
                comment: "Sources sort control label; argument is the active sort option in lowercase"
            ))
            return String(format: format, locale: .current, label)
        }

        static let emptyTitle = LocalizedStringResource(
            "sources.list.emptyTitle",
            defaultValue: "No sources yet",
            comment: "Empty-state title when the project has no Sources"
        )

        static let emptyMessage = LocalizedStringResource(
            "sources.list.emptyMessage",
            defaultValue: "Every fact should trace back to a record. Add the first source, then attach the scans and files it came from.",
            comment: "Empty-state body when the project has no Sources"
        )

        static let listAccessibilityLabel = LocalizedStringResource(
            "sources.list.accessibilityLabel",
            defaultValue: "Sources",
            comment: "Accessibility label for the Sources list"
        )

        static let columnSource = LocalizedStringResource(
            "sources.list.columnSource",
            defaultValue: "Source",
            comment: "Caption band above the Sources list left zone"
        )

        static let columnEvidenceGraph = LocalizedStringResource(
            "sources.list.columnEvidenceGraph",
            defaultValue: "Evidence graph",
            comment: "Caption band above the Sources list graph zone"
        )

        static let openGraph = LocalizedStringResource(
            "sources.list.openGraph",
            defaultValue: "Open graph",
            comment: "Sources list right-zone action when the Source has an Artifact"
        )

        static let needsArtifact = LocalizedStringResource(
            "sources.list.needsArtifact",
            defaultValue: "Needs an artifact",
            comment: "Blocked graph zone label when the Source has no Artifact"
        )

        static let needsArtifactTooltip = LocalizedStringResource(
            "sources.list.needsArtifactTooltip",
            defaultValue: "Add an artifact on the source page to build its evidence graph",
            comment: "Tooltip on the blocked Sources list graph zone"
        )

        static let openSourcePage = LocalizedStringResource(
            "sources.list.openSourcePage",
            defaultValue: "Open source",
            comment: "Accessibility label for the Sources list left (filing) zone"
        )

        static let addDialogTitle = LocalizedStringResource(
            "sources.add.title",
            defaultValue: "Add source",
            comment: "Add Source dialog title"
        )

        static let addDialogSubtitle = LocalizedStringResource(
            "sources.add.subtitle",
            defaultValue: "A thin record now — artifacts, notes and metadata live on the Source page.",
            comment: "Add Source dialog subtitle"
        )

        static let formType = LocalizedStringResource(
            "sources.add.formType",
            defaultValue: "Type",
            comment: "Add Source form label: source type"
        )

        static let typePlaceholder = LocalizedStringResource(
            "sources.add.typePlaceholder",
            defaultValue: "Search source types",
            comment: "Placeholder in the Add Source type combo box before a type is chosen"
        )

        static let typeNoMatch = LocalizedStringResource(
            "sources.add.typeNoMatch",
            defaultValue: "No type matches that search",
            comment: "Empty state in the Add Source type combo box when the query matches nothing"
        )

        static let typePoolEmpty = LocalizedStringResource(
            "sources.add.typePoolEmpty",
            defaultValue: "No source types in this project yet — add one under Source types.",
            comment: "Empty state in the Add Source type combo when the project has no types loaded"
        )

        static let formTitle = LocalizedStringResource(
            "sources.add.formTitle",
            defaultValue: "Title",
            comment: "Add Source form label: title"
        )

        static let formDescription = LocalizedStringResource(
            "sources.add.formDescription",
            defaultValue: "Description",
            comment: "Add Source form label: optional description"
        )

        static let formDescriptionHint = LocalizedStringResource(
            "sources.add.formDescriptionHint",
            defaultValue: "Optional — a short note on what this record is",
            comment: "Hint under the optional description field on Add Source"
        )

        static let createAction = LocalizedStringResource(
            "sources.add.create",
            defaultValue: "Create source",
            comment: "Primary button on the Add Source dialog"
        )

        static let cancelAction = LocalizedStringResource(
            "sources.add.cancel",
            defaultValue: "Cancel",
            comment: "Cancel button on the Add Source dialog"
        )

        static let typeRequired = LocalizedStringResource(
            "sources.add.typeRequired",
            defaultValue: "Choose a source type — it decides which fields the Source page shows.",
            comment: "Inline validation when Add Source is submitted without a type"
        )

        static let titleRequired = LocalizedStringResource(
            "sources.add.titleRequired",
            defaultValue: "Give the source a title — name the record as it identifies itself.",
            comment: "Inline validation when Add Source is submitted without a title"
        )

        static func toastCreatedTitle(ref: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sources.add.toastCreatedTitle",
                defaultValue: "%@",
                comment: "Toast title after creating a Source; argument is the SRC- ref"
            ))
            return String(format: format, locale: .current, ref)
        }

        static func toastCreatedBody(title: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sources.add.toastCreatedBody",
                defaultValue: "%@. Opening its Source page.",
                comment: "Toast body after creating a Source; argument is the title"
            ))
            return String(format: format, locale: .current, title)
        }

        // MARK: Source page (S2-18)

        static let pageTitleRequired = LocalizedStringResource(
            "sources.page.titleRequired",
            defaultValue: "A source needs a title.",
            comment: "Inline validation when Source page title is cleared"
        )

        static let descriptionPlaceholder = LocalizedStringResource(
            "sources.page.descriptionPlaceholder",
            defaultValue: "What this source is, and where you consulted it",
            comment: "Placeholder for the Source description field"
        )

        static let descriptionEmptyTitle = LocalizedStringResource(
            "sources.page.descriptionEmptyTitle",
            defaultValue: "No description yet",
            comment: "Empty-state title when a Source has no description"
        )

        static let descriptionEmptyMessage = LocalizedStringResource(
            "sources.page.descriptionEmptyMessage",
            defaultValue: "Say what this source is, and where you consulted it.",
            comment: "Empty-state body when a Source has no description"
        )

        static let descriptionHeading = LocalizedStringResource(
            "sources.page.descriptionHeading",
            defaultValue: "Description",
            comment: "Section heading above Source description"
        )

        static let credibilityHeading = LocalizedStringResource(
            "sources.page.credibilityHeading",
            defaultValue: "Credibility",
            comment: "Heading for the Source credibility control"
        )

        static let credibilityArgumentPlaceholder = LocalizedStringResource(
            "sources.page.credibilityArgumentPlaceholder",
            defaultValue: "Why this grade — optional",
            comment: "Placeholder for optional credibility argument"
        )

        static let credibilityHint = LocalizedStringResource(
            "sources.page.credibilityHint",
            defaultValue: "How far you trust this source as evidence.",
            comment: "Italic aside when a credibility assessment is saved"
        )

        static let credibilityHintUnset = LocalizedStringResource(
            "sources.page.credibilityHintUnset",
            defaultValue: "Not assessed — treated as standard",
            comment: "Italic aside when no credibility assessment row exists"
        )

        static let saveAssessment = LocalizedStringResource(
            "sources.page.saveAssessment",
            defaultValue: "Save assessment",
            comment: "Button to save credibility grade and argument"
        )

        static let cancelEdit = LocalizedStringResource(
            "sources.page.cancelEdit",
            defaultValue: "Cancel",
            comment: "Cancel an in-progress edit on the Source page"
        )

        static let saveAction = LocalizedStringResource(
            "sources.page.saveAction",
            defaultValue: "Save",
            comment: "Primary save for title or type edit"
        )

        static let editTitle = LocalizedStringResource(
            "sources.page.editTitle",
            defaultValue: "Edit title",
            comment: "Accessibility label for pencil to edit Source title"
        )

        static let editType = LocalizedStringResource(
            "sources.page.editType",
            defaultValue: "Edit type",
            comment: "Accessibility label for pencil to edit Source type"
        )

        static let editMetadataValue = LocalizedStringResource(
            "sources.page.editMetadataValue",
            defaultValue: "Edit value",
            comment: "Accessibility label for pencil to edit a saved text metadata value"
        )

        static let saveMetadataValue = LocalizedStringResource(
            "sources.page.saveMetadataValue",
            defaultValue: "Save value",
            comment: "Accessibility label for check to save an inline metadata text edit"
        )

        static let editMetadataDateValue = LocalizedStringResource(
            "sources.page.editMetadataDateValue",
            defaultValue: "Edit date value",
            comment: "Accessibility label for pencil to open the date metadata dialog"
        )

        static let editDescription = LocalizedStringResource(
            "sources.page.editDescription",
            defaultValue: "Edit",
            comment: "Button to enter description edit mode"
        )

        static let saveDescription = LocalizedStringResource(
            "sources.page.saveDescription",
            defaultValue: "Save description",
            comment: "Button to save Source description"
        )

        static let credibilitySavedStatus = LocalizedStringResource(
            "sources.page.credibilitySavedStatus",
            defaultValue: "Saved assessment",
            comment: "Status when credibility draft matches saved assessment"
        )

        static let credibilityUnsavedStatus = LocalizedStringResource(
            "sources.page.credibilityUnsavedStatus",
            defaultValue: "No assessment saved",
            comment: "Status when there is no credibility row and draft is clean at standard"
        )

        static let noUnsavedChanges = LocalizedStringResource(
            "sources.page.noUnsavedChanges",
            defaultValue: "No unsaved changes",
            comment: "Status under artifact fields when drafts match saved values"
        )

        static let addDateDialogTitle = LocalizedStringResource(
            "sources.page.addDateDialogTitle",
            defaultValue: "Add date value",
            comment: "Title of DateValue editor when creating structure"
        )

        static let editDateDialogTitle = LocalizedStringResource(
            "sources.page.editDateDialogTitle",
            defaultValue: "Edit date value",
            comment: "Title of DateValue editor when editing structure"
        )

        static let saveDateConfirm = LocalizedStringResource(
            "sources.page.saveDateConfirm",
            defaultValue: "Save value",
            comment: "Confirm button on the date metadata dialog"
        )

        static let dateValueAsWritten = LocalizedStringResource(
            "sources.page.dateValueAsWritten",
            defaultValue: "Value as written",
            comment: "Label for the plain-text wording field in the date metadata dialog"
        )

        static let dateValueAsWrittenHint = LocalizedStringResource(
            "sources.page.dateValueAsWrittenHint",
            defaultValue: "Keep the record's own wording; the structured date below is what search and sorting use",
            comment: "Hint under Value as written in the date metadata dialog"
        )

        static let dateKindLabel = LocalizedStringResource(
            "sources.page.dateKindLabel",
            defaultValue: "Kind",
            comment: "Label above DateValue kind segmented control"
        )

        static let dateKindPoint = LocalizedStringResource(
            "sources.page.dateKindPoint",
            defaultValue: "Single date",
            comment: "DateValue kind: point"
        )

        static let dateKindRange = LocalizedStringResource(
            "sources.page.dateKindRange",
            defaultValue: "Between two bounds",
            comment: "DateValue kind: range"
        )

        static let dateKindRangeHint = LocalizedStringResource(
            "sources.page.dateKindRangeHint",
            defaultValue: "A single date somewhere in this window — not how long something lasted",
            comment: "Hint under Between kind"
        )

        static let dateQualifierLabel = LocalizedStringResource(
            "sources.page.dateQualifierLabel",
            defaultValue: "Qualifier",
            comment: "Label above DateValue qualifier chips"
        )

        static let dateQualifierAsStated = LocalizedStringResource(
            "sources.page.dateQualifierAsStated",
            defaultValue: "As stated",
            comment: "DateValue qualifier empty"
        )

        static let dateQualifierAbout = LocalizedStringResource(
            "sources.page.dateQualifierAbout",
            defaultValue: "About",
            comment: "DateValue qualifier ABT"
        )

        static let dateQualifierBefore = LocalizedStringResource(
            "sources.page.dateQualifierBefore",
            defaultValue: "Before",
            comment: "DateValue qualifier BEF"
        )

        static let dateQualifierAfter = LocalizedStringResource(
            "sources.page.dateQualifierAfter",
            defaultValue: "After",
            comment: "DateValue qualifier AFT"
        )

        static let dateEarliestHeading = LocalizedStringResource(
            "sources.page.dateEarliestHeading",
            defaultValue: "Earliest — not before",
            comment: "Uppercase heading for range start side"
        )

        static let dateLatestHeading = LocalizedStringResource(
            "sources.page.dateLatestHeading",
            defaultValue: "Latest — not after",
            comment: "Uppercase heading for range end side"
        )

        static let datePointHeading = LocalizedStringResource(
            "sources.page.datePointHeading",
            defaultValue: "Date",
            comment: "Uppercase heading for point date cascade"
        )

        static let dateLeaveEmptyHint = LocalizedStringResource(
            "sources.page.dateLeaveEmptyHint",
            defaultValue: "Leave a part empty when the record does not say",
            comment: "Hint beside date cascade heading"
        )

        static let dateYear = LocalizedStringResource(
            "sources.page.dateYear",
            defaultValue: "Year",
            comment: "DateValue year field label"
        )

        static let dateMonth = LocalizedStringResource(
            "sources.page.dateMonth",
            defaultValue: "Month",
            comment: "DateValue month field label"
        )

        static let dateMonthNone = LocalizedStringResource(
            "sources.page.dateMonth.none",
            defaultValue: "—",
            comment: "DateValue month option for no month"
        )

        static let dateMonthJanuary = LocalizedStringResource(
            "sources.page.dateMonth.january",
            defaultValue: "January",
            comment: "DateValue January option"
        )
        static let dateMonthFebruary = LocalizedStringResource(
            "sources.page.dateMonth.february",
            defaultValue: "February",
            comment: "DateValue February option"
        )
        static let dateMonthMarch = LocalizedStringResource(
            "sources.page.dateMonth.march",
            defaultValue: "March",
            comment: "DateValue March option"
        )
        static let dateMonthApril = LocalizedStringResource(
            "sources.page.dateMonth.april",
            defaultValue: "April",
            comment: "DateValue April option"
        )
        static let dateMonthMay = LocalizedStringResource(
            "sources.page.dateMonth.may",
            defaultValue: "May",
            comment: "DateValue May option"
        )
        static let dateMonthJune = LocalizedStringResource(
            "sources.page.dateMonth.june",
            defaultValue: "June",
            comment: "DateValue June option"
        )
        static let dateMonthJuly = LocalizedStringResource(
            "sources.page.dateMonth.july",
            defaultValue: "July",
            comment: "DateValue July option"
        )
        static let dateMonthAugust = LocalizedStringResource(
            "sources.page.dateMonth.august",
            defaultValue: "August",
            comment: "DateValue August option"
        )
        static let dateMonthSeptember = LocalizedStringResource(
            "sources.page.dateMonth.september",
            defaultValue: "September",
            comment: "DateValue September option"
        )
        static let dateMonthOctober = LocalizedStringResource(
            "sources.page.dateMonth.october",
            defaultValue: "October",
            comment: "DateValue October option"
        )
        static let dateMonthNovember = LocalizedStringResource(
            "sources.page.dateMonth.november",
            defaultValue: "November",
            comment: "DateValue November option"
        )
        static let dateMonthDecember = LocalizedStringResource(
            "sources.page.dateMonth.december",
            defaultValue: "December",
            comment: "DateValue December option"
        )

        static let dateDay = LocalizedStringResource(
            "sources.page.dateDay",
            defaultValue: "Day",
            comment: "DateValue day field label"
        )

        static let dateHour = LocalizedStringResource(
            "sources.page.dateHour",
            defaultValue: "Hour",
            comment: "DateValue hour field label"
        )

        static let dateMinute = LocalizedStringResource(
            "sources.page.dateMinute",
            defaultValue: "Min",
            comment: "DateValue minute field label"
        )

        static let dateSecond = LocalizedStringResource(
            "sources.page.dateSecond",
            defaultValue: "Sec",
            comment: "DateValue second field label"
        )

        static let dateMillisecond = LocalizedStringResource(
            "sources.page.dateMillisecond",
            defaultValue: "Ms",
            comment: "DateValue millisecond field label"
        )

        static let dateTimeZone = LocalizedStringResource(
            "sources.page.dateTimeZone",
            defaultValue: "Time zone",
            comment: "DateValue free-text timezone label"
        )

        static let dateAddTime = LocalizedStringResource(
            "sources.page.dateAddTime",
            defaultValue: "Add time…",
            comment: "Link to reveal optional time fields"
        )

        static let dateHideTime = LocalizedStringResource(
            "sources.page.dateHideTime",
            defaultValue: "Hide time",
            comment: "Link to hide optional time fields"
        )

        static let dateShowAdvanced = LocalizedStringResource(
            "sources.page.dateShowAdvanced",
            defaultValue: "Calendar and phrase…",
            comment: "Link to reveal calendar and phrase"
        )

        static let dateHideAdvanced = LocalizedStringResource(
            "sources.page.dateHideAdvanced",
            defaultValue: "Hide calendar and phrase",
            comment: "Link to hide calendar and phrase"
        )

        static let dateCalendar = LocalizedStringResource(
            "sources.page.dateCalendar",
            defaultValue: "Calendar",
            comment: "DateValue calendar picker label"
        )

        static let dateCalendarGregorian = LocalizedStringResource(
            "sources.page.dateCalendar.gregorian",
            defaultValue: "Gregorian",
            comment: "DateValue calendar picker option: Gregorian"
        )

        static let dateCalendarJulian = LocalizedStringResource(
            "sources.page.dateCalendar.julian",
            defaultValue: "Julian",
            comment: "DateValue calendar picker option: Julian"
        )

        static let dateCalendarFrenchRepublican = LocalizedStringResource(
            "sources.page.dateCalendar.frenchRepublican",
            defaultValue: "French Republican",
            comment: "DateValue calendar picker option: French Republican"
        )

        static let dateCalendarHebrew = LocalizedStringResource(
            "sources.page.dateCalendar.hebrew",
            defaultValue: "Hebrew",
            comment: "DateValue calendar picker option: Hebrew"
        )

        static let datePhrase = LocalizedStringResource(
            "sources.page.datePhrase",
            defaultValue: "Phrase",
            comment: "DateValue phrase field label"
        )

        static let datePhrasePrompt = LocalizedStringResource(
            "sources.page.datePhrasePrompt",
            defaultValue: "Michaelmas term",
            comment: "Placeholder for DateValue phrase"
        )

        static let datePhraseHint = LocalizedStringResource(
            "sources.page.datePhraseHint",
            defaultValue: "A short gloss carried on the date itself — not the source's wording",
            comment: "Hint under DateValue phrase"
        )

        static let dateStoredAs = LocalizedStringResource(
            "sources.page.dateStoredAs",
            defaultValue: "Stored as",
            comment: "Label beside DateValue summary preview"
        )

        static let dateRangeOrderError = LocalizedStringResource(
            "sources.page.dateRangeOrderError",
            defaultValue: "The latest bound falls before the earliest bound. The date has to sit inside the window.",
            comment: "Inline error when range end precedes start"
        )

        static let dateYearOutOfRange = LocalizedStringResource(
            "sources.page.dateYearOutOfRange",
            defaultValue: "Enter a year between 1 and 9999.",
            comment: "Inline error when DateValue year is out of range"
        )

        static let dateMonthOutOfRange = LocalizedStringResource(
            "sources.page.dateMonthOutOfRange",
            defaultValue: "Month must be between 1 and 12.",
            comment: "Inline error when DateValue month is out of range"
        )

        static let dateDayOutOfRange = LocalizedStringResource(
            "sources.page.dateDayOutOfRange",
            defaultValue: "Day must be between 1 and 31.",
            comment: "Inline error when DateValue day is out of range"
        )

        static let dateHourOutOfRange = LocalizedStringResource(
            "sources.page.dateHourOutOfRange",
            defaultValue: "Hour must be between 0 and 23.",
            comment: "Inline error when DateValue hour is out of range"
        )

        static let dateMinuteOutOfRange = LocalizedStringResource(
            "sources.page.dateMinuteOutOfRange",
            defaultValue: "Minute must be between 0 and 59.",
            comment: "Inline error when DateValue minute is out of range"
        )

        static let dateSecondOutOfRange = LocalizedStringResource(
            "sources.page.dateSecondOutOfRange",
            defaultValue: "Second must be between 0 and 59.",
            comment: "Inline error when DateValue second is out of range"
        )

        static let dateMillisecondOutOfRange = LocalizedStringResource(
            "sources.page.dateMillisecondOutOfRange",
            defaultValue: "Millisecond must be between 0 and 999.",
            comment: "Inline error when DateValue millisecond is out of range"
        )

        static func metadataFieldCount(_ count: Int) -> String {
            if count == 1 {
                return String(localized: LocalizedStringResource(
                    "sources.page.metadataFieldCountOne",
                    defaultValue: "1 field",
                    comment: "Metadata section count when exactly one saved field"
                ))
            }
            let format = String(localized: LocalizedStringResource(
                "sources.page.metadataFieldCountMany",
                defaultValue: "%d fields",
                comment: "Metadata section count; argument is saved field count"
            ))
            return String(format: format, locale: .current, count)
        }

        static func artifactsCount(_ count: Int) -> String {
            if count == 1 {
                return String(localized: LocalizedStringResource(
                    "sources.page.artifactsCountOne",
                    defaultValue: "1 artifact",
                    comment: "Artifacts section count when exactly one"
                ))
            }
            let format = String(localized: LocalizedStringResource(
                "sources.page.artifactsCountMany",
                defaultValue: "%d artifacts",
                comment: "Artifacts section count; argument is artifact count"
            ))
            return String(format: format, locale: .current, count)
        }

        static let metadataHeading = LocalizedStringResource(
            "sources.page.metadataHeading",
            defaultValue: "Metadata",
            comment: "Heading for the Metadata section on the Source page"
        )

        static let addMetadata = LocalizedStringResource(
            "sources.page.addMetadata",
            defaultValue: "Add field",
            comment: "Button to open Add metadata field dialog"
        )

        static let metadataIntro = LocalizedStringResource(
            "sources.page.metadataIntro",
            defaultValue: "Metadata describes the source itself — what the record says about its own making: a registration number, a call number, the reel it sits on. Facts the record asserts about people or events are not metadata; those become citations and asserted facts, so they can move on to the entities they describe.",
            comment: "Prose under the Metadata section heading"
        )

        static let metadataEmptyTitle = LocalizedStringResource(
            "sources.page.metadataEmptyTitle",
            defaultValue: "No metadata yet",
            comment: "Empty-state title when a Source has no metadata rows"
        )

        static let metadataEmptyMessage = LocalizedStringResource(
            "sources.page.metadataEmptyMessage",
            defaultValue: "Accept a type suggestion or add a field from the vocabulary.",
            comment: "Empty-state body when a Source has no metadata rows"
        )

        static let dismissMetadataSuggestion = LocalizedStringResource(
            "sources.page.dismissMetadataSuggestion",
            defaultValue: "Dismiss this suggestion",
            comment: "Accessibility label for dismissing a type metadata suggestion"
        )

        static let addMetadataDialogTitle = LocalizedStringResource(
            "sources.page.addMetadataDialogTitle",
            defaultValue: "Add metadata",
            comment: "Title of the Add metadata dialog"
        )

        static let addMetadataDialogSubtitle = LocalizedStringResource(
            "sources.page.addMetadataDialogSubtitle",
            defaultValue: "Pick any field from this project’s vocabulary, then enter the value as written on the record.",
            comment: "Subtitle of the Add metadata dialog"
        )

        static let addMetadataConfirm = LocalizedStringResource(
            "sources.page.addMetadataConfirm",
            defaultValue: "Add",
            comment: "Confirm button on the Add metadata dialog"
        )

        static let metadataField = LocalizedStringResource(
            "sources.page.metadataField",
            defaultValue: "Field",
            comment: "Label for metadata field picker"
        )

        static let metadataFieldHint = LocalizedStringResource(
            "sources.page.metadataFieldHint",
            defaultValue: "Search the whole source metadata vocabulary",
            comment: "Hint under metadata field picker"
        )

        static let metadataValue = LocalizedStringResource(
            "sources.page.metadataValue",
            defaultValue: "Value",
            comment: "Label for metadata value entry"
        )

        static let metadataValueHint = LocalizedStringResource(
            "sources.page.metadataValueHint",
            defaultValue: "Enter it as written on the record",
            comment: "Hint under metadata value field"
        )

        static let metadataFieldRequired = LocalizedStringResource(
            "sources.page.metadataFieldRequired",
            defaultValue: "Choose a field.",
            comment: "Validation when Add metadata is submitted without a field"
        )

        static let metadataValueRequired = LocalizedStringResource(
            "sources.page.metadataValueRequired",
            defaultValue: "Enter a value.",
            comment: "Validation when Add metadata is submitted without a value"
        )

        static let metadataSuggestionPlaceholder = LocalizedStringResource(
            "sources.page.metadataSuggestionPlaceholder",
            defaultValue: "Add a value…",
            comment: "Placeholder on an empty type-suggestion metadata row"
        )

        static let metadataSuggestionsHeading = LocalizedStringResource(
            "sources.page.metadataSuggestionsHeading",
            defaultValue: "Suggested by this type",
            comment: "Subheading above type-suggested metadata fields without values"
        )

        static let saveMetadataSuggestion = LocalizedStringResource(
            "sources.page.saveMetadataSuggestion",
            defaultValue: "Save",
            comment: "Button to save a value on a type-suggested metadata row"
        )

        static let artifactsHeading = LocalizedStringResource(
            "sources.page.artifactsHeading",
            defaultValue: "Artifacts",
            comment: "Heading for the Artifacts section on the Source page"
        )

        static let addArtifact = LocalizedStringResource(
            "sources.page.addArtifact",
            defaultValue: "Add artifact",
            comment: "Button to open the Add Artifact dialog"
        )

        static let artifactsEmptyTitle = LocalizedStringResource(
            "sources.page.artifactsEmptyTitle",
            defaultValue: "No artifacts yet",
            comment: "Empty state title when a Source has no Artifacts"
        )

        static let artifactsEmptyMessage = LocalizedStringResource(
            "sources.page.artifactsEmptyMessage",
            defaultValue: "Add a scan, photo, or a fileless stand-in for something you only saw in person.",
            comment: "Empty state message for Artifacts"
        )

        static let artifactLabel = LocalizedStringResource(
            "sources.page.artifactLabel",
            defaultValue: "Label",
            comment: "Field label for an Artifact's required list headline"
        )

        static let artifactLabelHint = LocalizedStringResource(
            "sources.page.artifactLabelHint",
            defaultValue: "Name it so the row is recognisable in the list",
            comment: "Hint under Artifact label field"
        )

        static let artifactDescription = LocalizedStringResource(
            "sources.page.artifactDescription",
            defaultValue: "Description",
            comment: "Field label for optional Artifact description"
        )

        static let artifactDescriptionHint = LocalizedStringResource(
            "sources.page.artifactDescriptionHint",
            defaultValue: "Optional — folio, entry number, condition of the scan",
            comment: "Hint under Artifact description on the page"
        )

        static let artifactLabelRequired = LocalizedStringResource(
            "sources.page.artifactLabelRequired",
            defaultValue: "Give the artifact a label.",
            comment: "Validation when Add Artifact is submitted without a label"
        )

        static let saveArtifact = LocalizedStringResource(
            "sources.page.saveArtifact",
            defaultValue: "Save artifact",
            comment: "Button to save edited Artifact label and description"
        )

        static let openFile = LocalizedStringResource(
            "sources.page.openFile",
            defaultValue: "Open",
            comment: "Button to open an Artifact's primary File in an external app"
        )

        static let primaryFileHeading = LocalizedStringResource(
            "sources.page.primaryFileHeading",
            defaultValue: "Primary file",
            comment: "Uppercase section label above an Artifact's primary File card"
        )

        static let openFileCaption = LocalizedStringResource(
            "sources.page.openFileCaption",
            defaultValue: "Opens in the system's default app · the file is immutable once ingested",
            comment: "Caption under the Open file control"
        )

        static let addFile = LocalizedStringResource(
            "sources.page.addFile",
            defaultValue: "Add file…",
            comment: "Button to attach a first file to a fileless Artifact"
        )

        static let filelessHint = LocalizedStringResource(
            "sources.page.filelessHint",
            defaultValue: "Physical only — you have recorded the item without a scan. Attach a file when one exists; the artifact keeps its reference either way.",
            comment: "Callout when an Artifact has no primary File"
        )

        static let notesHeading = LocalizedStringResource(
            "sources.page.notesHeading",
            defaultValue: "Notes",
            comment: "Heading for the Notes stream on the Source page"
        )

        static let notesEmptyTitle = LocalizedStringResource(
            "sources.page.notesEmptyTitle",
            defaultValue: "No notes yet",
            comment: "Empty-state title when a Source has no notes"
        )

        static let notesEmptyMessage = LocalizedStringResource(
            "sources.page.notesEmptyMessage",
            defaultValue: "Record what the record itself cannot say — legibility, gaps in the film, what to order on the next visit.",
            comment: "Empty-state body under Notes"
        )

        static let notePlaceholder = LocalizedStringResource(
            "sources.page.notePlaceholder",
            defaultValue: "Add a note about this source",
            comment: "Placeholder for the new-note composer"
        )

        /// Composer byline beside the draft field (`Jake Robins · now`).
        static func noteComposerAttribution(displayName: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sources.page.noteComposerAttribution",
                defaultValue: "%@ · now",
                comment: "Note composer byline; argument is the session display name"
            ))
            return String(format: format, locale: .current, displayName)
        }

        static let addNote = LocalizedStringResource(
            "sources.page.addNote",
            defaultValue: "Add note",
            comment: "Button to save a new Source note"
        )

        static let editNote = LocalizedStringResource(
            "sources.page.editNote",
            defaultValue: "Edit note",
            comment: "Accessibility label for pencil to edit a Source note"
        )

        static let saveNote = LocalizedStringResource(
            "sources.page.saveNote",
            defaultValue: "Save note",
            comment: "Button to commit an edited Source note"
        )

        static let noteBodyRequired = LocalizedStringResource(
            "sources.page.noteBodyRequired",
            defaultValue: "A note needs some text.",
            comment: "Validation when saving an empty Source note body"
        )

        static let deleteNote = LocalizedStringResource(
            "sources.page.deleteNote",
            defaultValue: "Delete note",
            comment: "Accessibility label for deleting a Source note"
        )

        static let addArtifactDialogTitle = LocalizedStringResource(
            "sources.page.addArtifactDialogTitle",
            defaultValue: "Add artifact",
            comment: "Title of the Add Artifact dialog"
        )

        static let addArtifactDialogSubtitle = LocalizedStringResource(
            "sources.page.addArtifactDialogSubtitle",
            defaultValue: "A concrete representation of this source — scan, photo, or physical-only stand-in.",
            comment: "Subtitle of the Add Artifact dialog"
        )

        static let addArtifactConfirm = LocalizedStringResource(
            "sources.page.addArtifactConfirm",
            defaultValue: "Add artifact",
            comment: "Confirm button on the Add Artifact dialog"
        )

        static let optionalFile = LocalizedStringResource(
            "sources.page.optionalFile",
            defaultValue: "File",
            comment: "Label for optional file pick on Add Artifact"
        )

        static let chooseFile = LocalizedStringResource(
            "sources.page.chooseFile",
            defaultValue: "Choose file…",
            comment: "Button to pick a file in Add Artifact"
        )

        static let clearFile = LocalizedStringResource(
            "sources.page.clearFile",
            defaultValue: "Remove",
            comment: "Clear a chosen file before creating an Artifact"
        )

        static let filePickPrompt = LocalizedStringResource(
            "sources.page.filePickPrompt",
            defaultValue: "Choose",
            comment: "NSOpenPanel confirm button for Artifact file pick"
        )

        static let filePickMessage = LocalizedStringResource(
            "sources.page.filePickMessage",
            defaultValue: "Choose a file to attach to this artifact. Provenencia copies it into the project.",
            comment: "NSOpenPanel message for Artifact file pick"
        )

        static let ingestAllowedTypesCaption = LocalizedStringResource(
            "sources.page.ingestAllowedTypesCaption",
            defaultValue: "Images, PDF, Word, text (CSV, Markdown), audio, or video · 512 MB maximum",
            comment: "Always-visible caption under the ingest file drop row"
        )

        static let ingestDropIdleCreate = LocalizedStringResource(
            "sources.page.ingestDropIdleCreate",
            defaultValue: "Drop a file here, or skip — the artifact stays physical-only",
            comment: "Idle drop hint in Add Artifact when no file is chosen"
        )

        static let ingestDropIdleAttach = LocalizedStringResource(
            "sources.page.ingestDropIdleAttach",
            defaultValue: "Drop a file here, or choose one",
            comment: "Idle drop hint in Add File dialog"
        )

        static let ingestDropActive = LocalizedStringResource(
            "sources.page.ingestDropActive",
            defaultValue: "Drop to attach",
            comment: "Drop-target active hint while dragging a file over the ingest row"
        )

        static let addFileDialogTitle = LocalizedStringResource(
            "sources.page.addFileDialogTitle",
            defaultValue: "Add file",
            comment: "Title of the first-attach Add File dialog"
        )

        static let addFileDialogSubtitle = LocalizedStringResource(
            "sources.page.addFileDialogSubtitle",
            defaultValue: "First attach",
            comment: "Subtitle of the Add File dialog"
        )

        static let addFileConfirm = LocalizedStringResource(
            "sources.page.addFileConfirm",
            defaultValue: "Attach file",
            comment: "Confirm button on the Add File dialog"
        )

        static let addFileFirstAttachHint = LocalizedStringResource(
            "sources.page.addFileFirstAttachHint",
            defaultValue: "First attach only. Once this artifact has a file it keeps it — a better scan becomes a new artifact.",
            comment: "Info callout in the Add File dialog"
        )

        static let fileOpenMissing = LocalizedStringResource(
            "sources.page.fileOpenMissing",
            defaultValue: "That file is missing from the project folder.",
            comment: "Error when Open cannot find the object on disk"
        )

        static let pageFormType = LocalizedStringResource(
            "sources.page.formType",
            defaultValue: "Type",
            comment: "Source type field on the Source page"
        )

        static func toastArtifactCreatedTitle(ref: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sources.page.toastArtifactCreatedTitle",
                defaultValue: "%@",
                comment: "Toast title after creating an Artifact; argument is ART- ref"
            ))
            return String(format: format, locale: .current, ref)
        }

        static let toastArtifactCreatedFileless = String(localized: LocalizedStringResource(
            "sources.page.toastArtifactCreatedFileless",
            defaultValue: "Created with no file yet — physical only.",
            comment: "Toast body after creating a fileless Artifact"
        ))

        static let toastArtifactCreatedWithFile = String(localized: LocalizedStringResource(
            "sources.page.toastArtifactCreatedWithFile",
            defaultValue: "Created and the file was ingested.",
            comment: "Toast body after creating an Artifact with a file"
        ))

        static let toastFileAttachedTitle = String(localized: LocalizedStringResource(
            "sources.page.toastFileAttachedTitle",
            defaultValue: "File attached",
            comment: "Toast title after first-attach ingest"
        ))

        static func toastFileAttachedBody(name: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sources.page.toastFileAttachedBody",
                defaultValue: "%@ was ingested into the project.",
                comment: "Toast body after ingest; argument is original filename"
            ))
            return String(format: format, locale: .current, name)
        }

        static let coverBadge = LocalizedStringResource(
            "sources.page.coverBadge",
            defaultValue: "Cover",
            comment: "Badge on the Artifact currently used as the Source thumbnail"
        )

        static let useAsThumbnail = LocalizedStringResource(
            "sources.page.useAsThumbnail",
            defaultValue: "Use as thumbnail",
            comment: "Button to pin an Artifact as the Source list/identity cover"
        )

        static let thumbnailMenuTitle = LocalizedStringResource(
            "sources.page.thumbnailMenuTitle",
            defaultValue: "Source thumbnail",
            comment: "Context menu title on the Source identity cover"
        )

        static let thumbnailMenuHint = LocalizedStringResource(
            "sources.page.thumbnailMenuHint",
            defaultValue: "Right-click for thumbnail options",
            comment: "Accessibility hint on the Source identity cover thumbnail"
        )

        static let thumbnailRevertToDefault = LocalizedStringResource(
            "sources.page.thumbnailRevertToDefault",
            defaultValue: "Revert to default",
            comment: "Context menu action to use the Source type icon as cover"
        )

        static let thumbnailDefaultInUse = LocalizedStringResource(
            "sources.page.thumbnailDefaultInUse",
            defaultValue: "Default icon is in use",
            comment: "Disabled context menu item when the type icon is already the cover"
        )

        static let toastThumbnailUpdatedTitle = String(localized: LocalizedStringResource(
            "sources.page.toastThumbnailUpdatedTitle",
            defaultValue: "Thumbnail updated",
            comment: "Toast title after changing Source cover"
        ))

        static func toastThumbnailUpdatedArtifactBody(ref: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sources.page.toastThumbnailUpdatedArtifactBody",
                defaultValue: "Using %@ as the Source thumbnail.",
                comment: "Toast body after pinning an Artifact; argument is ART- ref"
            ))
            return String(format: format, locale: .current, ref)
        }

        static let toastThumbnailUpdatedTypeIconBody = String(localized: LocalizedStringResource(
            "sources.page.toastThumbnailUpdatedTypeIconBody",
            defaultValue: "Using the Source type icon as the thumbnail.",
            comment: "Toast body after reverting cover to the type icon"
        ))
    }

    enum SourceFields {
        static let description = LocalizedStringResource(
            "sourceFields.list.description",
            defaultValue: "The metadata a source can carry in this project. Provenencia seeds the common fields; you add the ones your records actually use.",
            comment: "Explanatory copy under the Source fields page title"
        )

        static func countLine(total: Int, seeded: Int, user: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceFields.list.countLine",
                defaultValue: "%lld fields · %lld seeded · %lld yours",
                comment: "Source fields count summary; arguments are total, seeded (provenencia), and user field counts"
            ))
            return String(format: format, locale: .current, total, seeded, user)
        }

        static func countLineWithPlugin(total: Int, seeded: Int, user: Int, plugin: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceFields.list.countLineWithPlugin",
                defaultValue: "%lld fields · %lld seeded · %lld yours · %lld plugin",
                comment: "Source fields count summary including plugin-origin fields; arguments are total, seeded, user, and plugin field counts"
            ))
            return String(format: format, locale: .current, total, seeded, user, plugin)
        }

        static let addField = LocalizedStringResource(
            "sourceFields.list.addField",
            defaultValue: "Add field",
            comment: "Button: add a new Source field (toolbar, empty state, and add-form submit)"
        )

        static let columnLabel = LocalizedStringResource(
            "sourceFields.list.columnLabel",
            defaultValue: "Label",
            comment: "Source fields list column header: label"
        )

        static let columnKey = LocalizedStringResource(
            "sourceFields.list.columnKey",
            defaultValue: "Key",
            comment: "Source fields list column header: key"
        )

        static let columnDataType = LocalizedStringResource(
            "sourceFields.list.columnDataType",
            defaultValue: "Data type",
            comment: "Source fields list column header: data type"
        )

        static let dataTypeText = LocalizedStringResource(
            "sourceFields.dataType.text",
            defaultValue: "text",
            comment: "Source field data type badge/option: text"
        )

        static let dataTypeDate = LocalizedStringResource(
            "sourceFields.dataType.date",
            defaultValue: "date",
            comment: "Source field data type badge/option: date"
        )

        static let dataTypeUrl = LocalizedStringResource(
            "sourceFields.dataType.url",
            defaultValue: "url",
            comment: "Source field data type badge/option: url"
        )

        static let deleteField = LocalizedStringResource(
            "sourceFields.delete.action",
            defaultValue: "Delete field",
            comment: "Delete button in the Source fields detail pane, and the confirm dialog's destructive button"
        )

        static let deleteOwnedByPlugin = LocalizedStringResource(
            "sourceFields.delete.ownedByPlugin",
            defaultValue: "Owned by the plugin",
            comment: "Tooltip on the disabled delete button when the selected field comes from a plugin"
        )

        static func deleteInUse(count: Int) -> LocalizedStringResource {
            count == 1 ? deleteInUseOne : deleteInUseOther(count: count)
        }

        private static let deleteInUseOne = LocalizedStringResource(
            "sourceFields.delete.inUseOne",
            defaultValue: "In use on 1 source",
            comment: "Tooltip on the disabled delete button when exactly one source carries a value for the field"
        )

        private static func deleteInUseOther(count: Int) -> LocalizedStringResource {
            LocalizedStringResource(
                "sourceFields.delete.inUseOther",
                defaultValue: "In use on \(count) sources",
                comment: "Tooltip on the disabled delete button; argument is how many sources carry a value for the field"
            )
        }

        static func deleteConfirmTitle(label: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceFields.delete.confirmTitle",
                defaultValue: "Delete %@?",
                comment: "Title of the delete-field confirmation dialog; argument is the field label"
            ))
            return String(format: format, locale: .current, label)
        }

        static let deleteConfirmMessage = LocalizedStringResource(
            "sourceFields.delete.confirmMessage",
            defaultValue: "No source in this project carries a value for this field, so nothing is lost. The key is released and can be minted again by a later field with the same label.",
            comment: "Message of the delete-field confirmation sheet: what is and is not lost"
        )

        static let deleteKeyReleased = LocalizedStringResource(
            "sourceFields.delete.keyReleased",
            defaultValue: "Key released",
            comment: "Micro-caps label beside the key a field delete releases, in the confirmation sheet"
        )

        static let deleteKeep = LocalizedStringResource(
            "sourceFields.delete.keep",
            defaultValue: "Keep field",
            comment: "Button that closes the delete-field confirmation without deleting"
        )

        static let toastDeletedTitle = LocalizedStringResource(
            "sourceFields.toast.deletedTitle",
            defaultValue: "Field deleted",
            comment: "Toast title after a source field is deleted"
        )

        static func toastDeletedBody(label: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceFields.toast.deletedBody",
                defaultValue: "%@ is no longer in this project's vocabulary.",
                comment: "Toast body after a source field is deleted; argument is the field label"
            ))
            return String(format: format, locale: .current, label)
        }

        static let emptyProjectTitle = LocalizedStringResource(
            "sourceFields.emptyProject.title",
            defaultValue: "No source fields yet",
            comment: "Title of the empty state when the project has zero metadata fields"
        )

        static let emptyProjectBody = LocalizedStringResource(
            "sourceFields.emptyProject.body",
            defaultValue: "This project has no metadata vocabulary. Add the fields your records actually carry — a certificate number, a photographer, an album code.",
            comment: "Body of the empty state when the project has zero metadata fields"
        )

        static func resultLine(shown: Int, total: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceFields.list.resultLineAll",
                defaultValue: "%lld fields",
                comment: "Footer result count for the Source fields list; argument is the total"
            ))
            return String(format: format, locale: .current, total)
        }

        static let detailEyebrowField = LocalizedStringResource(
            "sourceFields.detail.eyebrowField",
            defaultValue: "Field",
            comment: "Eyebrow label above an existing field's detail panel"
        )

        static let detailEyebrowNewField = LocalizedStringResource(
            "sourceFields.detail.eyebrowNewField",
            defaultValue: "New field",
            comment: "Eyebrow label above the add-field panel"
        )

        static let keyHintAdd = LocalizedStringResource(
            "sourceFields.detail.keyHintAdd",
            defaultValue: "Provenencia mints the key from the label when the field is added",
            comment: "Hint under the live key preview while adding a field"
        )

        static let keyHintEdit = LocalizedStringResource(
            "sourceFields.detail.keyHintEdit",
            defaultValue: "The key is minted once from the label and never changes — renaming the field keeps existing sources attached",
            comment: "Hint under the key on an existing field's detail panel"
        )

        static func lockedNotePlugin(pluginID: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceFields.detail.lockedNotePlugin",
                defaultValue: "Supplied by the %@ plugin. The plugin owns this definition — Provenencia will not edit it.",
                comment: "Callout explaining why a plugin-origin field can't be edited; argument is the plugin id"
            ))
            return String(format: format, locale: .current, pluginID)
        }

        static let dataTypeSectionLabel = LocalizedStringResource(
            "sourceFields.detail.dataTypeSectionLabel",
            defaultValue: "Data type",
            comment: "Section label above the read-only data type line on a locked field's detail"
        )

        static let descriptionSectionLabel = LocalizedStringResource(
            "sourceFields.detail.descriptionSectionLabel",
            defaultValue: "Description",
            comment: "Section label above the read-only description on a locked field's detail"
        )

        static let descriptionEmptyPlaceholder = LocalizedStringResource(
            "sourceFields.detail.descriptionEmptyPlaceholder",
            defaultValue: "—",
            comment: "Shown in place of a locked field's description when it has none"
        )

        static let panelEmptyTitle = LocalizedStringResource(
            "sourceFields.detail.panelEmptyTitle",
            defaultValue: "No field selected",
            comment: "Title of the empty state shown in the detail panel before any field is selected"
        )

        static let panelEmptyBody = LocalizedStringResource(
            "sourceFields.detail.panelEmptyBody",
            defaultValue: "Select a field to read or edit its definition. Fields supplied by a plugin are read-only; the rest of this project's vocabulary stays editable.",
            comment: "Body of the empty state shown in the detail panel before any field is selected"
        )

        static let formLabel = LocalizedStringResource(
            "sourceFields.form.label",
            defaultValue: "Label",
            comment: "Add/edit form field: label"
        )

        static let formLabelPlaceholder = LocalizedStringResource(
            "sourceFields.form.labelPlaceholder",
            defaultValue: "Grandma’s album code",
            comment: "Placeholder text for the add-field label input"
        )

        static let formDataType = LocalizedStringResource(
            "sourceFields.form.dataType",
            defaultValue: "Data type",
            comment: "Add/edit form field: data type picker"
        )

        static let formDataTypeHint = LocalizedStringResource(
            "sourceFields.form.dataTypeHint",
            defaultValue: "Text, date, or url — chosen at create and immutable afterward",
            comment: "Hint under the data type picker on add"
        )

        static let formDataTypeImmutableHint = LocalizedStringResource(
            "sourceFields.form.dataTypeImmutableHint",
            defaultValue: "Data type is fixed when the field is created so existing values stay valid.",
            comment: "Hint under the read-only data type on edit"
        )

        static let formDescription = LocalizedStringResource(
            "sourceFields.form.description",
            defaultValue: "Description",
            comment: "Add/edit form field: description"
        )

        static let formDescriptionHint = LocalizedStringResource(
            "sourceFields.form.descriptionHint",
            defaultValue: "What a researcher should put in this field, in your own words",
            comment: "Hint under the description field"
        )

        static let formDescriptionPlaceholder = LocalizedStringResource(
            "sourceFields.form.descriptionPlaceholder",
            defaultValue: "Pencil code on the back of prints from the album",
            comment: "Placeholder text for the add-field description input"
        )

        static let errorLabelRequired = LocalizedStringResource(
            "sourceFields.form.errorLabelRequired",
            defaultValue: "A label is required — it is how the field reads on a source.",
            comment: "Inline validation error when the label is blank"
        )

        static let errorUnslugifiable = LocalizedStringResource(
            "sourceFields.form.errorUnslugifiable",
            defaultValue: "That label cannot be turned into a key. Use at least one letter or number.",
            comment: "Inline validation error when the label has no letters or digits to slug"
        )

        static let saveSaving = LocalizedStringResource(
            "sourceFields.form.saveSaving",
            defaultValue: "Saving",
            comment: "Primary button label while a Source field add/edit is in flight"
        )

        static let saveChanges = LocalizedStringResource(
            "sourceFields.form.saveChanges",
            defaultValue: "Save changes",
            comment: "Primary button label for committing an edit to an existing field"
        )

        static let cancel = LocalizedStringResource(
            "sourceFields.form.cancel",
            defaultValue: "Cancel",
            comment: "Secondary button label that dismisses the add-field form"
        )

        static let revert = LocalizedStringResource(
            "sourceFields.form.revert",
            defaultValue: "Revert",
            comment: "Secondary button label that discards unsaved edits to an existing field"
        )

        static let toastAddedTitle = LocalizedStringResource(
            "sourceFields.toast.addedTitle",
            defaultValue: "Field added",
            comment: "Success toast title after creating a Source field"
        )

        static func toastAddedBody(label: String, key: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceFields.toast.addedBody",
                defaultValue: "%@ is in this project’s vocabulary as %@.",
                comment: "Success toast body after creating a Source field; arguments are label then minted key"
            ))
            return String(format: format, locale: .current, label, key)
        }

        static let toastUpdatedTitle = LocalizedStringResource(
            "sourceFields.toast.updatedTitle",
            defaultValue: "Field updated",
            comment: "Success toast title after editing a Source field"
        )

        static func toastUpdatedBody(label: String, key: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceFields.toast.updatedBody",
                defaultValue: "%@ — the key stays %@.",
                comment: "Success toast body after editing a Source field; arguments are label then key"
            ))
            return String(format: format, locale: .current, label, key)
        }
    }

    enum SubjectFields {
        static let description = LocalizedStringResource(
            "subjectFields.list.description",
            defaultValue: "The properties a subject can carry, and which of the seven subject types carry them.",
            comment: "Explanatory copy under the Subject fields page title"
        )
        static let allProperties = LocalizedStringResource(
            "subjectFields.strip.all",
            defaultValue: "All properties",
            comment: "Type strip card that clears the subject-type filter"
        )
        static let bridgeRole = LocalizedStringResource(
            "subjectFields.strip.bridge",
            defaultValue: "bridge",
            comment: "Micro-label on bridge subject-type strip cards; rendered uppercase (BRIDGE)"
        )
        static let typeStripAccessibility = LocalizedStringResource(
            "subjectFields.strip.accessibility",
            defaultValue: "Subject types",
            comment: "Accessibility label for the type strip pressed-button group"
        )
        static func stripFieldCount(count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "subjectFields.strip.fieldCount",
                defaultValue: "%lld fields",
                comment: "Type strip count under a subject type; argument is binding count"
            ))
            return String(format: format, locale: .current, count)
        }
        static let searchPlaceholder = LocalizedStringResource(
            "subjectFields.search.placeholder",
            defaultValue: "Search properties",
            comment: "List-card search field placeholder on Subject fields"
        )
        static let originUserShort = LocalizedStringResource(
            "subjectFields.table.originUser",
            defaultValue: "user",
            comment: "Table origin column for researcher-created properties"
        )
        static let originSeededShort = LocalizedStringResource(
            "subjectFields.table.originSeeded",
            defaultValue: "seeded",
            comment: "Table origin column for Provenencia-seeded properties"
        )
        static let valueTypeText = LocalizedStringResource(
            "subjectFields.valueType.text",
            defaultValue: "Text",
            comment: "Property value type label: text"
        )
        static let valueTypeInteger = LocalizedStringResource(
            "subjectFields.valueType.integer",
            defaultValue: "Integer",
            comment: "Property value type label: integer"
        )
        static let valueTypeDate = LocalizedStringResource(
            "subjectFields.valueType.date",
            defaultValue: "Date",
            comment: "Property value type label: date"
        )
        static let valueTypeName = LocalizedStringResource(
            "subjectFields.valueType.name",
            defaultValue: "Name",
            comment: "Property value type label: name"
        )
        static let valueTypeSubject = LocalizedStringResource(
            "subjectFields.valueType.subject",
            defaultValue: "Subject",
            comment: "Property value type label: subject"
        )
        static let valueTypeTerm = LocalizedStringResource(
            "subjectFields.valueType.term",
            defaultValue: "Term",
            comment: "Property value type label: term (registry-only)"
        )
        static let columnOn = LocalizedStringResource(
            "subjectFields.table.columnOn",
            defaultValue: "On",
            comment: "Subject fields table column: binding toggle for the focused type"
        )
        static let columnProperty = LocalizedStringResource(
            "subjectFields.table.columnProperty",
            defaultValue: "Property",
            comment: "Subject fields table column: property label"
        )
        static let columnValueType = LocalizedStringResource(
            "subjectFields.table.columnValueType",
            defaultValue: "Value type",
            comment: "Subject fields table column: value type"
        )
        static let columnOrigin = LocalizedStringResource(
            "subjectFields.table.columnOrigin",
            defaultValue: "Origin",
            comment: "Subject fields table column: origin"
        )
        static let columnBoundTo = LocalizedStringResource(
            "subjectFields.table.columnBoundTo",
            defaultValue: "Bound to",
            comment: "Subject fields table column: bound subject types"
        )
        static func boundOverflow(count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "subjectFields.table.boundOverflow",
                defaultValue: "+%lld",
                comment: "Overflow when more than three Bound-to chips; argument is remaining count"
            ))
            return String(format: format, locale: .current, count)
        }
        static func rowBoundAnnouncement(count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "subjectFields.table.rowBoundAnnouncement",
                defaultValue: "bound to %lld types",
                comment: "VoiceOver fragment for how many types a property is bound to"
            ))
            return String(format: format, locale: .current, count)
        }
        static func emptySearchTitle(query: String) -> String {
            let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
            if trimmed.isEmpty {
                return String(localized: LocalizedStringResource(
                    "subjectFields.table.emptySearchTitle",
                    defaultValue: "No properties match",
                    comment: "Empty table title when no properties are visible"
                ))
            }
            let format = String(localized: LocalizedStringResource(
                "subjectFields.table.emptySearchTitleQuery",
                defaultValue: "No property matches “%@”",
                comment: "Empty table title when search matches nothing; argument is the query"
            ))
            return String(format: format, locale: .current, trimmed)
        }
        static let emptySearch = LocalizedStringResource(
            "subjectFields.table.emptySearch",
            defaultValue: "Clear the search, or create it as a user property",
            comment: "Empty state when search/filter matches no properties"
        )
        static let newProperty = LocalizedStringResource(
            "subjectFields.toolbar.newProperty",
            defaultValue: "New property",
            comment: "Toolbar button to open create Property sheet"
        )
        static func addPropertyPlaceholder(typeLabel: String) -> LocalizedStringResource {
            LocalizedStringResource(
                "subjectFields.toolbar.addPropertyPlaceholder",
                defaultValue: "Add a property to \(typeLabel)",
                comment: "ComboBox placeholder when a subject type is focused; argument is type label"
            )
        }
        static let addPropertyEmpty = LocalizedStringResource(
            "subjectFields.toolbar.addPropertyEmpty",
            defaultValue: "No unbound property matches that name",
            comment: "ComboBox empty state when binding an existing property to the focused type"
        )
        static let inspectorAccessibility = LocalizedStringResource(
            "subjectFields.inspector.accessibility",
            defaultValue: "Property inspector",
            comment: "Accessibility label for the property inspector card"
        )
        static let inspectorEmpty = LocalizedStringResource(
            "subjectFields.inspector.empty",
            defaultValue: "Select a property to see its description and bindings.",
            comment: "Inspector empty state"
        )
        static let inspectorValueType = LocalizedStringResource(
            "subjectFields.inspector.valueType",
            defaultValue: "Value type",
            comment: "Inspector meta label: value type"
        )
        static let valueTypeImmutable = LocalizedStringResource(
            "subjectFields.inspector.valueTypeImmutable",
            defaultValue: "Immutable after create",
            comment: "Accessibility label for the lock beside value type in the inspector"
        )
        static let inspectorOrigin = LocalizedStringResource(
            "subjectFields.inspector.origin",
            defaultValue: "Origin",
            comment: "Inspector meta label: origin"
        )
        static let inspectorOriginUser = LocalizedStringResource(
            "subjectFields.inspector.originUser",
            defaultValue: "User — you created this",
            comment: "Inspector origin line for researcher-created properties"
        )
        static let inspectorOriginSeeded = LocalizedStringResource(
            "subjectFields.inspector.originSeeded",
            defaultValue: "Seeded by Provenencia",
            comment: "Inspector origin line for product-seeded properties"
        )
        static let inspectorValuesRecorded = LocalizedStringResource(
            "subjectFields.inspector.valuesRecorded",
            defaultValue: "Values recorded",
            comment: "Inspector meta label: use / binding count from engine"
        )
        static func valuesRecordedCount(count: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "subjectFields.inspector.valuesRecordedCount",
                defaultValue: "%lld",
                comment: "Inspector values-recorded count"
            ))
            return String(format: format, locale: .current, count)
        }
        static let bindingsSection = LocalizedStringResource(
            "subjectFields.inspector.bindings",
            defaultValue: "Bound to",
            comment: "Inspector section label for bindings list"
        )
        static let bindingsAccessibility = LocalizedStringResource(
            "subjectFields.inspector.bindingsAccessibility",
            defaultValue: "Subject type bindings",
            comment: "Accessibility label for the Bound-to checkbox list"
        )
        static func bindCount(bound: Int, total: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "subjectFields.inspector.bindCount",
                defaultValue: "%lld of %lld",
                comment: "Inspector bound-type count beside Bound to; bound, then total types"
            ))
            return String(format: format, locale: .current, bound, total)
        }
        static let termNote = LocalizedStringResource(
            "subjectFields.inspector.termNote",
            defaultValue: "Values come from the product vocabulary. You choose one when citing this property.",
            comment: "Inspector note for term-typed seeded properties"
        )
        static let deleteProperty = LocalizedStringResource(
            "subjectFields.inspector.delete",
            defaultValue: "Delete property",
            comment: "Delete button in property inspector"
        )
        static let deleteInUse = LocalizedStringResource(
            "subjectFields.inspector.deleteInUse",
            defaultValue: "This property is still bound to one or more subject types.",
            comment: "Why delete is disabled when usedBy > 0"
        )
        static let deleteSeeded = LocalizedStringResource(
            "subjectFields.inspector.deleteSeeded",
            defaultValue: "Seeded properties cannot be deleted.",
            comment: "Why delete is disabled for provenencia-origin properties"
        )
        static let deleteUnused = LocalizedStringResource(
            "subjectFields.inspector.deleteUnused",
            defaultValue: "Not in use. Deleting removes it from every type it is bound to.",
            comment: "Inspector note when a user property can be deleted"
        )
        static func lockedBindingReason(typeLabel: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "subjectFields.inspector.lockedBinding",
                defaultValue: "The Interpretation subject registry requires this property on %@. The binding cannot be removed.",
                comment: "Callout when activating a locked binding; argument is subject type label"
            ))
            return String(format: format, locale: .current, typeLabel)
        }
        static let bindingLocked = LocalizedStringResource(
            "subjectFields.inspector.bindingLocked",
            defaultValue: "Locked",
            comment: "Accessibility / badge for a locked subject-type binding"
        )
        static let bindingBound = LocalizedStringResource(
            "subjectFields.inspector.bindingBound",
            defaultValue: "Bound",
            comment: "Accessibility label for an active unbound-able binding"
        )
        static let bindingNotBound = LocalizedStringResource(
            "subjectFields.inspector.bindingNotBound",
            defaultValue: "not bound",
            comment: "Accessibility label for an inactive binding checkbox"
        )
        static let bindingRegistry = LocalizedStringResource(
            "subjectFields.inspector.bindingRegistry",
            defaultValue: "registry",
            comment: "Micro-label beside a locked Bound-to row (registry-held)"
        )
        static let createTitle = LocalizedStringResource(
            "subjectFields.create.title",
            defaultValue: "New property",
            comment: "Create property sheet title"
        )
        static let createOriginNote = LocalizedStringResource(
            "subjectFields.create.originNote",
            defaultValue: "Origin is recorded as user — seeded properties come from Provenencia",
            comment: "Create property sheet note about origin"
        )
        static let createBindSection = LocalizedStringResource(
            "subjectFields.create.bindSection",
            defaultValue: "Bind to subject types",
            comment: "Create property sheet: bind checklist section"
        )
        static let createLabel = LocalizedStringResource(
            "subjectFields.create.label",
            defaultValue: "Label",
            comment: "Create property form: label field"
        )
        static let createLabelHint = LocalizedStringResource(
            "subjectFields.create.labelHint",
            defaultValue: "What a researcher sees on the subject",
            comment: "Hint under create property label"
        )
        static let createLabelPlaceholder = LocalizedStringResource(
            "subjectFields.create.labelPlaceholder",
            defaultValue: "Burial ground",
            comment: "Placeholder for create property label field"
        )
        static let createKey = LocalizedStringResource(
            "subjectFields.create.key",
            defaultValue: "Key",
            comment: "Create property form: machine key field"
        )
        static let createKeyHint = LocalizedStringResource(
            "subjectFields.create.keyHint",
            defaultValue: "Generated from the label",
            comment: "Hint under create property key; key is minted server-side from label"
        )
        static let createKeyPlaceholder = LocalizedStringResource(
            "subjectFields.create.keyPlaceholder",
            defaultValue: "burial-ground",
            comment: "Placeholder for create property key preview"
        )
        static let createValueType = LocalizedStringResource(
            "subjectFields.create.valueType",
            defaultValue: "Value type",
            comment: "Create property form: value type picker"
        )
        static let createValueTypeHint = LocalizedStringResource(
            "subjectFields.create.valueTypeHint",
            defaultValue: "Cannot be changed once the property exists",
            comment: "Hint under create property value type chips"
        )
        static let createDescription = LocalizedStringResource(
            "subjectFields.create.description",
            defaultValue: "Description",
            comment: "Create property form: description"
        )
        static let createDescriptionPlaceholder = LocalizedStringResource(
            "subjectFields.create.descriptionPlaceholder",
            defaultValue: "How the value should be read from the record",
            comment: "Placeholder for create property description"
        )
        static let createCancel = LocalizedStringResource(
            "subjectFields.create.cancel",
            defaultValue: "Cancel",
            comment: "Create property sheet cancel"
        )
        static let createSubmit = LocalizedStringResource(
            "subjectFields.create.submit",
            defaultValue: "Create property",
            comment: "Create property sheet primary action"
        )
        static let errorLabelRequired = LocalizedStringResource(
            "subjectFields.create.errorLabelRequired",
            defaultValue: "Enter a label for this property.",
            comment: "Validation when create label is empty"
        )
        static let errorValueType = LocalizedStringResource(
            "subjectFields.create.errorValueType",
            defaultValue: "Choose a researcher value type. Term properties are product vocabulary only.",
            comment: "Validation when create value type is invalid"
        )
        static let toastCreatedTitle = LocalizedStringResource(
            "subjectFields.toast.createdTitle",
            defaultValue: "Property created",
            comment: "Success toast title after creating a property"
        )
        static func toastCreatedBody(label: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "subjectFields.toast.createdBody",
                defaultValue: "%@ is ready to bind.",
                comment: "Success toast body after creating a property; argument is label"
            ))
            return String(format: format, locale: .current, label)
        }
        static let toastDeletedTitle = LocalizedStringResource(
            "subjectFields.toast.deletedTitle",
            defaultValue: "Property deleted",
            comment: "Success toast title after deleting a property"
        )
        static func toastDeletedBody(label: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "subjectFields.toast.deletedBody",
                defaultValue: "%@ was removed from this project.",
                comment: "Success toast body after deleting a property; argument is label"
            ))
            return String(format: format, locale: .current, label)
        }
        static let deleteConfirmTitle = LocalizedStringResource(
            "subjectFields.delete.confirmTitle",
            defaultValue: "Delete this property?",
            comment: "Confirm sheet title for deleting a user property"
        )
        static let deleteConfirmMessage = LocalizedStringResource(
            "subjectFields.delete.confirmMessage",
            defaultValue: "This removes the property definition. It cannot be undone.",
            comment: "Confirm sheet message for deleting a user property"
        )
        static let deleteKeep = LocalizedStringResource(
            "subjectFields.delete.keep",
            defaultValue: "Keep",
            comment: "Confirm sheet cancel for delete property"
        )
    }

    /// Maps stable Go/FFI error codes to localized user-facing copy.
    enum SourceTypes {
        static let description = LocalizedStringResource(
            "sourceTypes.list.description",
            defaultValue: "The kinds of record this project cites, and the fields each kind usually carries. The fields are suggestions — a source of this type may leave any of them blank.",
            comment: "Explanatory copy under the Source types page title"
        )

        static func countLine(total: Int, seeded: Int, user: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceTypes.list.countLine",
                defaultValue: "%lld types · %lld seeded · %lld yours",
                comment: "Source types count summary; arguments are total, seeded (provenencia), and user type counts"
            ))
            return String(format: format, locale: .current, total, seeded, user)
        }

        static func countLineWithPlugin(total: Int, seeded: Int, user: Int, plugin: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceTypes.list.countLineWithPlugin",
                defaultValue: "%lld types · %lld seeded · %lld yours · %lld plugin",
                comment: "Source types count summary including plugin-origin types; arguments are total, seeded, user, and plugin type counts"
            ))
            return String(format: format, locale: .current, total, seeded, user, plugin)
        }

        static let addType = LocalizedStringResource(
            "sourceTypes.list.addType",
            defaultValue: "Add type",
            comment: "Button: add a new Source type (toolbar, empty state, and add-form submit)"
        )

        static let columnLabel = LocalizedStringResource(
            "sourceTypes.list.columnLabel",
            defaultValue: "Label",
            comment: "Source types list column header: label"
        )

        static let columnKey = LocalizedStringResource(
            "sourceTypes.list.columnKey",
            defaultValue: "Key",
            comment: "Source types list column header: key"
        )

        static let columnFields = LocalizedStringResource(
            "sourceTypes.list.columnFields",
            defaultValue: "Associated fields",
            comment: "Source types list column header: how many metadata fields the type suggests"
        )

        static let fieldCountNone = LocalizedStringResource(
            "sourceTypes.list.fieldCountNone",
            defaultValue: "—",
            comment: "Shown in the associated-fields column when a type suggests no fields"
        )

        static func resultLine(shown: Int, total: Int) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceTypes.list.resultLineAll",
                defaultValue: "%lld types",
                comment: "Footer result count for the Source types list; argument is the total"
            ))
            return String(format: format, locale: .current, total)
        }

        static let emptyProjectTitle = LocalizedStringResource(
            "sourceTypes.emptyProject.title",
            defaultValue: "No source types yet",
            comment: "Title of the empty state when the project has zero source types"
        )

        static let emptyProjectBody = LocalizedStringResource(
            "sourceTypes.emptyProject.body",
            defaultValue: "This project has no record classes to cite against. Add the kinds of record you actually hold — a parish register, a scrapbook, a headstone photograph.",
            comment: "Body of the empty state when the project has zero source types"
        )

        static let detailEyebrowType = LocalizedStringResource(
            "sourceTypes.detail.eyebrowType",
            defaultValue: "Source type",
            comment: "Eyebrow label above an existing type's detail panel"
        )

        static let detailEyebrowNewType = LocalizedStringResource(
            "sourceTypes.detail.eyebrowNewType",
            defaultValue: "New type",
            comment: "Eyebrow label above the add-type panel"
        )

        static let keyHintAdd = LocalizedStringResource(
            "sourceTypes.detail.keyHintAdd",
            defaultValue: "Provenencia mints the key from the label when the type is added",
            comment: "Hint under the live key preview while adding a type"
        )

        static let keyHintEdit = LocalizedStringResource(
            "sourceTypes.detail.keyHintEdit",
            defaultValue: "The key is minted once from the label and never changes — renaming the type keeps existing sources attached",
            comment: "Hint under the key on an existing type's detail panel"
        )

        static func usage(count: Int) -> LocalizedStringResource {
            switch count {
            case 0: usageNone
            case 1: usageOne
            default: usageOther(count: count)
            }
        }

        private static let usageNone = LocalizedStringResource(
            "sourceTypes.detail.usageNone",
            defaultValue: "no sources yet",
            comment: "Line under a type's title when no source is classified as it"
        )

        private static let usageOne = LocalizedStringResource(
            "sourceTypes.detail.usageOne",
            defaultValue: "in use on 1 source",
            comment: "Line under a type's title when exactly one source is classified as it"
        )

        private static func usageOther(count: Int) -> LocalizedStringResource {
            LocalizedStringResource(
                "sourceTypes.detail.usageOther",
                defaultValue: "in use on \(count) sources",
                comment: "Line under a type's title; argument is how many sources are classified as it"
            )
        }

        static let panelEmptyTitle = LocalizedStringResource(
            "sourceTypes.detail.panelEmptyTitle",
            defaultValue: "No type selected",
            comment: "Title of the empty state shown in the detail panel before any type is selected"
        )

        static let panelEmptyBody = LocalizedStringResource(
            "sourceTypes.detail.panelEmptyBody",
            defaultValue: "Select a type to read its description and the fields it suggests. Types seeded by Provenencia are yours to edit; only plugin-owned types are fixed.",
            comment: "Body of the empty state shown in the detail panel before any type is selected"
        )

        static func lockedNotePlugin(pluginID: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceTypes.detail.lockedNotePlugin",
                defaultValue: "Supplied by the %@ plugin. The plugin owns this type, its description and the fields it suggests — Provenencia will not edit or delete them.",
                comment: "Callout explaining why a plugin-origin type can't be edited; argument is the plugin id"
            ))
            return String(format: format, locale: .current, pluginID)
        }

        static let descriptionSectionLabel = LocalizedStringResource(
            "sourceTypes.detail.descriptionSectionLabel",
            defaultValue: "Description",
            comment: "Section label above the read-only description on a locked type's detail"
        )

        static let descriptionEmptyPlaceholder = LocalizedStringResource(
            "sourceTypes.detail.descriptionEmptyPlaceholder",
            defaultValue: "—",
            comment: "Shown in place of a locked type's description when it has none"
        )

        static let formLabel = LocalizedStringResource(
            "sourceTypes.form.label",
            defaultValue: "Label",
            comment: "Add/edit form field: label"
        )

        static let formLabelPlaceholder = LocalizedStringResource(
            "sourceTypes.form.labelPlaceholder",
            defaultValue: "Parish register",
            comment: "Placeholder text for the add-type label input"
        )

        static let formDescription = LocalizedStringResource(
            "sourceTypes.form.description",
            defaultValue: "Description",
            comment: "Add/edit form field: description"
        )

        static let formDescriptionHint = LocalizedStringResource(
            "sourceTypes.form.descriptionHint",
            defaultValue: "What kind of record belongs to this type, in your own words",
            comment: "Hint under the description field"
        )

        static let formDescriptionPlaceholder = LocalizedStringResource(
            "sourceTypes.form.descriptionPlaceholder",
            defaultValue: "A bound register of baptisms, marriages or burials kept by a parish",
            comment: "Placeholder text for the add-type description input"
        )

        static let formIcon = LocalizedStringResource(
            "sourceTypes.form.icon",
            defaultValue: "Icon",
            comment: "Add/edit form field: evidence icon for this type"
        )

        static let formIconHint = LocalizedStringResource(
            "sourceTypes.form.iconHint",
            defaultValue: "The mark that stands for this type wherever a source of it is listed",
            comment: "Hint under the source-type icon field"
        )

        static let formIconChange = LocalizedStringResource(
            "sourceTypes.form.iconChange",
            defaultValue: "Change",
            comment: "Link-styled control that opens the icon picker dialog"
        )

        static func formIconChangeAccessibility(name: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceTypes.form.iconChangeAccessibility",
                defaultValue: "Icon: %@ — choose a different one",
                comment: "Accessibility label for the icon field button; argument is the current mark name"
            ))
            return String(format: format, locale: .current, name)
        }

        static let iconPickerTitle = LocalizedStringResource(
            "sourceTypes.iconPicker.title",
            defaultValue: "Choose an icon",
            comment: "Title of the Source-type icon picker dialog"
        )

        static let iconPickerSubtitle = LocalizedStringResource(
            "sourceTypes.iconPicker.subtitle",
            defaultValue: "The mark that stands for this type wherever a source of it is listed. Twenty-one marks name a kind of record, not a file format.",
            comment: "Subtitle of the Source-type icon picker dialog"
        )

        static let iconPickerConfirm = LocalizedStringResource(
            "sourceTypes.iconPicker.confirm",
            defaultValue: "Use icon",
            comment: "Commits the selected icon in the Source-type icon picker form dialog"
        )

        static let iconPickerCancel = LocalizedStringResource(
            "sourceTypes.iconPicker.cancel",
            defaultValue: "Keep current",
            comment: "Dismisses the Source-type icon picker without changing the draft icon"
        )

        static let iconPickerGroupLabel = LocalizedStringResource(
            "sourceTypes.iconPicker.group",
            defaultValue: "Icon for this source type",
            comment: "Accessibility label for the icon picker radio group"
        )

        static let errorLabelRequired = LocalizedStringResource(
            "sourceTypes.form.errorLabelRequired",
            defaultValue: "A label is required — it is how the type reads on a source.",
            comment: "Inline validation error when the label is blank"
        )

        static let errorUnslugifiable = LocalizedStringResource(
            "sourceTypes.form.errorUnslugifiable",
            defaultValue: "That label cannot be turned into a key. Use at least one letter or number.",
            comment: "Inline validation error when the label has no letters or digits to slug"
        )

        static let saveSaving = LocalizedStringResource(
            "sourceTypes.form.saveSaving",
            defaultValue: "Saving",
            comment: "Primary button label while a Source type add/edit is in flight"
        )

        static let saveChanges = LocalizedStringResource(
            "sourceTypes.form.saveChanges",
            defaultValue: "Save changes",
            comment: "Primary button label for committing an edit to an existing type"
        )

        static let cancel = LocalizedStringResource(
            "sourceTypes.form.cancel",
            defaultValue: "Cancel",
            comment: "Secondary button label that dismisses the add-type form"
        )

        static let revert = LocalizedStringResource(
            "sourceTypes.form.revert",
            defaultValue: "Revert",
            comment: "Secondary button label that discards unsaved edits to an existing type"
        )

        static let addSuggestionsNote = LocalizedStringResource(
            "sourceTypes.form.addSuggestionsNote",
            defaultValue: "Suggested fields are assigned after the type is saved.",
            comment: "Callout in the add-type form explaining that associations come later"
        )

        static let suggestedSectionLabel = LocalizedStringResource(
            "sourceTypes.suggested.sectionLabel",
            defaultValue: "Suggested fields",
            comment: "Section label above the fields a type suggests"
        )

        static let suggestedHint = LocalizedStringResource(
            "sourceTypes.suggested.hint",
            defaultValue: "Suggestions, not a schema — a source of this type may leave any of them blank",
            comment: "Hint under the suggested fields section label"
        )

        static func assignedCount(count: Int) -> LocalizedStringResource {
            count == 1 ? assignedCountOne : assignedCountOther(count: count)
        }

        private static let assignedCountOne = LocalizedStringResource(
            "sourceTypes.suggested.countOne",
            defaultValue: "1 field",
            comment: "Count beside the suggested fields section label when the type suggests exactly one field"
        )

        private static func assignedCountOther(count: Int) -> LocalizedStringResource {
            LocalizedStringResource(
                "sourceTypes.suggested.countOther",
                defaultValue: "\(count) fields",
                comment: "Count beside the suggested fields section label; argument is how many fields the type suggests"
            )
        }

        static let noAssignedBody = LocalizedStringResource(
            "sourceTypes.suggested.noneBody",
            defaultValue: "No suggested fields yet — a source of this type will offer nothing but the standard citation. Assign the fields these records usually carry.",
            comment: "Body of the panel shown when a type suggests no fields yet"
        )

        static let assignPlaceholder = LocalizedStringResource(
            "sourceTypes.suggested.assignPlaceholder",
            defaultValue: "Field label or key",
            comment: "Placeholder in the assign-field combo box, naming both things it searches"
        )

        static let assignNoMatch = LocalizedStringResource(
            "sourceTypes.suggested.assignNoMatch",
            defaultValue: "No field in the vocabulary matches that — add it in Source fields first",
            comment: "Shown inside the assign-field combo box list when the typed query matches no field"
        )

        static let assignFieldLabel = LocalizedStringResource(
            "sourceTypes.suggested.assignFieldLabel",
            defaultValue: "Source field to assign",
            comment: "Accessibility label for the assign-field combo box"
        )

        static let assignField = LocalizedStringResource(
            "sourceTypes.suggested.assignField",
            defaultValue: "Assign field",
            comment: "Spoken label for the assign button when no field is picked yet"
        )

        static func assignFieldNamed(field: String, type: String) -> LocalizedStringResource {
            LocalizedStringResource(
                "sourceTypes.suggested.assignFieldNamed",
                defaultValue: "Assign field \(field) to \(type)",
                comment: "Spoken label for the assign button; arguments are the picked field label then the type label"
            )
        }

        static let assignTipPoolEmpty = LocalizedStringResource(
            "sourceTypes.suggested.assignTipPoolEmpty",
            defaultValue: "Every field is already assigned",
            comment: "Tooltip on the disabled assign button when the type already suggests the whole vocabulary"
        )

        static let assignTipChoose = LocalizedStringResource(
            "sourceTypes.suggested.assignTipChoose",
            defaultValue: "Choose a field to assign",
            comment: "Tooltip on the disabled assign button before a field is picked"
        )

        static func assignTipField(label: String) -> LocalizedStringResource {
            LocalizedStringResource(
                "sourceTypes.suggested.assignTipField",
                defaultValue: "Assign \(label) to this type",
                comment: "Tooltip on the enabled assign button; argument is the picked field label"
            )
        }

        static let poolHint = LocalizedStringResource(
            "sourceTypes.suggested.poolHint",
            defaultValue: "The pool is the Source fields vocabulary — add a new field there first if it is missing",
            comment: "Hint under the assign-field picker naming where the pool comes from"
        )

        static let poolHintEmpty = LocalizedStringResource(
            "sourceTypes.suggested.poolHintEmpty",
            defaultValue: "Every field in the vocabulary is already suggested for this type",
            comment: "Hint under the assign-field picker when nothing is left to assign"
        )

        static func removeSuggestion(label: String) -> LocalizedStringResource {
            LocalizedStringResource(
                "sourceTypes.suggested.remove",
                defaultValue: "Remove \(label) from this type",
                comment: "Accessibility label and tooltip on the control that detaches one suggested field; argument is the field label"
            )
        }

        static let toastAssignedTitle = LocalizedStringResource(
            "sourceTypes.toast.assignedTitle",
            defaultValue: "Field assigned",
            comment: "Toast title after attaching a field to a type"
        )

        static func toastAssignedBody(field: String, type: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceTypes.toast.assignedBody",
                defaultValue: "%1$@ is now suggested for %2$@.",
                comment: "Toast body after attaching a field to a type; arguments are the field label then the type label"
            ))
            return String(format: format, locale: .current, field, type)
        }

        static let toastRemovedTitle = LocalizedStringResource(
            "sourceTypes.toast.removedTitle",
            defaultValue: "Suggestion removed",
            comment: "Toast title after detaching a field from a type"
        )

        /// The reassurance that carries T-20: detaching the join deletes
        /// neither the field nor the values sources already hold for it.
        static func toastRemovedBody(field: String, type: String, valueCount: Int) -> String {
            if valueCount == 0 {
                return String(format: String(localized: removedBodyNoValues), locale: .current, field, type)
            }
            return String(format: String(localized: removedBodyWithValues), locale: .current, field, type, valueCount)
        }

        private static let removedBodyNoValues = LocalizedStringResource(
            "sourceTypes.toast.removedBodyNoValues",
            defaultValue: "%1$@ is no longer suggested for %2$@. The field stays in this project’s vocabulary.",
            comment: "Toast body after detaching a field no source carries a value for; arguments are the field label then the type label"
        )

        private static let removedBodyWithValues = LocalizedStringResource(
            "sourceTypes.toast.removedBodyWithValues",
            defaultValue: "%1$@ is no longer suggested for %2$@. The field stays in this project’s vocabulary, and the %3$lld sources already carrying a value keep it.",
            comment: "Toast body after detaching a field sources still carry values for; arguments are the field label, the type label, then how many sources hold a value"
        )

        static let toastAddedTitle = LocalizedStringResource(
            "sourceTypes.toast.addedTitle",
            defaultValue: "Type added",
            comment: "Success toast title after creating a Source type"
        )

        static func toastAddedBody(label: String, key: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceTypes.toast.addedBody",
                defaultValue: "%1$@ is in this project’s vocabulary as %2$@. Assign the fields it should suggest.",
                comment: "Success toast body after creating a Source type; arguments are label then minted key"
            ))
            return String(format: format, locale: .current, label, key)
        }

        static let toastUpdatedTitle = LocalizedStringResource(
            "sourceTypes.toast.updatedTitle",
            defaultValue: "Type updated",
            comment: "Success toast title after editing a Source type"
        )

        static func toastUpdatedBody(label: String, key: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceTypes.toast.updatedBody",
                defaultValue: "%1$@ — the key stays %2$@.",
                comment: "Success toast body after editing a Source type; arguments are label then key"
            ))
            return String(format: format, locale: .current, label, key)
        }

        static let deleteType = LocalizedStringResource(
            "sourceTypes.delete.action",
            defaultValue: "Delete type",
            comment: "Delete button in the Source types detail pane, and the confirm dialog's destructive button"
        )

        static let deleteOwnedByPlugin = LocalizedStringResource(
            "sourceTypes.delete.ownedByPlugin",
            defaultValue: "Owned by the plugin",
            comment: "Tooltip on the disabled delete button when the selected type comes from a plugin"
        )

        static func deleteInUse(count: Int) -> LocalizedStringResource {
            count == 1 ? deleteInUseOne : deleteInUseOther(count: count)
        }

        private static let deleteInUseOne = LocalizedStringResource(
            "sourceTypes.delete.inUseOne",
            defaultValue: "In use on 1 source",
            comment: "Tooltip on the disabled delete button when exactly one source is classified as the type"
        )

        private static func deleteInUseOther(count: Int) -> LocalizedStringResource {
            LocalizedStringResource(
                "sourceTypes.delete.inUseOther",
                defaultValue: "In use on \(count) sources",
                comment: "Tooltip on the disabled delete button; argument is how many sources are classified as the type"
            )
        }

        static func deleteConfirmTitle(label: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceTypes.delete.confirmTitle",
                defaultValue: "Delete %@?",
                comment: "Title of the delete-type confirmation dialog; argument is the type label"
            ))
            return String(format: format, locale: .current, label)
        }

        static let deleteConfirmMessage = LocalizedStringResource(
            "sourceTypes.delete.confirmMessage",
            defaultValue: "No source in this project is classified as this type, so no citation loses its class. The field suggestions attached to it go with it; the fields themselves stay in the vocabulary.",
            comment: "Message of the delete-type confirmation sheet: what does and does not cascade"
        )

        static let deleteKeyReleased = LocalizedStringResource(
            "sourceTypes.delete.keyReleased",
            defaultValue: "Key released",
            comment: "Micro-caps label beside the key a type delete releases, in the confirmation sheet"
        )

        static let deleteKeep = LocalizedStringResource(
            "sourceTypes.delete.keep",
            defaultValue: "Keep type",
            comment: "Button that closes the delete-type confirmation without deleting"
        )

        static let toastDeletedTitle = LocalizedStringResource(
            "sourceTypes.toast.deletedTitle",
            defaultValue: "Type deleted",
            comment: "Toast title after a source type is deleted"
        )

        static func toastDeletedBody(label: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "sourceTypes.toast.deletedBody",
                defaultValue: "%@ is no longer in this project’s vocabulary. Its field suggestions went with it; the fields did not.",
                comment: "Toast body after a source type is deleted; argument is the type label"
            ))
            return String(format: format, locale: .current, label)
        }
    }

    enum Errors {
        static let catalogAlreadyExists = LocalizedStringResource(
            "error.catalog.already_exists",
            defaultValue: "Project already exists.",
            comment: "FFI error catalog.already_exists"
        )
        static let catalogAlreadyOpen = LocalizedStringResource(
            "error.catalog.already_open",
            defaultValue: "Catalog session conflict — something opened the project database outside the app session.",
            comment: "FFI error catalog.already_open; bypass of held catalogsession"
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
        static let catalogSchemaMismatch = LocalizedStringResource(
            "error.catalog.schema_mismatch",
            defaultValue: "This catalog’s schema does not match this version of Provenencia.",
            comment: "FFI error catalog.schema_mismatch"
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
        static let auditInvalid = LocalizedStringResource(
            "error.audit.invalid",
            defaultValue: "Invalid audit record.",
            comment: "FFI error audit.invalid"
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
        static let sourcesInvalid = LocalizedStringResource(
            "error.sources.invalid",
            defaultValue: "Invalid source.",
            comment: "FFI error sources.invalid"
        )
        static let subjectsInvalid = LocalizedStringResource(
            "error.subjects.invalid",
            defaultValue: "Invalid subject.",
            comment: "FFI error subjects.invalid"
        )
        static let subjectsInUse = LocalizedStringResource(
            "error.subjects.in_use",
            defaultValue: "This subject still has citations or is used as a property value.",
            comment: "FFI error subjects.in_use when Observations still reference the subject"
        )
        static let subjectPositionsInvalid = LocalizedStringResource(
            "error.subjectpositions.invalid",
            defaultValue: "Invalid subject position.",
            comment: "FFI error subjectpositions.invalid"
        )
        static let subjectTypesInvalid = LocalizedStringResource(
            "error.subjecttypes.invalid",
            defaultValue: "Invalid subject type.",
            comment: "FFI error subjecttypes.invalid"
        )
        static let subjectTypesDuplicatePrefix = LocalizedStringResource(
            "error.subjecttypes.duplicate_prefix",
            defaultValue: "That subject-type prefix is already in use.",
            comment: "FFI error subjecttypes.duplicate_prefix"
        )
        static let artifactsInvalid = LocalizedStringResource(
            "error.artifacts.invalid",
            defaultValue: "Invalid artifact.",
            comment: "FFI error artifacts.invalid"
        )
        static let artifactsFileAlreadyAttached = LocalizedStringResource(
            "error.artifacts.file_already_attached",
            defaultValue: "This artifact already has a file. Add a new artifact for a better scan.",
            comment: "FFI error artifacts.file_already_attached"
        )
        static let sourceCredibilityInvalid = LocalizedStringResource(
            "error.sourcecredibility.invalid",
            defaultValue: "Invalid source credibility assessment.",
            comment: "FFI error sourcecredibility.invalid"
        )
        static let sourceMetadataInvalid = LocalizedStringResource(
            "error.sourcemetadata.invalid",
            defaultValue: "Invalid source metadata.",
            comment: "FFI error sourcemetadata.invalid"
        )
        static let filesInvalid = LocalizedStringResource(
            "error.files.invalid",
            defaultValue: "Invalid file.",
            comment: "FFI error files.invalid"
        )
        static let ingestInvalid = LocalizedStringResource(
            "error.ingest.invalid",
            defaultValue: "Could not ingest that file.",
            comment: "FFI error ingest.invalid"
        )
        static let ingestPermissionDeniedTitle = LocalizedStringResource(
            "error.ingest.permission_denied.title",
            defaultValue: "The file could not be read.",
            comment: "Callout title for ingest.permission_denied"
        )
        static let ingestPermissionDeniedHelp = LocalizedStringResource(
            "error.ingest.permission_denied.help",
            defaultValue: "Check that the volume is mounted and that Provenencia has access to the folder, then try again.",
            comment: "Callout help for ingest.permission_denied"
        )
        static let ingestPermissionDenied = LocalizedStringResource(
            "error.ingest.permission_denied",
            defaultValue: "Permission denied reading that file.",
            comment: "FFI error ingest.permission_denied (banner fallback)"
        )
        static let ingestUnsupportedOfficeTitle = LocalizedStringResource(
            "error.ingest.unsupported_office.title",
            defaultValue: "Spreadsheets and presentations aren’t accepted as artifact files",
            comment: "Callout title for ingest.unsupported_office"
        )
        static let ingestUnsupportedOfficeHelp = LocalizedStringResource(
            "error.ingest.unsupported_office.help",
            defaultValue: "Word documents are fine — Excel, PowerPoint, and similar apps are not. Export to PDF or attach a Word file, scan, or image instead (up to 512 MB).",
            comment: "Callout help for ingest.unsupported_office"
        )
        static let ingestUnsupportedArchiveTitle = LocalizedStringResource(
            "error.ingest.unsupported_archive.title",
            defaultValue: "Archives aren’t accepted as artifact files",
            comment: "Callout title for ingest.unsupported_archive"
        )
        static let ingestUnsupportedArchiveHelp = LocalizedStringResource(
            "error.ingest.unsupported_archive.help",
            defaultValue: "Unzip the archive and attach the scan, PDF, or recording inside. Artifacts hold images, PDFs, Word documents, plain text (including CSV and Markdown), audio, or video — up to 512 MB.",
            comment: "Callout help for ingest.unsupported_archive"
        )
        static let ingestUnsupportedExecutableTitle = LocalizedStringResource(
            "error.ingest.unsupported_executable.title",
            defaultValue: "Programs and installers can’t be attached",
            comment: "Callout title for ingest.unsupported_executable"
        )
        static let ingestUnsupportedExecutableHelp = LocalizedStringResource(
            "error.ingest.unsupported_executable.help",
            defaultValue: "Attach the evidence file itself — a scan, PDF, transcription, or recording — not an application.",
            comment: "Callout help for ingest.unsupported_executable"
        )
        static let ingestUnsupportedTypeHelp = LocalizedStringResource(
            "error.ingest.unsupported_type.help",
            defaultValue: "Artifacts hold images, PDFs, Word documents, plain text (including CSV and Markdown), audio, or video — up to 512 MB. Choose a different file.",
            comment: "Callout help for ingest.unsupported_type"
        )
        static let ingestUnsupportedTypeGenericTitle = LocalizedStringResource(
            "error.ingest.unsupported_type.genericTitle",
            defaultValue: "That file type isn’t accepted",
            comment: "Callout title when unsupported type has no short label"
        )
        static let ingestUnidentifiedTitle = LocalizedStringResource(
            "error.ingest.unidentified.title",
            defaultValue: "We couldn’t identify that file",
            comment: "Callout title for ingest.unidentified"
        )
        static let ingestUnidentifiedHelp = LocalizedStringResource(
            "error.ingest.unidentified.help",
            defaultValue: "Artifacts hold images, PDFs, Word documents, plain text (including CSV and Markdown), audio, or video — up to 512 MB. Try a clearer export from the original app.",
            comment: "Callout help for ingest.unidentified"
        )
        static let ingestEmptyTitle = LocalizedStringResource(
            "error.ingest.empty.title",
            defaultValue: "That file is empty",
            comment: "Callout title for ingest.empty"
        )
        static let ingestEmptyHelp = LocalizedStringResource(
            "error.ingest.empty.help",
            defaultValue: "Choose a real scan, PDF, or recording with content.",
            comment: "Callout help for ingest.empty"
        )
        static let ingestTooLargeTitle = LocalizedStringResource(
            "error.ingest.too_large.title",
            defaultValue: "That file is over the 512 MB cap",
            comment: "Callout title for ingest.too_large"
        )
        static let ingestTooLargeHelp = LocalizedStringResource(
            "error.ingest.too_large.help",
            defaultValue: "This file is %@. Downsample or export a smaller copy, then choose that file.",
            comment: "Callout help for ingest.too_large; argument is human-readable size"
        )
        static let ingestNotAFileTitle = LocalizedStringResource(
            "error.ingest.not_a_file.title",
            defaultValue: "That isn’t a regular file",
            comment: "Callout title for ingest.not_a_file"
        )
        static let ingestNotAFileHelp = LocalizedStringResource(
            "error.ingest.not_a_file.help",
            defaultValue: "Choose a single file — not a folder or special device.",
            comment: "Callout help for ingest.not_a_file"
        )
        static let ingestSymlinkTitle = LocalizedStringResource(
            "error.ingest.symlink.title",
            defaultValue: "Symbolic links can’t be ingested",
            comment: "Callout title for ingest.symlink"
        )
        static let ingestSymlinkHelp = LocalizedStringResource(
            "error.ingest.symlink.help",
            defaultValue: "Choose the real file the link points to.",
            comment: "Callout help for ingest.symlink"
        )
        static let ingestMissingTitle = LocalizedStringResource(
            "error.ingest.missing.title",
            defaultValue: "That file is no longer available",
            comment: "Callout title for ingest.missing"
        )
        static let ingestMissingHelp = LocalizedStringResource(
            "error.ingest.missing.help",
            defaultValue: "It may have moved or the volume was unmounted. Choose the file again.",
            comment: "Callout help for ingest.missing"
        )
        static let ingestMultiFileTitle = LocalizedStringResource(
            "error.ingest.multi_file.title",
            defaultValue: "Drop a single file",
            comment: "Callout title when multiple files are dropped"
        )
        static let ingestMultiFileHelp = LocalizedStringResource(
            "error.ingest.multi_file.help",
            defaultValue: "Artifacts attach one file at a time. Drop or choose a single scan, PDF, or recording.",
            comment: "Callout help when multiple files are dropped"
        )
        static let ingestUnreadableTitle = LocalizedStringResource(
            "error.ingest.unreadable.title",
            defaultValue: "The file could not be read.",
            comment: "Callout title when a dropped/picked URL cannot be resolved"
        )
        static let ingestUnreadableHelp = LocalizedStringResource(
            "error.ingest.unreadable.help",
            defaultValue: "Check that the volume is mounted and that Provenencia has access to the folder, then try again.",
            comment: "Callout help when a dropped/picked URL cannot be resolved"
        )

        /// Title + body for ingest reject Callouts (local precheck or FFI).
        struct IngestCallout: Equatable {
            var title: LocalizedStringResource
            var message: String
        }

        static func ingestCallout(
            reason: IngestMediaPolicy.Reason,
            sizeLabel: String? = nil,
            typeLabel: String? = nil
        ) -> IngestCallout {
            switch reason {
            case .office:
                return IngestCallout(
                    title: ingestUnsupportedOfficeTitle,
                    message: String(localized: ingestUnsupportedOfficeHelp)
                )
            case .archive:
                return IngestCallout(
                    title: ingestUnsupportedArchiveTitle,
                    message: String(localized: ingestUnsupportedArchiveHelp)
                )
            case .executable:
                return IngestCallout(
                    title: ingestUnsupportedExecutableTitle,
                    message: String(localized: ingestUnsupportedExecutableHelp)
                )
            case .empty:
                return IngestCallout(
                    title: ingestEmptyTitle,
                    message: String(localized: ingestEmptyHelp)
                )
            case .tooLarge:
                let size = sizeLabel ?? "?"
                let format = String(localized: ingestTooLargeHelp)
                return IngestCallout(
                    title: ingestTooLargeTitle,
                    message: String(format: format, locale: .current, size)
                )
            case .unidentified:
                return IngestCallout(
                    title: ingestUnidentifiedTitle,
                    message: String(localized: ingestUnidentifiedHelp)
                )
            case .disallowedSniff, .genericType:
                if let typeLabel, !typeLabel.isEmpty, typeLabel != "That" {
                    return IngestCallout(
                        title: ingestUnsupportedTypeGenericTitle,
                        message: String(format: String(localized: LocalizedStringResource(
                            "error.ingest.unsupported_type.helpWithLabel",
                            defaultValue: "%@ files aren’t accepted. Artifacts hold images, PDFs, Word documents, plain text (including CSV and Markdown), audio, or video — up to 512 MB. Choose a different file.",
                            comment: "Callout help for unsupported type with label; argument is short type"
                        )), locale: .current, typeLabel)
                    )
                }
                return IngestCallout(
                    title: ingestUnsupportedTypeGenericTitle,
                    message: String(localized: ingestUnsupportedTypeHelp)
                )            case .notAFile:
                return IngestCallout(
                    title: ingestNotAFileTitle,
                    message: String(localized: ingestNotAFileHelp)
                )
            case .symlink:
                return IngestCallout(
                    title: ingestSymlinkTitle,
                    message: String(localized: ingestSymlinkHelp)
                )
            case .missing:
                return IngestCallout(
                    title: ingestMissingTitle,
                    message: String(localized: ingestMissingHelp)
                )
            case .permission, .unreadable:
                return IngestCallout(
                    title: ingestPermissionDeniedTitle,
                    message: String(localized: ingestPermissionDeniedHelp)
                )
            case .multiFile:
                return IngestCallout(
                    title: ingestMultiFileTitle,
                    message: String(localized: ingestMultiFileHelp)
                )
            }
        }

        static func ingestCallout(code: String, params: [String] = []) -> IngestCallout? {
            switch code {
            case "ingest.unsupported_office":
                return ingestCallout(reason: .office)
            case "ingest.unsupported_archive":
                return ingestCallout(reason: .archive)
            case "ingest.unsupported_executable":
                return ingestCallout(reason: .executable)
            case "ingest.unsupported_type":
                return ingestCallout(reason: .disallowedSniff, typeLabel: params.first)
            case "ingest.unidentified":
                return ingestCallout(reason: .unidentified)
            case "ingest.empty":
                return ingestCallout(reason: .empty)
            case "ingest.too_large":
                return ingestCallout(reason: .tooLarge, sizeLabel: params.first)
            case "ingest.not_a_file":
                return ingestCallout(reason: .notAFile)
            case "ingest.symlink":
                return ingestCallout(reason: .symlink)
            case "ingest.missing":
                return ingestCallout(reason: .missing)
            case "ingest.permission_denied":
                return ingestCallout(reason: .permission)
            case "ingest.invalid":
                return IngestCallout(
                    title: LocalizedStringResource(
                        "error.ingest.invalid.title",
                        defaultValue: "Could not ingest that file",
                        comment: "Callout title for ingest.invalid"
                    ),
                    message: String(localized: ingestInvalid)
                )
            default:
                return nil
            }
        }
        static let sourceTypesInvalid = LocalizedStringResource(
            "error.sourcetypes.invalid",
            defaultValue: "Invalid source type.",
            comment: "FFI error sourcetypes.invalid"
        )
        static let sourceTypesInUse = LocalizedStringResource(
            "error.sourcetypes.in_use",
            defaultValue: "That source type is still used by one or more sources.",
            comment: "FFI error sourcetypes.in_use"
        )
        static func sourceTypesDuplicateKey(key: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "error.sourcetypes.duplicate_key",
                defaultValue: "You already have a type with the key %@. Give this one a different label.",
                comment: "FFI error sourcetypes.duplicate_key; argument is the colliding key"
            ))
            return String(format: format, locale: .current, key)
        }
        static let sourceFieldsInvalid = LocalizedStringResource(
            "error.sourcefields.invalid",
            defaultValue: "Invalid metadata field.",
            comment: "FFI error sourcefields.invalid"
        )
        static func sourceFieldsDuplicateKey(key: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "error.sourcefields.duplicate_key",
                defaultValue: "You already have a field with the key %@. Give this one a different label.",
                comment: "FFI error sourcefields.duplicate_key; argument is the colliding key"
            ))
            return String(format: format, locale: .current, key)
        }
        static let sourceFieldsInUse = LocalizedStringResource(
            "error.sourcefields.in_use",
            defaultValue: "That metadata field is still used on one or more sources.",
            comment: "FFI error sourcefields.in_use"
        )
        static let sourceVocabInvalid = LocalizedStringResource(
            "error.sourcevocab.invalid",
            defaultValue: "Invalid source vocabulary.",
            comment: "FFI error sourcevocab.invalid"
        )
        static let propertiesInvalid = LocalizedStringResource(
            "error.properties.invalid",
            defaultValue: "Invalid property.",
            comment: "FFI error properties.invalid"
        )
        static func propertiesDuplicateKey(key: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "error.properties.duplicate_key",
                defaultValue: "You already have a property with the key %@. Give this one a different label.",
                comment: "FFI error properties.duplicate_key; argument is the colliding key"
            ))
            return String(format: format, locale: .current, key)
        }
        static let propertiesInUse = LocalizedStringResource(
            "error.properties.in_use",
            defaultValue: "That property is still bound to one or more subject types.",
            comment: "FFI error properties.in_use"
        )
        static let propertyTermsInvalid = LocalizedStringResource(
            "error.propertyterms.invalid",
            defaultValue: "Invalid property term.",
            comment: "FFI error propertyterms.invalid"
        )
        static func propertyTermsDuplicateKey(key: String) -> String {
            let format = String(localized: LocalizedStringResource(
                "error.propertyterms.duplicate_key",
                defaultValue: "You already have a term with the key %@. Give this one a different label.",
                comment: "FFI error propertyterms.duplicate_key; argument is the colliding key"
            ))
            return String(format: format, locale: .current, key)
        }
        static let propertyTermsLocked = LocalizedStringResource(
            "error.propertyterms.locked",
            defaultValue: "That term is locked by the product vocabulary and cannot be changed.",
            comment: "FFI error propertyterms.locked"
        )
        static let propertyTermsInUse = LocalizedStringResource(
            "error.propertyterms.in_use",
            defaultValue: "That term is still used by one or more observations.",
            comment: "FFI error propertyterms.in_use"
        )
        static let subjectVocabInvalid = LocalizedStringResource(
            "error.subjectvocab.invalid",
            defaultValue: "Invalid subject vocabulary.",
            comment: "FFI error subjectvocab.invalid"
        )
        static let subjectVocabLocked = LocalizedStringResource(
            "error.subjectvocab.locked",
            defaultValue: "That binding is locked by the product vocabulary and cannot be removed.",
            comment: "FFI error subjectvocab.locked"
        )
        static let connectInvalid = LocalizedStringResource(
            "error.connect.invalid",
            defaultValue: "That connection isn’t valid.",
            comment: "FFI error connect.invalid"
        )
        static let connectRefused = LocalizedStringResource(
            "error.connect.refused",
            defaultValue: "Those subjects can’t be connected.",
            comment: "FFI error connect.refused"
        )
        static let locatorInvalid = LocalizedStringResource(
            "error.locator.invalid",
            defaultValue: "That citation locator isn’t valid.",
            comment: "FFI error locator.invalid"
        )
        static let citationsInvalid = LocalizedStringResource(
            "error.citations.invalid",
            defaultValue: "That citation isn’t valid.",
            comment: "FFI error citations.invalid"
        )
        static let observationsInvalid = LocalizedStringResource(
            "error.observations.invalid",
            defaultValue: "That observation isn’t valid.",
            comment: "FFI error observations.invalid"
        )

        static let observationsEdgeLocked = LocalizedStringResource(
            "error.observations.edge_locked",
            defaultValue: "That connection edge cannot be edited or deleted.",
            comment: "FFI error observations.edge_locked"
        )
        static let dateValuesInvalid = LocalizedStringResource(
            "error.datevalues.invalid",
            defaultValue: "Invalid date value.",
            comment: "FFI error datevalues.invalid"
        )
        static let fileDerivativesInvalid = LocalizedStringResource(
            "error.filederivatives.invalid",
            defaultValue: "Invalid file derivative.",
            comment: "FFI error filederivatives.invalid"
        )
        static let fileDerivativesUnprocessable = LocalizedStringResource(
            "error.filederivatives.unprocessable",
            defaultValue: "Cannot generate a preview for that file.",
            comment: "FFI error filederivatives.unprocessable"
        )
        static let fileDerivativesCorruptObject = LocalizedStringResource(
            "error.filederivatives.corrupt_object",
            defaultValue: "Stored file object is corrupt.",
            comment: "FFI error filederivatives.corrupt_object"
        )
        static let unknown = LocalizedStringResource(
            "error.internal.unknown",
            defaultValue: "Something went wrong. Please try again.",
            comment: "FFI error internal.unknown and other unmapped codes"
        )
        static let internalUnknownMethod = LocalizedStringResource(
            "error.internal.unknown_method",
            defaultValue: "That action isn’t supported in this version of Provenencia.",
            comment: "FFI error internal.unknown_method"
        )
        static let internalMigrations = LocalizedStringResource(
            "error.internal.migrations",
            defaultValue: "Couldn’t update this catalog’s database. Try again.",
            comment: "FFI error internal.migrations"
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
            case "catalog.schema_mismatch":
                return String(localized: catalogSchemaMismatch)
            case "project.invalid_metadata":
                return String(localized: projectInvalidMetadata)
            case "project.missing_metadata":
                return String(localized: projectMissingMetadata)
            case "users.invalid":
                return String(localized: usersInvalid)
            case "audit.invalid":
                return String(localized: auditInvalid)
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
            case "sources.invalid":
                return String(localized: sourcesInvalid)
            case "subjects.invalid":
                return String(localized: subjectsInvalid)
            case "subjects.in_use":
                return String(localized: subjectsInUse)
            case "subjectpositions.invalid":
                return String(localized: subjectPositionsInvalid)
            case "subjecttypes.invalid":
                return String(localized: subjectTypesInvalid)
            case "subjecttypes.duplicate_prefix":
                return String(localized: subjectTypesDuplicatePrefix)
            case "artifacts.invalid":
                return String(localized: artifactsInvalid)
            case "artifacts.file_already_attached":
                return String(localized: artifactsFileAlreadyAttached)
            case "sourcecredibility.invalid":
                return String(localized: sourceCredibilityInvalid)
            case "sourcemetadata.invalid":
                return String(localized: sourceMetadataInvalid)
            case "files.invalid":
                return String(localized: filesInvalid)
            case "ingest.invalid":
                return String(localized: ingestInvalid)
            case "ingest.permission_denied":
                return String(localized: ingestPermissionDenied)
            case "ingest.unsupported_office":
                return String(localized: ingestUnsupportedOfficeHelp)
            case "ingest.unsupported_archive":
                return String(localized: ingestUnsupportedArchiveHelp)
            case "ingest.unsupported_executable":
                return String(localized: ingestUnsupportedExecutableHelp)
            case "ingest.unsupported_type":
                return ingestCallout(reason: .disallowedSniff, typeLabel: params.first).message
            case "ingest.unidentified":
                return String(localized: ingestUnidentifiedHelp)
            case "ingest.empty":
                return String(localized: ingestEmptyHelp)
            case "ingest.too_large":
                return ingestCallout(reason: .tooLarge, sizeLabel: params.first).message
            case "ingest.not_a_file":
                return String(localized: ingestNotAFileHelp)
            case "ingest.symlink":
                return String(localized: ingestSymlinkHelp)
            case "ingest.missing":
                return String(localized: ingestMissingHelp)
            case "sourcetypes.invalid":
                return String(localized: sourceTypesInvalid)
            case "sourcetypes.in_use":
                return String(localized: sourceTypesInUse)
            case "sourcetypes.duplicate_key":
                return sourceTypesDuplicateKey(key: params.first ?? "?")
            case "sourcefields.invalid":
                return String(localized: sourceFieldsInvalid)
            case "sourcefields.duplicate_key":
                return sourceFieldsDuplicateKey(key: params.first ?? "?")
            case "sourcefields.in_use":
                return String(localized: sourceFieldsInUse)
            case "sourcevocab.invalid":
                return String(localized: sourceVocabInvalid)
            case "properties.invalid":
                return String(localized: propertiesInvalid)
            case "properties.duplicate_key":
                return propertiesDuplicateKey(key: params.first ?? "?")
            case "properties.in_use":
                return String(localized: propertiesInUse)
            case "propertyterms.invalid":
                return String(localized: propertyTermsInvalid)
            case "propertyterms.duplicate_key":
                return propertyTermsDuplicateKey(key: params.first ?? "?")
            case "propertyterms.locked":
                return String(localized: propertyTermsLocked)
            case "propertyterms.in_use":
                return String(localized: propertyTermsInUse)
            case "subjectvocab.invalid":
                return String(localized: subjectVocabInvalid)
            case "subjectvocab.locked":
                return String(localized: subjectVocabLocked)
            case "connect.invalid":
                return String(localized: connectInvalid)
            case "connect.refused":
                return String(localized: connectRefused)
            case "locator.invalid":
                return String(localized: locatorInvalid)
            case "citations.invalid":
                return String(localized: citationsInvalid)
            case "observations.invalid":
                return String(localized: observationsInvalid)
            case "observations.edge_locked":
                return String(localized: observationsEdgeLocked)
            case "datevalues.invalid":
                return String(localized: dateValuesInvalid)
            case "filederivatives.invalid":
                return String(localized: fileDerivativesInvalid)
            case "filederivatives.unprocessable":
                return String(localized: fileDerivativesUnprocessable)
            case "filederivatives.corrupt_object":
                return String(localized: fileDerivativesCorruptObject)
            case "internal.unknown":
                return String(localized: unknown)
            case "internal.unknown_method":
                return String(localized: internalUnknownMethod)
            case "internal.migrations":
                return String(localized: internalMigrations)
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

        /// `ref.invalid` / `ref.invalid_prefix` / `ref.reserved_prefix` are
        /// minting / reserved-prefix failures and intentionally fall through to
        /// `unknown` unless a future client surfaces them as distinct copy.
    }
}
