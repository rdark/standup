package markdown

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/mvdan/xurls"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type Parser struct {
	md goldmark.Markdown
}

// NewParser creates a new Markdown Parser.
func NewParser() *Parser {
	return &Parser{
		md: goldmark.New(
			goldmark.WithExtensions(
				meta.Meta,
				extension.NewLinkify(
					extension.WithLinkifyAllowedProtocols([][]byte{
						[]byte("http:"),
						[]byte("https:"),
					}),
					extension.WithLinkifyURLRegexp(
						xurls.Strict,
					),
				),
			),
		),
	}
}

func (p *Parser) ParseNoteContent(content string, skipText []string, noteType NoteType) (*NoteContent, error) {
	bytes := []byte(content)

	bodyStart := findBodyStartAfterFrontMatter(bytes)

	body := strings.TrimSpace(
		string(bytes[bodyStart:]),
	)

	bytes = []byte(body)

	context := parser.NewContext()
	root := p.md.Parser().Parse(
		text.NewReader(bytes),
		parser.WithContext(context),
	)

	sections, err := parseSections(root, bytes, skipText)
	if err != nil {
		return nil, err
	}

	adjacentLinks, err := parseAdjacentLinks(root, bytes, noteType)
	if err != nil {
		return nil, err
	}

	return &NoteContent{
		Body:          body,
		Sections:      sections,
		AdjacentLinks: adjacentLinks,
	}, nil
}

var frontmatterRegex = regexp.MustCompile(`(?ms)^\s*-+\s*$.*?^\s*-+\s*$`)
var fileDateRegex = regexp.MustCompile(`^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])(?:\.md)?$`)

var relativeStandupRegex = regexp.MustCompile(`^\.\.\/standup\/\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])(?:\.md)?$`)
var relativeJournalRegex = regexp.MustCompile(`^\.\.\/journal\/\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])(?:\.md)?$`)

func findBodyStartAfterFrontMatter(source []byte) int {
	index := frontmatterRegex.FindIndex(source)
	if index == nil {
		return 0
	}

	return index[1]
}

// parse
func parseAdjacentLinks(root ast.Node, source []byte, sourceNoteType NoteType) ([]AdjacentLink, error) {
	adjacentLinks := make([]AdjacentLink, 0)

	err := ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind() {

			case ast.KindLink:
				link := n.(*ast.Link)
				if !isURL(string(link.Destination)) {
					var targetNoteType NoteType

					// work out the target note type and skip if it isn't one in scope for an AdjacentLink
					// TODO: make this less ugly
					if fileDateRegex.Match(link.Destination) {
						targetNoteType = sourceNoteType
						// match oddly linked of same type
					} else if sourceNoteType == NoteTypeJournal && relativeJournalRegex.Match(link.Destination) {
						targetNoteType = sourceNoteType
					} else if sourceNoteType == NoteTypeJournal && relativeStandupRegex.Match(link.Destination) {
						targetNoteType = NoteTypeStandup
						// match oddly linked of same type
					} else if sourceNoteType == NoteTypeStandup && relativeStandupRegex.Match(link.Destination) {
						targetNoteType = sourceNoteType
					} else if sourceNoteType == NoteTypeStandup && relativeJournalRegex.Match(link.Destination) {
						targetNoteType = NoteTypeJournal
					} else {
						return ast.WalkContinue, nil
					}

					for child := link.FirstChild(); child != nil; child = child.NextSibling() {

						if text := child.(*ast.Text); text != nil {
							adjacentLink := AdjacentLink{
								SourceNoteType:   sourceNoteType,
								TargetNoteType:   targetNoteType,
								Title:            string(text.Segment.Value(source)),
								Target:           string(link.Destination),
							}

							if lines := link.Parent().Lines(); lines != nil && lines.Len() > 0 {
								adjacentLink.LinkStart = lines.At(0).Start
								adjacentLink.LinkEnd = lines.At(0).Stop
							} else {
								return ast.WalkContinue, nil
							}

							adjacentLinks = append(adjacentLinks, adjacentLink)
						}
					}
				}
			}
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, err
	}

	return adjacentLinks, nil
}

