import Foundation

// MARK: - Resolution

extension L10n {
    /// Resolves `resource` to a `String` through the bundle's cached string table.
    ///
    /// Use this, not `String(localized:)`, whenever copy is needed as a `String`.
    /// `String(localized: LocalizedStringResource)` asks the bundle for explicit
    /// localizations, a path that re-reads and re-parses the whole
    /// `Localizable.strings` table on every call (~1.5 ms against ~1 µs here).
    /// Called from view bodies, that cost lands on every frame. `Text(resource)`
    /// resolves through SwiftUI and does not need this.
    ///
    /// Only static resources resolve here: a resource with interpolated
    /// arguments carries values this lookup cannot read, so give it a `%@`
    /// catalog format and a ``format(_:_:)`` helper instead. Every key must be
    /// in the catalog (checked by `scripts/check-localizable-xcstrings.py`);
    /// a missing key resolves to the key itself, as it does in Foundation.
    static func string(_ resource: LocalizedStringResource) -> String {
        string(resource, in: .main)
    }

    /// ``string(_:)`` against a specific bundle — the seam tests use to resolve
    /// through a bundle with other localizations.
    static func string(_ resource: LocalizedStringResource, in bundle: Bundle) -> String {
        let resolved = bundle.localizedString(forKey: resource.key, value: nil, table: resource.table)
        // Only semantic keys (`evidenceGraph.subject.addProperty`) belong in the
        // catalog; English-as-key resources in previews and test fixtures
        // ("Delete field") resolve to their key by design.
        assert(
            resolved != resource.key || !isSemanticKey(resource.key),
            "L10n key missing from the String Catalog: \(resource.key)"
        )
        return resolved
    }

    /// A dotted catalog key such as `evidenceGraph.subject.addProperty`, as
    /// `scripts/check-localizable-xcstrings.py` defines it.
    private static func isSemanticKey(_ key: String) -> Bool {
        key.first?.isLowercase == true && key.contains(".") && !key.contains(" ")
    }

    /// Resolves the catalog format behind `resource` and fills in `arguments`,
    /// in a locale whose language matches the localization that was resolved.
    ///
    /// Plural and other catalog variations pick their form from the locale's
    /// language, so formatting with a locale in another language silently applies
    /// the wrong rules. Always format L10n copy here rather than calling
    /// `String(format:)` yourself.
    static func format(_ resource: LocalizedStringResource, _ arguments: any CVarArg...) -> String {
        format(resource, in: .main, locale: formattingLocale, arguments: arguments)
    }

    /// ``format(_:_:)`` with a caller-chosen formatting locale, for helpers whose
    /// callers format values (such as dates) for a specific locale. The copy itself
    /// still resolves from the bundle's localization.
    static func format(_ resource: LocalizedStringResource, locale: Locale, _ arguments: any CVarArg...) -> String {
        format(resource, in: .main, locale: locale, arguments: arguments)
    }

    /// The one place L10n copy is formatted. `locale` must share the language of
    /// the localization `bundle` resolves, or plural variations pick the wrong form.
    static func format(
        _ resource: LocalizedStringResource,
        in bundle: Bundle,
        locale: Locale,
        arguments: [any CVarArg]
    ) -> String {
        String(format: string(resource, in: bundle), locale: locale, arguments: arguments)
    }

    /// `Locale.current`, with its language replaced by the bundle's resolved
    /// localization when the two differ (for example, a per-app language set in
    /// System Settings). Region formatting preferences are kept.
    static var formattingLocale: Locale {
        formattingLocale(current: .current, resolvedLocalization: Bundle.main.preferredLocalizations.first)
    }

