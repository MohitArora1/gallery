package controller

import (
	"context"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/MohitArora1/gallery/models"
	"github.com/MohitArora1/gallery/utils"
	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/square/go-jose.v2/json"
)

// swagger:operation POST /albums Albums createAlbum
//
// Create Album
//
// Create a new album
//
// ---
// produces:
// - application/json
// consumes:
// - application/json
// parameters:
// - name: Album
//   in: body
//   description: Album
//   required: true
//   schema:
//     "$ref": "#/definitions/Album"
// responses:
//  '201':
//    description: Success, record created
//    schema:
//      "$ref": "#/definitions/Album"

func albumPostHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("Unexpected error %v\n", string(debug.Stack()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}()

	client := utils.GetMongoClient()
	defer client.Disconnect(context.Background())
	database := client.Database(utils.Config.DatabaseName)
	album, err := saveAlbum(database, r)
	if err != nil {
		if err == ErrBadRequest {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if utils.IsDuplicate(err) {
			http.Error(w, "Duplicate", http.StatusConflict)
			return
		}
		log.Printf("unexpected error %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	utils.WriteJSON(w, album)
}

func saveAlbum(database *mongo.Database, r *http.Request) (album models.Album, err error) {
	err = json.NewDecoder(r.Body).Decode(&album)
	if err != nil {
		log.Printf("error while decoding album data :%v\n", err)
		err = ErrBadRequest
		return
	}
	insertResult, err := database.Collection(utils.Album).InsertOne(context.Background(), &album)
	if err != nil {
		return
	}
	album.ID = insertResult.InsertedID.(primitive.ObjectID)
	return
}

// swagger:operation DELETE /albums/{albumID} Albums deleteAlbum
//
// Delete Album
//
// Delete a existing album
//
// ---
// produces:
// - application/json
// consumes:
// - application/json
// parameters:
// - name: albumID
//   in: path
//   description: albumID is needed to delete album
//   required: true
//   type: string
// responses:
//  '200':
//    description: Success

func albumDeleteHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("Unexpected error %v\n", string(debug.Stack()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}()

	client := utils.GetMongoClient()
	defer client.Disconnect(context.Background())
	database := client.Database(utils.Config.DatabaseName)
	err := deleteAlbumData(database, r)
	if err != nil {
		log.Printf("unexpected error :%v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func deleteAlbumData(database *mongo.Database, r *http.Request) (err error) {
	vars := mux.Vars(r)
	id := vars["albumID"]
	albumID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		err = ErrBadRequest
		return
	}
	cursor, err := database.Collection(utils.Image).Find(context.Background(), primitive.M{"album_id": id})
	if err != nil {
		log.Printf("error in find album data")
		return
	}
	for cursor.Next(context.Background()) {
		var image models.Image
		err = cursor.Decode(&image)
		if err != nil {
			continue
		}
		log.Printf("deleting image with id :%v", image.ID.Hex())
		_, err = database.Collection(utils.Image).DeleteOne(context.Background(), primitive.M{"_id": image.ID})
		if err != nil {
			log.Printf("error in remove image :%v", err)
			return err
		}
		ext := strings.Split(image.URL, ".")[1]
		_ = os.Remove(utils.Config.Storage + "/" + image.ID.Hex() + "." + ext)

	}
	_, err = database.Collection(utils.Album).DeleteOne(context.Background(), primitive.M{"_id": albumID})
	return
}

// swagger:operation GET /albums Albums getAlbums
//
// get Albums List
//
// Get Album List
//
// ---
// produces:
// - application/json
// consumes:
// - application/json
// parameters:
// - name: limit
//   in: query
//   description: to limit the result default is 10 and max could be 100
//   required: false
//   type: string
// - name: offset
//   in: query
//   description: to skip items
//   required: false
//   type: string
// responses:
//  '200':
//    description: Success
//    schema:
//      "$ref": "#/definitions/AlbumResponse"

func albumGetHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("Unexpected error %v\n", string(debug.Stack()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}()

	client := utils.GetMongoClient()
	defer client.Disconnect(context.Background())
	pagination, err := utils.GetPaginationParams(r)
	if err != nil {
		log.Printf("pagination error :%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	database := client.Database(utils.Config.DatabaseName)
	albums, err := getAlbums(database, r, pagination)
	if err != nil {
		log.Printf("unexpected error :%v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	pagination.Count = int64(len(albums))
	albumResponse := models.AlbumResponse{
		Pagination: *pagination,
		Albums:     albums,
	}
	utils.WriteJSON(w, albumResponse)
}

func getAlbums(database *mongo.Database, r *http.Request, pagination *models.Pagination) (albums []models.Album, err error) {
	opts := options.Find().SetLimit(pagination.Limit).SetSkip(pagination.Offset)
	pagination.Total, _ = database.Collection(utils.Album).CountDocuments(context.Background(), primitive.M{})
	cursor, err := database.Collection(utils.Album).Find(context.Background(), primitive.M{}, opts)
	if err != nil {
		log.Printf("error in get album :%v", err)
		return
	}
	defer cursor.Close(context.Background())
	err = cursor.All(context.Background(), &albums)
	return
}
