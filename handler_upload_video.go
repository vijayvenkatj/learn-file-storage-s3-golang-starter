package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/helpers"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	
	uploadLimit := int64(1 << 30)

	r.Body = http.MaxBytesReader(w, r.Body, uploadLimit)

	pathVal := r.PathValue("videoID")

	videoID := uuid.MustParse(pathVal)

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

	videoDetails, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find Video", err)
		return
	}
	if videoDetails.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorised", err)
		return
	}


	videoFile, videoFileHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Issue getting video", err)
		return
	}
	defer videoFile.Close()

	mimeType := videoFileHeader.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(mimeType)
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Invalid file format",err)
		return
	}

	uploadFile, err := os.CreateTemp(cfg.assetsRoot, "tubeby-upload-*.mp4")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to store file",err)
		return
	}
	defer os.Remove(uploadFile.Name())
	defer uploadFile.Close()
	io.Copy(uploadFile,videoFile)

	uploadFile.Seek(0,io.SeekStart)

	processedFilePath, err := helpers.ProcessVideoForFastStart(uploadFile.Name())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to process file",err)
		fmt.Print(err)
		return
	}

	aspectRatio, err := helpers.GetVideoAspectRatio(processedFilePath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get aspectRatio",err)
		return
	}

	key := filepath.Base(processedFilePath) 

	if aspectRatio == "16:9" {
		key = "landscape/" + key
	} else if aspectRatio == "9:16" {
		key = "portrait/" + key
	} else {
		key = "other/" + key
	}

	processedFile, err := os.Open(processedFilePath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error opening processed file",err)
		return
	}
	defer processedFile.Close()

	_, err = cfg.s3Client.PutObject(r.Context(),&s3.PutObjectInput{
		Bucket: &cfg.s3Bucket,
		Key: &key,
		Body: processedFile,
		ContentType: &mimeType,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to store file in s3",err)
		return
	}

	fileUrl := fmt.Sprintf("%s,%s",cfg.s3Bucket,key)

	videoDetails.VideoURL = &fileUrl

	err = cfg.db.UpdateVideo(videoDetails)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update url in db",err)
		return
	}

	video, err := cfg.dbVideoToSignedVideo(videoDetails)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get video", err)
		return
	}

	respondWithJSON(w,200,video)
}
