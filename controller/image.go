package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/MohitArora1/gallery/models"
	"github.com/MohitArora1/gallery/utils"
	"github.com/gorilla/mux"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// swagger:operation POST /albums/{albumID}/images Images postImage
//
// Post Image
//
// Post a new Image
//
// ---
// produces:
// - application/json
// consumes:
// - multipart/form-data
// parameters:
// - name: albumID
//   in: path
//   description: Image to upload
//   required: true
//   type: string
// - name: image
//   in: formData
//   description: Image to upload
//   required: true
//   type: file
// responses:
//  '201':
//    description: Success, record created
//    schema:
//      "$ref": "#/definitions/Album"

func imagePostHandler(w http.ResponseWriter, r *http.Request) {
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

	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("image")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	var ext string
	switch handler.Header.Get("Content-Type") {
	case "image/png":
		ext = ".png"
	case "image/jpeg":
		ext = ".jpeg"
	default:
		http.Error(w, "media type not supported", http.StatusBadRequest)
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	id := primitive.NewObjectID()
	err = ioutil.WriteFile(utils.Config.Storage+"/"+id.Hex()+ext, fileBytes, 0644)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	url := "http://" + r.Host + "/data/" + id.Hex() + ext
	fmt.Printf(url)
	albumID := vars["albumID"]
	image := models.Image{
		ID:      id,
		AlbumID: albumID,
		URL:     url,
	}
	insertRecord, err := database.Collection(utils.Image).InsertOne(context.Background(), &image)
	if err != nil {
		if utils.IsDuplicate(err) {
			log.Printf("duplicate error :%v", err)
			http.Error(w, "Duplicate", http.StatusConflict)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	image.ID = insertRecord.InsertedID.(primitive.ObjectID)
	msg, _ := json.Marshal(image)
	produceMessage(string(msg))
	utils.WriteJSON(w, image)

}

func produceMessage(msg string) {
	topic := "myTopic"

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": utils.Config.KafkaURL})
	if err != nil {
		log.Printf("Not able to connect with kafka")
		return
	}

	defer p.Close()

	// Produce messages to topic (asynchronously)
	for _, word := range []string{msg} {
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}, nil)
	}

}

// swagger:operation GET /albums/{albumID}/images Images getAlbumsImages
//
// get Albums Image List
//
// Get Album Image List
//
// ---
// produces:
// - application/json
// consumes:
// - application/json
// parameters:
// - name: albumID
//   in: path
//   description: albumID to list images from album
//   required: true
//   type: string
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
//      "$ref": "#/definitions/ImageResponse"

func imageGetHandler(w http.ResponseWriter, r *http.Request) {
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
	images, err := getAlbumImages(database, r, pagination)
	if err != nil {
		log.Printf("unexpected error :%v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	pagination.Count = int64(len(images))
	imagesResponse := models.ImageResponse{
		Pagination: *pagination,
		Images:     images,
	}
	utils.WriteJSON(w, imagesResponse)
}

func getAlbumImages(database *mongo.Database, r *http.Request, pagination *models.Pagination) (images []models.Image, err error) {
	vars := mux.Vars(r)
	albumID := vars["albumID"]
	opts := options.Find().SetLimit(pagination.Limit).SetSkip(pagination.Offset)
	pagination.Total, _ = database.Collection(utils.Image).CountDocuments(context.Background(), primitive.M{"album_id": albumID})
	cursor, err := database.Collection(utils.Image).Find(context.Background(), primitive.M{"album_id": albumID}, opts)
	if err != nil {
		log.Printf("error in get album images :%v", err)
		return
	}
	defer cursor.Close(context.Background())
	err = cursor.All(context.Background(), &images)
	return
}

// swagger:operation DELETE /albums/{albumID}/images/{imageID} Images deleteImage
//
// Delete Image
//
// Delete a existing image
//
// ---
// produces:
// - application/json
// consumes:
// - application/json
// parameters:
// - name: albumID
//   in: path
//   description: albumID is needed to delete image
//   required: true
//   type: string
// - name: imageID
//   in: path
//   description: imageID is needed to delete image
//   required: true
//   type: string
// responses:
//  '200':
//    description: Success

func imageDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	err := deleteImage(database, r)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		log.Printf("unexpected error :%v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func deleteImage(database *mongo.Database, r *http.Request) (err error) {
	vars := mux.Vars(r)
	id := vars["imageID"]
	albumID := vars["albumID"]
	imageID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("wrong id :%v", err)
		err = ErrBadRequest
		return
	}
	var image models.Image
	err = database.Collection(utils.Image).FindOne(context.Background(), primitive.M{"_id": imageID, "album_id": albumID}).Decode(&image)
	if err != nil {
		return
	}
	ext := strings.Split(image.URL, ".")[1]
	err = os.Remove(utils.Config.Storage + "/" + id + "." + ext)
	_, err = database.Collection(utils.Image).DeleteOne(context.Background(), primitive.M{"_id": imageID, "album_id": albumID})
	if err != nil {
		log.Printf("error in delete image")
		return
	}
	return
}
