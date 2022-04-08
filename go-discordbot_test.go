package discrdbt_test

import (
	"testing"

	"github.com/dougrich/go-discordbot"
)

func TestPlaceholder(t *testing.T) {
	if discrdbt.Placeholder() != 2 {
		t.Fatal("Expected a placeholder value of 2")
	}
}
