package syntax

import (
	"context"
	"testing"

	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
)

func TestStructuralReplacements(t *testing.T) {
	tests := []struct {
		name       string
		original   string
		proposed   string
		references []Reference
		want       string
		parser     *tree_sitter.Parser
	}{
		{
			name: "single reference in function",
			original: `
		func processUser(id int) error {
		    validate(id)
		    startTx()
		    updateUser(id)
		    commit()
		    return nil
		}`,
			proposed: `
		func processUser(id int) error {
		    // ... existing code ...
		    log.Info("processing user")
		    return nil
		}`,
			references: []Reference{
				3,
			},
			want: `
		func processUser(id int) error {
		    validate(id)
		    startTx()
		    updateUser(id)
		    commit()
		    log.Info("processing user")
		    return nil
		}`,
			parser: getParserForLanguage("go"),
		},
		{
			name: "modified function signature, ref at end of function",
			original: `
		func processUser(id int) error {
		    validate(id)
		    validateAgain(id)
		    prepForUpdate(id)
		    return update(id)
		}`,
			proposed: `
		func processUser(id int, force bool) (error, bool) {
		    if force {
		        log.Warn("will force update")
		    }
		    // ... existing code ...
		}`,
			references: []Reference{
				6,
			},
			want: `
		func processUser(id int, force bool) (error, bool) {
		    if force {
		        log.Warn("will force update")
		    }
		    validate(id)
		    validateAgain(id)
		    prepForUpdate(id)
		    return update(id)
		}`,
			parser: getParserForLanguage("go"),
		},
		{
			name: "modified function signature, ref at beginning of function",
			original: `
		func processUser(id int) error {
		    validate(id)
		    validateAgain(id)
		    prepForUpdate(id)
		    return update(id)
		}`,
			proposed: `
		func processUser(id int, force bool) (error, bool) {
		    // ... existing code ...
		    if force {
		        log.Warn("will force update")
		    }
		    return update(id)
		}`,
			references: []Reference{
				3,
			},
			want: `
		func processUser(id int, force bool) (error, bool) {
		    validate(id)
		    validateAgain(id)
		    prepForUpdate(id)
		    if force {
		        log.Warn("will force update")
		    }
		    return update(id)
		}`,
			parser: getParserForLanguage("go"),
		},
		{
			name: "bad formatting",
			original: `
		func processUser(id int) error {
		validate(id)
		validateAgain(id)
		prepForUpdate(id)
		return update(id)
		}`,
			proposed: `
		func processUser(id int, force bool) (error, bool) {
		// ... existing code ...
		if force {
		log.Warn("will force update")
		}
		return update(id)
		}`,
			references: []Reference{
				3,
			},
			want: `
		func processUser(id int, force bool) (error, bool) {
		validate(id)
		validateAgain(id)
		prepForUpdate(id)
		if force {
		log.Warn("will force update")
		}
		return update(id)
		}`,
			parser: getParserForLanguage("go"),
		},
		{
			name: "multiple refs in class/nested structures",
			original: `
		type UserService struct {
		    db *DB
		    cache *Cache
		}

		func (s *UserService) Process() {
		    s.validate()
		    s.update()
		    s.notify()
		}

		func (s *UserService) Update() {
		    s.db.begin()
		    s.db.exec()
		    s.db.commit()
		}`,
			proposed: `
		type UserService struct {
		    // ... existing fields ...
		    metrics *Metrics
		}

		func (s *UserService) Process() {
		    // ... existing validation ...
				s.metrics.Record()
		    // ... existing processing ...
		}

		func (s *UserService) Update() {
		    // ... existing code ...
		}`,
			references: []Reference{
				3, 8, 10, 14,
			},
			want: `
		type UserService struct {
		    db *DB
		    cache *Cache
		    metrics *Metrics
		}

		func (s *UserService) Process() {
		    s.validate()
		    s.update()
				s.metrics.Record()
		    s.notify()
		}

		func (s *UserService) Update() {
		    s.db.begin()
		    s.db.exec()
		    s.db.commit()
		}`,
			parser: getParserForLanguage("go"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyReferences(context.Background(), tt.original, tt.proposed, tt.references, tt.parser)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
