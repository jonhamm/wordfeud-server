package localize

import "golang.org/x/text/language"

func Localized(lang language.Tag, text string) string {
	switch lang {
	case language.Danish:
		return danish(text)
	}
	return text
}

func danish(text string) string {
	switch text {
	case "Game completed after %d moves as %s has no more tiles in rack":
		return "Spillet afsluttet efter %d træk da %s ikke har flere bogstaver"
	case "Game completed after %d moves as there has been %d conequtive passes":
		return "Spillet afsluttet efter %d træk da der har været %d uafbrudte pas meldinger"
	case "Game is a draw between %d players: %s":
		return "Spillet er uafgjort mellem %d spillere: %s"
	case "Game is won by %s":
		return "Spillet er vundet af %s"
	case "Wrote game file after move %d \"%s\"":
		return "Skrev spil fil efter træk nummer %d \"%s\""
	case "Game file is %s":
		return "Spil filen er %s"
	case "%s scored %d with remaining rack %s":
		return "%s har scoret %d point og har disse bogstaver tilbage %s"
	}
	return text
}
