package utils

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/MohitArora1/gallery/models"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type config struct {
	DatabaseName string `mapstructure:"database_name"`
	DatabaseURL  string `mapstructure:"database_url"`
	KafkaURL     string `mapstructure:"kafka_url"`
	Storage      string `mapstructure:"storage"`
}

var (
	// Config is used to get config varibales
	Config config

	// Album is collection name
	Album = "album"

	//Image is collection name
	Image = "image"
)

const (
	ConstDefaultLimit int64 = 10
	ConstOffset             = "offset"
	ConstMaxLimit     int64 = 100
	ConstLimit              = "limit"
)

// LoadConfig is to load config
func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetEnvPrefix("GALLERY")
	viper.AutomaticEnv()
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&Config)
	if err != nil {
		panic(err)
	}
}

// GetMongoClient to crate client
func GetMongoClient() *mongo.Client {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(Config.DatabaseURL))
	if err != nil {
		panic(err)
	}
	return client
}

//IsDuplicate is to identify duplicate mongodb error
func IsDuplicate(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}

// WriteJSON will write json in writer
func WriteJSON(w http.ResponseWriter, model interface{}) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	w.Header().Set("Content-Type", "application/json")
	enc.Encode(model)
}

//GetPaginationParams reads the request data and provide pagination details
func GetPaginationParams(r *http.Request) (pagination *models.Pagination, pagingError error) {
	pagination = &models.Pagination{}
	var err error
	pagination.Limit = ConstDefaultLimit

	queryLimit := r.URL.Query().Get(ConstLimit)
	if queryLimit != "" {
		pagination.Limit, err = strconv.ParseInt(queryLimit, 10, 64)
		if err != nil {
			return nil, errors.New("bad reqeuest")
		} else if pagination.Limit < 1 {
			return nil, errors.New("limit can't be less the 1")
		} else if pagination.Limit > ConstMaxLimit {
			return nil, errors.New("max limit could be 100")
		}
	} else {
		pagination.Limit = ConstDefaultLimit
	}

	queryOffset := r.URL.Query().Get(ConstOffset)
	if queryOffset != "" {
		pagination.Offset, err = strconv.ParseInt(queryOffset, 10, 64)
		if err != nil {
			return nil, errors.New("bad reqeuest")
		} else if pagination.Offset < 0 {
			return nil, errors.New("offset can't be negative")
		}
	} else {
		pagination.Offset = 0
	}

	return pagination, nil
}
