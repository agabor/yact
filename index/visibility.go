package index

type Visibility int

const (
	VisibilityPublic Visibility = iota
	VisibilityPrivate
)

func languagesForVisibility(visibility Visibility) []indexLanguage {
	languages := make([]indexLanguage, len(indexLanguages))
	copy(languages, indexLanguages)
	for position := range languages {
		languages[position].Rules.PublicOnly = visibility == VisibilityPublic
	}
	return languages
}

func indexScopeDescription(visibility Visibility) string {
	switch visibility {
	case VisibilityPublic:
		return publicIndexScope
	case VisibilityPrivate:
		return privateIndexScope
	}
	return ""
}
