package database

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/pkg/env"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	// DefaultMessageCollection adalah nama collection default untuk pesan chat.
	DefaultMessageCollection = "messages"
	// defaultMessageDatabase adalah nama database default jika env tidak diisi.
	defaultMessageDatabase = "simple_messaging_app"
	// mongoConnectTimeout membatasi waktu koneksi & ping saat startup.
	mongoConnectTimeout = 10 * time.Second
	// namespaceExistsCode adalah kode error MongoDB saat collection sudah ada.
	namespaceExistsCode = 48
)

var (
	// MongoClient menyimpan handle client MongoDB.
	MongoClient *mongo.Client
	// Mongo menyimpan handle database MongoDB yang dipakai aplikasi.
	Mongo *mongo.Database
)

// SetupMongo membuka koneksi ke MongoDB, memverifikasinya dengan ping, lalu
// menyiapkan collection pesan beserta index-nya. Collection dibuat eksplisit
// agar langsung terlihat di database meskipun belum ada pesan yang disimpan.
func SetupMongo() {
	uri := env.GetEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := env.GetEnv("MONGODB_DB", defaultMessageDatabase)

	ctx, cancel := context.WithTimeout(context.Background(), mongoConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB: ", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		log.Fatal("Failed to ping MongoDB: ", err)
	}

	MongoClient = client
	Mongo = client.Database(dbName)

	if err := ensureMessageCollection(ctx); err != nil {
		log.Fatal("Failed to prepare messages collection: ", err)
	}

	log.Printf("Successfully connected to MongoDB database %q!", dbName)
}

// ensureMessageCollection membuat collection pesan (idempoten) dan index
// room_id + created_at yang dipakai saat mengambil riwayat chat.
func ensureMessageCollection(ctx context.Context) error {
	collectionName := MessageCollectionName()
	collection := Mongo.Collection(collectionName)

	if err := Mongo.CreateCollection(ctx, collectionName); err != nil {
		var commandErr mongo.CommandError
		if !errors.As(err, &commandErr) || !commandErr.HasErrorCode(namespaceExistsCode) {
			return err
		}
	}

	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "room_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
		Options: options.Index().SetName("room_id_created_at"),
	})

	return err
}

// MessageCollectionName mengembalikan nama collection pesan dari env.
func MessageCollectionName() string {
	return env.GetEnv("MONGODB_COLLECTION", DefaultMessageCollection)
}

// MessageCollection mengembalikan handle collection pesan.
func MessageCollection() *mongo.Collection {
	return Mongo.Collection(MessageCollectionName())
}

// DisconnectMongo menutup koneksi MongoDB. Dipanggil saat aplikasi berhenti.
func DisconnectMongo(ctx context.Context) error {
	if MongoClient == nil {
		return nil
	}

	return MongoClient.Disconnect(ctx)
}