    static func formattingLocale(current: Locale, resolvedLocalization: String?) -> Locale {
        guard let resolvedLocalization else { return current }
        let resolved = Locale.Language(identifier: resolvedLocalization)
        guard current.language.languageCode != resolved.languageCode else { return current }
        var components = Locale.Components(locale: current)
        // Language and script only: languageComponents also carries the region.
        components.languageComponents.languageCode = resolved.languageCode
        components.languageComponents.script = resolved.script
        return Locale(components: components)
    }
}

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
            return L10n.format(LocalizedStringResource(
                "designSystem.table.filterColumn",
                defaultValue: "Filter %@",
                comment: "Accessibility label for a PVTable column's filter menu; argument is the column title"
            ), column)
        }

        static let disclosureState = LocalizedStringResource(
            "designSystem.disclosure.state",
            defaultValue: "State",
            comment: "VoiceOver custom-content key for whether something is shown: a disclosure, a menu, an expandable row"
        )

        static let disclosureExpanded = LocalizedStringResource(
            "designSystem.disclosure.expanded",
            defaultValue: "Expanded",
            comment: "VoiceOver value when a disclosure, menu or expandable row is open"
        )

        static let disclosureCollapsed = LocalizedStringResource(
            "designSystem.disclosure.collapsed",
            defaultValue: "Collapsed",
            comment: "VoiceOver value when a disclosure, menu or expandable row is closed"
        )

        static let selectPosition = LocalizedStringResource(
            "designSystem.select.position",
            defaultValue: "Position",
            comment: "VoiceOver custom-content key for the highlighted PVSelect option index"
        )

        static func selectOptionPosition(current: Int, count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "designSystem.select.optionPosition",
                defaultValue: "%1$lld of %2$lld",
                comment: "VoiceOver position of the highlighted PVSelect option; arguments are 1-based index and count"
            ), current, count)
        }

        static func tableFilterOptionCount(label: String, count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "designSystem.table.filterOptionCount",
                defaultValue: "%1$@ (%2$lld)",
                comment: "A PVTable filter menu option with its row count; arguments are the option label and the count"
            ), label, count)
        }

        static let thumbnailEmpty = LocalizedStringResource(
            "designSystem.thumbnail.empty",
            defaultValue: "No preview",
            comment: "Accessibility label for an empty PVThumbnail placeholder"
        )

        static let listLoading = LocalizedStringResource(
            "designSystem.list.loading",
            defaultValue: "Loading",
            comment: "Accessibility label for a PVList skeleton shown on first load"
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
        static let markSubjectPlaceRelationship = LocalizedStringResource(
            "designSystem.mark.subjectPlaceRelationship",
            defaultValue: "Place relationship bridge"
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
            return L10n.format(LocalizedStringResource(
                "onboarding.chooseFile.bodySignedIn",
                defaultValue: "You're signed in as %@. Open a project you already have, or create a new one.",
                comment: "Onboarding choose-file body when researcher is locked; argument is display name"
            ), displayName)
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
            return L10n.format(LocalizedStringResource(
                "onboarding.identify.folderNamePreview",
                defaultValue: "Folder: %@",
                comment: "Live preview of kebab-case project folder name; argument is folder basename"
            ), folderName)
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
            return L10n.format(LocalizedStringResource(
                "onboarding.identify.contributorsBody",
                defaultValue: "These are the contributors already in %@. Choose yourself to keep the same ID, or add a new contributor.",
                comment: "Open-mode identify body; argument is project folder name"
            ), projectName)
        }

        static func contributorOption(displayName: String, ref: String) -> String {
            if ref.isEmpty {
                return displayName
            }
            return L10n.format(LocalizedStringResource(
                "onboarding.identify.contributorOption",
                defaultValue: "%1$@ (%2$@)",
                comment: "Contributor accessibility/combined label; arguments are display name then USR-… ref"
            ), displayName, ref)
        }

        static func contributorRef(ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "onboarding.identify.contributorRef",
                defaultValue: "(%@)",
                comment: "Parenthesized short ref beside a display name; argument is USR-… ref"
            ), ref)
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
            return L10n.format(LocalizedStringResource(
                "onboarding.home.folder",
                defaultValue: "Folder: %@",
                comment: "Home screen project folder line; argument is folder basename"
            ), folderName)
        }

        static func homeCreated(date: String) -> String {
            return L10n.format(LocalizedStringResource(
                "onboarding.home.created",
                defaultValue: "Created %@",
                comment: "Home screen created date; argument is localized date"
            ), date)
        }

        static func homeUpdated(date: String) -> String {
            return L10n.format(LocalizedStringResource(
                "onboarding.home.updated",
                defaultValue: "Updated %@",
                comment: "Home screen updated date; argument is localized date"
            ), date)
        }

        static func homeUpdatedBy(displayName: String, ref: String) -> String {
            if ref.isEmpty {
                return L10n.format(LocalizedStringResource(
                    "onboarding.home.updatedByName",
                    defaultValue: "Last edited by %@",
                    comment: "Home screen last editor without ref; argument is display name"
                ), displayName)
            }
            return L10n.format(LocalizedStringResource(
                "onboarding.home.updatedBy",
                defaultValue: "Last edited by %1$@ (%2$@)",
                comment: "Home screen last editor; arguments are display name then USR-… ref"
            ), displayName, ref)
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

        static let metadataTitle = LocalizedStringResource(
            "workspace.section.metadata.title",
            defaultValue: "Metadata",
            comment: "Workspace sidebar destination and page title: Metadata"
        )

        static let propertiesTitle = LocalizedStringResource(
            "workspace.section.properties.title",
            defaultValue: "Properties",
            comment: "Workspace sidebar destination and page title: Properties"
        )

        static let personsTitle = LocalizedStringResource(
            "workspace.section.persons.title",
            defaultValue: "Persons",
            comment: "Workspace sidebar destination and page title: Persons (Conclude section)"
        )

        static let eventsTitle = LocalizedStringResource(
            "workspace.section.events.title",
            defaultValue: "Events",
            comment: "Workspace sidebar destination and page title: Events (Conclude section)"
        )

        static let placesTitle = LocalizedStringResource(
            "workspace.section.places.title",
            defaultValue: "Places",
            comment: "Workspace sidebar destination and page title: Places (Conclude section)"
        )

        /// Places list header meta: "1 place · by name" / "N places · by name".
        static func placeCount(_ count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "workspace.places.count",
                defaultValue: "%lld places · by name",
                comment: "Places list header meta; argument is how many Places are listed. Includes the name sort."
            ), count)
        }

        /// Header meta while a stale Places list reloads.
        static func placeCountRefreshing(_ count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "workspace.places.countRefreshing",
                defaultValue: "%@ · refreshing",
                comment: "Places list header meta while the list reloads; argument is the place count text"
            ), placeCount(count))
        }

        static let placesEmptyTitle = LocalizedStringResource(
            "workspace.places.emptyTitle",
            defaultValue: "No places yet",
            comment: "Places list empty state title"
        )

        static let placesEmptyMessage = LocalizedStringResource(
            "workspace.places.emptyMessage",
            defaultValue: "A Place is created when you promote a place subject from a card on an Evidence graph. Promoted places appear here, by name.",
            comment: "Places list empty state body; explains Promote"
        )

        /// VoiceOver label for one Places row. Extra names and the chain are
        /// omitted when the row has none.
        static func placeRowAccessibility(title: String, extra: Int, chain: String, ref: String) -> String {
            if extra > 0 && !chain.isEmpty {
                return L10n.format(LocalizedStringResource(
                    "workspace.places.rowAccessibilityExtraInChain",
                    defaultValue: "%1$@, and %#@extra@, in %3$@, %4$@",
                    comment: "VoiceOver label for a Places row with other names and a parent chain; arguments are the title, the extra-name count, the chain, and the ref"
                ), title, extra, chain, ref)
            }
            if extra > 0 {
                return L10n.format(LocalizedStringResource(
                    "workspace.places.rowAccessibilityExtra",
                    defaultValue: "%1$@, and %#@extra@, %3$@",
                    comment: "VoiceOver label for a Places row with other names and no chain; arguments are the title, the extra-name count, and the ref"
                ), title, extra, ref)
            }
            if !chain.isEmpty {
                return L10n.format(LocalizedStringResource(
                    "workspace.places.rowAccessibilityInChain",
                    defaultValue: "%1$@, in %2$@, %3$@",
                    comment: "VoiceOver label for a Places row with a parent chain; arguments are the title, the chain, and the ref"
                ), title, chain, ref)
            }
            return L10n.format(LocalizedStringResource(
                "workspace.places.rowAccessibility",
                defaultValue: "%1$@, %2$@",
                comment: "VoiceOver label for a Places row with one name and no chain; arguments are the title and the ref"
            ), title, ref)
        }

        static let sidebarSourceTitle = LocalizedStringResource(
            "workspace.sidebar.section.source",
            defaultValue: "Source",
            comment: "Sidebar section title above Sources (workflow stage)"
        )

        static let sidebarConcludeTitle = LocalizedStringResource(
            "workspace.sidebar.section.conclude",
            defaultValue: "Conclude",
            comment: "Sidebar section title above Persons, Events and Places (workflow stage)"
        )

        static let sidebarConfigureTitle = LocalizedStringResource(
            "workspace.sidebar.section.configure",
            defaultValue: "Configure",
            comment: "Sidebar section title above Source types, Metadata and Properties"
        )

        /// Persons list header meta: "1 person" / "N persons".
        static func personCount(_ count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "workspace.persons.count",
                defaultValue: "%lld persons",
                comment: "Persons list header meta; argument is how many Persons are listed"
            ), count)
        }

        /// Header meta while a stale list reloads: "N persons · refreshing".
        static func personCountRefreshing(_ count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "workspace.persons.countRefreshing",
                defaultValue: "%@ · refreshing",
                comment: "Persons list header meta while the list reloads; argument is the person count text"
            ), personCount(count))
        }

        static let personsEmptyTitle = LocalizedStringResource(
            "workspace.persons.emptyTitle",
            defaultValue: "No persons yet",
            comment: "Persons list empty state title"
        )

        static let personsEmptyMessage = LocalizedStringResource(
            "workspace.persons.emptyMessage",
            defaultValue: "A Person is created when you promote a subject from a card on an Evidence graph. Promoted subjects appear here, one row per Person.",
            comment: "Persons list empty state body; explains Promote"
        )

        /// VoiceOver label for one Persons row: "James Robins, PER-7KD45".
        static func personRowAccessibility(title: String, ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "workspace.persons.rowAccessibility",
                defaultValue: "%1$@, %2$@",
                comment: "VoiceOver label for a Persons list row; arguments are the title and the ref"
            ), title, ref)
        }

        /// Events list header meta: "1 event · by date" / "N events · by date".
        static func eventCount(_ count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "workspace.events.count",
                defaultValue: "%lld events · by date",
                comment: "Events list header meta; argument is how many Events are listed. Includes the date sort."
            ), count)
        }

        /// Header meta while a stale Events list reloads.
        static func eventCountRefreshing(_ count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "workspace.events.countRefreshing",
                defaultValue: "%@ · refreshing",
                comment: "Events list header meta while the list reloads; argument is the event count text"
            ), eventCount(count))
        }

        static let eventsEmptyTitle = LocalizedStringResource(
            "workspace.events.emptyTitle",
            defaultValue: "No events yet",
            comment: "Events list empty state title"
        )

        static let eventsEmptyMessage = LocalizedStringResource(
            "workspace.events.emptyMessage",
            defaultValue: "An Event is created when you promote an event subject from a card on an Evidence graph. Promoted events appear here, in date order.",
            comment: "Events list empty state body; explains Promote"
        )

        /// VoiceOver label for one Events row. The date is omitted when the row has none.
        static func eventRowAccessibility(title: String, date: String, ref: String) -> String {
            if date.isEmpty {
                return L10n.format(LocalizedStringResource(
                    "workspace.events.rowAccessibility",
                    defaultValue: "%1$@, %2$@",
                    comment: "VoiceOver label for an Events list row with no date; arguments are the title and the ref"
                ), title, ref)
            }
            return L10n.format(LocalizedStringResource(
                "workspace.events.rowAccessibilityDated",
                defaultValue: "%1$@, %2$@, %3$@",
                comment: "VoiceOver label for an Events list row; arguments are the title, the date, and the ref"
            ), title, date, ref)
        }

        static let evidenceGraphTitle = LocalizedStringResource(
            "workspace.section.evidenceGraph.title",
            defaultValue: "Evidence graph",
            comment: "Evidence graph deep place title (toolbar breadcrumb and stub)"
        )

        static func evidenceGraphFor(sourceTitle: String) -> String {
            return L10n.format(LocalizedStringResource(
                "workspace.section.evidenceGraph.forSource",
                defaultValue: "Evidence graph for %@",
                comment: "Composer breadcrumb segment; argument is the Source title"
            ), sourceTitle)
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

        static let omnibarKindMetadataField = LocalizedStringResource(
            "workspace.omnibar.kind.metadataField",
            defaultValue: "Metadata field",
            comment: "Kind chip on an omnibar hit for a metadata field"
        )

        static let omnibarKindPerson = LocalizedStringResource(
            "workspace.omnibar.kind.person",
            defaultValue: "Person",
            comment: "Kind chip on an omnibar hit for a Person"
        )

        static let omnibarKindEvent = LocalizedStringResource(
            "workspace.omnibar.kind.event",
            defaultValue: "Event",
            comment: "Kind chip on an omnibar hit for an Event"
        )

        static let omnibarKindPlace = LocalizedStringResource(
            "workspace.omnibar.kind.place",
            defaultValue: "Place",
            comment: "Kind chip on an omnibar hit for a Place"
        )

        static func omnibarNoMatchesTitle(query: String) -> String {
            return L10n.format(LocalizedStringResource(
                "workspace.omnibar.noMatchesTitle",
                defaultValue: "No matches for “%@”.",
                comment: "Omnibar empty-state title; argument is the typed query"
            ), query)
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
            return L10n.format(LocalizedStringResource(
                "workspace.omnibar.match.note",
                defaultValue: "Note: %@",
                comment: "Omnibar match context when the query hit a Source note; argument is a short snippet"
            ), snippet)
        }

        /// Omnibar match-context line for metadata text; argument is the raw snippet.
        static func omnibarMatchMetadata(snippet: String) -> String {
            return L10n.format(LocalizedStringResource(
                "workspace.omnibar.match.metadata",
                defaultValue: "Metadata: %@",
                comment: "Omnibar match context when the query hit Source metadata; argument is a short snippet"
            ), snippet)
        }

        /// Omnibar match-context line for a filename/artifact label; argument is the raw snippet.
        static func omnibarMatchFilename(snippet: String) -> String {
            return L10n.format(LocalizedStringResource(
                "workspace.omnibar.match.filename",
                defaultValue: "Filename: %@",
                comment: "Omnibar match context when the query hit a filename or artifact label; argument is a short snippet"
            ), snippet)
        }

        static func omnibarMatchName(snippet: String) -> String {
            return L10n.format(LocalizedStringResource(
                "workspace.omnibar.match.name",
                defaultValue: "Name: %@",
                comment: "Omnibar match context when an alternate name matched; argument is that name"
            ), snippet)
        }

        static func omnibarMatchPlace(snippet: String) -> String {
            return L10n.format(LocalizedStringResource(
                "workspace.omnibar.match.place",
                defaultValue: "Place: %@",
                comment: "Omnibar match context when an alternate place name matched; argument is that name"
            ), snippet)
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

    /// Promote flow place (S9-11, board S9-D9). Generated table; kind-specific
    /// copy switches on the subject's kind.
    enum Promote {
        static let pageTitle = LocalizedStringResource("promote.page.title", defaultValue: "Promote", comment: "Promote page title")
        static func mapRest(count: Int) -> String {
            L10n.format(LocalizedStringResource("promote.page.mapRest", defaultValue: "Map the rest of this graph (%lld)", comment: "Adds every other person, event, and place; argument is how many"), count)
        }
        static let groupKind = LocalizedStringResource("promote.page.group.kind", defaultValue: "Kind", comment: "Group Promote rows by person, event, or place")
        static let groupAssessment = LocalizedStringResource("promote.page.group.assessment", defaultValue: "Assessment", comment: "Group Promote rows by how sure the match is")
        static let groupLabel = LocalizedStringResource("promote.page.group.label", defaultValue: "Group rows", comment: "Accessibility label for the Promote grouping menu")
        static let kindPerson = LocalizedStringResource("promote.page.kind.person", defaultValue: "People", comment: "Promote section of person rows")
        static let kindEvent = LocalizedStringResource("promote.page.kind.event", defaultValue: "Events", comment: "Promote section of event rows")
        static let kindPlace = LocalizedStringResource("promote.page.kind.place", defaultValue: "Places", comment: "Promote section of place rows")
        static let assessmentStrong = LocalizedStringResource("promote.assessment.strong", defaultValue: "Strong", comment: "Promote match assessment")
        static let assessmentMedium = LocalizedStringResource("promote.assessment.medium", defaultValue: "Medium", comment: "Promote match assessment between strong and weak")
        static let assessmentWeak = LocalizedStringResource("promote.assessment.weak", defaultValue: "Weak", comment: "Promote match assessment")
        static let assessmentNone = LocalizedStringResource("promote.assessment.none", defaultValue: "No match", comment: "Promote match assessment when nothing lines up")
        static let decided = LocalizedStringResource("promote.row.decided", defaultValue: "Decided", comment: "A Promote row the researcher has changed")
        static let updated = LocalizedStringResource("promote.row.updated", defaultValue: "Updated", comment: "A suggested Promote row that moved after a re-proposal")
        static let targetPlaceholder = LocalizedStringResource("promote.target.placeholder", defaultValue: "Select where this files", comment: "Promote target menu when no record, New, or Skip is chosen yet")
        static let skip = LocalizedStringResource("promote.target.skip", defaultValue: "Skip — not filed yet", comment: "Leave this subject unfiled; Skip is not a rejection")
        static func via(neighbor: String, ref: String, role: String) -> String {
            L10n.format(LocalizedStringResource("promote.reason.via", defaultValue: "via %1$@ → %2$@ (%3$@)", comment: "Why a row was matched: neighbor name, neighbor ref, relationship"), neighbor, ref, role)
        }
        static let alreadyFiled = LocalizedStringResource("promote.reason.filed", defaultValue: "Already filed", comment: "Reason on a read-only promoted row")
        static func agreesOn(property: String) -> String {
            L10n.format(LocalizedStringResource("promote.reason.agrees", defaultValue: "Agrees on %@", comment: "Reason naming the property that agrees most; argument is the property's label"), property)
        }
        static func viaNeighbor(neighbor: String, ref: String) -> String {
            L10n.format(LocalizedStringResource("promote.reason.viaNeighbor", defaultValue: "via %1$@ → %2$@", comment: "Why a row was matched when the connection has no phrase: neighbor name, neighbor's handle ref"), neighbor, ref)
        }
        static let reasonVia = LocalizedStringResource("promote.reason.viaUnknown", defaultValue: "Reached from a matched neighbor", comment: "Why a row was matched, when the neighbor row is not on the page")
        static let reasonDecided = LocalizedStringResource("promote.reason.decided", defaultValue: "Your choice", comment: "Reason on a row whose target the researcher chose")
        static let reasonWeak = LocalizedStringResource("promote.reason.weak", defaultValue: "Too little agrees to suggest a match", comment: "Reason on a row whose best candidate is below the bar")
        static let reasonTaken = LocalizedStringResource("promote.reason.taken", defaultValue: "Another row took the best match", comment: "Reason on a row whose best handle went to another row on the page")
        static let reasonNoMatch = LocalizedStringResource("promote.reason.noMatch", defaultValue: "No matching record", comment: "Reason on a row with no candidate handle")
        static let reasonEmpty = LocalizedStringResource("promote.reason.empty", defaultValue: "Nothing to match on", comment: "Reason on a row whose subject has no values to compare")
        static func updatedDetail(_ reason: String) -> String {
            L10n.format(LocalizedStringResource("promote.row.updatedDetail", defaultValue: "Updated · %@", comment: "A suggested row moved on the last proposal; argument is why, e.g. via a neighbor"), reason)
        }
        static let claimHeading = LocalizedStringResource("promote.sheet.claim", defaultValue: "Claim", comment: "Heading above status, confidence, and argument on the evidence sheet")
        static let backToRows = LocalizedStringResource("promote.sheet.back", defaultValue: "Back to all rows", comment: "Closes the Promote evidence sheet")
        static let sheetNeighbor = LocalizedStringResource("promote.sheet.neighbor", defaultValue: "a neighbor", comment: "Evidence group name when the neighbor row is not on the page; shown after 'Through'")
        static let ownRecords = LocalizedStringResource("promote.sheet.own", defaultValue: "This subject", comment: "Evidence group for the subject's own records")
        static let status = LocalizedStringResource("promote.sheet.status", defaultValue: "Status", comment: "Claim status field")
        static let confidence = LocalizedStringResource("promote.sheet.confidence", defaultValue: "Confidence", comment: "Claim confidence field")
        static let confidenceUnset = LocalizedStringResource("promote.sheet.confidence.unset", defaultValue: "Not stated", comment: "No confidence grade chosen")
        static let argument = LocalizedStringResource("promote.sheet.argument", defaultValue: "Argument", comment: "Claim argument field")
        static func pinLine(property: String, here: String, there: String) -> String {
            L10n.format(LocalizedStringResource("promote.sheet.pinLine", defaultValue: "Pin %1$@: %2$@ and %3$@", comment: "VoiceOver label for one pin box; 1 = property, 2 = this source's value, 3 = the matched record's value"), property, here, there)
        }
        static func evidenceFor(name: String, assessment: String, reason: String) -> String {
            L10n.format(LocalizedStringResource("promote.row.evidenceFor", defaultValue: "Evidence for %1$@: %2$@, %3$@", comment: "VoiceOver label for the button that opens a row's evidence; 1 = subject name, 2 = assessment, 3 = reason"), name, assessment, reason)
        }
        static func outcome(_ outcome: String) -> LocalizedStringResource {
            switch outcome {
            case "agree": LocalizedStringResource("promote.sheet.agrees", defaultValue: "Agrees", comment: "A comparison that agrees")
            case "partial": LocalizedStringResource("promote.sheet.similar", defaultValue: "Similar", comment: "A comparison whose text is a spelling variant or shares words")
            case "conflict": LocalizedStringResource("promote.sheet.conflicts", defaultValue: "Conflicts", comment: "A comparison that disagrees")
            default: LocalizedStringResource("promote.sheet.unknown", defaultValue: "Unknown", comment: "A comparison with nothing to compare")
            }
        }
        static func connections(count: Int) -> String {
            L10n.format(LocalizedStringResource("promote.connections.count", defaultValue: "%lld connections will be filed", comment: "Bridge summary; argument is how many will be filed"), count)
        }
        static let connectionFiled = LocalizedStringResource("promote.connections.filed", defaultValue: "Filed", comment: "Switch label for a connection that will be filed")
        static func connectionState(_ state: PromoteFlow.ConnectionState) -> LocalizedStringResource {
            switch state {
            case .files: connectionFiled
            case .off: LocalizedStringResource("promote.connections.off", defaultValue: "Not filed this time", comment: "A connection the researcher switched off")
            case .selfLink: LocalizedStringResource("promote.connections.self", defaultValue: "Not filed · both ends are the same record", comment: "A connection whose ends resolve to one handle")
            case .endSkipped: LocalizedStringResource("promote.connections.skipped", defaultValue: "Not filed · an end is skipped", comment: "A connection that stays unfiled because an end is not being filed")
            }
        }
        static func doneSummary(file: Int, skip: Int, connections: Int) -> String {
            L10n.format(LocalizedStringResource("promote.done.summary", defaultValue: "File %1$lld · Skip %2$lld · %3$lld connections", comment: "Footer: how many rows will be filed, skipped, and how many connections"), file, skip, connections)
        }
        static let filing = LocalizedStringResource("promote.done.filing", defaultValue: "Filing", comment: "Done button while the batch is writing")
        static let done = LocalizedStringResource("promote.done.action", defaultValue: "Done", comment: "Files the Promote page")
        static let cancel = LocalizedStringResource("promote.cancel", defaultValue: "Cancel", comment: "Leaves the Promote page")
        static let staleTitle = LocalizedStringResource("promote.stale.title", defaultValue: "Nothing was filed — the catalog changed while this page was open", comment: "Callout title when Done loses the revision race")
        static let stale = LocalizedStringResource("promote.stale.notice", defaultValue: "Provenencia proposed the rows again. Rows you decided are kept. Review the updated rows and press Done.", comment: "Callout body when Done loses the revision race")
        static func filedTitle(filed: Int, connections: Int) -> String {
            L10n.format(LocalizedStringResource("promote.toast.filedTitle", defaultValue: "Filed %1$lld subjects and %2$lld connections", comment: "Toast title after Done; 1 = subjects filed, 2 = connections filed"), filed, connections)
        }
        static func filedBody(skipped: Int) -> String {
            L10n.format(LocalizedStringResource("promote.toast.filedBody", defaultValue: "%lld subjects are still on the graph, not filed yet", comment: "Toast body after Done; skipped subjects are not a rejection"), skipped)
        }
        static let leaveTitle = LocalizedStringResource("promote.leave.pageTitle", defaultValue: "Discard your changes to this mapping?", comment: "Leave guard title when the Promote page has manual changes")
        static func leaveDetail(rows: Int, connections: Int) -> String {
            L10n.format(LocalizedStringResource("promote.leave.pageMessage", defaultValue: "You chose targets on %1$lld rows and switched off %2$lld connections. Nothing has been filed; opening Promote again proposes the graph afresh.", comment: "Leave guard body; 1 = rows changed, 2 = connections switched off"), rows, connections)
        }
        static func openedFrom(name: String) -> String {
            L10n.format(LocalizedStringResource("promote.page.openedFrom", defaultValue: "Opened from %@ · weak and unmatched rows start as Skip", comment: "Shown once the rest of the graph is mapped; argument is the subject the page opened on"), name)
        }
        static func promoting(count: Int) -> String {
            L10n.format(LocalizedStringResource("promote.page.promoting", defaultValue: "%lld subjects on this graph", comment: "Header count once the whole graph is mapped"), count)
        }
        static func rowSummary(rows: Int, filed: Int) -> String {
            L10n.format(LocalizedStringResource("promote.page.rowSummary", defaultValue: "%1$lld rows · %2$lld already filed", comment: "Header once the whole graph is mapped; 1 = row count, 2 = already-filed anchors"), rows, filed)
        }
        static func singleSummary(more: Int) -> String {
            L10n.format(LocalizedStringResource("promote.page.singleSummary", defaultValue: "1 row · %lld more on this graph", comment: "Header before the rest of the graph is mapped; argument is how many other subjects"), more)
        }
        static let mapRestHint = LocalizedStringResource("promote.page.mapRestHint", defaultValue: "Adds a row for every other person, event and place on this Source", comment: "Beside the control that maps the rest of the graph")
        static let columnSource = LocalizedStringResource("promote.page.column.source", defaultValue: "On this Source", comment: "Promote row column: the subject on this Source")
        static let columnTarget = LocalizedStringResource("promote.page.column.target", defaultValue: "Files on", comment: "Promote row column: the handle the subject will file on")
        static let columnAssessment = LocalizedStringResource("promote.page.column.assessment", defaultValue: "Assessment", comment: "Promote row column: how sure the match is")
        static let groupedByKind = LocalizedStringResource("promote.page.grouped.kind", defaultValue: "Grouped by kind", comment: "Closed label of the grouping menu when rows are grouped by kind")
        static let groupedByAssessment = LocalizedStringResource("promote.page.grouped.assessment", defaultValue: "Grouped by assessment", comment: "Closed label of the grouping menu when rows are grouped by assessment")
        static func sectionCount(_ count: Int) -> String {
            L10n.format(LocalizedStringResource("promote.section.count", defaultValue: "%lld", comment: "How many rows are in a Promote section"), count)
        }
        static func sectionCountFiled(rows: Int, filed: Int) -> String {
            L10n.format(LocalizedStringResource("promote.section.countFiled", defaultValue: "%1$lld · %2$lld already filed", comment: "Promote section count when some rows were filed earlier; 1 = rows, 2 = already filed"), rows, filed)
        }
        static let alreadyFiledHeading = LocalizedStringResource("promote.row.alreadyFiledHeading", defaultValue: "Already filed", comment: "Heading above read-only rows filed from this Source earlier")
        static let filedEarlier = LocalizedStringResource("promote.row.filedEarlier", defaultValue: "Filed from this Source earlier", comment: "Assessment stand-in on a read-only already-filed row")
        static func filesOn(name: String) -> String {
            L10n.format(LocalizedStringResource("promote.target.filesOn", defaultValue: "Files %@ on", comment: "Accessibility label for a row's target menu; argument is the subject's name"), name)
        }
        static func duplicateTitle(name: String, ref: String) -> String {
            L10n.format(LocalizedStringResource("promote.warning.duplicateTitle", defaultValue: "%1$@ is also filed on %2$@", comment: "Warning title when two rows share a handle; 1 = the other subject's name, 2 = the handle ref"), name, ref)
        }
        static func duplicateNewTitle(name: String) -> String {
            L10n.format(LocalizedStringResource("promote.warning.duplicateNewTitle", defaultValue: "%@ is also filed New and looks the same", comment: "Warning title when two New rows on the page match each other; argument is the other subject's name"), name)
        }
        static let duplicateBody = LocalizedStringResource("promote.warning.duplicateBody", defaultValue: "These may be the same record. Combine them on the Evidence graph, or pick another target for one row.", comment: "Warning body when two rows may be the same record")
        static func conflictTitle(ref: String) -> String {
            L10n.format(LocalizedStringResource("promote.warning.conflictTitle", defaultValue: "A stronger match is %@", comment: "Warning title when a decided row disagrees with a new proposal; argument is the other handle ref"), ref)
        }
        static let conflictBody = LocalizedStringResource("promote.warning.conflictBody", defaultValue: "This choice is kept.", comment: "Warning body when a decided row disagrees with a new proposal")
        static let connectionsTitle = LocalizedStringResource("promote.connections.title", defaultValue: "Connections", comment: "Title of the connections list")
        static func connectionsMeta(filed: Int, off: Int, unfiled: Int) -> String {
            L10n.format(LocalizedStringResource("promote.connections.meta", defaultValue: "%1$lld filed · %2$lld switched off · %3$lld not filed", comment: "Connections list summary; 1 = filed, 2 = switched off, 3 = not filed"), filed, off, unfiled)
        }
        static let connectionsHint = LocalizedStringResource("promote.connections.hint", defaultValue: "Switch off a relationship you don’t accept from this Source; it stays on the graph", comment: "Hint in the connections list")
        static let sheetHere = LocalizedStringResource("promote.sheet.column.here", defaultValue: "This source", comment: "Evidence sheet column: the value on this Source")
        static let sheetPin = LocalizedStringResource("promote.sheet.column.pin", defaultValue: "Pin", comment: "Evidence sheet column: whether the comparison is pinned")
        static let sheetCompared = LocalizedStringResource("promote.sheet.column.compared", defaultValue: "Compared", comment: "Evidence sheet column: what is being compared")
        static let sheetResult = LocalizedStringResource("promote.sheet.column.result", defaultValue: "Result", comment: "Evidence sheet column: agrees, conflicts, or unknown")
        static let sheetWeight = LocalizedStringResource("promote.sheet.column.weight", defaultValue: "Weight", comment: "Evidence sheet column: how much the comparison counted")
        static let sheetStatusHint = LocalizedStringResource("promote.sheet.status.hint", defaultValue: "Provisional and rejected arrive in a later spike", comment: "Hint under claim status; only Accepted can be filed")
        static func argumentHint(pinned: Int) -> String {
            L10n.format(LocalizedStringResource("promote.sheet.argument.hint", defaultValue: "Drafted from the %lld pinned comparisons · edit before filing", comment: "Hint under the claim argument; argument is how many comparisons are pinned"), pinned)
        }
        static func sheetSubtitle(assessment: String, reason: String) -> String {
            L10n.format(LocalizedStringResource("promote.sheet.subtitle", defaultValue: "%1$@ · %2$@. Agreeing comparisons are pinned as this claim’s evidence.", comment: "Evidence sheet subtitle; 1 = assessment, 2 = why this row matched"), assessment, reason)
        }
        static func through(_ neighbor: String) -> String {
            L10n.format(LocalizedStringResource("promote.sheet.through", defaultValue: "Through %@", comment: "Evidence group reached through a neighbor; argument is the neighbor's name"), neighbor)
        }
        static let conflictCalloutTitle = LocalizedStringResource("promote.sheet.conflict.title", defaultValue: "A comparison disagrees", comment: "Evidence sheet callout when one comparison conflicts")
        static let conflictCalloutBody = LocalizedStringResource("promote.sheet.conflict.body", defaultValue: "The conflicting comparison is not pinned. The claim can still be filed.", comment: "Evidence sheet callout body for a conflict")
        static func newOption(_ kind: EvidencePrimaryKind) -> LocalizedStringResource {
            let resource: LocalizedStringResource = switch kind {
            case .person: LocalizedStringResource("promote.new.person", defaultValue: "New Person", comment: "Promote target choice: mint a new handle")
            case .event: LocalizedStringResource("promote.new.event", defaultValue: "New Event", comment: "Promote target choice: mint a new handle")
            case .place: LocalizedStringResource("promote.new.place", defaultValue: "New Place", comment: "Promote target choice: mint a new handle")
            }
            return resource
        }
        static func leaveTitle(name: String) -> String {
            L10n.format(LocalizedStringResource("promote.leave.title", defaultValue: "Leave Promote without filing %@?", comment: "Promote leave guard title; argument is the subject's name or ref"), name)
        }
        static let leaveConfirm = LocalizedStringResource("promote.leave.confirm", defaultValue: "Discard changes", comment: "Promote leave guard: leave and drop the mapping")
        static let leaveCancel = LocalizedStringResource("promote.leave.cancel", defaultValue: "Keep mapping", comment: "Promote leave guard: stay on the page")
        static let statusAccepted = LocalizedStringResource("promote.claim.status.accepted", defaultValue: "Accepted", comment: "Promote claim status option: the subject is this handle")
        static func breadcrumb(ref: String) -> String {
            L10n.format(LocalizedStringResource("promote.breadcrumb", defaultValue: "Promote %@", comment: "Toolbar breadcrumb for the Promote place; argument is the subject ref"), ref)
        }
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
        static let editPlaceRelationshipTitle = LocalizedStringResource(
            "evidenceGraph.edit.placeRelationshipTitle",
            defaultValue: "Edit place relationship",
            comment: "Edit-subject dialog title for a place_relationship bridge"
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

        /// VoiceOver for card trash: "Delete {kind} {label or sentence}, {ref}".
        static func deleteAccessibility(kind: String, label: String, ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.subject.deleteAccessibility",
                defaultValue: "Delete %1$@ %2$@, %3$@",
                comment: "VoiceOver for graph card trash; arguments are kind, label or sentence, catalog ref"
            ), kind, label, ref)
        }

        static let addProperty = LocalizedStringResource(
            "evidenceGraph.subject.addProperty",
            defaultValue: "Add property",
            comment: "Quiet footer control on primary cards that opens the citation composer"
        )

        static let addPropertyUnavailable = LocalizedStringResource(
            "evidenceGraph.subject.addPropertyUnavailable",
            defaultValue: "Add property, unavailable — this Source has no Artifacts",
            comment: "VoiceOver when Add property is disabled because the Source has no Artifact"
        )

        // MARK: Promote + membership footer (S9-D8 / S9-04)

        static let promote = LocalizedStringResource(
            "evidenceGraph.subject.promote",
            defaultValue: "Promote",
            comment: "Footer button on an unpromoted primary card that creates its Person, Event or Place"
        )

        static func promoteAccessibility(label: String, ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.subject.promoteAccessibility",
                defaultValue: "Promote %1$@, %2$@",
                comment: "VoiceOver for the Promote footer; arguments are the subject label and its candidate ref"
            ), label, ref)
        }

        static func promoteHelp(kind: EvidencePrimaryKind) -> LocalizedStringResource {
            switch kind {
            case .person:
                LocalizedStringResource(
                    "evidenceGraph.subject.promoteHelp.person",
                    defaultValue: "Promote to a person",
                    comment: "Tooltip on the Promote footer button"
                )
            case .event:
                LocalizedStringResource(
                    "evidenceGraph.subject.promoteHelp.event",
                    defaultValue: "Promote to an event",
                    comment: "Tooltip on the Promote footer button"
                )
            case .place:
                LocalizedStringResource(
                    "evidenceGraph.subject.promoteHelp.place",
                    defaultValue: "Promote to a place",
                    comment: "Tooltip on the Promote footer button"
                )
            }
        }

        static func openHandlePage(kind: EvidencePrimaryKind) -> LocalizedStringResource {
            switch kind {
            case .person:
                LocalizedStringResource(
                    "evidenceGraph.subject.openHandlePage.person",
                    defaultValue: "Open person page",
                    comment: "Membership row text beside the handle ref until the auto-reconciled name lands (S9-09)"
                )
            case .event:
                LocalizedStringResource(
                    "evidenceGraph.subject.openHandlePage.event",
                    defaultValue: "Open event page",
                    comment: "Membership row text beside the handle ref until the auto-reconciled name lands (S9-09)"
                )
            case .place:
                LocalizedStringResource(
                    "evidenceGraph.subject.openHandlePage.place",
                    defaultValue: "Open place page",
                    comment: "Membership row text beside the handle ref until the auto-reconciled name lands (S9-09)"
                )
            }
        }

        static func openHandleAccessibility(kind: EvidencePrimaryKind, ref: String) -> String {
            let resource: LocalizedStringResource = switch kind {
            case .person:
                LocalizedStringResource(
                    "evidenceGraph.subject.openHandleAccessibility.person",
                    defaultValue: "Open person %@",
                    comment: "VoiceOver action on a promoted card that opens its handle page; argument is the handle ref"
                )
            case .event:
                LocalizedStringResource(
                    "evidenceGraph.subject.openHandleAccessibility.event",
                    defaultValue: "Open event %@",
                    comment: "VoiceOver action on a promoted card that opens its handle page; argument is the handle ref"
                )
            case .place:
                LocalizedStringResource(
                    "evidenceGraph.subject.openHandleAccessibility.place",
                    defaultValue: "Open place %@",
                    comment: "VoiceOver action on a promoted card that opens its handle page; argument is the handle ref"
                )
            }
            return L10n.format(resource, ref)
        }

        static let promoteAll = LocalizedStringResource(
            "evidenceGraph.header.promoteAll",
            defaultValue: "Promote all",
            comment: "Secondary header button that opens Promote with every subject expanded; also the breadcrumb for that entry"
        )

        static let jumpToSourcePage = LocalizedStringResource(
            "evidenceGraph.header.jumpToSourcePage",
            defaultValue: "Jump to Source page",
            comment: "Secondary header button that opens this Source’s filing page"
        )

        static func jumpToSourcePageAccessibility(title: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.header.jumpToSourcePageAccessibility",
                defaultValue: "Jump to Source page — %@",
                comment: "VoiceOver for the header jump; argument is the Source title"
            ), title)
        }

        static let negatedPrefix = LocalizedStringResource(
            "evidenceGraph.row.negatedPrefix",
            defaultValue: "Not",
            comment: "Micro-caps prefix before a negative Observation value on a cited graph row"
        )

        static let negatedAccessibility = LocalizedStringResource(
            "evidenceGraph.row.negatedAccessibility",
            defaultValue: "Negated",
            comment: "VoiceOver mark when a cited graph row has polarity negative"
        )

        static func conflictOneOf(count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.row.conflictOneOf",
                defaultValue: "one of %lld values",
                comment: "VoiceOver fragment when a cited row shares its Property; argument is how many"
            ), count)
        }

        static let citedRowEditHint = LocalizedStringResource(
            "evidenceGraph.row.editHint",
            defaultValue: "Edit property",
            comment: "Trailing VoiceOver hint on a cited graph row that opens the composer"
        )

        static let legendConflict = LocalizedStringResource(
            "evidenceGraph.legend.conflict",
            defaultValue: "Conflict — competing values",
            comment: "Canvas legend line for the ochre conflict bracket"
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
        static let addPlaceRelationshipTitle = LocalizedStringResource(
            "evidenceGraph.create.placeRelationshipTitle",
            defaultValue: "Add place relationship",
            comment: "Create-bridge dialog title for a place_relationship mid-card"
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
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.summary.location",
                defaultValue: "%1$@ took place in %2$@",
                comment: "Location edge summary; arguments are event label then place label"
            ), event, place)
        }

        /// Location mid-phrase when endpoint labels are incomplete.
        static let bridgeSummaryLocationBare = LocalizedStringResource(
            "evidenceGraph.bridge.summary.locationBare",
            defaultValue: "Took place in",
            comment: "Location edge summary when event/place endpoint labels are missing"
        )

        /// Relationship: "{person} is the {type} of {related_to}".
        static func bridgeSummaryRelationship(person: String, type: String, related: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.summary.relationship",
                defaultValue: "%1$@ is the %2$@ of %3$@",
                comment: "Relationship edge summary; arguments are person, relationship_type, related_to"
            ), person, type, related)
        }

        /// Relationship without type: "{person} is related to {related_to}".
        static func bridgeSummaryRelationshipFallback(person: String, related: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.summary.relationshipFallback",
                defaultValue: "%1$@ is related to %2$@",
                comment: "Relationship edge summary without relationship_type; person then related_to"
            ), person, related)
        }

        /// Relationship mid-phrase when only the type term is known.
        static func bridgeSummaryRelationshipTypeOnly(type: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.summary.relationshipTypeOnly",
                defaultValue: "Is the %@ of",
                comment: "Relationship edge summary when endpoint labels are missing; argument is type"
            ), type)
        }

        static let bridgeSummaryRelationshipBare = LocalizedStringResource(
            "evidenceGraph.bridge.summary.relationshipBare",
            defaultValue: "Is related to",
            comment: "Relationship edge summary when type and endpoint labels are missing"
        )

        /// Place relationship: "{from} is {type} {to}".
        static func bridgeSummaryPlaceRelationship(from: String, type: String, to: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.summary.placeRelationship",
                defaultValue: "%1$@ is %2$@ %3$@",
                comment: "Place relationship edge summary; arguments are from place, type, to place"
            ), from, type, to)
        }

        static func bridgeSummaryPlaceRelationshipFallback(from: String, to: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.summary.placeRelationshipFallback",
                defaultValue: "%1$@ relates to %2$@",
                comment: "Place relationship edge summary without type; from then to"
            ), from, to)
        }

        static func bridgeSummaryPlaceRelationshipTypeOnly(type: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.summary.placeRelationshipTypeOnly",
                defaultValue: "Is %@ of",
                comment: "Place relationship mid-phrase when endpoints missing; argument is type"
            ), type)
        }

        static let bridgeSummaryPlaceRelationshipBare = LocalizedStringResource(
            "evidenceGraph.bridge.summary.placeRelationshipBare",
            defaultValue: "Relates to",
            comment: "Place relationship edge summary when type and endpoints are missing"
        )

        /// Participation with role: "{person} participated as {role} at {event}".
        static func bridgeSummaryParticipation(person: String, role: String, event: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.summary.participation",
                defaultValue: "%1$@ participated as %2$@ at %3$@",
                comment: "Participation edge summary; arguments are person, role, event"
            ), person, role, event)
        }

        /// Participation without role: "{person} participated in {event}".
        static func bridgeSummaryParticipationFallback(person: String, event: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.summary.participationFallback",
                defaultValue: "%1$@ participated in %2$@",
                comment: "Participation edge summary without role; arguments are person then event"
            ), person, event)
        }

        /// Participation mid-phrase when only the role term is known.
        static func bridgeSummaryParticipationRoleOnly(role: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.summary.participationRoleOnly",
                defaultValue: "Participated as %@",
                comment: "Participation edge summary when endpoint labels are missing; argument is role"
            ), role)
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
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.nounTypeAndRef",
                defaultValue: "%1$@ %2$@",
                comment: "Bridge endpoint noun when the working label is blank; type label then ref"
            ), type, ref)
        }

        /// Whole-name fallback: "{kind phrase} · {ref}".
        static func bridgeNameKindAndRef(phrase: String, ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "evidenceGraph.bridge.nameKindAndRef",
                defaultValue: "%1$@ · %2$@",
                comment: "Bridge name when edges cannot be read; kind phrase then bridge ref"
            ), phrase, ref)
        }

        static func subjectCount(count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "evidenceGraph.header.subjectCount",
                defaultValue: "%lld subjects",
                comment: "Evidence graph header count; argument is how many placed primaries are on the canvas"
            ), count)
        }
    }

    /// Product Property-term display names (`propertyTerm.<propertyKey>.<termKey>`).
    enum PropertyTerm {
        static func resource(propertyKey: String, termKey: String) -> LocalizedStringResource? {
            switch (propertyKey, termKey) {
            case ("event_type", "birth"): return eventTypeBirth
            case ("event_type", "death"): return eventTypeDeath
            case ("event_type", "marriage"): return eventTypeMarriage
            case ("event_type", "baptism"): return eventTypeBaptism
            case ("event_type", "burial"): return eventTypeBurial
            case ("event_type", "census"): return eventTypeCensus
            case ("event_type", "residence"): return eventTypeResidence
            case ("event_type", "migration"): return eventTypeMigration
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
            case ("place_relationship_type", "part_of"): return placeRelationshipPartOf
            case ("place_relationship_type", "succeeded_by"): return placeRelationshipSucceededBy
            default: return nil
            }
        }

        static let eventTypeBirth = LocalizedStringResource(
            "propertyTerm.event_type.birth",
            defaultValue: "Birth",
            comment: "Product event_type term: birth"
        )
        static let eventTypeDeath = LocalizedStringResource(
            "propertyTerm.event_type.death",
            defaultValue: "Death",
            comment: "Product event_type term: death"
        )
        static let eventTypeMarriage = LocalizedStringResource(
            "propertyTerm.event_type.marriage",
            defaultValue: "Marriage",
            comment: "Product event_type term: marriage"
        )
        static let eventTypeBaptism = LocalizedStringResource(
            "propertyTerm.event_type.baptism",
            defaultValue: "Baptism",
            comment: "Product event_type term: baptism"
        )
        static let eventTypeBurial = LocalizedStringResource(
            "propertyTerm.event_type.burial",
            defaultValue: "Burial",
            comment: "Product event_type term: burial"
        )
        static let eventTypeCensus = LocalizedStringResource(
            "propertyTerm.event_type.census",
            defaultValue: "Census Enumeration",
            comment: "Product event_type term: census"
        )
        static let eventTypeResidence = LocalizedStringResource(
            "propertyTerm.event_type.residence",
            defaultValue: "Residence",
            comment: "Product event_type term: residence"
        )
        static let eventTypeMigration = LocalizedStringResource(
            "propertyTerm.event_type.migration",
            defaultValue: "Migration",
            comment: "Product event_type term: migration"
        )
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
        static let placeRelationshipPartOf = LocalizedStringResource(
            "propertyTerm.place_relationship_type.part_of",
            defaultValue: "Part of",
            comment: "Product place_relationship_type term: part_of"
        )
        static let placeRelationshipSucceededBy = LocalizedStringResource(
            "propertyTerm.place_relationship_type.succeeded_by",
            defaultValue: "Succeeded by",
            comment: "Product place_relationship_type term: succeeded_by"
        )
    }

    /// Shared genealogical DateValue display (list rows, previews). Not Source-page-owned.
    enum Dates {
        static func displayAbout(_ date: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.display.about",
                defaultValue: "About %@",
                comment: "DateValue summary prefix for ABT; argument is the formatted point date"
            )
            return L10n.format(resource, locale: locale, date)
        }

        static func displayBefore(_ date: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.display.before",
                defaultValue: "Before %@",
                comment: "DateValue summary prefix for BEF; argument is the formatted point date"
            )
            return L10n.format(resource, locale: locale, date)
        }

        static func displayAfter(_ date: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.display.after",
                defaultValue: "After %@",
                comment: "DateValue summary prefix for AFT; argument is the formatted point date"
            )
            return L10n.format(resource, locale: locale, date)
        }

        static func displayBetween(start: String, end: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.display.between",
                defaultValue: "Between %1$@ and %2$@",
                comment: "DateValue summary for a range; arguments are formatted start then end"
            )
            return L10n.format(resource, locale: locale, start, end)
        }

        static func rowAbout(_ date: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.row.about",
                defaultValue: "abt %@",
                comment: "Events list date prefix for ABT; argument is the formatted point date"
            )
            return L10n.format(resource, locale: locale, date)
        }

        static func rowBefore(_ date: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.row.before",
                defaultValue: "bef %@",
                comment: "Events list date prefix for BEF; argument is the formatted point date"
            )
            return L10n.format(resource, locale: locale, date)
        }

        static func rowAfter(_ date: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.row.after",
                defaultValue: "aft %@",
                comment: "Events list date prefix for AFT; argument is the formatted point date"
            )
            return L10n.format(resource, locale: locale, date)
        }

        static func rowBetween(start: String, end: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.row.between",
                defaultValue: "bet %1$@ and %2$@",
                comment: "Events list date for a range; arguments are the formatted start and end"
            )
            return L10n.format(resource, locale: locale, start, end)
        }

        static func rowSpan(start: String, end: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.row.span",
                defaultValue: "%1$@–%2$@",
                comment: "Events list date when an event has a start and an end and no single date. Arguments are the two formatted dates. The separator is an en dash."
            )
            return L10n.format(resource, locale: locale, start, end)
        }

        static func rowDayMonthYear(day: Int32, month: String, year: Int32, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.row.dayMonthYear",
                defaultValue: "%1$@ %2$@ %3$@",
                comment: "Row date with day, abbreviated month, and year (14 May 1817). Day and year are digits, never grouped. Reorder for your locale."
            )
            return L10n.format(resource, locale: locale, String(day), month, String(year))
        }

        static func rowMonthYear(month: String, year: Int32, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.row.monthYear",
                defaultValue: "%1$@ %2$@",
                comment: "Row date with abbreviated month and year (Mar 1790). The year is digits, never grouped."
            )
            return L10n.format(resource, locale: locale, month, String(year))
        }

        static func rowDayMonth(day: Int32, month: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.row.dayMonth",
                defaultValue: "%1$@ %2$@",
                comment: "Row date with day and abbreviated month, no year (14 May)."
            )
            return L10n.format(resource, locale: locale, String(day), month)
        }

        static func rowDayYear(day: Int32, year: Int32, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.row.dayYear",
                defaultValue: "%1$@ %2$@",
                comment: "Row date with a day and a year but no month. The year is digits, never grouped."
            )
            return L10n.format(resource, locale: locale, String(day), String(year))
        }

        static func rowWithPhrase(_ date: String, phrase: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "dates.row.withPhrase",
                defaultValue: "%1$@ · %2$@",
                comment: "Row date followed by its recorded phrase. Arguments are the formatted date and the phrase."
            )
            return L10n.format(resource, locale: locale, date, phrase)
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
                return L10n.string(LocalizedStringResource(
                    "nameValue.parts.count.none",
                    defaultValue: "none",
                    comment: "Parts heading count when the NameValue has no parts"
                ))
            }
            return L10n.format(LocalizedStringResource(
                "nameValue.parts.count",
                defaultValue: "%lld parts",
                comment: "Parts heading count; argument is the part count"
            ), count)
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
            return L10n.format(LocalizedStringResource(
                "nameValue.part.error.emptyValue",
                defaultValue: "Part %lld cannot be empty — remove it instead.",
                comment: "Validation when a NameValue part value is blank; argument is 1-based index"
            ), position)
        }

        static func partAccessibility(position: Int, of count: Int, typeLabel: String) -> String {
            return L10n.format(LocalizedStringResource(
                "a11y.nameValue.part",
                defaultValue: "Part %1$lld of %2$lld, %3$@",
                comment: "VoiceOver name for a NameValue part row; arguments are index, count, type label"
            ), position, count, typeLabel)
        }

        static func partMoveUp(position: Int) -> String {
            L10n.format(LocalizedStringResource(
                "a11y.nameValue.part.moveUp",
                defaultValue: "Move part %lld up",
                comment: "VoiceOver for move-up; argument is 1-based part index"
            ), position)
        }

        static func partMoveDown(position: Int) -> String {
            L10n.format(LocalizedStringResource(
                "a11y.nameValue.part.moveDown",
                defaultValue: "Move part %lld down",
                comment: "VoiceOver for move-down; argument is 1-based part index"
            ), position)
        }

        static func partRemove(position: Int) -> String {
            L10n.format(LocalizedStringResource(
                "a11y.nameValue.part.remove",
                defaultValue: "Remove part %lld",
                comment: "VoiceOver for remove; argument is 1-based part index"
            ), position)
        }

        static func partMoved(position: Int, of count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "a11y.nameValue.part.moved",
                defaultValue: "Part moved to position %1$lld of %2$lld",
                comment: "Announcement after Option-arrow reorder; arguments are new index and count"
            ), position, count)
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
            return L10n.format(LocalizedStringResource(
                "artifactViewer.pageOf",
                defaultValue: "of %lld",
                comment: "Artifact viewer page count suffix; argument is total pages"
            ), total)
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

        static func zoomPercent(percent: Int) -> String {
            L10n.format(LocalizedStringResource(
                "artifactViewer.zoomPercent",
                defaultValue: "Zoom %lld percent",
                comment: "VoiceOver label for the live zoom percentage"
            ), percent)
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

        static func pageSet(page: Int) -> String {
            L10n.format(LocalizedStringResource(
                "artifactViewer.pageSet",
                defaultValue: "Page %lld set",
                comment: "Set page button after the locator page matches the viewer; argument is page number"
            ), page)
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

        static let findToggle = LocalizedStringResource(
            "artifactViewer.findToggle",
            defaultValue: "Find in PDF",
            comment: "Opens PDF Find on the citation composer tool strip; also the find field prompt"
        )

        static let findUnavailable = LocalizedStringResource(
            "artifactViewer.findUnavailable",
            defaultValue: "Find unavailable, no text layer",
            comment: "VoiceOver label for PDF Find when the document has no extractable text"
        )

        static let findField = LocalizedStringResource(
            "artifactViewer.findField",
            defaultValue: "Find in PDF",
            comment: "Accessibility label for the PDF Find keyword field"
        )

        static let findPrevious = LocalizedStringResource(
            "artifactViewer.findPrevious",
            defaultValue: "Previous match",
            comment: "PDF Find: go to the previous matching string"
        )

        static let findNext = LocalizedStringResource(
            "artifactViewer.findNext",
            defaultValue: "Next match",
            comment: "PDF Find: go to the next matching string"
        )

        static let findDone = LocalizedStringResource(
            "artifactViewer.findDone",
            defaultValue: "Done",
            comment: "Closes the PDF Find row"
        )

        static func findMatchOf(current: Int, total: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "artifactViewer.findMatchOf",
                defaultValue: "%1$lld of %2$lld",
                comment: "Find field suffix; arguments are current match index then total matches"
            ), current, total)
        }

        static let findNoteNoTextLayer = LocalizedStringResource(
            "artifactViewer.findNoteNoTextLayer",
            defaultValue: "This PDF has no text layer — it’s a scanned image, so there is nothing to find",
            comment: "PDF Find note when every page has empty extractable text"
        )

        static func findNoteNoMatches(pageCount: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "artifactViewer.findNoteNoMatches",
                defaultValue: "No matches in %lld pages",
                comment: "PDF Find note when the query hits nothing; argument is page count"
            ), pageCount)
        }

        static func findNoteJumped(page: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "artifactViewer.findNoteJumped",
                defaultValue: "Jumped to page %lld.",
                comment: "PDF Find note after the active hit changes the viewer page"
            ), page)
        }

        static func findNoteNextPage(page: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "artifactViewer.findNoteNextPage",
                defaultValue: "Next match is on page %lld.",
                comment: "PDF Find note when the following hit is on another page"
            ), page)
        }

        static func findNoteWrappedFirst(page: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "artifactViewer.findNoteWrappedFirst",
                defaultValue: "Wrapped to the first match, page %lld",
                comment: "PDF Find note after next wraps from the last hit to the first"
            ), page)
        }

        static func findNoteWrappedLast(page: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "artifactViewer.findNoteWrappedLast",
                defaultValue: "Wrapped to the last match, page %lld",
                comment: "PDF Find note after previous wraps from the first hit to the last"
            ), page)
        }
    }

    /// Citation composer place (S7-08 thin submit path; board-aligned shell).
    enum CitationComposer {
        static func breadcrumbCitationFor(scope: String) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.breadcrumb.citationFor",
                defaultValue: "Citation for %@",
                comment: "Composer breadcrumb leaf; argument is subject label or bridge edge sentence"
            ), scope)
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

        static let deleteCitation = LocalizedStringResource(
            "citationComposer.deleteCitation",
            defaultValue: "Delete citation",
            comment: "Citation fields button that presents DeleteImpact for the saved Citation"
        )

        static func deleteCitationAccessibility(ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.deleteCitationAccessibility",
                defaultValue: "Delete citation %@",
                comment: "VoiceOver for Delete citation; argument is CIT-…"
            ), ref)
        }

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

        static let negatedBadge = LocalizedStringResource(
            "citationComposer.observation.negatedBadge",
            defaultValue: "Not",
            comment: "Danger badge left of a negated Observation value; PVBadge uppercases it"
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

        static let connectionPlaceRelationship = LocalizedStringResource(
            "citationComposer.connectionPlaceRelationship",
            defaultValue: "Type",
            comment: "Term picker label on a place_relationship connection row (part of / succeeded by)"
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
            return L10n.format(LocalizedStringResource(
                "citationComposer.newSubjectOption",
                defaultValue: "New %@…",
                comment: "Subject picker option; argument is person, event, or place"
            ), typeKey)
        }

        static func unsavedSummary(
            citationDirty: Bool,
            observationCount: Int,
            connectionTouched: Bool
        ) -> String {
            var parts: [String] = []
            if citationDirty {
                parts.append(L10n.string(LocalizedStringResource(
                    "citationComposer.unsavedCitationPart",
                    defaultValue: "the citation",
                    comment: "Unsaved-summary clause for dirty citation fields"
                )))
            }
            if observationCount > 0 {
                parts.append(L10n.format(LocalizedStringResource(
                    "citationComposer.unsavedObservations",
                    defaultValue: "%lld observations",
                    comment: "Unsaved-summary clause for dirty observation rows; argument is how many"
                ), observationCount))
            }
            if connectionTouched {
                parts.append(L10n.string(LocalizedStringResource(
                    "citationComposer.unsavedConnectionPart",
                    defaultValue: "a new connection",
                    comment: "Unsaved-summary clause for a touched pending connection"
                )))
            }
            let joined = parts.joined(separator: ", ")
            return L10n.format(LocalizedStringResource(
                "citationComposer.unsavedSummary",
                defaultValue: "Unsaved: %@",
                comment: "Footer and leave-guard body; argument is the joined unsaved parts"
            ), joined)
        }

        static func connectionAccessibility(
            sentence: String,
            termLabel: String,
            term: String,
            status: String
        ) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.connectionAccessibility",
                defaultValue: "Connection, %1$@. %2$@ %3$@. %4$@",
                comment: "VoiceOver for a relationship or participation connection"
            ), sentence, termLabel, term, status)
        }

        static func connectionAccessibilityLocation(status: String) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.connectionAccessibilityLocation",
                defaultValue: "Connection, Location. %@",
                comment: "VoiceOver for a location connection; argument is New/Edited/Saved ref"
            ), status)
        }

        static func connectionSavedStatus(ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.connectionSavedStatus",
                defaultValue: "Saved, %@",
                comment: "VoiceOver status for a saved connection; argument is the bridge ref"
            ), ref)
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
        static let untitledCitation = LocalizedStringResource(
            "citationComposer.untitledCitation",
            defaultValue: "Untitled citation",
            comment: "DeleteImpact title when the citation has no transcription"
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

        static let autoTranscribe = LocalizedStringResource(
            "citationComposer.autoTranscribe",
            defaultValue: "Auto transcribe",
            comment: "Button that fills transcription from Vision OCR of the image"
        )

        static let autoTranscribeReading = LocalizedStringResource(
            "citationComposer.autoTranscribeReading",
            defaultValue: "Reading text…",
            comment: "Busy label on Auto transcribe while OCR is running"
        )

        static let autoTranscribeDrawnRegion = LocalizedStringResource(
            "citationComposer.autoTranscribeDrawnRegion",
            defaultValue: "Auto transcribe, drawn region",
            comment: "VoiceOver name when Auto transcribe will read the locator region"
        )

        static let autoTranscribeHintWholeImage = LocalizedStringResource(
            "citationComposer.autoTranscribeHintWholeImage",
            defaultValue: "Reads the whole image. Draw a region to transcribe a crop.",
            comment: "Hint under transcription when OCR will read the full image"
        )

        static let autoTranscribeHintRegion = LocalizedStringResource(
            "citationComposer.autoTranscribeHintRegion",
            defaultValue: "Reads the drawn region.",
            comment: "Hint under transcription when a locator region will be cropped"
        )

        static let pasteTranscription = LocalizedStringResource(
            "citationComposer.pasteTranscription",
            defaultValue: "Paste transcription from selection",
            comment: "Button that copies the PDF I-beam selection into transcription"
        )

        static let pasteHintSelect = LocalizedStringResource(
            "citationComposer.pasteHintSelect",
            defaultValue: "Select text on the page, then paste it here",
            comment: "Hint when the PDF has a text layer but no I-beam selection"
        )

        static let pasteHintNoTextLayer = LocalizedStringResource(
            "citationComposer.pasteHintNoTextLayer",
            defaultValue: "This PDF has no text layer, so there is no text to select",
            comment: "Hint when every PDF page has empty extractable text"
        )

        static func pasteHintSelected(lines: Int, page: Int) -> String {
            L10n.format(LocalizedStringResource(
                "citationComposer.pasteHintSelected",
                defaultValue: "Pastes the %1$lld lines selected on page %2$lld",
                comment: "Hint when an I-beam selection can be pasted; arguments are line count and page"
            ), lines, page)
        }

        static func pasteHintAfter(page: Int) -> String {
            L10n.format(LocalizedStringResource(
                "citationComposer.pasteHintAfter",
                defaultValue: "Pasted from page %lld. Check it against the page, then Save citation",
                comment: "Hint after a successful paste; argument is the selection page"
            ), page)
        }

        static func pasteUnavailable(hint: String) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.pasteUnavailable",
                defaultValue: "Paste transcription from selection, unavailable. %@",
                comment: "VoiceOver when Paste is disabled; argument is the current field hint"
            ), hint)
        }

        static let pasteReplaceConfirm = LocalizedStringResource(
            "citationComposer.pasteReplaceConfirm",
            defaultValue: "Replace",
            comment: "Confirm button that overwrites transcription with the PDF selection"
        )

        static func pasteReplaceMessage(selectedLines: Int, page: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.pasteReplaceMessage",
                defaultValue: "Paste replaces the text already in the field with the %1$lld lines selected on page %2$lld. The current text isn’t kept. Nothing is written until Save citation.",
                comment: "Confirm message when paste would overwrite; arguments are selected line count then page"
            ), selectedLines, page)
        }

        static let autoTranscribeHintAudio = LocalizedStringResource(
            "citationComposer.autoTranscribeHintAudio",
            defaultValue: "Audio cannot be read as text.",
            comment: "Disabled hint when the Artifact is audio"
        )

        static let autoTranscribeHintVideo = LocalizedStringResource(
            "citationComposer.autoTranscribeHintVideo",
            defaultValue: "Video cannot be read as text.",
            comment: "Disabled hint when the Artifact is video"
        )

        static let autoTranscribeHintMissingFile = LocalizedStringResource(
            "citationComposer.autoTranscribeHintMissingFile",
            defaultValue: "There is no file to read.",
            comment: "Disabled hint when the Artifact file is missing"
        )

        static let autoTranscribeHintNoRaster = LocalizedStringResource(
            "citationComposer.autoTranscribeHintNoRaster",
            defaultValue: "There is no image to read.",
            comment: "Disabled hint when the image raster failed to load"
        )

        static let autoTranscribeReplaceTitle = LocalizedStringResource(
            "citationComposer.autoTranscribeReplaceTitle",
            defaultValue: "Replace the transcription?",
            comment: "Confirm title when Auto transcribe would overwrite existing text"
        )

        static let autoTranscribeReplaceMessage = LocalizedStringResource(
            "citationComposer.autoTranscribeReplaceMessage",
            defaultValue: "Auto transcribe will replace the text in this field. Nothing is written until you save the citation.",
            comment: "Confirm message when replacing a non-empty transcription"
        )

        static let autoTranscribeWholePageTitle = LocalizedStringResource(
            "citationComposer.autoTranscribeWholePageTitle",
            defaultValue: "Transcribe the whole page?",
            comment: "Confirm title when OCR would read a large image with no region"
        )

        static let autoTranscribeWholePageMessage = LocalizedStringResource(
            "citationComposer.autoTranscribeWholePageMessage",
            defaultValue: "This image is large and has no region. Reading the whole page can be slow and inaccurate. Draw a region, or continue.",
            comment: "Confirm message for a large no-region Auto transcribe job"
        )

        static let autoTranscribeReplaceAndWholePageMessage = LocalizedStringResource(
            "citationComposer.autoTranscribeReplaceAndWholePageMessage",
            defaultValue: "Auto transcribe will replace the text in this field. This image is large and has no region, so the whole page will be read — that can be slow and inaccurate. Nothing is written until you save the citation.",
            comment: "Confirm message when replacing text and reading a large whole page"
        )

        static let autoTranscribeKeep = LocalizedStringResource(
            "citationComposer.autoTranscribeKeep",
            defaultValue: "Keep",
            comment: "Cancel Auto transcribe confirm and leave the existing transcription"
        )

        static let autoTranscribeConfirm = LocalizedStringResource(
            "citationComposer.autoTranscribeConfirm",
            defaultValue: "Transcribe",
            comment: "Confirm button that starts Auto transcribe after a warning"
        )

        static let autoTranscribeFailed = LocalizedStringResource(
            "citationComposer.autoTranscribeFailed",
            defaultValue: "Provenencia couldn’t read this image. The transcription is unchanged.",
            comment: "Callout when Vision OCR fails"
        )

        static let autoTranscribeNothingFound = LocalizedStringResource(
            "citationComposer.autoTranscribeNothingFound",
            defaultValue: "Nothing was found in this image. The transcription is unchanged.",
            comment: "Callout when OCR returns no text"
        )

        static let autoTranscribeDismiss = LocalizedStringResource(
            "citationComposer.autoTranscribeDismiss",
            defaultValue: "Dismiss",
            comment: "Action that clears the Auto transcribe failure callout"
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
            return L10n.format(LocalizedStringResource(
                "citationComposer.artifactIndexOf",
                defaultValue: "Artifact %1$lld of %2$lld · %3$@",
                comment: "Viewer header index; arguments are 1-based index, total, and media kind"
            ), index, total, kind)
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
            return L10n.format(LocalizedStringResource(
                "citationComposer.mediaCaptionNoFileDetail",
                defaultValue: "No file · %@",
                comment: "Caption under a fileless Artifact tile; argument is the Artifact description"
            ), detail)
        }

        static func locatorPage(_ page: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.locatorPage",
                defaultValue: "Page %lld",
                comment: "Locator list row for a committed page; argument is page number"
            ), page)
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
            return L10n.format(LocalizedStringResource(
                "citationComposer.locatorPoints",
                defaultValue: "%lld points",
                comment: "Locator list helper for a region; argument is vertex count"
            ), count)
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
            return L10n.format(LocalizedStringResource(
                "citationComposer.removePage",
                defaultValue: "Remove page %lld",
                comment: "Accessibility name to remove a page locator; argument is page number"
            ), page)
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

        static let citationMissing = LocalizedStringResource(
            "citationComposer.citationMissing",
            defaultValue: "That citation no longer exists. It may have been deleted.",
            comment: "Neutral callout when an edit link or history entry points at a missing citation"
        )

        static let catalogUnavailable = LocalizedStringResource(
            "citationComposer.catalogUnavailable",
            defaultValue: "The catalog did not finish loading. Go back to the evidence graph and try again.",
            comment: "Load-failed callout when composer catalog query handles never became ready"
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
            return L10n.format(LocalizedStringResource(
                "citationComposer.artifactMenuLabel",
                defaultValue: "Artifact, %@",
                comment: "VoiceOver name for the Artifact identity control; argument is Artifact title"
            ), title)
        }

        static func artifactMenuMeta(kind: String, count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.artifactMenuMeta",
                defaultValue: "%1$@ · %2$lld citations",
                comment: "Artifact menu row meta; arguments are media kind and citation count"
            ), kind, count)
        }

        static func citationMenuCount(count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.citationMenuCount",
                defaultValue: "%lld obs",
                comment: "Observation count on a Citation menu row; argument is count"
            ), count)
        }

        static func citationMenuRef(ref: String, count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.citationMenuRef",
                defaultValue: "Citation, %1$@, %2$lld observations",
                comment: "VoiceOver name for the Citation identity control; arguments are ref and count"
            ), ref, count)
        }

        static func identityChangedCitation(ref: String, count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.identityChangedCitation",
                defaultValue: "Now citing %1$@, %2$lld observations",
                comment: "VoiceOver announcement after switching Citation; arguments are ref and count"
            ), ref, count)
        }

        static func identityChangedArtifact(title: String) -> String {
            return L10n.format(LocalizedStringResource(
                "citationComposer.identityChangedArtifact",
                defaultValue: "Now citing %@",
                comment: "VoiceOver announcement after switching Artifact; argument is Artifact title"
            ), title)
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

        static func pluginPill(pluginID: String) -> String {
            L10n.format(LocalizedStringResource(
                "origin.pill.plugin",
                defaultValue: "Supplied by the %@ plugin",
                comment: "Accessibility label and tooltip for the pill marking a plugin-owned row in a vocabulary list; argument is the plugin id"
            ), pluginID)
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
            return L10n.format(LocalizedStringResource(
                "sources.list.countLine",
                defaultValue: "%lld sources",
                comment: "Sources list count when unfiltered; argument is total"
            ), total)
        }

        static func countLineFiltered(visible: Int, total: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "sources.list.countLineFiltered",
                defaultValue: "%1$lld of %2$lld sources",
                comment: "Sources list count when search/filter narrows the list; arguments are visible then total"
            ), visible, total)
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
            return L10n.format(LocalizedStringResource(
                "sources.list.sortedBy",
                defaultValue: "Sorted by %@",
                comment: "Sources sort control label; argument is the active sort option in lowercase"
            ), label)
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

        static let graphNotStarted = LocalizedStringResource(
            "sources.list.graphNotStarted",
            defaultValue: "not started",
            comment: "Italic graph-zone state when a Source has zero canvas subjects"
        )

        static let graphCountsLoading = LocalizedStringResource(
            "sources.list.graphCountsLoading",
            defaultValue: "counts loading",
            comment: "VoiceOver while Sources list graph-progress counts have not arrived"
        )

        static func graphSubjectCount(_ count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "sources.list.subjectCount",
                defaultValue: "%lld subjects",
                comment: "Sources list graph-zone subject count; argument is canvas subject count"
            ), count)
        }

        static func graphObservationCount(_ count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "sources.list.observationCount",
                defaultValue: "%lld observations",
                comment: "Sources list graph-zone observation count; argument is Observation count"
            ), count)
        }

        static func graphCountLine(subjects: Int, observations: Int) -> String {
            "\(graphSubjectCount(subjects)) · \(graphObservationCount(observations))"
        }

        static func graphZoneAccessibility(
            progress: SourceGraphProgress?,
            countsLoading: Bool
        ) -> String {
            let open = L10n.string(openGraph)
            if countsLoading && progress == nil {
                return "\(open), \(L10n.string(graphCountsLoading))"
            }
            let subjects = progress?.subjectCount ?? 0
            if subjects == 0 {
                return "\(open), \(graphSubjectCount(0)), \(L10n.string(graphNotStarted))"
            }
            return [
                open,
                graphSubjectCount(subjects),
                graphObservationCount(progress?.observationCount ?? 0),
            ].joined(separator: ", ")
        }

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
            return L10n.format(LocalizedStringResource(
                "sources.add.toastCreatedTitle",
                defaultValue: "%@",
                comment: "Toast title after creating a Source; argument is the SRC- ref"
            ), ref)
        }

        static func toastCreatedBody(title: String) -> String {
            return L10n.format(LocalizedStringResource(
                "sources.add.toastCreatedBody",
                defaultValue: "%@. Opening its Source page.",
                comment: "Toast body after creating a Source; argument is the title"
            ), title)
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

        static let jumpToEvidenceGraph = LocalizedStringResource(
            "sources.page.jumpToEvidenceGraph",
            defaultValue: "Jump to graph",
            comment: "Identity-header secondary button that opens this Source's Evidence graph"
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

        static let deleteMetadataValue = LocalizedStringResource(
            "sources.page.deleteMetadataValue",
            defaultValue: "Delete value",
            comment: "Accessibility label for trash on a saved metadata row"
        )

        static let openMetadataURL = LocalizedStringResource(
            "sources.page.openMetadataURL",
            defaultValue: "Open in browser",
            comment: "Tooltip on a clickable saved url metadata value"
        )

        static func deleteMetadataConfirmTitle(label: String) -> String {
            return L10n.format(LocalizedStringResource(
                "sources.page.deleteMetadataConfirmTitle",
                defaultValue: "Delete %@?",
                comment: "Confirm title when clearing a saved metadata value; argument is the field label"
            ), label)
        }

        static let deleteMetadataConfirmMessage = LocalizedStringResource(
            "sources.page.deleteMetadataConfirmMessage",
            defaultValue: "This removes the saved value. Suggested fields return to the dashed list; extra fields disappear.",
            comment: "Confirm message when clearing a saved metadata value"
        )

        static let deleteMetadataConfirm = LocalizedStringResource(
            "sources.page.deleteMetadataConfirm",
            defaultValue: "Delete value",
            comment: "Confirm button that clears a saved metadata value"
        )

        static let deleteMetadataKeep = LocalizedStringResource(
            "sources.page.deleteMetadataKeep",
            defaultValue: "Keep value",
            comment: "Cancel button on the metadata-value delete confirm"
        )

        static let deleteSource = LocalizedStringResource(
            "sources.page.deleteSource",
            defaultValue: "Delete source",
            comment: "Accessibility label for deleting the Source"
        )

        static let deleteArtifact = LocalizedStringResource(
            "sources.page.deleteArtifact",
            defaultValue: "Delete artifact",
            comment: "Accessibility label for deleting an Artifact"
        )

        static let toastArtifactDeletedTitle = LocalizedStringResource(
            "sources.page.toastArtifactDeletedTitle",
            defaultValue: "Artifact deleted",
            comment: "Toast title after an Artifact is erased"
        )

        static func toastArtifactDeletedBody(ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "sources.page.toastArtifactDeletedBody",
                defaultValue: "%@ was erased from this source.",
                comment: "Toast body after Artifact erase; argument is ART-…"
            ), ref)
        }

        static let deleteNoteConfirmTitle = LocalizedStringResource(
            "sources.page.deleteNoteConfirmTitle",
            defaultValue: "Delete this note?",
            comment: "Confirm title when deleting a Source note"
        )

        static let deleteNoteConfirmMessage = LocalizedStringResource(
            "sources.page.deleteNoteConfirmMessage",
            defaultValue: "This removes the note from the source.",
            comment: "Confirm message when deleting a Source note"
        )

        static let deleteNoteConfirm = LocalizedStringResource(
            "sources.page.deleteNoteConfirm",
            defaultValue: "Delete note",
            comment: "Confirm button that deletes a Source note"
        )

        static let deleteNoteKeep = LocalizedStringResource(
            "sources.page.deleteNoteKeep",
            defaultValue: "Keep note",
            comment: "Cancel button on the note delete confirm"
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
            return L10n.format(LocalizedStringResource(
                "sources.page.metadataFieldCount",
                defaultValue: "%lld fields",
                comment: "Metadata section count; argument is saved field count"
            ), count)
        }

        static func artifactsCount(_ count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "sources.page.artifactsCount",
                defaultValue: "%lld artifacts",
                comment: "Artifacts section count; argument is artifact count"
            ), count)
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
            return L10n.format(LocalizedStringResource(
                "sources.page.noteComposerAttribution",
                defaultValue: "%@ · now",
                comment: "Note composer byline; argument is the session display name"
            ), displayName)
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
            return L10n.format(LocalizedStringResource(
                "sources.page.toastArtifactCreatedTitle",
                defaultValue: "%@",
                comment: "Toast title after creating an Artifact; argument is ART- ref"
            ), ref)
        }

        static let toastArtifactCreatedFileless = L10n.string(LocalizedStringResource(
            "sources.page.toastArtifactCreatedFileless",
            defaultValue: "Created with no file yet — physical only.",
            comment: "Toast body after creating a fileless Artifact"
        ))

        static let toastArtifactCreatedWithFile = L10n.string(LocalizedStringResource(
            "sources.page.toastArtifactCreatedWithFile",
            defaultValue: "Created and the file was ingested.",
            comment: "Toast body after creating an Artifact with a file"
        ))

        static let toastFileAttachedTitle = L10n.string(LocalizedStringResource(
            "sources.page.toastFileAttachedTitle",
            defaultValue: "File attached",
            comment: "Toast title after first-attach ingest"
        ))

        static func toastFileAttachedBody(name: String) -> String {
            return L10n.format(LocalizedStringResource(
                "sources.page.toastFileAttachedBody",
                defaultValue: "%@ was ingested into the project.",
                comment: "Toast body after ingest; argument is original filename"
            ), name)
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

        static let toastThumbnailUpdatedTitle = L10n.string(LocalizedStringResource(
            "sources.page.toastThumbnailUpdatedTitle",
            defaultValue: "Thumbnail updated",
            comment: "Toast title after changing Source cover"
        ))

        static func toastThumbnailUpdatedArtifactBody(ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "sources.page.toastThumbnailUpdatedArtifactBody",
                defaultValue: "Using %@ as the Source thumbnail.",
                comment: "Toast body after pinning an Artifact; argument is ART- ref"
            ), ref)
        }

        static let toastThumbnailUpdatedTypeIconBody = L10n.string(LocalizedStringResource(
            "sources.page.toastThumbnailUpdatedTypeIconBody",
            defaultValue: "Using the Source type icon as the thumbnail.",
            comment: "Toast body after reverting cover to the type icon"
        ))
    }

    enum Metadata {
        static let description = LocalizedStringResource(
            "metadata.list.description",
            defaultValue: "The metadata a source can carry in this project. Provenencia seeds the common fields; you add the ones your records actually use.",
            comment: "Explanatory copy under the Metadata page title"
        )

        static func countLine(total: Int, seeded: Int, user: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "metadata.list.countLine",
                defaultValue: "%1$lld fields · %2$lld seeded · %3$lld yours",
                comment: "Metadata count summary; arguments are total, seeded (provenencia), and user field counts"
            ), total, seeded, user)
        }

        static func countLineWithPlugin(total: Int, seeded: Int, user: Int, plugin: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "metadata.list.countLineWithPlugin",
                defaultValue: "%1$lld fields · %2$lld seeded · %3$lld yours · %4$lld plugin",
                comment: "Metadata count summary including plugin-origin fields; arguments are total, seeded, user, and plugin field counts"
            ), total, seeded, user, plugin)
        }

        static let addField = LocalizedStringResource(
            "metadata.list.addField",
            defaultValue: "Add field",
            comment: "Button: add a new Metadata field (toolbar, empty state, and add-form submit)"
        )

        static let columnLabel = LocalizedStringResource(
            "metadata.list.columnLabel",
            defaultValue: "Label",
            comment: "Metadata list column header: label"
        )

        static let columnKey = LocalizedStringResource(
            "metadata.list.columnKey",
            defaultValue: "Key",
            comment: "Metadata list column header: key"
        )

        static let columnDataType = LocalizedStringResource(
            "metadata.list.columnDataType",
            defaultValue: "Data type",
            comment: "Metadata list column header: data type"
        )

        static let dataTypeText = LocalizedStringResource(
            "metadata.dataType.text",
            defaultValue: "text",
            comment: "Metadata field data type badge/option: text"
        )

        static let dataTypeDate = LocalizedStringResource(
            "metadata.dataType.date",
            defaultValue: "date",
            comment: "Metadata field data type badge/option: date"
        )

        static let dataTypeUrl = LocalizedStringResource(
            "metadata.dataType.url",
            defaultValue: "url",
            comment: "Metadata field data type badge/option: url"
        )

        static let deleteField = LocalizedStringResource(
            "metadata.delete.action",
            defaultValue: "Delete field",
            comment: "Tooltip on the Metadata inspector trash"
        )

        static func deleteFieldAccessibility(label: String) -> String {
            return L10n.format(LocalizedStringResource(
                "metadata.delete.accessibility",
                defaultValue: "Delete field %@",
                comment: "VoiceOver for Metadata trash; argument is the field label"
            ), label)
        }

        static let toastDeletedTitle = LocalizedStringResource(
            "metadata.toast.deletedTitle",
            defaultValue: "Field deleted",
            comment: "Toast title after a metadata field is deleted"
        )

        static func toastDeletedBody(label: String) -> String {
            return L10n.format(LocalizedStringResource(
                "metadata.toast.deletedBody",
                defaultValue: "%@ is no longer in this project's vocabulary.",
                comment: "Toast body after a metadata field is deleted; argument is the field label"
            ), label)
        }

        static let emptyProjectTitle = LocalizedStringResource(
            "metadata.emptyProject.title",
            defaultValue: "No metadata fields yet",
            comment: "Title of the empty state when the project has zero metadata fields"
        )

        static let emptyProjectBody = LocalizedStringResource(
            "metadata.emptyProject.body",
            defaultValue: "This project has no metadata vocabulary. Add the fields your records actually carry — a certificate number, a photographer, an album code.",
            comment: "Body of the empty state when the project has zero metadata fields"
        )

        static func resultLine(shown: Int, total: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "metadata.list.resultLineAll",
                defaultValue: "%lld fields",
                comment: "Footer result count for the Metadata list; argument is the total"
            ), total)
        }

        static let detailEyebrowField = LocalizedStringResource(
            "metadata.detail.eyebrowField",
            defaultValue: "Field",
            comment: "Eyebrow label above an existing field's detail panel"
        )

        static let detailEyebrowNewField = LocalizedStringResource(
            "metadata.detail.eyebrowNewField",
            defaultValue: "New field",
            comment: "Eyebrow label above the add-field panel"
        )

        static let keyHintAdd = LocalizedStringResource(
            "metadata.detail.keyHintAdd",
            defaultValue: "Provenencia mints the key from the label when the field is added",
            comment: "Hint under the live key preview while adding a field"
        )

        static let keyHintEdit = LocalizedStringResource(
            "metadata.detail.keyHintEdit",
            defaultValue: "The key is minted once from the label and never changes — renaming the field keeps existing sources attached",
            comment: "Hint under the key on an existing field's detail panel"
        )

        static func lockedNotePlugin(pluginID: String) -> String {
            return L10n.format(LocalizedStringResource(
                "metadata.detail.lockedNotePlugin",
                defaultValue: "Supplied by the %@ plugin. The plugin owns this definition — Provenencia will not edit it.",
                comment: "Callout explaining why a plugin-origin field can't be edited; argument is the plugin id"
            ), pluginID)
        }

        static let dataTypeSectionLabel = LocalizedStringResource(
            "metadata.detail.dataTypeSectionLabel",
            defaultValue: "Data type",
            comment: "Section label above the read-only data type line on a locked field's detail"
        )

        static let descriptionSectionLabel = LocalizedStringResource(
            "metadata.detail.descriptionSectionLabel",
            defaultValue: "Description",
            comment: "Section label above the read-only description on a locked field's detail"
        )

        static let descriptionEmptyPlaceholder = LocalizedStringResource(
            "metadata.detail.descriptionEmptyPlaceholder",
            defaultValue: "—",
            comment: "Shown in place of a locked field's description when it has none"
        )

        static let panelEmptyTitle = LocalizedStringResource(
            "metadata.detail.panelEmptyTitle",
            defaultValue: "No field selected",
            comment: "Title of the empty state shown in the detail panel before any field is selected"
        )

        static let panelEmptyBody = LocalizedStringResource(
            "metadata.detail.panelEmptyBody",
            defaultValue: "Select a field to read or edit its definition. Fields supplied by a plugin are read-only; the rest of this project's vocabulary stays editable.",
            comment: "Body of the empty state shown in the detail panel before any field is selected"
        )

        static let formLabel = LocalizedStringResource(
            "metadata.form.label",
            defaultValue: "Label",
            comment: "Add/edit form field: label"
        )

        static let formLabelPlaceholder = LocalizedStringResource(
            "metadata.form.labelPlaceholder",
            defaultValue: "Grandma’s album code",
            comment: "Placeholder text for the add-field label input"
        )

        static let formDataType = LocalizedStringResource(
            "metadata.form.dataType",
            defaultValue: "Data type",
            comment: "Add/edit form field: data type picker"
        )

        static let formDataTypeHint = LocalizedStringResource(
            "metadata.form.dataTypeHint",
            defaultValue: "Text or url — chosen at create and immutable afterward",
            comment: "Hint under the data type picker on add"
        )

        static let formDataTypeImmutableHint = LocalizedStringResource(
            "metadata.form.dataTypeImmutableHint",
            defaultValue: "Data type is fixed when the field is created so existing values stay valid.",
            comment: "Hint under the read-only data type on edit"
        )

        static let formDescription = LocalizedStringResource(
            "metadata.form.description",
            defaultValue: "Description",
            comment: "Add/edit form field: description"
        )

        static let formDescriptionHint = LocalizedStringResource(
            "metadata.form.descriptionHint",
            defaultValue: "What a researcher should put in this field, in your own words",
            comment: "Hint under the description field"
        )

        static let formDescriptionPlaceholder = LocalizedStringResource(
            "metadata.form.descriptionPlaceholder",
            defaultValue: "Pencil code on the back of prints from the album",
            comment: "Placeholder text for the add-field description input"
        )

        static let errorLabelRequired = LocalizedStringResource(
            "metadata.form.errorLabelRequired",
            defaultValue: "A label is required — it is how the field reads on a source.",
            comment: "Inline validation error when the label is blank"
        )

        static let errorUnslugifiable = LocalizedStringResource(
            "metadata.form.errorUnslugifiable",
            defaultValue: "That label cannot be turned into a key. Use at least one letter or number.",
            comment: "Inline validation error when the label has no letters or digits to slug"
        )

        static let saveSaving = LocalizedStringResource(
            "metadata.form.saveSaving",
            defaultValue: "Saving",
            comment: "Primary button label while a Metadata field add/edit is in flight"
        )

        static let saveChanges = LocalizedStringResource(
            "metadata.form.saveChanges",
            defaultValue: "Save changes",
            comment: "Primary button label for committing an edit to an existing field"
        )

        static let cancel = LocalizedStringResource(
            "metadata.form.cancel",
            defaultValue: "Cancel",
            comment: "Secondary button label that dismisses the add-field form"
        )

        static let revert = LocalizedStringResource(
            "metadata.form.revert",
            defaultValue: "Revert",
            comment: "Secondary button label that discards unsaved edits to an existing field"
        )

        static let toastAddedTitle = LocalizedStringResource(
            "metadata.toast.addedTitle",
            defaultValue: "Field added",
            comment: "Success toast title after creating a Metadata field"
        )

        static func toastAddedBody(label: String, key: String) -> String {
            return L10n.format(LocalizedStringResource(
                "metadata.toast.addedBody",
                defaultValue: "%1$@ is in this project’s vocabulary as %2$@.",
                comment: "Success toast body after creating a Metadata field; arguments are label then minted key"
            ), label, key)
        }

        static let toastUpdatedTitle = LocalizedStringResource(
            "metadata.toast.updatedTitle",
            defaultValue: "Field updated",
            comment: "Success toast title after editing a Metadata field"
        )

        static func toastUpdatedBody(label: String, key: String) -> String {
            return L10n.format(LocalizedStringResource(
                "metadata.toast.updatedBody",
                defaultValue: "%1$@ — the key stays %2$@.",
                comment: "Success toast body after editing a Metadata field; arguments are label then key"
            ), label, key)
        }
    }

    enum Properties {
        static let description = LocalizedStringResource(
            "properties.list.description",
            defaultValue: "The properties a subject can carry, and which of the seven subject types carry them.",
            comment: "Explanatory copy under the Properties page title"
        )
        static let allProperties = LocalizedStringResource(
            "properties.strip.all",
            defaultValue: "All properties",
            comment: "Type strip card that clears the subject-type filter"
        )
        static let bridgeRole = LocalizedStringResource(
            "properties.strip.bridge",
            defaultValue: "bridge",
            comment: "Micro-label on bridge subject-type strip cards; rendered uppercase (BRIDGE)"
        )
        static let typeStripAccessibility = LocalizedStringResource(
            "properties.strip.accessibility",
            defaultValue: "Subject types",
            comment: "Accessibility label for the type strip pressed-button group"
        )
        static func stripFieldCount(count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "properties.strip.fieldCount",
                defaultValue: "%lld fields",
                comment: "Type strip count under a subject type; argument is binding count"
            ), count)
        }
        static let searchPlaceholder = LocalizedStringResource(
            "properties.search.placeholder",
            defaultValue: "Search properties",
            comment: "List-card search field placeholder on Properties"
        )
        static let originUserShort = LocalizedStringResource(
            "properties.table.originUser",
            defaultValue: "user",
            comment: "Table origin column for researcher-created properties"
        )
        static let originSeededShort = LocalizedStringResource(
            "properties.table.originSeeded",
            defaultValue: "seeded",
            comment: "Table origin column for Provenencia-seeded properties"
        )
        static let valueTypeText = LocalizedStringResource(
            "properties.valueType.text",
            defaultValue: "Text",
            comment: "Property value type label: text"
        )
        static let valueTypeInteger = LocalizedStringResource(
            "properties.valueType.integer",
            defaultValue: "Integer",
            comment: "Property value type label: integer"
        )
        static let valueTypeDate = LocalizedStringResource(
            "properties.valueType.date",
            defaultValue: "Date",
            comment: "Property value type label: date"
        )
        static let valueTypeName = LocalizedStringResource(
            "properties.valueType.name",
            defaultValue: "Name",
            comment: "Property value type label: name"
        )
        static let valueTypeSubject = LocalizedStringResource(
            "properties.valueType.subject",
            defaultValue: "Subject",
            comment: "Property value type label: subject"
        )
        static let valueTypeTerm = LocalizedStringResource(
            "properties.valueType.term",
            defaultValue: "Term",
            comment: "Property value type label: term (registry-only)"
        )
        static let columnOn = LocalizedStringResource(
            "properties.table.columnOn",
            defaultValue: "On",
            comment: "Properties table column: binding toggle for the focused type"
        )
        static let columnProperty = LocalizedStringResource(
            "properties.table.columnProperty",
            defaultValue: "Property",
            comment: "Properties table column: property label"
        )
        static let columnValueType = LocalizedStringResource(
            "properties.table.columnValueType",
            defaultValue: "Value type",
            comment: "Properties table column: value type"
        )
        static let columnOrigin = LocalizedStringResource(
            "properties.table.columnOrigin",
            defaultValue: "Origin",
            comment: "Properties table column: origin"
        )
        static let columnBoundTo = LocalizedStringResource(
            "properties.table.columnBoundTo",
            defaultValue: "Bound to",
            comment: "Properties table column: bound subject types"
        )
        static func boundOverflow(count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "properties.table.boundOverflow",
                defaultValue: "+%lld",
                comment: "Overflow when more than three Bound-to chips; argument is remaining count"
            ), count)
        }
        static func rowBoundAnnouncement(count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "properties.table.rowBoundAnnouncement",
                defaultValue: "bound to %lld types",
                comment: "VoiceOver fragment for how many types a property is bound to"
            ), count)
        }
        static func emptySearchTitle(query: String) -> String {
            let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
            if trimmed.isEmpty {
                return L10n.string(LocalizedStringResource(
                    "properties.table.emptySearchTitle",
                    defaultValue: "No properties match",
                    comment: "Empty table title when no properties are visible"
                ))
            }
            return L10n.format(LocalizedStringResource(
                "properties.table.emptySearchTitleQuery",
                defaultValue: "No property matches “%@”",
                comment: "Empty table title when search matches nothing; argument is the query"
            ), trimmed)
        }
        static let emptySearch = LocalizedStringResource(
            "properties.table.emptySearch",
            defaultValue: "Clear the search, or create it as a user property",
            comment: "Empty state when search/filter matches no properties"
        )
        static let newProperty = LocalizedStringResource(
            "properties.toolbar.newProperty",
            defaultValue: "New property",
            comment: "Toolbar button to open create Property sheet"
        )
        static func addPropertyPlaceholder(typeLabel: String) -> String {
            L10n.format(LocalizedStringResource(
                "properties.toolbar.addPropertyPlaceholder",
                defaultValue: "Add a property to %@",
                comment: "ComboBox placeholder when a subject type is focused; argument is type label"
            ), typeLabel)
        }
        static let addPropertyEmpty = LocalizedStringResource(
            "properties.toolbar.addPropertyEmpty",
            defaultValue: "No unbound property matches that name",
            comment: "ComboBox empty state when binding an existing property to the focused type"
        )
        static let inspectorAccessibility = LocalizedStringResource(
            "properties.inspector.accessibility",
            defaultValue: "Property inspector",
            comment: "Accessibility label for the property inspector card"
        )
        static let inspectorEmpty = LocalizedStringResource(
            "properties.inspector.empty",
            defaultValue: "Select a property to see its description and bindings.",
            comment: "Inspector empty state"
        )
        static let inspectorValueType = LocalizedStringResource(
            "properties.inspector.valueType",
            defaultValue: "Value type",
            comment: "Inspector meta label: value type"
        )
        static let valueTypeImmutable = LocalizedStringResource(
            "properties.inspector.valueTypeImmutable",
            defaultValue: "Immutable after create",
            comment: "Accessibility label for the lock beside value type in the inspector"
        )
        static let holds = LocalizedStringResource(
            "properties.inspector.holds",
            defaultValue: "Holds",
            comment: "Inspector and create-form label for property cardinality"
        )
        static let holdsOne = LocalizedStringResource(
            "properties.inspector.holdsOne",
            defaultValue: "One value",
            comment: "Cardinality choice: the engine narrows records to one value"
        )
        static let holdsOneDescription = LocalizedStringResource(
            "properties.inspector.holdsOneDescription",
            defaultValue: "Records are narrowed to one value, or shown as mixed",
            comment: "Explanation under the one-value cardinality choice"
        )
        static let holdsSeveral = LocalizedStringResource(
            "properties.inspector.holdsSeveral",
            defaultValue: "Several values",
            comment: "Cardinality choice: every distinct surviving value is kept"
        )
        static let holdsSeveralDescription = LocalizedStringResource(
            "properties.inspector.holdsSeveralDescription",
            defaultValue: "Every distinct value is kept; spellings still merge",
            comment: "Explanation under the several-values cardinality choice"
        )
        static let holdsEventsHint = LocalizedStringResource(
            "properties.inspector.holdsEventsHint",
            defaultValue: "Facts that change over time, like occupation or residence, usually belong in events",
            comment: "Standing hint under Holds, and the create-form field hint"
        )
        static func holdsReconciled(label: String) -> String {
            L10n.format(LocalizedStringResource(
                "properties.inspector.holdsReconciled",
                defaultValue: "Every Person, Event and Place carrying %@ is reconciled again; no values are deleted",
                comment: "Hint after a cardinality change until another property is selected; argument is the property label"
            ), label)
        }
        static let holdsLocked = LocalizedStringResource(
            "properties.inspector.holdsLocked",
            defaultValue: "Set by Provenencia",
            comment: "Accessibility label for the lock beside a seeded property's Holds value"
        )
        static func holdsGroup(label: String) -> String {
            L10n.format(LocalizedStringResource(
                "properties.inspector.holdsGroup",
                defaultValue: "%@ holds",
                comment: "Accessibility label for the Holds radio group; argument is the property label"
            ), label)
        }
        static let inspectorOrigin = LocalizedStringResource(
            "properties.inspector.origin",
            defaultValue: "Origin",
            comment: "Inspector meta label: origin"
        )
        static let inspectorOriginUser = LocalizedStringResource(
            "properties.inspector.originUser",
            defaultValue: "User — you created this",
            comment: "Inspector origin line for researcher-created properties"
        )
        static let inspectorOriginSeeded = LocalizedStringResource(
            "properties.inspector.originSeeded",
            defaultValue: "Seeded by Provenencia",
            comment: "Inspector origin line for product-seeded properties"
        )
        static let inspectorUsedOn = LocalizedStringResource(
            "properties.inspector.usedOn",
            defaultValue: "Used on",
            comment: "Inspector meta label: observation count for the selected property"
        )
        static func usedOnCount(count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "properties.inspector.usedOnCount",
                defaultValue: "%lld observations",
                comment: "Inspector observation count; argument is UsedBy"
            ), count)
        }
        static let bindingsSection = LocalizedStringResource(
            "properties.inspector.bindings",
            defaultValue: "Bound to",
            comment: "Inspector section label for bindings list"
        )
        static let bindingsAccessibility = LocalizedStringResource(
            "properties.inspector.bindingsAccessibility",
            defaultValue: "Subject type bindings",
            comment: "Accessibility label for the Bound-to checkbox list"
        )
        static func bindCount(bound: Int, total: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "properties.inspector.bindCount",
                defaultValue: "%1$lld of %2$lld",
                comment: "Inspector bound-type count beside Bound to; bound, then total types"
            ), bound, total)
        }
        static let termNote = LocalizedStringResource(
            "properties.inspector.termNote",
            defaultValue: "Values come from the product vocabulary. You choose one when citing this property.",
            comment: "Inspector note for term-typed seeded properties"
        )
        static let deleteProperty = LocalizedStringResource(
            "properties.delete.action",
            defaultValue: "Delete property",
            comment: "Tooltip on the Properties inspector trash"
        )
        static func deletePropertyAccessibility(label: String) -> String {
            return L10n.format(LocalizedStringResource(
                "properties.delete.accessibility",
                defaultValue: "Delete property %@",
                comment: "VoiceOver for Properties trash; argument is the property label"
            ), label)
        }
        static let editAction = LocalizedStringResource(
            "properties.edit.action",
            defaultValue: "Edit label and description",
            comment: "Tooltip on the Properties inspector pencil"
        )
        static let editSave = LocalizedStringResource(
            "properties.edit.save",
            defaultValue: "Save",
            comment: "Save tooltip for in-place property label/description edit"
        )
        static let editCancel = LocalizedStringResource(
            "properties.edit.cancel",
            defaultValue: "Cancel editing",
            comment: "Cancel tooltip for in-place property label/description edit"
        )
        static func editKeyStays(key: String) -> String {
            return L10n.format(LocalizedStringResource(
                "properties.edit.keyStays",
                defaultValue: "The key stays %@ — Observations already cite it",
                comment: "Lock line under the label input while editing; argument is the property key"
            ), key)
        }
        static let editDescriptionPlaceholder = LocalizedStringResource(
            "properties.edit.descriptionPlaceholder",
            defaultValue: "Say what this property records",
            comment: "Placeholder for the in-place property description textarea"
        )
        static func lockedBindingReason(typeLabel: String) -> String {
            return L10n.format(LocalizedStringResource(
                "properties.inspector.lockedBinding",
                defaultValue: "The Interpretation subject registry requires this property on %@. The binding cannot be removed.",
                comment: "Callout when activating a locked binding; argument is subject type label"
            ), typeLabel)
        }
        static let bindingLocked = LocalizedStringResource(
            "properties.inspector.bindingLocked",
            defaultValue: "Locked",
            comment: "Accessibility / badge for a locked subject-type binding"
        )
        static let bindingBound = LocalizedStringResource(
            "properties.inspector.bindingBound",
            defaultValue: "Bound",
            comment: "Accessibility label for an active unbound-able binding"
        )
        static let bindingNotBound = LocalizedStringResource(
            "properties.inspector.bindingNotBound",
            defaultValue: "not bound",
            comment: "Accessibility label for an inactive binding checkbox"
        )
        static let bindingRegistry = LocalizedStringResource(
            "properties.inspector.bindingRegistry",
            defaultValue: "registry",
            comment: "Micro-label beside a locked Bound-to row (registry-held)"
        )
        static let createTitle = LocalizedStringResource(
            "properties.create.title",
            defaultValue: "New property",
            comment: "Create property sheet title"
        )
        static let createOriginNote = LocalizedStringResource(
            "properties.create.originNote",
            defaultValue: "Origin is recorded as user — seeded properties come from Provenencia",
            comment: "Create property sheet note about origin"
        )
        static let createBindSection = LocalizedStringResource(
            "properties.create.bindSection",
            defaultValue: "Bind to subject types",
            comment: "Create property sheet: bind checklist section"
        )
        static let createLabel = LocalizedStringResource(
            "properties.create.label",
            defaultValue: "Label",
            comment: "Create property form: label field"
        )
        static let createLabelHint = LocalizedStringResource(
            "properties.create.labelHint",
            defaultValue: "What a researcher sees on the subject",
            comment: "Hint under create property label"
        )
        static let createLabelPlaceholder = LocalizedStringResource(
            "properties.create.labelPlaceholder",
            defaultValue: "Burial ground",
            comment: "Placeholder for create property label field"
        )
        static let createKey = LocalizedStringResource(
            "properties.create.key",
            defaultValue: "Key",
            comment: "Create property form: machine key field"
        )
        static let createKeyHint = LocalizedStringResource(
            "properties.create.keyHint",
            defaultValue: "Generated from the label",
            comment: "Hint under create property key; key is minted server-side from label"
        )
        static let createKeyPlaceholder = LocalizedStringResource(
            "properties.create.keyPlaceholder",
            defaultValue: "burial-ground",
            comment: "Placeholder for create property key preview"
        )
        static let createValueType = LocalizedStringResource(
            "properties.create.valueType",
            defaultValue: "Value type",
            comment: "Create property form: value type picker"
        )
        static let createValueTypeHint = LocalizedStringResource(
            "properties.create.valueTypeHint",
            defaultValue: "Cannot be changed once the property exists",
            comment: "Hint under create property value type chips"
        )
        static let createDescription = LocalizedStringResource(
            "properties.create.description",
            defaultValue: "Description",
            comment: "Create property form: description"
        )
        static let createDescriptionPlaceholder = LocalizedStringResource(
            "properties.create.descriptionPlaceholder",
            defaultValue: "How the value should be read from the record",
            comment: "Placeholder for create property description"
        )
        static let createCancel = LocalizedStringResource(
            "properties.create.cancel",
            defaultValue: "Cancel",
            comment: "Create property sheet cancel"
        )
        static let createSubmit = LocalizedStringResource(
            "properties.create.submit",
            defaultValue: "Create property",
            comment: "Create property sheet primary action"
        )
        static let errorLabelRequired = LocalizedStringResource(
            "properties.create.errorLabelRequired",
            defaultValue: "Enter a label for this property.",
            comment: "Validation when create label is empty"
        )
        static let errorValueType = LocalizedStringResource(
            "properties.create.errorValueType",
            defaultValue: "Choose a researcher value type. Term properties are product vocabulary only.",
            comment: "Validation when create value type is invalid"
        )
        static let toastCreatedTitle = LocalizedStringResource(
            "properties.toast.createdTitle",
            defaultValue: "Property created",
            comment: "Success toast title after creating a property"
        )
        static func toastCreatedBody(label: String) -> String {
            return L10n.format(LocalizedStringResource(
                "properties.toast.createdBody",
                defaultValue: "%@ is ready to bind.",
                comment: "Success toast body after creating a property; argument is label"
            ), label)
        }
        static let toastDeletedTitle = LocalizedStringResource(
            "properties.toast.deletedTitle",
            defaultValue: "Property deleted",
            comment: "Success toast title after deleting a property"
        )
        static func toastDeletedBody(label: String) -> String {
            return L10n.format(LocalizedStringResource(
                "properties.toast.deletedBody",
                defaultValue: "%@ was removed from this project.",
                comment: "Success toast body after deleting a property; argument is label"
            ), label)
        }
        static let toastUpdatedTitle = LocalizedStringResource(
            "properties.toast.updatedTitle",
            defaultValue: "Property updated",
            comment: "Success toast title after editing a property label or description"
        )
        static func toastUpdatedBody(label: String, key: String) -> String {
            return L10n.format(LocalizedStringResource(
                "properties.toast.updatedBody",
                defaultValue: "%1$@ — the key stays %2$@.",
                comment: "Success toast body after editing a property; arguments are label then key"
            ), label, key)
        }
    }

    /// Maps stable Go/FFI error codes to localized user-facing copy.
    enum SourceTypes {
        static let description = LocalizedStringResource(
            "sourceTypes.list.description",
            defaultValue: "The kinds of record this project cites, and the fields each kind usually carries. The fields are suggestions — a source of this type may leave any of them blank.",
            comment: "Explanatory copy under the Source types page title"
        )

        static func countLine(total: Int, seeded: Int, user: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "sourceTypes.list.countLine",
                defaultValue: "%1$lld types · %2$lld seeded · %3$lld yours",
                comment: "Source types count summary; arguments are total, seeded (provenencia), and user type counts"
            ), total, seeded, user)
        }

        static func countLineWithPlugin(total: Int, seeded: Int, user: Int, plugin: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "sourceTypes.list.countLineWithPlugin",
                defaultValue: "%1$lld types · %2$lld seeded · %3$lld yours · %4$lld plugin",
                comment: "Source types count summary including plugin-origin types; arguments are total, seeded, user, and plugin type counts"
            ), total, seeded, user, plugin)
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
            return L10n.format(LocalizedStringResource(
                "sourceTypes.list.resultLineAll",
                defaultValue: "%lld types",
                comment: "Footer result count for the Source types list; argument is the total"
            ), total)
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

        static func usage(count: Int) -> String {
            switch count {
            case 0: L10n.string(usageNone)
            default: L10n.format(LocalizedStringResource(
                "sourceTypes.detail.usage",
                defaultValue: "in use on %lld sources",
                comment: "Line under a type's title; argument is how many sources are classified as it"
            ), count)
            }
        }

        private static let usageNone = LocalizedStringResource(
            "sourceTypes.detail.usageNone",
            defaultValue: "no sources yet",
            comment: "Line under a type's title when no source is classified as it"
        )

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
            return L10n.format(LocalizedStringResource(
                "sourceTypes.detail.lockedNotePlugin",
                defaultValue: "Supplied by the %@ plugin. The plugin owns this type, its description and the fields it suggests — Provenencia will not edit or delete them.",
                comment: "Callout explaining why a plugin-origin type can't be edited; argument is the plugin id"
            ), pluginID)
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
            return L10n.format(LocalizedStringResource(
                "sourceTypes.form.iconChangeAccessibility",
                defaultValue: "Icon: %@ — choose a different one",
                comment: "Accessibility label for the icon field button; argument is the current mark name"
            ), name)
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

        static func assignedCount(count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "sourceTypes.suggested.count",
                defaultValue: "%lld fields",
                comment: "Count beside the suggested fields section label; argument is how many fields the type suggests"
            ), count)
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
            defaultValue: "No field in the vocabulary matches that — add it in Metadata first",
            comment: "Shown inside the assign-field combo box list when the typed query matches no field"
        )

        static let assignFieldLabel = LocalizedStringResource(
            "sourceTypes.suggested.assignFieldLabel",
            defaultValue: "Metadata field to assign",
            comment: "Accessibility label for the assign-field combo box"
        )

        static let assignField = LocalizedStringResource(
            "sourceTypes.suggested.assignField",
            defaultValue: "Assign field",
            comment: "Spoken label for the assign button when no field is picked yet"
        )

        static func assignFieldNamed(field: String, type: String) -> String {
            L10n.format(LocalizedStringResource(
                "sourceTypes.suggested.assignFieldNamed",
                defaultValue: "Assign field %1$@ to %2$@",
                comment: "Spoken label for the assign button; arguments are the picked field label then the type label"
            ), field, type)
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

        static func assignTipField(label: String) -> String {
            L10n.format(LocalizedStringResource(
                "sourceTypes.suggested.assignTipField",
                defaultValue: "Assign %@ to this type",
                comment: "Tooltip on the enabled assign button; argument is the picked field label"
            ), label)
        }

        static let poolHint = LocalizedStringResource(
            "sourceTypes.suggested.poolHint",
            defaultValue: "The pool is the Metadata vocabulary — add a new field there first if it is missing",
            comment: "Hint under the assign-field picker naming where the pool comes from"
        )

        static let poolHintEmpty = LocalizedStringResource(
            "sourceTypes.suggested.poolHintEmpty",
            defaultValue: "Every field in the vocabulary is already suggested for this type",
            comment: "Hint under the assign-field picker when nothing is left to assign"
        )

        static func removeSuggestion(label: String) -> String {
            L10n.format(LocalizedStringResource(
                "sourceTypes.suggested.remove",
                defaultValue: "Remove %@ from this type",
                comment: "Accessibility label and tooltip on the control that detaches one suggested field; argument is the field label"
            ), label)
        }

        static let toastAssignedTitle = LocalizedStringResource(
            "sourceTypes.toast.assignedTitle",
            defaultValue: "Field assigned",
            comment: "Toast title after attaching a field to a type"
        )

        static func toastAssignedBody(field: String, type: String) -> String {
            return L10n.format(LocalizedStringResource(
                "sourceTypes.toast.assignedBody",
                defaultValue: "%1$@ is now suggested for %2$@.",
                comment: "Toast body after attaching a field to a type; arguments are the field label then the type label"
            ), field, type)
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
                return L10n.format(removedBodyNoValues, field, type)
            }
            return L10n.format(removedBodyWithValues, field, type, valueCount)
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
            return L10n.format(LocalizedStringResource(
                "sourceTypes.toast.addedBody",
                defaultValue: "%1$@ is in this project’s vocabulary as %2$@. Assign the fields it should suggest.",
                comment: "Success toast body after creating a Source type; arguments are label then minted key"
            ), label, key)
        }

        static let toastUpdatedTitle = LocalizedStringResource(
            "sourceTypes.toast.updatedTitle",
            defaultValue: "Type updated",
            comment: "Success toast title after editing a Source type"
        )

        static func toastUpdatedBody(label: String, key: String) -> String {
            return L10n.format(LocalizedStringResource(
                "sourceTypes.toast.updatedBody",
                defaultValue: "%1$@ — the key stays %2$@.",
                comment: "Success toast body after editing a Source type; arguments are label then key"
            ), label, key)
        }

        static let deleteType = LocalizedStringResource(
            "sourceTypes.delete.action",
            defaultValue: "Delete type",
            comment: "Tooltip on the Source types inspector trash"
        )

        static func deleteTypeAccessibility(label: String) -> String {
            return L10n.format(LocalizedStringResource(
                "sourceTypes.delete.accessibility",
                defaultValue: "Delete type %@",
                comment: "VoiceOver for Source types trash; argument is the type label"
            ), label)
        }

        static let toastDeletedTitle = LocalizedStringResource(
            "sourceTypes.toast.deletedTitle",
            defaultValue: "Type deleted",
            comment: "Toast title after a source type is deleted"
        )

        static func toastDeletedBody(label: String) -> String {
            return L10n.format(LocalizedStringResource(
                "sourceTypes.toast.deletedBody",
                defaultValue: "%@ is no longer in this project’s vocabulary. Its field suggestions went with it; the fields did not.",
                comment: "Toast body after a source type is deleted; argument is the type label"
            ), label)
        }
    }

    enum DeleteImpact {
        static func noun(_ kind: String, count: Int = 1) -> String {
            let pair: (LocalizedStringResource, LocalizedStringResource)
            switch kind {
            case "source":
                pair = (nounSource, nounSources)
            case "artifact":
                pair = (nounArtifact, nounArtifacts)
            case "citation":
                pair = (nounCitation, nounCitations)
            case "observation":
                pair = (nounObservation, nounObservations)
            case "subject":
                pair = (nounSubject, nounSubjects)
            case "source_type":
                pair = (nounSourceType, nounSourceTypes)
            case "metadata_field":
                pair = (nounMetadataField, nounMetadata)
            case "source_credibility_grade":
                pair = (nounCredibilityGrade, nounCredibilityGrades)
            case "subject_type":
                pair = (nounSubjectType, nounSubjectTypes)
            case "property":
                pair = (nounProperty, nounProperties)
            case "property_term":
                pair = (nounPropertyTerm, nounPropertyTerms)
            case "canonical_entity":
                pair = (nounCanonicalEntity, nounCanonicalEntities)
            case "user":
                pair = (nounUser, nounUsers)
            case "project":
                pair = (nounProject, nounProjects)
            case "file":
                pair = (nounFile, nounFiles)
            case "narrative":
                pair = (nounNarrative, nounNarratives)
            case "reconciliation_claim":
                pair = (nounReconciliationClaim, nounReconciliationClaims)
            default:
                pair = (nounItem, nounItems)
            }
            return L10n.string(count == 1 ? pair.0 : pair.1)
        }

        static let nounSource = LocalizedStringResource(
            "deleteImpact.kind.source.one",
            defaultValue: "source",
            comment: "Singular Impact kind noun"
        )
        static let nounSources = LocalizedStringResource(
            "deleteImpact.kind.source.other",
            defaultValue: "sources",
            comment: "Plural Impact kind noun"
        )
        static let nounArtifact = LocalizedStringResource(
            "deleteImpact.kind.artifact.one",
            defaultValue: "artifact",
            comment: "Singular Impact kind noun"
        )
        static let nounArtifacts = LocalizedStringResource(
            "deleteImpact.kind.artifact.other",
            defaultValue: "artifacts",
            comment: "Plural Impact kind noun"
        )
        static let nounCitation = LocalizedStringResource(
            "deleteImpact.kind.citation.one",
            defaultValue: "citation",
            comment: "Singular Impact kind noun"
        )
        static let nounCitations = LocalizedStringResource(
            "deleteImpact.kind.citation.other",
            defaultValue: "citations",
            comment: "Plural Impact kind noun"
        )
        static let nounObservation = LocalizedStringResource(
            "deleteImpact.kind.observation.one",
            defaultValue: "observation",
            comment: "Singular Impact kind noun"
        )
        static let nounObservations = LocalizedStringResource(
            "deleteImpact.kind.observation.other",
            defaultValue: "observations",
            comment: "Plural Impact kind noun"
        )
        static let nounSubject = LocalizedStringResource(
            "deleteImpact.kind.subject.one",
            defaultValue: "subject",
            comment: "Singular Impact kind noun"
        )
        static let nounSubjects = LocalizedStringResource(
            "deleteImpact.kind.subject.other",
            defaultValue: "subjects",
            comment: "Plural Impact kind noun"
        )
        static let nounSourceType = LocalizedStringResource(
            "deleteImpact.kind.sourceType.one",
            defaultValue: "source type",
            comment: "Singular Impact kind noun"
        )
        static let nounSourceTypes = LocalizedStringResource(
            "deleteImpact.kind.sourceType.other",
            defaultValue: "source types",
            comment: "Plural Impact kind noun"
        )
        static let nounMetadataField = LocalizedStringResource(
            "deleteImpact.kind.metadataField.one",
            defaultValue: "metadata field",
            comment: "Singular Impact kind noun"
        )
        static let nounMetadata = LocalizedStringResource(
            "deleteImpact.kind.metadataField.other",
            defaultValue: "metadata fields",
            comment: "Plural Impact kind noun"
        )
        static let nounCredibilityGrade = LocalizedStringResource(
            "deleteImpact.kind.credibilityGrade.one",
            defaultValue: "credibility grade",
            comment: "Singular Impact kind noun"
        )
        static let nounCredibilityGrades = LocalizedStringResource(
            "deleteImpact.kind.credibilityGrade.other",
            defaultValue: "credibility grades",
            comment: "Plural Impact kind noun"
        )
        static let nounSubjectType = LocalizedStringResource(
            "deleteImpact.kind.subjectType.one",
            defaultValue: "subject type",
            comment: "Singular Impact kind noun"
        )
        static let nounSubjectTypes = LocalizedStringResource(
            "deleteImpact.kind.subjectType.other",
            defaultValue: "subject types",
            comment: "Plural Impact kind noun"
        )
        static let nounProperty = LocalizedStringResource(
            "deleteImpact.kind.property.one",
            defaultValue: "property",
            comment: "Singular Impact kind noun"
        )
        static let nounProperties = LocalizedStringResource(
            "deleteImpact.kind.property.other",
            defaultValue: "properties",
            comment: "Plural Impact kind noun"
        )
        static let nounPropertyTerm = LocalizedStringResource(
            "deleteImpact.kind.propertyTerm.one",
            defaultValue: "property term",
            comment: "Singular Impact kind noun"
        )
        static let nounPropertyTerms = LocalizedStringResource(
            "deleteImpact.kind.propertyTerm.other",
            defaultValue: "property terms",
            comment: "Plural Impact kind noun"
        )
        static let nounCanonicalEntity = LocalizedStringResource(
            "deleteImpact.kind.canonicalEntity.one",
            defaultValue: "person, event, or place",
            comment: "Singular Impact kind noun for a Conclusion handle (PER-…, EVT-…, PLC-…)"
        )
        static let nounCanonicalEntities = LocalizedStringResource(
            "deleteImpact.kind.canonicalEntity.other",
            defaultValue: "people, events, or places",
            comment: "Plural Impact kind noun for Conclusion handles"
        )
        static let nounUser = LocalizedStringResource(
            "deleteImpact.kind.user.one",
            defaultValue: "user",
            comment: "Singular infra Impact kind noun"
        )
        static let nounUsers = LocalizedStringResource(
            "deleteImpact.kind.user.other",
            defaultValue: "users",
            comment: "Plural infra Impact kind noun"
        )
        static let nounProject = LocalizedStringResource(
            "deleteImpact.kind.project.one",
            defaultValue: "project",
            comment: "Singular infra Impact kind noun"
        )
        static let nounProjects = LocalizedStringResource(
            "deleteImpact.kind.project.other",
            defaultValue: "projects",
            comment: "Plural infra Impact kind noun"
        )
        static let nounFile = LocalizedStringResource(
            "deleteImpact.kind.file.one",
            defaultValue: "file",
            comment: "Singular Impact kind noun"
        )
        static let nounFiles = LocalizedStringResource(
            "deleteImpact.kind.file.other",
            defaultValue: "files",
            comment: "Plural Impact kind noun"
        )
        static let nounNarrative = LocalizedStringResource(
            "deleteImpact.kind.narrative.one",
            defaultValue: "narrative",
            comment: "Singular Impact kind noun"
        )
        static let nounNarratives = LocalizedStringResource(
            "deleteImpact.kind.narrative.other",
            defaultValue: "narratives",
            comment: "Plural Impact kind noun"
        )
        static let nounReconciliationClaim = LocalizedStringResource(
            "deleteImpact.kind.reconciliation_claim.one",
            defaultValue: "reconciliation claim",
            comment: "Singular Impact kind noun"
        )
        static let nounReconciliationClaims = LocalizedStringResource(
            "deleteImpact.kind.reconciliation_claim.other",
            defaultValue: "reconciliation claims",
            comment: "Plural Impact kind noun"
        )
        static let nounItem = LocalizedStringResource(
            "deleteImpact.kind.item.one",
            defaultValue: "item",
            comment: "Singular fallback Impact kind noun"
        )
        static let nounItems = LocalizedStringResource(
            "deleteImpact.kind.item.other",
            defaultValue: "items",
            comment: "Plural fallback Impact kind noun"
        )

        static func confirmTitle(noun: String, ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.confirm.title",
                defaultValue: "Delete %1$@ %2$@?",
                comment: "Allowed-delete confirm title; arguments are kind noun and ref"
            ), noun, ref)
        }

        /// Allowed Subject delete: the handle(s) the Subject leaves. Arguments: handle refs (joined).
        static func leavesHandle(handleRefs: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.confirm.leavesHandle",
                defaultValue: "It’s removed from %@. That record stays.",
                comment: "Allowed Subject delete confirm, appended after the consequence; argument is the handle ref(s), e.g. PER-7KD45"
            ), handleRefs)
        }

        /// Fallback for a cascade via the Mac has no specific copy for. Arguments: kind noun, refs.
        static func alsoAffects(noun: String, refs: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.confirm.alsoAffects",
                defaultValue: "This also changes %1$@ %2$@.",
                comment: "Allowed-delete confirm line for a non-blocking cascade without specific copy; arguments are kind noun and ref list"
            ), noun, refs)
        }

        /// Allowed Observation delete: the handle(s) whose claims pinned it. Arguments: handle refs (joined).
        static func leavesEvidence(handleRefs: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.confirm.leavesEvidence",
                defaultValue: "It’s also removed from the evidence for %@. Those claims stay, with less evidence.",
                comment: "Allowed Observation delete confirm, appended after the consequence; argument is the handle ref(s) whose Identity Claims pinned it"
            ), handleRefs)
        }

        static let confirmMessage = LocalizedStringResource(
            "deleteImpact.confirm.message",
            defaultValue: "It’s erased from the catalog. This can’t be undone.",
            comment: "Allowed-delete confirm consequence"
        )

        static let nothingReferences = LocalizedStringResource(
            "deleteImpact.confirm.nothingReferences",
            defaultValue: "Nothing else references it.",
            comment: "Quiet line on an allowed delete confirm"
        )

        static func deleteAction(noun: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.confirm.delete",
                defaultValue: "Delete %@",
                comment: "Allowed-delete confirm button; argument is kind noun"
            ), noun)
        }

        static func keepAction(noun: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.confirm.keep",
                defaultValue: "Keep %@",
                comment: "Allowed-delete keep button; argument is kind noun"
            ), noun)
        }

        static func noticeTitle(noun: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.notice.title",
                defaultValue: "You can’t delete this %@",
                comment: "Blocked-delete notice title; argument is kind noun"
            ), noun)
        }

        static let noticeSubtitle = LocalizedStringResource(
            "deleteImpact.notice.subtitle",
            defaultValue: "The following items need to be deleted before this one can.",
            comment: "Blocked-delete notice subtitle when inbound groups are listed"
        )

        static let done = LocalizedStringResource(
            "deleteImpact.notice.done",
            defaultValue: "Done",
            comment: "Dismiss the blocked-delete notice"
        )

        static func overflow(count: Int, kind: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.overflow",
                defaultValue: "And %1$lld more %2$@",
                comment: "Overflow under a capped inbound list; arguments are remainder and plural kind noun"
            ), count, noun(kind, count: count))
        }

        static func viaHeading(via: String, kind: String, total: Int) -> String {
            if let format = viaFormats[via] {
                return L10n.format(format, total)
            }
            return viaFallback(kind: kind, total: total)
        }

        static func isKnownVia(_ via: String) -> Bool {
            viaFormats[via] != nil
        }

        static let viaUnknownOne = LocalizedStringResource(
            "deleteImpact.via.unknown.one",
            defaultValue: "%1$lld %2$@ references this",
            comment: "Unknown-via group heading, singular; arguments are total and kind noun"
        )
        static let viaUnknownOther = LocalizedStringResource(
            "deleteImpact.via.unknown.other",
            defaultValue: "%1$lld %2$@ reference this",
            comment: "Unknown-via group heading, plural; arguments are total and kind noun"
        )

        static func viaFallback(kind: String, total: Int) -> String {
            return L10n.format(total == 1 ? viaUnknownOne : viaUnknownOther, total, noun(kind, count: total))
        }

        private static let viaFormats: [String: LocalizedStringResource] = [
            "observations.citation_id": LocalizedStringResource(
                "deleteImpact.via.observationsCitationId",
                defaultValue: "%lld observations still belong to this",
                comment: "Inbound heading for observations.citation_id; argument is how many"
            ),
            "observations.subject_id": LocalizedStringResource(
                "deleteImpact.via.observationsSubjectId",
                defaultValue: "This subject has %lld observations",
                comment: "Inbound heading for observations.subject_id; argument is how many"
            ),
            "observations.value_subject_id": LocalizedStringResource(
                "deleteImpact.via.observationsValueSubjectId",
                defaultValue: "%lld observations use this as an endpoint",
                comment: "Inbound heading for observations.value_subject_id; argument is how many"
            ),
            "artifacts.source_id": LocalizedStringResource(
                "deleteImpact.via.artifactsSourceId",
                defaultValue: "%lld artifacts belong to this source",
                comment: "Inbound heading for artifacts.source_id; argument is how many"
            ),
            "subjects.source_id": LocalizedStringResource(
                "deleteImpact.via.subjectsSourceId",
                defaultValue: "%lld subjects belong to this source",
                comment: "Inbound heading for subjects.source_id; argument is how many"
            ),
            "citations.artifact_id": LocalizedStringResource(
                "deleteImpact.via.citationsArtifactId",
                defaultValue: "%lld citations are drawn from this artifact",
                comment: "Inbound heading for citations.artifact_id; argument is how many"
            ),
            "sources.source_type_id": LocalizedStringResource(
                "deleteImpact.via.sourcesSourceTypeId",
                defaultValue: "%lld sources have this type",
                comment: "Inbound heading for sources.source_type_id; argument is how many"
            ),
            "source_metadata.field_id": LocalizedStringResource(
                "deleteImpact.via.sourceMetadataFieldId",
                defaultValue: "%lld sources record a value for this field",
                comment: "Inbound heading for source_metadata.field_id; argument is how many"
            ),
            "source_credibility_assessments.credibility_grade_id": LocalizedStringResource(
                "deleteImpact.via.credibilityGradeId",
                defaultValue: "%lld sources use this grade",
                comment: "Inbound heading for source_credibility_assessments.credibility_grade_id; argument is how many"
            ),
            "subjects.subject_type_id": LocalizedStringResource(
                "deleteImpact.via.subjectsSubjectTypeId",
                defaultValue: "%lld subjects have this type",
                comment: "Inbound heading for subjects.subject_type_id; argument is how many"
            ),
            "observations.property_id": LocalizedStringResource(
                "deleteImpact.via.observationsPropertyId",
                defaultValue: "%lld observations use this property",
                comment: "Inbound heading for observations.property_id; argument is how many"
            ),
            "property_terms.property_id": LocalizedStringResource(
                "deleteImpact.via.propertyTermsPropertyId",
                defaultValue: "%lld terms belong to this property",
                comment: "Inbound heading for property_terms.property_id; argument is how many"
            ),
            "observations.value_term_id": LocalizedStringResource(
                "deleteImpact.via.observationsValueTermId",
                defaultValue: "%lld observations use this term",
                comment: "Inbound heading for observations.value_term_id; argument is how many"
            ),
            "artifacts.file_id": LocalizedStringResource(
                "deleteImpact.via.artifactsFileId",
                defaultValue: "%lld artifacts use this file",
                comment: "Inbound heading for artifacts.file_id; argument is how many"
            ),
        ]

        static let gateNotFoundTitle = LocalizedStringResource(
            "deleteImpact.gate.notFound.title",
            defaultValue: "Already gone",
            comment: "Extra-gate title for not_found"
        )
        static func gateNotFoundBody(ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.gate.notFound.body",
                defaultValue: "%@ no longer exists. There is nothing left to delete.",
                comment: "Extra-gate body for not_found; argument is the target ref"
            ), ref)
        }

        static let gateEdgeLockedTitle = LocalizedStringResource(
            "deleteImpact.gate.edgeLocked.title",
            defaultValue: "This observation is part of a connection",
            comment: "Extra-gate title for edge_locked"
        )
        static let gateEdgeLockedBody = LocalizedStringResource(
            "deleteImpact.gate.edgeLocked.body",
            defaultValue: "It ties a bridge to its endpoints, so you can’t delete it here. Delete the bridge on the evidence graph instead.",
            comment: "Extra-gate body for edge_locked"
        )

        static let gateOriginLockedTitle = LocalizedStringResource(
            "deleteImpact.gate.originLocked.title",
            defaultValue: "This is built-in vocabulary",
            comment: "Extra-gate title for origin_locked"
        )
        static let gateOriginLockedBody = LocalizedStringResource(
            "deleteImpact.gate.originLocked.body",
            defaultValue: "Provenencia-seeded properties and terms can’t be deleted. Plugin vocabulary is removed from the plugin manager, not here.",
            comment: "Extra-gate body for origin_locked"
        )

        static let gateInfraTitle = LocalizedStringResource(
            "deleteImpact.gate.infra.title",
            defaultValue: "This can’t be deleted here",
            comment: "Extra-gate title for infra"
        )
        static let gateInfraBody = LocalizedStringResource(
            "deleteImpact.gate.infra.body",
            defaultValue: "This record is part of the project itself, not a research record.",
            comment: "Extra-gate body for infra"
        )

        static let gateUnknownTitle = LocalizedStringResource(
            "deleteImpact.gate.unknown.title",
            defaultValue: "Not available right now",
            comment: "Fallback extra-gate title"
        )
        static func gateUnknownBody(ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.gate.unknown.body",
                defaultValue: "%@ can’t be deleted right now.",
                comment: "Fallback extra-gate body; argument is the target ref"
            ), ref)
        }

        static func summaryAllowed(noun: String, ref: String, title: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.summary.allowed",
                defaultValue: "Delete %1$@ %2$@, %3$@. This erases it and can’t be undone.",
                comment: "VoiceOver for allowed delete; arguments are noun, ref, title"
            ), noun, ref, title)
        }

        static func summaryBlocked(noun: String, ref: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.summary.blocked",
                defaultValue: "Blocked. You can’t delete %1$@ %2$@.",
                comment: "VoiceOver lead-in for a blocked delete; arguments are noun and ref"
            ), noun, ref)
        }

        static func summaryGroup(heading: String, refs: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.summary.group",
                defaultValue: "%1$@: %2$@.",
                comment: "VoiceOver one inbound group; arguments are heading and comma-separated refs"
            ), heading, refs)
        }

        static func summaryMore(count: Int) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.summary.more",
                defaultValue: ", and %lld more",
                comment: "VoiceOver remainder after the first named refs"
            ), count)
        }

        static func rowAccessibility(ref: String, title: String) -> String {
            return L10n.format(LocalizedStringResource(
                "deleteImpact.row.goTo",
                defaultValue: "%1$@, %2$@. Go to it.",
                comment: "Blocker row accessibility; arguments are ref and title"
            ), ref, title)
        }
    }

    /// Event title templates (S9-22). `EventTitleDisplay` fills them. The
    /// type word is the product name for the type key, else the catalog
    /// label, else `fallbackType` when the event has no type.
    enum EventTitle {
        static func ofOne(type: String, subject: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "eventTitle.ofOne",
                defaultValue: "%1$@ of %2$@",
                comment: "Event title for one subject. Arguments are the type word, then the subject."
            )
            return L10n.format(resource, locale: locale, type, subject)
        }

        static func marriage(a: String, b: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "eventTitle.marriage",
                defaultValue: "Marriage of %1$@ and %2$@",
                comment: "Event title for a marriage with two subjects. Arguments are the two people."
            )
            return L10n.format(resource, locale: locale, a, b)
        }

        static func etAl(type: String, first: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "eventTitle.etAl",
                defaultValue: "%1$@ of %2$@ et al.",
                comment: "Event title for several subjects, or three or more on a marriage. Arguments are the type word, then the first subject."
            )
            return L10n.format(resource, locale: locale, type, first)
        }

        static func atPlace(type: String, place: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "eventTitle.atPlace",
                defaultValue: "%1$@ at %2$@",
                comment: "Event title when there is no recorded name, subject, or label, and there is a place. Arguments are the type word, then the place."
            )
            return L10n.format(resource, locale: locale, type, place)
        }

        static func unspecified(type: String, locale: Locale = .autoupdatingCurrent) -> String {
            let resource = LocalizedStringResource(
                "eventTitle.unspecified",
                defaultValue: "Unspecified %@",
                comment: "Event title when there is no recorded name, subject, label, or place. Argument is the type word."
            )
            return L10n.format(resource, locale: locale, type)
        }

        static let unnamedPerson = LocalizedStringResource(
            "eventTitle.unnamedPerson",
            defaultValue: "unnamed person",
            comment: "Subject name in an event title when that person has no reconciled name"
        )

        static let fallbackType = LocalizedStringResource(
            "eventTitle.fallbackType",
            defaultValue: "Event",
            comment: "Event title type word when the event has no event type"
        )
    }

    /// Conclusion detail copy (S9-15, S9-16): field states, auto-reconciler
    /// outcomes, the Why table and the Person page. `ReconciledValueDisplay`
    /// and `PersonDetailContent` compose them.
    enum Conclusions {
        static let stateEmpty = LocalizedStringResource(
            "conclusions.state.empty",
            defaultValue: "Nothing recorded",
            comment: "Empty Conclusion field with no specific wording for its Property"
        )

        static let badgeMerged = LocalizedStringResource(
            "conclusions.badge.merged",
            defaultValue: "Merged",
            comment: "Badge on a Conclusion field whose one value several Sources agree on"
        )

        static let badgeMixed = LocalizedStringResource(
            "conclusions.badge.mixed",
            defaultValue: "Mixed",
            comment: "Badge on a Conclusion field whose records disagree, so several values are shown"
        )

        static let badgeConcluded = LocalizedStringResource(
            "conclusions.badge.concluded",
            defaultValue: "Concluded",
            comment: "Badge on a Conclusion field whose value the researcher concluded"
        )

        static let countConcluded = LocalizedStringResource(
            "conclusions.count.concluded",
            defaultValue: "by you",
            comment: "Support column of a concluded Conclusion field: the researcher set the value"
        )

        static let outcomeKept = LocalizedStringResource(
            "conclusions.outcome.kept",
            defaultValue: "kept",
            comment: "Auto-reconciler outcome: the record counted toward the value shown"
        )

        static let outcomeFolded = LocalizedStringResource(
            "conclusions.outcome.folded",
            defaultValue: "folded",
            comment: "Auto-reconciler outcome: the record's value was merged into another value"
        )

        static let outcomeOutvoted = LocalizedStringResource(
            "conclusions.outcome.outvoted",
            defaultValue: "outvoted",
            comment: "Auto-reconciler outcome: more Sources read another spelling"
        )

        static let outcomeWeak = LocalizedStringResource(
            "conclusions.outcome.weak",
            defaultValue: "weak",
            comment: "Auto-reconciler outcome: weak evidence that stronger evidence disagrees with"
        )

        static let outcomeDenied = LocalizedStringResource(
            "conclusions.outcome.denied",
            defaultValue: "denied",
            comment: "Auto-reconciler outcome: a stronger negative record says the value is wrong"
        )

        static let outcomeProvisional = LocalizedStringResource(
            "conclusions.outcome.provisional",
            defaultValue: "provisional member",
            comment: "Auto-reconciler outcome: the record's Subject is only a provisional member, so it is reasoning only"
        )

        static let outcomeNoEvidence = LocalizedStringResource(
            "conclusions.outcome.noEvidence",
            defaultValue: "no usable value",
            comment: "Auto-reconciler outcome: the record states nothing usable for this field, such as a name without parts"
        )

        static let outcomeAgainst = LocalizedStringResource(
            "conclusions.outcome.against",
            defaultValue: "disagrees · did not eliminate",
            comment: "Auto-reconciler outcome of a negative record that eliminated nothing; it still counts against the value"
        )

        static let weakLowTrustSource = LocalizedStringResource(
            "conclusions.weak.lowTrustSource",
            defaultValue: "low-trust Source",
            comment: "Why evidence is weak: the Source's credibility is below standard"
        )

        static let weakUncertainTranscription = LocalizedStringResource(
            "conclusions.weak.uncertainTranscription",
            defaultValue: "uncertain transcription",
            comment: "Why evidence is weak: the Citation's transcription is marked uncertain"
        )

        static let weakLowConfidenceClaim = LocalizedStringResource(
            "conclusions.weak.lowConfidenceClaim",
            defaultValue: "low-confidence claim",
            comment: "Why evidence is weak: the member's Identity Claim confidence is below moderate"
        )

        static let readAsNone = LocalizedStringResource(
            "conclusions.readAs.none",
            defaultValue: "—",
            comment: "Why table, Read as column: the record states nothing for this field"
        )

        static let whyReadAs = LocalizedStringResource(
            "conclusions.why.readAs",
            defaultValue: "Read as",
            comment: "Why table column heading: what the record says"
        )

        static let whySource = LocalizedStringResource(
            "conclusions.why.source",
            defaultValue: "Source",
            comment: "Why table column heading: the Source the record is cited under"
        )

        static let whyOutcome = LocalizedStringResource(
            "conclusions.why.outcome",
            defaultValue: "Outcome",
            comment: "Why table column heading: what the auto-reconciler did with the record"
        )

        static let whyButton = LocalizedStringResource(
            "conclusions.why.button",
            defaultValue: "Why",
            comment: "Button on a Conclusion field that shows the records behind its value"
        )

        static let openEvent = LocalizedStringResource(
            "conclusions.detail.openEvent",
            defaultValue: "Open event",
            comment: "Icon button on a Person's birth or death date row that opens the event the date is read from"
        )

        static func openPlace(_ place: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.detail.openPlace",
                defaultValue: "Open %@",
                comment: "Icon button on a derived place row that opens the Place page; argument is the place name"
            ), place)
        }

        static func moreCount(_ count: Int) -> String {
            let resource = LocalizedStringResource(
                "conclusions.list.more",
                defaultValue: "+%lld",
                comment: "Badge after a list row title or line: how many more names or places there are. Argument is the count."
            )
            return L10n.format(resource, count)
        }

        static func lifeSpan(born: String, died: String) -> String {
            let resource = LocalizedStringResource(
                "conclusions.life.span",
                defaultValue: "%1$@ – %2$@",
                comment: "Persons row: birth date, then death date. Arguments are the two formatted dates."
            )
            return L10n.format(resource, born, died)
        }

        static func lifeSpanOpen(born: String) -> String {
            let resource = LocalizedStringResource(
                "conclusions.life.spanOpen",
                defaultValue: "%@ –",
                comment: "Persons row with a birth and no death: the birth date and an open dash. Argument is the formatted date."
            )
            return L10n.format(resource, born)
        }

        static func lifePlaces(born: String, died: String) -> String {
            let resource = LocalizedStringResource(
                "conclusions.life.places",
                defaultValue: "%1$@ → %2$@",
                comment: "Persons row: birth place, then death place. Arguments are the two place names."
            )
            return L10n.format(resource, born, died)
        }

        static func lineJoin(first: String, rest: String) -> String {
            let resource = LocalizedStringResource(
                "conclusions.line.join",
                defaultValue: "%1$@ · %2$@",
                comment: "Joins two parts of one secondary line (dates · places, a date · its phrase). Arguments are the two parts."
            )
            return L10n.format(resource, first, rest)
        }

        static let chainSeparator = LocalizedStringResource(
            "conclusions.place.chainSeparator",
            defaultValue: ", ",
            comment: "Separator between a Place and its parents in the parent chain (York, Upper Canada)"
        )

        static func chainOrPair(first: String, second: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.place.chainOr",
                defaultValue: "%1$@ or %2$@",
                comment: "Joins two simultaneous parent places when the date cannot pick one"
            ), first, second)
        }

        static func placePeriodOpen(start: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.place.periodOpen",
                defaultValue: "%@ –",
                comment: "Place period with a start and no end"
            ), start)
        }

        static func placePeriodUntil(end: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.place.periodUntil",
                defaultValue: "– %@",
                comment: "Place period with an end and no start"
            ), end)
        }

        static let placeSucceeded = LocalizedStringResource(
            "conclusions.place.succeeded",
            defaultValue: "Succeeded",
            comment: "Role on a Succession row: this Place succeeded the named Place"
        )

        static let placeSucceededBy = LocalizedStringResource(
            "conclusions.place.succeededBy",
            defaultValue: "Succeeded by",
            comment: "Role on a Succession row: this Place was succeeded by the named Place"
        )

        static func placeContainsMore(_ count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.place.containsMore",
                defaultValue: "%lld more",
                comment: "Remainder when Contains truncates a long list of child Places"
            ), count)
        }

        static func morePlaces(count: Int) -> String {
            let resource = LocalizedStringResource(
                "conclusions.a11y.morePlaces",
                defaultValue: "%lld more places",
                comment: "VoiceOver: other places a row stands for beyond the one it shows. Argument is the count."
            )
            return L10n.format(resource, count)
        }

        static let detailsSection = LocalizedStringResource(
            "conclusions.detail.details",
            defaultValue: "Details",
            comment: "Section heading over a Person's fields"
        )

        static let personNoLikeness = LocalizedStringResource(
            "conclusions.person.noLikeness",
            defaultValue: "No likeness recorded",
            comment: "Accessibility label of the empty portrait slot on a Person page"
        )

        static let eventNoImage = LocalizedStringResource(
            "conclusions.event.noImage",
            defaultValue: "No image recorded",
            comment: "Thumbnail placeholder on an Event page when no image is recorded"
        )

        static let placeNoImage = LocalizedStringResource(
            "conclusions.place.noImage",
            defaultValue: "No image recorded",
            comment: "Thumbnail placeholder on a Place page when no image is recorded"
        )

        static let personBornAbbr = LocalizedStringResource(
            "conclusions.person.born",
            defaultValue: "b.",
            comment: "Abbreviation before a Person's birth date and place in the page header"
        )

        static let personDiedAbbr = LocalizedStringResource(
            "conclusions.person.died",
            defaultValue: "d.",
            comment: "Abbreviation before a Person's death date and place in the page header"
        )

        static let dateUnknown = LocalizedStringResource(
            "conclusions.detail.dateUnknown",
            defaultValue: "date unknown",
            comment: "Page header: the Person's birth or death date is not known"
        )

        static let placeUnknown = LocalizedStringResource(
            "conclusions.detail.placeUnknown",
            defaultValue: "place unknown",
            comment: "Page header: the Person's birth or death place is not known"
        )

        static let eventDate = LocalizedStringResource(
            "conclusions.event.date",
            defaultValue: "Date",
            comment: "Row label on an Event page"
        )

        static let eventPlace = LocalizedStringResource(
            "conclusions.event.place",
            defaultValue: "Place",
            comment: "Row label on an Event page"
        )

        static let emptyEventDate = LocalizedStringResource(
            "conclusions.event.emptyDate",
            defaultValue: "No date recorded",
            comment: "Empty Date row on an Event page"
        )

        static let emptyEventPlace = LocalizedStringResource(
            "conclusions.event.emptyPlace",
            defaultValue: "No place recorded",
            comment: "Empty Place row on an Event page until places are filled"
        )

        static let placeNames = LocalizedStringResource(
            "conclusions.place.names",
            defaultValue: "Names",
            comment: "Row label for a Place's toponyms"
        )

        static let emptyPlaceNames = LocalizedStringResource(
            "conclusions.place.emptyNames",
            defaultValue: "No names recorded",
            comment: "Empty Names row on a Place page"
        )

        static let placePeriod = LocalizedStringResource(
            "conclusions.place.period",
            defaultValue: "Period",
            comment: "Row label for when a Place was current"
        )

        static let emptyPlacePeriod = LocalizedStringResource(
            "conclusions.place.emptyPeriod",
            defaultValue: "No period recorded",
            comment: "Empty Period row on a Place page until periods are filled"
        )

        static let placeNoParent = LocalizedStringResource(
            "conclusions.place.noParent",
            defaultValue: "No parent place recorded",
            comment: "Header line when a Place has no parent in the hierarchy"
        )

        static let placePartOf = LocalizedStringResource(
            "conclusions.place.partOf",
            defaultValue: "Part of",
            comment: "Section heading for the places this Place belongs to"
        )

        static let placePartOfAside = LocalizedStringResource(
            "conclusions.place.partOfAside",
            defaultValue: "Span is when the link holds",
            comment: "Aside under the Part of heading on a Place page"
        )

        static let placePartOfEmpty = LocalizedStringResource(
            "conclusions.place.partOfEmpty",
            defaultValue: "Not part of any recorded place",
            comment: "Empty Part of section on a Place page"
        )

        static let placeContains = LocalizedStringResource(
            "conclusions.place.contains",
            defaultValue: "Contains",
            comment: "Section heading for the places inside this Place"
        )

        static let placeContainsAside = LocalizedStringResource(
            "conclusions.place.containsAside",
            defaultValue: "Direct children only",
            comment: "Aside under the Contains heading on a Place page"
        )

        static let placeContainsEmpty = LocalizedStringResource(
            "conclusions.place.containsEmpty",
            defaultValue: "No places recorded as part of this one",
            comment: "Empty Contains section on a Place page"
        )

        static let placeSuccession = LocalizedStringResource(
            "conclusions.place.succession",
            defaultValue: "Succession",
            comment: "Section heading for a Place's renames and mergers"
        )

        static let placeSuccessionAside = LocalizedStringResource(
            "conclusions.place.successionAside",
            defaultValue: "Renames and mergers · not part of the hierarchy",
            comment: "Aside under the Succession heading on a Place page"
        )

        static let placeSuccessionEmpty = LocalizedStringResource(
            "conclusions.place.successionEmpty",
            defaultValue: "No predecessor or successor recorded",
            comment: "Empty Succession section on a Place page"
        )

        static let personName = LocalizedStringResource(
            "conclusions.person.name",
            defaultValue: "Name",
            comment: "Row label on a Person page"
        )

        static let personBirthDate = LocalizedStringResource(
            "conclusions.person.birthDate",
            defaultValue: "Birth date",
            comment: "Row label on a Person page"
        )

        static let personBirthPlace = LocalizedStringResource(
            "conclusions.person.birthPlace",
            defaultValue: "Birth place",
            comment: "Row label on a Person page"
        )

        static let personDeathDate = LocalizedStringResource(
            "conclusions.person.deathDate",
            defaultValue: "Death date",
            comment: "Row label on a Person page"
        )

        static let personDeathPlace = LocalizedStringResource(
            "conclusions.person.deathPlace",
            defaultValue: "Death place",
            comment: "Row label on a Person page"
        )

        static let emptyName = LocalizedStringResource(
            "conclusions.empty.name",
            defaultValue: "No name recorded",
            comment: "Empty Name row on a Person page"
        )

        static let emptySexAtBirth = LocalizedStringResource(
            "conclusions.empty.sexAtBirth",
            defaultValue: "No sex at birth recorded",
            comment: "Empty Sex at birth row on a Person page"
        )

        static let emptyBirthDate = LocalizedStringResource(
            "conclusions.empty.birthDate",
            defaultValue: "No birth date recorded",
            comment: "Empty Birth date row on a Person page"
        )

        static let emptyBirthPlace = LocalizedStringResource(
            "conclusions.empty.birthPlace",
            defaultValue: "No birth place recorded",
            comment: "Empty Birth place row on a Person page"
        )

        static let emptyDeathDate = LocalizedStringResource(
            "conclusions.empty.deathDate",
            defaultValue: "No death date recorded",
            comment: "Empty Death date row on a Person page"
        )

        static let emptyDeathPlace = LocalizedStringResource(
            "conclusions.empty.deathPlace",
            defaultValue: "No death place recorded",
            comment: "Empty Death place row on a Person page"
        )

        static let a11yConcluded = LocalizedStringResource(
            "conclusions.a11y.concluded",
            defaultValue: "concluded",
            comment: "VoiceOver state of a concluded Conclusion field"
        )

        static let a11yBorn = LocalizedStringResource(
            "conclusions.a11y.born",
            defaultValue: "Born",
            comment: "VoiceOver word before a Person's birth date and place"
        )

        static let a11yDied = LocalizedStringResource(
            "conclusions.a11y.died",
            defaultValue: "Died",
            comment: "VoiceOver word before a Person's death date and place"
        )

        static func sourceCount(_ count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.sourceCount",
                defaultValue: "%lld Sources",
                comment: "How many Sources support a value; argument is the count"
            ), count)
        }

        static func againstCount(_ count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.againstCount",
                defaultValue: "%lld records disagree",
                comment: "Negative records that count against a Conclusion field's value; argument is the count"
            ), count)
        }

        static func otherValues(_ count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.otherValues",
                defaultValue: "%lld other values",
                comment: "Button on a mixed Conclusion field that shows its other values; argument is how many"
            ), count)
        }

        static func memberCount(_ count: Int) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.memberCount",
                defaultValue: "%lld members",
                comment: "Page header: how many Subjects are members of this handle; argument is the count"
            ), count)
        }

        static func outcomeFoldedInto(_ value: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.outcome.foldedInto",
                defaultValue: "folded into %@",
                comment: "Auto-reconciler outcome: the record's value was merged into another; argument is that value"
            ), value)
        }

        static func outcomeOutvotedBy(_ support: Int, sources: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.outcome.outvotedBy",
                defaultValue: "outvoted (%1$lld of %2$@)",
                comment: "Auto-reconciler outcome: a majority read otherwise; arguments are the winning value's Source count and every voting Source as text (e.g. 3 Sources)"
            ), support, sources)
        }

        static func outcomeWeakBecause(_ causes: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.outcome.weakBecause",
                defaultValue: "weak · %@",
                comment: "Auto-reconciler outcome: weak evidence; argument lists why (low-trust Source, uncertain transcription, low-confidence claim)"
            ), causes)
        }

        static func outcomeDeniedBy(_ source: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.outcome.deniedBy",
                defaultValue: "denied by %@",
                comment: "Auto-reconciler outcome: a stronger negative record says the value is wrong; argument is that record's Source"
            ), source)
        }

        static func readAsNot(_ value: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.readAs.not",
                defaultValue: "not %@",
                comment: "Why table, Read as column: a negative record saying the value is not this; argument is the value"
            ), value)
        }

        static func whyTitle(_ value: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.why.title",
                defaultValue: "Why “%@”",
                comment: "Heading of the records behind a value; argument is the value"
            ), value)
        }

        static func whyThese(_ count: Int, label: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.why.these",
                defaultValue: "Why these %1$lld %2$@",
                comment: "Heading of the records behind several kept values; arguments are the count and the lowercased field label"
            ), count, label.localizedLowercase)
        }

        static func whyRecords(_ label: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.why.records",
                defaultValue: "Records considered for %@",
                comment: "VoiceOver label of the records behind a field; argument is the field label"
            ), label)
        }

        static func otherValuesList(_ label: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.otherValues.list",
                defaultValue: "Other values of %@",
                comment: "VoiceOver label of a mixed field's other values; argument is the field label"
            ), label)
        }

        static func a11ySingle(_ sources: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.a11y.single",
                defaultValue: "single, %@",
                comment: "VoiceOver state of a field with one value from one Source; argument is the Source count text"
            ), sources)
        }

        static func a11yMerged(_ sources: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.a11y.merged",
                defaultValue: "merged from %@",
                comment: "VoiceOver state of a field whose value several Sources agree on; argument is the Source count text"
            ), sources)
        }

        static func a11yMixed(_ sources: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.a11y.mixed",
                defaultValue: "mixed from %@",
                comment: "VoiceOver state of a field showing several values; argument is the Source count text"
            ), sources)
        }

        static func a11yMultiple(_ count: Int, sources: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.a11y.multiple",
                defaultValue: "%#@values@ from %2$@",
                comment: "VoiceOver state of a field that keeps several values; arguments are how many values and the Source count text"
            ), count, sources)
        }

        static func a11yEmpty(_ label: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.a11y.empty",
                defaultValue: "%@, empty",
                comment: "VoiceOver label of an empty field; argument is the field label"
            ), label)
        }

        static func a11yList(_ first: String, rest: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.a11y.list",
                defaultValue: "%1$@, %2$@",
                comment: "Joins two parts of a spoken Conclusion field (label, value, state, disagreement)"
            ), first, rest)
        }

        static func a11yVital(_ word: String, date: String, place: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.a11y.vital",
                defaultValue: "%1$@ %2$@, %3$@",
                comment: "VoiceOver for a Person header line; arguments are Born/Died, the date, and the place"
            ), word, date, place)
        }

        static func a11yRef(_ ref: String) -> String {
            L10n.format(LocalizedStringResource(
                "conclusions.a11y.ref",
                defaultValue: "Reference %@",
                comment: "VoiceOver label of a handle's ref (PER-…)"
            ), ref)
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
            return L10n.format(LocalizedStringResource(
                "error.catalog.unsupported_version",
                defaultValue: "Unsupported catalog version (%@).",
                comment: "FFI error catalog.unsupported_version; argument is catalog user_version"
            ), version)
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
        static let sourcesInUse = LocalizedStringResource(
            "error.sources.in_use",
            defaultValue: "This source still has artifacts or subjects.",
            comment: "FFI error sources.in_use"
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
        static let identityClaimsInvalid = LocalizedStringResource(
            "error.identityclaims.invalid",
            defaultValue: "Invalid identity claim.",
            comment: "FFI error identityclaims.invalid"
        )
        static let identityClaimsAlreadyMember = LocalizedStringResource(
            "error.identityclaims.already_member",
            defaultValue: "This subject already belongs to a person, event, or place.",
            comment: "FFI error identityclaims.already_member when Promote targets a subject that already has an accepted claim"
        )
        static let identityClaimsTypeMismatch = LocalizedStringResource(
            "error.identityclaims.type_mismatch",
            defaultValue: "A subject can only join a record of its own kind.",
            comment: "FFI error identityclaims.type_mismatch when the handle's type differs from the subject's"
        )
        static let canonicalEntitiesInvalid = LocalizedStringResource(
            "error.canonicalentities.invalid",
            defaultValue: "Invalid person, event, or place record.",
            comment: "FFI error canonicalentities.invalid"
        )
        static let promoteInvalid = LocalizedStringResource(
            "error.promote.invalid",
            defaultValue: "This subject can’t be promoted.",
            comment: "FFI error promote.invalid"
        )
        static let promoteUnsupportedType = LocalizedStringResource(
            "error.promote.unsupported_type",
            defaultValue: "Only people, events, and places can be promoted.",
            comment: "FFI error promote.unsupported_type for bridge subjects"
        )
        static let promoteStale = LocalizedStringResource(
            "error.promote.stale",
            defaultValue: "The catalog changed while this proposal was open. Review it again.",
            comment: "FFI error promote.stale when a Done loses the revision race"
        )
        static let conclusionDetailsNotFound = LocalizedStringResource(
            "error.conclusiondetails.not_found",
            defaultValue: "This record no longer exists. It may have been merged into another.",
            comment: "FFI error conclusiondetails.not_found: the handle was merged or deleted"
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
        static let artifactsInUse = LocalizedStringResource(
            "error.artifacts.in_use",
            defaultValue: "This artifact still has citations.",
            comment: "FFI error artifacts.in_use"
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
                    message: L10n.string(ingestUnsupportedOfficeHelp)
                )
            case .archive:
                return IngestCallout(
                    title: ingestUnsupportedArchiveTitle,
                    message: L10n.string(ingestUnsupportedArchiveHelp)
                )
            case .executable:
                return IngestCallout(
                    title: ingestUnsupportedExecutableTitle,
                    message: L10n.string(ingestUnsupportedExecutableHelp)
                )
            case .empty:
                return IngestCallout(
                    title: ingestEmptyTitle,
                    message: L10n.string(ingestEmptyHelp)
                )
            case .tooLarge:
                let size = sizeLabel ?? "?"
                return IngestCallout(
                    title: ingestTooLargeTitle,
                    message: L10n.format(ingestTooLargeHelp, size)
                )
            case .unidentified:
                return IngestCallout(
                    title: ingestUnidentifiedTitle,
                    message: L10n.string(ingestUnidentifiedHelp)
                )
            case .disallowedSniff, .genericType:
                if let typeLabel, !typeLabel.isEmpty, typeLabel != "That" {
                    return IngestCallout(
                        title: ingestUnsupportedTypeGenericTitle,
                        message: L10n.format(LocalizedStringResource(
                            "error.ingest.unsupported_type.helpWithLabel",
                            defaultValue: "%@ files aren’t accepted. Artifacts hold images, PDFs, Word documents, plain text (including CSV and Markdown), audio, or video — up to 512 MB. Choose a different file.",
                            comment: "Callout help for unsupported type with label; argument is short type"
                        ), typeLabel)
                    )
                }
                return IngestCallout(
                    title: ingestUnsupportedTypeGenericTitle,
                    message: L10n.string(ingestUnsupportedTypeHelp)
                )            case .notAFile:
                return IngestCallout(
                    title: ingestNotAFileTitle,
                    message: L10n.string(ingestNotAFileHelp)
                )
            case .symlink:
                return IngestCallout(
                    title: ingestSymlinkTitle,
                    message: L10n.string(ingestSymlinkHelp)
                )
            case .missing:
                return IngestCallout(
                    title: ingestMissingTitle,
                    message: L10n.string(ingestMissingHelp)
                )
            case .permission, .unreadable:
                return IngestCallout(
                    title: ingestPermissionDeniedTitle,
                    message: L10n.string(ingestPermissionDeniedHelp)
                )
            case .multiFile:
                return IngestCallout(
                    title: ingestMultiFileTitle,
                    message: L10n.string(ingestMultiFileHelp)
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
                    message: L10n.string(ingestInvalid)
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
        static let sourceTypesOriginLocked = LocalizedStringResource(
            "error.sourcetypes.origin_locked",
            defaultValue: "A plugin owns that source type, so it cannot be deleted.",
            comment: "FFI error sourcetypes.origin_locked"
        )
        static func sourceTypesDuplicateKey(key: String) -> String {
            return L10n.format(LocalizedStringResource(
                "error.sourcetypes.duplicate_key",
                defaultValue: "You already have a type with the key %@. Give this one a different label.",
                comment: "FFI error sourcetypes.duplicate_key; argument is the colliding key"
            ), key)
        }
        static let metadataInvalid = LocalizedStringResource(
            "error.metadatafields.invalid",
            defaultValue: "Invalid metadata field.",
            comment: "FFI error metadatafields.invalid"
        )
        static func metadataDuplicateKey(key: String) -> String {
            return L10n.format(LocalizedStringResource(
                "error.metadatafields.duplicate_key",
                defaultValue: "You already have a field with the key %@. Give this one a different label.",
                comment: "FFI error metadatafields.duplicate_key; argument is the colliding key"
            ), key)
        }
        static let metadataInUse = LocalizedStringResource(
            "error.metadatafields.in_use",
            defaultValue: "That metadata field is still used on one or more sources.",
            comment: "FFI error metadatafields.in_use"
        )
        static let metadataOriginLocked = LocalizedStringResource(
            "error.metadatafields.origin_locked",
            defaultValue: "A plugin owns that metadata field, so it cannot be deleted.",
            comment: "FFI error metadatafields.origin_locked"
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
            return L10n.format(LocalizedStringResource(
                "error.properties.duplicate_key",
                defaultValue: "You already have a property with the key %@. Give this one a different label.",
                comment: "FFI error properties.duplicate_key; argument is the colliding key"
            ), key)
        }
        static let propertiesInUse = LocalizedStringResource(
            "error.properties.in_use",
            defaultValue: "That property is still used on observations or still has terms.",
            comment: "FFI error properties.in_use"
        )
        static let propertiesOriginLocked = LocalizedStringResource(
            "error.properties.origin_locked",
            defaultValue: "Provenencia-seeded and plugin properties cannot be deleted here.",
            comment: "FFI error properties.origin_locked"
        )
        static let propertyTermsInvalid = LocalizedStringResource(
            "error.propertyterms.invalid",
            defaultValue: "Invalid property term.",
            comment: "FFI error propertyterms.invalid"
        )
        static func propertyTermsDuplicateKey(key: String) -> String {
            return L10n.format(LocalizedStringResource(
                "error.propertyterms.duplicate_key",
                defaultValue: "You already have a term with the key %@. Give this one a different label.",
                comment: "FFI error propertyterms.duplicate_key; argument is the colliding key"
            ), key)
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
        static let deleteImpactInvalid = LocalizedStringResource(
            "error.deleteimpact.invalid",
            defaultValue: "That delete preview isn’t valid.",
            comment: "FFI error deleteimpact.invalid"
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
        static let citationsInUse = LocalizedStringResource(
            "error.citations.in_use",
            defaultValue: "This citation still has observations.",
            comment: "FFI error citations.in_use"
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
        static let observationsInUse = LocalizedStringResource(
            "error.observations.in_use",
            defaultValue: "This observation is still in use.",
            comment: "FFI error observations.in_use"
        )
        static let dateValuesInvalid = LocalizedStringResource(
            "error.datevalues.invalid",
            defaultValue: "Invalid date value.",
            comment: "FFI error datevalues.invalid"
        )
        static let nameValuesInvalid = LocalizedStringResource(
            "error.namevalues.invalid",
            defaultValue: "Invalid name value.",
            comment: "FFI error namevalues.invalid"
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
                return L10n.string(catalogAlreadyExists)
            case "catalog.already_open":
                return L10n.string(catalogAlreadyOpen)
            case "catalog.not_a_project":
                return L10n.string(catalogNotAProject)
            case "catalog.unsupported_version":
                return catalogUnsupportedVersion(version: params.first ?? "?")
            case "catalog.invalid_folder_name":
                return L10n.string(catalogInvalidFolderName)
            case "catalog.closed":
                return L10n.string(catalogClosed)
            case "catalog.schema_mismatch":
                return L10n.string(catalogSchemaMismatch)
            case "project.invalid_metadata":
                return L10n.string(projectInvalidMetadata)
            case "project.missing_metadata":
                return L10n.string(projectMissingMetadata)
            case "users.invalid":
                return L10n.string(usersInvalid)
            case "audit.invalid":
                return L10n.string(auditInvalid)
            case "identity.not_found":
                return L10n.string(identityNotFound)
            case "identity.invalid_name":
                return L10n.string(identityInvalidName)
            case "identity.invalid_id":
                return L10n.string(identityInvalidID)
            case "identity.invalid_ref":
                return L10n.string(identityInvalidRef)
            case "install.not_found":
                return L10n.string(installNotFound)
            case "install.invalid":
                return L10n.string(installInvalid)
            case "onboarding.blank_name":
                return L10n.string(onboardingBlankName)
            case "onboarding.invalid_family_name":
                return L10n.string(onboardingInvalidFamilyName)
            case "onboarding.unknown_user":
                return L10n.string(onboardingUnknownUser)
            case "file.not_found":
                return L10n.string(fileNotFound)
            case "sources.invalid":
                return L10n.string(sourcesInvalid)
            case "sources.in_use":
                return L10n.string(sourcesInUse)
            case "subjects.invalid":
                return L10n.string(subjectsInvalid)
            case "subjects.in_use":
                return L10n.string(subjectsInUse)
            case "identityclaims.invalid":
                return L10n.string(identityClaimsInvalid)
            case "identityclaims.already_member":
                return L10n.string(identityClaimsAlreadyMember)
            case "identityclaims.type_mismatch":
                return L10n.string(identityClaimsTypeMismatch)
            case "canonicalentities.invalid":
                return L10n.string(canonicalEntitiesInvalid)
            case "promote.invalid":
                return L10n.string(promoteInvalid)
            case "promote.unsupported_type":
                return L10n.string(promoteUnsupportedType)
            case "promote.stale":
                return L10n.string(promoteStale)
            case "conclusiondetails.not_found":
                return L10n.string(conclusionDetailsNotFound)
            case "subjectpositions.invalid":
                return L10n.string(subjectPositionsInvalid)
            case "subjecttypes.invalid":
                return L10n.string(subjectTypesInvalid)
            case "subjecttypes.duplicate_prefix":
                return L10n.string(subjectTypesDuplicatePrefix)
            case "artifacts.invalid":
                return L10n.string(artifactsInvalid)
            case "artifacts.in_use":
                return L10n.string(artifactsInUse)
            case "artifacts.file_already_attached":
                return L10n.string(artifactsFileAlreadyAttached)
            case "sourcecredibility.invalid":
                return L10n.string(sourceCredibilityInvalid)
            case "sourcemetadata.invalid":
                return L10n.string(sourceMetadataInvalid)
            case "files.invalid":
                return L10n.string(filesInvalid)
            case "ingest.invalid":
                return L10n.string(ingestInvalid)
            case "ingest.permission_denied":
                return L10n.string(ingestPermissionDenied)
            case "ingest.unsupported_office":
                return L10n.string(ingestUnsupportedOfficeHelp)
            case "ingest.unsupported_archive":
                return L10n.string(ingestUnsupportedArchiveHelp)
            case "ingest.unsupported_executable":
                return L10n.string(ingestUnsupportedExecutableHelp)
            case "ingest.unsupported_type":
                return ingestCallout(reason: .disallowedSniff, typeLabel: params.first).message
            case "ingest.unidentified":
                return L10n.string(ingestUnidentifiedHelp)
            case "ingest.empty":
                return L10n.string(ingestEmptyHelp)
            case "ingest.too_large":
                return ingestCallout(reason: .tooLarge, sizeLabel: params.first).message
            case "ingest.not_a_file":
                return L10n.string(ingestNotAFileHelp)
            case "ingest.symlink":
                return L10n.string(ingestSymlinkHelp)
            case "ingest.missing":
                return L10n.string(ingestMissingHelp)
            case "sourcetypes.invalid":
                return L10n.string(sourceTypesInvalid)
            case "sourcetypes.in_use":
                return L10n.string(sourceTypesInUse)
            case "sourcetypes.origin_locked":
                return L10n.string(sourceTypesOriginLocked)
            case "sourcetypes.duplicate_key":
                return sourceTypesDuplicateKey(key: params.first ?? "?")
            case "metadatafields.invalid":
                return L10n.string(metadataInvalid)
            case "metadatafields.duplicate_key":
                return metadataDuplicateKey(key: params.first ?? "?")
            case "metadatafields.in_use":
                return L10n.string(metadataInUse)
            case "metadatafields.origin_locked":
                return L10n.string(metadataOriginLocked)
            case "sourcevocab.invalid":
                return L10n.string(sourceVocabInvalid)
            case "properties.invalid":
                return L10n.string(propertiesInvalid)
            case "properties.duplicate_key":
                return propertiesDuplicateKey(key: params.first ?? "?")
            case "properties.in_use":
                return L10n.string(propertiesInUse)
            case "properties.origin_locked":
                return L10n.string(propertiesOriginLocked)
            case "propertyterms.invalid":
                return L10n.string(propertyTermsInvalid)
            case "propertyterms.duplicate_key":
                return propertyTermsDuplicateKey(key: params.first ?? "?")
            case "propertyterms.locked":
                return L10n.string(propertyTermsLocked)
            case "propertyterms.in_use":
                return L10n.string(propertyTermsInUse)
            case "subjectvocab.invalid":
                return L10n.string(subjectVocabInvalid)
            case "subjectvocab.locked":
                return L10n.string(subjectVocabLocked)
            case "connect.invalid":
                return L10n.string(connectInvalid)
            case "connect.refused":
                return L10n.string(connectRefused)
            case "locator.invalid":
                return L10n.string(locatorInvalid)
            case "citations.invalid":
                return L10n.string(citationsInvalid)
            case "citations.in_use":
                return L10n.string(citationsInUse)
            case "observations.invalid":
                return L10n.string(observationsInvalid)
            case "observations.edge_locked":
                return L10n.string(observationsEdgeLocked)
            case "observations.in_use":
                return L10n.string(observationsInUse)
            case "datevalues.invalid":
                return L10n.string(dateValuesInvalid)
            case "namevalues.invalid":
                return L10n.string(nameValuesInvalid)
            case "deleteimpact.invalid":
                return L10n.string(deleteImpactInvalid)
            case "filederivatives.invalid":
                return L10n.string(fileDerivativesInvalid)
            case "filederivatives.unprocessable":
                return L10n.string(fileDerivativesUnprocessable)
            case "filederivatives.corrupt_object":
                return L10n.string(fileDerivativesCorruptObject)
            case "internal.unknown":
                return L10n.string(unknown)
            case "internal.unknown_method":
                return L10n.string(internalUnknownMethod)
            case "internal.migrations":
                return L10n.string(internalMigrations)
            default:
                return L10n.string(unknown)
            }
        }

        static func message(for error: Error) -> String {
            if let coded = error as? CoreInvokeError {
                switch coded {
                case .coded(_, let code, _, let params):
                    return message(code: code, params: params)
                case .failed:
                    return L10n.string(unknown)
                }
            }
            return L10n.string(unknown)
        }

        /// `ref.invalid` / `ref.invalid_prefix` / `ref.reserved_prefix` are
        /// minting / reserved-prefix failures and intentionally fall through to
        /// `unknown` unless a future client surfaces them as distinct copy.
    }
}