// parseSections extracts each section from the body delimited by a heading
func parseSections(root ast.Node, source []byte, skipText []string) ([]Section, error) {
	const maxListDepth = 3
	sections := make([]Section, 0)
	var currentSection *Section
	listLevel := 0
	skipNext := false

	err := ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if skipNext {
			skipNext = false
			return ast.WalkSkipChildren, nil
		}

		if entering {
			switch n.Kind() {
			case ast.KindHeading:
				// Start a new section when we hit a heading
				if currentSection != nil {
					if lines := n.Lines(); lines != nil && lines.Len() > 0 {
						currentSection.ContentEnd = lines.At(0).Start
					}
					sections = append(sections, *currentSection)
				}

				// Extract heading text
				heading := n.(*ast.Heading)
				var title strings.Builder
				for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
					// TODO: Link handling within headings?
					if child.Kind() == ast.KindText {
						if text := child.(*ast.Text); text != nil {
							title.Write(text.Segment.Value(source))
						}
					}
				}

				// Create new section starting after the heading
				if lines := n.Lines(); lines != nil && lines.Len() > 0 {
					currentSection = &Section{
						Title:        title.String(),
						ContentStart: lines.At(lines.Len()-1).Stop + 1, // Start after the heading's newline
					}
				}

			case ast.KindFencedCodeBlock:
				if currentSection != nil {
					code := n.(*ast.FencedCodeBlock)
					if len(currentSection.Content) > 0 {
						currentSection.Content += "\n"
					}
					currentSection.Content += "```"
					if code.Info != nil {
						currentSection.Content += string(code.Info.Segment.Value(source))
					}
					currentSection.Content += "\n"
				}

			case ast.KindCodeBlock:
				if currentSection != nil {
					if len(currentSection.Content) > 0 {
						currentSection.Content += "\n"
					}
					currentSection.Content += "```\n"
				}

			case ast.KindBlockquote:
				if currentSection != nil {
					if len(currentSection.Content) > 0 {
						currentSection.Content += "\n"
					}
					currentSection.Content += "> "
				}

			case ast.KindText:
				if currentSection != nil {
					text := n.(*ast.Text)
					// no headings or paragraphs
					if !isParentKind(n, ast.KindHeading) && !isParentKind(n, ast.KindParagraph) {
						content := text.Segment.Value(source)
						if !slices.Contains(skipText, string(content)) && len(content) > 0 {
							appendToSnippet(currentSection, string(content), false)
						}
					}
				}

			case ast.KindLink:
				if currentSection != nil && !isParentKind(n, ast.KindHeading) {
					link := n.(*ast.Link)
					if isURL(string(link.Destination)) {
						appendToSnippet(currentSection, "[", false)
						for child := link.FirstChild(); child != nil; child = child.NextSibling() {
							if text := child.(*ast.Text); text != nil {
								appendToSnippet(currentSection, string(text.Segment.Value(source)), false)
							}
						}
						appendToSnippet(currentSection, "]("+strings.TrimSpace(string(link.Destination))+")", false)
						skipNext = true
					}
				}

			case ast.KindList:
				listLevel++
				if listLevel > maxListDepth {
					return ast.WalkStop, fmt.Errorf("list nesting too deep: maximum allowed is %d levels", maxListDepth)
				}
				return ast.WalkContinue, nil

			case ast.KindListItem:
				if currentSection != nil {
					if len(currentSection.Content) > 0 {
						currentSection.Content += "\n"
					}
					level := listLevel
					if level > maxListDepth {
						level = maxListDepth
					}
					currentSection.Content += strings.Repeat("    ", level-1) + "* "
				}
			}
		} else { // exiting node
			switch n.Kind() {
			case ast.KindList:
				listLevel--
				if currentSection != nil && listLevel == 0 {
					currentSection.Content += "\n"
				}
			case ast.KindFencedCodeBlock, ast.KindCodeBlock:
				if currentSection != nil {
					currentSection.Content += "\n```\n"
				}
			case ast.KindBlockquote:
				if currentSection != nil {
					currentSection.Content += "\n"
				}
			}
		}
		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, err
	}

	// Add the last section if exists
	if currentSection != nil {
		if lines := root.Lines(); lines != nil && lines.Len() > 0 {
			currentSection.ContentEnd = lines.At(lines.Len() - 1).Stop
		}
		sections = append(sections, *currentSection)
	}

	return sections, nil
}

// Helper function to check if a node's parent is of a specific kind
func isParentKind(n ast.Node, kind ast.NodeKind) bool {
	parent := n.Parent()
	return parent != nil && parent.Kind() == kind
}

func appendToSnippet(section *Section, content string, addSpace bool) {
	if section == nil {
		return
	}

	if addSpace && len(section.Content) > 0 &&
		!strings.HasSuffix(section.Content, "\n") &&
		!strings.HasSuffix(section.Content, " ") &&
		!strings.HasPrefix(content, "\n") {
		section.Content += " "
	}
	section.Content += content
}

// isURL returns whether the given string is a valid URL.
func isURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

type NoteType int

// Note types
const (
	NoteTypeJournal NoteType = iota
	NoteTypeStandup
)

func (t NoteType) String() string {
    switch t {
    case NoteTypeJournal:
        return "NoteTypeJournal"
    case NoteTypeStandup:
        return "NoteTypeStandup"
    default:
        return fmt.Sprintf("%d", int(t))
    }
}

// Section represents a portion of the overall document delimited by a heading
type Section struct {
	// Label of the link.
	Title string
	// Content of the section
	Content string
	// Start byte offset of the content
	ContentStart int
	// End byte offset of the content
	ContentEnd int
}

// AdjacentLink represents a link to an adjacent note
type AdjacentLink struct {
	// Type of the note where the link is defined
	SourceNoteType NoteType
	// Type of the note where the link points to
	TargetNoteType NoteType
	// The title of the link, matched from config; used to determine the type
	Title            string
	// The target of the link
	Target string
	// Start byte offset of the link as defined in the body
	LinkStart int
	// End byte offset of the link as defined in the body
	LinkEnd int
}

// NoteContent holds the data parsed from the note content.
type NoteContent struct {
	// Body is the content of the note
	Body string
	// Sections is a list of the sections within the body
	Sections []Section
	// A list of adjacent links
	AdjacentLinks []AdjacentLink
}
