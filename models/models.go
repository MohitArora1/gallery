package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Album is to store album metadata
// swagger:model Album
type Album struct {
	ID   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name"`
}

// Image is to store image metadata
type Image struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AlbumID string             `json:"album_id" bson:"album_id"`
	URL     string             `json:"url" bson:"url"`
}

//Pagination ...
type Pagination struct {
	Count  int64 `json:"count"`
	Total  int64 `json:"total"`
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
}

//ImageResponse to return images response
// swagger:model ImageResponse
type ImageResponse struct {
	Pagination Pagination `json:"pagination"`
	Images     []Image    `json:"images"`
}

//AlbumResponse to return images response
// swagger:model AlbumResponse
type AlbumResponse struct {
	Pagination Pagination `json:"pagination"`
	Albums     []Album    `json:"albums"`
}
