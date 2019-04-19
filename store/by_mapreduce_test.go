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
	var name = doc.name;
	if(name.toLowerCase().substring(0, 1) == "a") {
		emit(doc.name);
	}
}
	`)
	assert.NoError(t, err)
	fmt.Println(lines)
	assert.Len(t, lines, 1)
}
