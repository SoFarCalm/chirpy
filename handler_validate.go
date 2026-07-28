package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleanedBody := replaceBadWords(params.Body, []string{"kerfuffle", "sharbert", "fornax"})

	respondWithJSON(w, http.StatusOK, returnVals{
		Body: cleanedBody,
	})
}

func replaceBadWords(s string, badWords []string) string {
	words := strings.Split(s, " ")

	for i, word := range words {
		for _, badWord := range badWords {
			if strings.EqualFold(word, badWord) {
				words[i] = "****"
			}
		}
	}

	cleanedWords := strings.Join(words, " ")

	return cleanedWords
}

// Assuming the length validation passed, replace any of the following words in the Chirp with the static 4-character string ****:
// kerfuffle
// sharbert
// fornax

// Be sure to match against uppercase versions of the words as well, but not punctuation. "Sharbert!" does not need to be replaced, we'll consider it a different word due to the exclamation point.

// Finally, instead of the valid boolean, your handler should always return the cleaned version of the text in a JSON response, even if nothing changed:
