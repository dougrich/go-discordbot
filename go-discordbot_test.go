package discordbot_test

import (
	"errors"
	"testing"

	"github.com/dougrich/go-discordbot"
	"github.com/stretchr/testify/assert"
)

type testPackage func(b *discordbot.Bot) error

func (t testPackage) Name() string {
	return "TestPackage"
}

func (t testPackage) Register(b *discordbot.Bot) error {
	return t(b)
}

func TestInclude(t *testing.T) {
	b := &discordbot.Bot{}
	b.Include(testPackage(func(b *discordbot.Bot) error {
		return errors.New("Internal error")
	}))
	assert.Error(t, b.ErrRegistration, "discordbot: unexpected error registering TestPackage ->\nInternal Error")
}
