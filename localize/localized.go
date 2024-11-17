package localize

import (
	"golang.org/x/text/language"
)

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
		return "Spillet afsluttet efter %d træk da %s ikke har flere brikker"
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
	case "%s scored %d and has %s left":
		return "%s har scoret %d point og har %s tilbage"
	case "Scrabble game":
		return "Scrabble spil"
	case "Random number generator seed:":
		return "Tilfældigtalsgenerator frø:"
	case "Remaining free tiles:":
		return "Tilbageværende frie brikker:"
	case "Number of moves in game:":
		return "Antal træk i spillet:"
	case "%d tiles":
		return "%d brikker"
	case "1 tile":
		return "1 brik"
	case "no tiles":
		return "ingen brikker"
	case "%s move number %d %s %s..%s \"%s\" gives score %d":
		return "%s træk nummer %d %s %s..%s \"%s\" giver %d point"
	case "%s has total score %d and %s":
		return "%s har samlet %d point og %s"
	}
	return text
}
