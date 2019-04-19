package store

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByMapReduce(t *testing.T) {
	s, err := fixture()
	assert.NoError(t, err)
	lines, err := s.ByMapReduce(PROJECT, []string{"drugstore", "user"}, `
function map(doc) {
	emit(doc.name);
}
	`)
	assert.NoError(t, err)
	fmt.Println(lines)
	assert.Len(t, lines, 2)
}
