package headers

import "net/http"

// Generator handles header generation
type Generator struct {
	DefaultProfile Profile
}

// NewGenerator creates a generator with a default profile
func NewGenerator(profile Profile) *Generator {
	return &Generator{
		DefaultProfile: profile,
	}
}

// GetHeaders returns a populated http.Header based on the profile
func (g *Generator) GetHeaders() http.Header {
	h := make(http.Header)
	g.DefaultProfile.Apply(h)
	// Add other standard tracking headers or randomization here if needed
	return h
}
