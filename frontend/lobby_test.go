package frontend

import (
	"testing"

	"github.com/1996krishna/guessthisstagings/api"
	"github.com/1996krishna/guessthisstagings/game"
)

func TestCreateLobby(t *testing.T) {
	data := api.CreateLobbyData(&game.Lobby{
		LobbyID: "TEST",
	})

	var previousSize uint8 = 0
	for _, suggestedSize := range data.SuggestedBrushSizes {
		if suggestedSize < previousSize {
			t.Error("Sorting in SuggestedBrushSizes is incorrect")
		}
	}

	for _, suggestedSize := range data.SuggestedBrushSizes {
		if suggestedSize < game.MinBrushSize {
			t.Errorf("suggested brushsize %d is below MinBrushSize %d", suggestedSize, game.MinBrushSize)
		}

		if suggestedSize > game.MaxBrushSize {
			t.Errorf("suggested brushsize %d is above MaxBrushSize %d", suggestedSize, game.MaxBrushSize)
		}
	}
}
