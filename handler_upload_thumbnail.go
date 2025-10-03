package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	maxMemory := 10 << 20;

	err = r.ParseMultipartForm(int64(maxMemory))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse MultiPart Form", err)
		return
	}

	thumbnailFile, fileHeader , err := r.FormFile("thumbnail")
	ext := filepath.Ext(fileHeader.Filename)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get Thumbnail file", err)
		return
	}

	mimeType := fileHeader.Header.Get("Content-Type");
	mediaType, _, err := mime.ParseMediaType(mimeType)
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		fmt.Print(mediaType)
		respondWithError(w, http.StatusForbidden, "Invalid file format",err)
		return
	}


	videoDetails, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting video details", err)
		return
	}

	if videoDetails.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Un-Authorised", err)
		return
	}

	randomData := make([]byte,32)
	rand.Read(randomData)

	thumbnailFileName := filepath.Join(cfg.assetsRoot,fmt.Sprintf("%s%s",base64.RawStdEncoding.EncodeToString(randomData),ext))

	osFile, err := os.Create(thumbnailFileName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error writing file", err)
		return
	}

	thumbnailUrl := fmt.Sprintf("http://localhost:8091/%s",thumbnailFileName)
	videoDetails.ThumbnailURL = &thumbnailUrl
	
	fmt.Print(thumbnailUrl)

	io.Copy(osFile,thumbnailFile)



	err = cfg.db.UpdateVideo(videoDetails)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating thumbnail", err)
		return
	}


	respondWithJSON(w, http.StatusOK, videoDetails)
}
