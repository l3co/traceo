package slug_test

import (
	"testing"

	"github.com/l3co/traceo-api/pkg/slug"
	"github.com/stretchr/testify/assert"
)

func TestGenerate_BasicName(t *testing.T) {
	result := slug.Generate("João Silva", "abc12345def")
	assert.Equal(t, "joao-silva-abc12345", result)
}

func TestGenerate_Accents(t *testing.T) {
	result := slug.Generate("José María González", "id123456")
	assert.Equal(t, "jose-maria-gonzalez-id123456", result)
}

func TestGenerate_SpecialChars(t *testing.T) {
	result := slug.Generate("Ana & Maria! @#$", "abcd1234")
	assert.Equal(t, "ana-maria-abcd1234", result)
}

func TestGenerate_MultipleSpaces(t *testing.T) {
	result := slug.Generate("Pedro   de   Souza", "xyz98765")
	assert.Equal(t, "pedro-de-souza-xyz98765", result)
}

func TestGenerate_EmptyName(t *testing.T) {
	result := slug.Generate("", "abc12345")
	assert.Equal(t, "missing-abc12345", result)
}

func TestGenerate_ShortID(t *testing.T) {
	result := slug.Generate("Ana", "ab")
	assert.Equal(t, "ana-ab", result)
}

func TestGenerate_EmptyID(t *testing.T) {
	result := slug.Generate("Ana Maria", "")
	assert.Equal(t, "ana-maria", result)
}
