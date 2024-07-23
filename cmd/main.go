package main

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"time"

	"be-story/repository/entity"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	collection *mongo.Collection
	ctx        context.Context

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func init() {
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("StoryGram").Collection("stories")
}

func main() {
	defer client.Disconnect(ctx)

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/new-story", handleNewStory)

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	go listenForNewStories()

	select {}

	// Example usage
	//userID := "user123"
	//newStory := entity.Story{
	//	UserID:    userID,
	//	Content:   "Hello, this is my first story!",
	//	Views:     0,
	//	ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	//}
	//
	//err := createStory(newStory)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//stories, err := getStories(userID)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//for _, story := range stories {
	//	fmt.Printf("Story ID: %s, Content: %s\n", story.ID.Hex(), story.Content)
	//}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Example: Broadcast new stories to all connected clients
	for {
		// Read message from client
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Received message: %s\n", msg)
	}
}

func listenForNewStories() {
	// Example: Continuously monitor for new stories and broadcast them
	pipeline := mongo.Pipeline{bson.D{{"$match", bson.D{{"operationType", "insert"}}}}}

	changeStream, err := collection.Watch(ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}
	defer changeStream.Close(ctx)

	for changeStream.Next(ctx) {
		var changeEvent struct {
			OperationType string       `bson:"operationType"`
			FullDocument  entity.Story `bson:"fullDocument"`
		}
		if err := changeStream.Decode(&changeEvent); err != nil {
			log.Println(err)
			continue
		}

		// Broadcast new story to all connected clients
		broadcastNewStory(changeEvent.FullDocument)
	}
}

func broadcastNewStory(story entity.Story) {
	// Implement broadcasting to all connected clients
	// Example:
	log.Printf("Broadcasting new story: %s\n", story.Content)
}

func handleNewStory(w http.ResponseWriter, r *http.Request) {
	// Handle creating a new story via HTTP request
	// Example:
	userID := "user123"
	newStory := entity.Story{
		UserID:    userID,
		Content:   "Hello, this is a new story!",
		Views:     0,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	_, err := collection.InsertOne(ctx, newStory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "New story created successfully")
}

func createStory(story entity.Story) error {
	_, err := collection.InsertOne(ctx, story)
	if err != nil {
		return err
	}
	return nil
}

func getStories(userID string) ([]entity.Story, error) {
	var stories []entity.Story
	cursor, err := collection.Find(ctx, bson.M{"userID": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var story entity.Story
		if err := cursor.Decode(&story); err != nil {
			return nil, err
		}
		stories = append(stories, story)
	}
	return stories, nil
}
