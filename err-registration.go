package discordbot

import "fmt"

type errRegistration struct {
	packageName string
	err         error
}

func (e errRegistration) Error() string {
	return fmt.Sprintf("discordbot: unexpected error registering %s ->\n%v", e.packageName, e.err)
}
