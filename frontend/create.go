package frontend

import (
	"log"
	"net/http"

	"github.com/1996krishna/guessthis_new/api"
	"github.com/1996krishna/guessthis_new/game"
	"github.com/1996krishna/guessthis_new/state"
	"github.com/1996krishna/guessthis_new/translations"
)

//This file contains the API for the official web client.

func homePage(w http.ResponseWriter, r *http.Request) {
	translation, locale := determineTranslation(r)
	createPageData := createDefaultLobbyCreatePageData()
	createPageData.Translation = translation
	createPageData.Locale = locale

	err := pageTemplates.ExecuteTemplate(w, "lobby-create-page", createPageData)
	if err != nil {
		log.Printf("Error templating home page: %s\n", err)
	}
}

func createDefaultLobbyCreatePageData() *LobbyCreatePageData {
	return &LobbyCreatePageData{
		BasePageConfig:    currentBasePageConfig,
		SettingBounds:     game.LobbySettingBounds,
		Languages:         game.SupportedLanguages,
		Public:            "false",
		DrawingTime:       "120",
		Rounds:            "4",
		MaxPlayers:        "12",
		CustomWordsChance: "50",
		ClientsPerIPLimit: "1",
		EnableVotekick:    "true",
		Language:          "english",
	}
}

// LobbyCreatePageData defines all non-static data for the lobby create page.
type LobbyCreatePageData struct {
	*BasePageConfig
	*game.SettingBounds
	Translation       translations.Translation
	Locale            string
	Errors            []string
	Languages         map[string]string
	Public            string
	DrawingTime       string
	Rounds            string
	MaxPlayers        string
	CustomWords       string
	CustomWordsChance string
	ClientsPerIPLimit string
	EnableVotekick    string
	Language          string
}

// ssrCreateLobby allows creating a lobby, optionally returning errors that
// occurred during creation.
func ssrCreateLobby(w http.ResponseWriter, r *http.Request) {
	formParseError := r.ParseForm()
	if formParseError != nil {
		http.Error(w, formParseError.Error(), http.StatusBadRequest)
		return
	}

	language, languageInvalid := api.ParseLanguage(r.Form.Get("language"))
	drawingTime, drawingTimeInvalid := api.ParseDrawingTime(r.Form.Get("drawing_time"))
	rounds, roundsInvalid := api.ParseRounds(r.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := api.ParseMaxPlayers("12")
	customWords, customWordsInvalid := api.ParseCustomWords("")
	customWordChance, customWordChanceInvalid := api.ParseCustomWordsChance("0")
	clientsPerIPLimit, clientsPerIPLimitInvalid := api.ParseClientsPerIPLimit("3")
	enableVotekick, enableVotekickInvalid := api.ParseBoolean("enable votekick", "true")
	publicLobby, publicLobbyInvalid := api.ParseBoolean("public", "true")

	//Prevent resetting the form, since that would be annoying as hell.
	pageData := LobbyCreatePageData{
		BasePageConfig:    currentBasePageConfig,
		SettingBounds:     game.LobbySettingBounds,
		Languages:         game.SupportedLanguages,
		Public:            "true",
		DrawingTime:       r.Form.Get("drawing_time"),
		Rounds:            r.Form.Get("rounds"),
		MaxPlayers:        "12",
		CustomWords:       "",
		CustomWordsChance: "0",
		ClientsPerIPLimit: "3",
		EnableVotekick:    "true",
		Language:          r.Form.Get("language"),
	}

	if languageInvalid != nil {
		pageData.Errors = append(pageData.Errors, languageInvalid.Error())
	}
	if drawingTimeInvalid != nil {
		pageData.Errors = append(pageData.Errors, drawingTimeInvalid.Error())
	}
	if roundsInvalid != nil {
		pageData.Errors = append(pageData.Errors, roundsInvalid.Error())
	}
	if maxPlayersInvalid != nil {
		pageData.Errors = append(pageData.Errors, maxPlayersInvalid.Error())
	}
	if customWordsInvalid != nil {
		pageData.Errors = append(pageData.Errors, customWordsInvalid.Error())
	}
	if customWordChanceInvalid != nil {
		pageData.Errors = append(pageData.Errors, customWordChanceInvalid.Error())
	}
	if clientsPerIPLimitInvalid != nil {
		pageData.Errors = append(pageData.Errors, clientsPerIPLimitInvalid.Error())
	}
	if enableVotekickInvalid != nil {
		pageData.Errors = append(pageData.Errors, enableVotekickInvalid.Error())
	}
	if publicLobbyInvalid != nil {
		pageData.Errors = append(pageData.Errors, publicLobbyInvalid.Error())
	}

	translation, locale := determineTranslation(r)
	pageData.Translation = translation
	pageData.Locale = locale

	if len(pageData.Errors) != 0 {
		err := pageTemplates.ExecuteTemplate(w, "lobby-create-page", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var playerName = r.Form.Get("playername")

    var playerDisplayImage = r.Form.Get("activeimage")
    log.Printf("Image name: %s\n", playerDisplayImage)
    
	player, lobby, createError := game.CreateLobby(playerName, playerDisplayImage, language, publicLobby, drawingTime, rounds, maxPlayers, customWordChance, clientsPerIPLimit, customWords, enableVotekick)
	if createError != nil {
		pageData.Errors = append(pageData.Errors, createError.Error())
		templateError := pageTemplates.ExecuteTemplate(w, "lobby-create-page", pageData)
		if templateError != nil {
			userFacingError(w, templateError.Error())
		}

		return
	}

	lobby.WriteJSON = api.WriteJSON
	player.SetLastKnownAddress(api.GetIPAddressFromRequest(r))

	api.SetUsersessionCookie(w, player)

	//We only add the lobby if we could do all necessary pre-steps successfully.
	state.AddLobby(lobby)

	http.Redirect(w, r, currentBasePageConfig.RootPath+"/ssrEnterLobby?lobby_id="+lobby.LobbyID, http.StatusFound)
}
