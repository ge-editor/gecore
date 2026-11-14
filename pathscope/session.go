package pathscope

// Resolver builds the Matcher for a kind. The workspace package supplies it,
// so this package does not depend on workspace.
type Resolver func(Kind) *Matcher

// Session is the temporary scope of one mode (the Mark mode). It starts from
// the setting of the workspace, can be changed while the mode is open, and is
// dropped with the mode: it never writes back to the workspace.
type Session struct {
	kind    Kind
	resolve Resolver
	m       *Matcher
}

// NewSession starts a Session with the kind initial (All when it is not
// valid). resolve may be nil: the Matcher then has no workspaces, which is
// enough for All.
func NewSession(initial Kind, resolve Resolver) *Session {
	if !initial.Valid() {
		initial = All
	}
	s := &Session{kind: initial, resolve: resolve}
	s.m = s.build()
	return s
}

func (s *Session) build() *Matcher {
	if s.resolve != nil {
		if m := s.resolve(s.kind); m != nil {
			return m
		}
	}
	return New(s.kind, nil, nil)
}

// Kind returns the current kind of the Session.
func (s *Session) Kind() Kind { return s.kind }

// Matcher returns the Matcher of the current kind.
func (s *Session) Matcher() *Matcher { return s.m }

// Set changes the kind and builds the Matcher again. A kind that is not valid
// is ignored.
func (s *Session) Set(k Kind) {
	if !k.Valid() {
		return
	}
	s.kind = k
	s.m = s.build()
}
