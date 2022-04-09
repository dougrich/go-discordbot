package discordbot

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestArgumentsInt64(t *testing.T) {
	assert := assert.New(t)
	args := &Arguments{
		data: discordgo.ApplicationCommandInteractionData{
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(5),
				},
			},
		},
	}
	v := int64(0)
	err := args.Scan(&v)
	assert.NoError(err)
	assert.Equal(int64(5), v)
}
func TestArgumentsInt64_2(t *testing.T) {
	assert := assert.New(t)
	args := &Arguments{
		data: discordgo.ApplicationCommandInteractionData{
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(5),
				},
				{
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(7),
				},
			},
		},
	}
	a := int64(0)
	b := int64(0)
	err := args.Scan(&a, &b)

	assert.NoError(err)
	assert.Equal(int64(5), a)
	assert.Equal(int64(7), b)
}
func TestArgumentsInt64_underflow(t *testing.T) {
	assert := assert.New(t)
	args := &Arguments{
		data: discordgo.ApplicationCommandInteractionData{
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(5),
				},
			},
		},
	}
	a := int64(0)
	b := int64(0)
	err := args.Scan(&a, &b)

	assert.NoError(err)
	assert.Equal(int64(5), a)
	assert.Equal(int64(0), b)
}
func TestArgumentsInt64_overflow(t *testing.T) {
	assert := assert.New(t)
	args := &Arguments{
		data: discordgo.ApplicationCommandInteractionData{
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(5),
				},
				{
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(7),
				},
				{
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(9),
				},
			},
		},
	}
	a := int64(0)
	b := int64(0)
	err := args.Scan(&a, &b)

	assert.NoError(err)
	assert.Equal(int64(5), a)
	assert.Equal(int64(7), b)
}
