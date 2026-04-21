package helpers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt-project/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jdeepak1218/golang-jwt-project/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	User_type  string
	jwt.RegisteredClaims
}

var UserCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokesn(email, first_name, last_name, userType, uid string) (signedToken, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_name: first_name,
		Last_name:  last_name,
		Uid:        uid,
		User_type:  userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	refreshClaims := &SignedDetails{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethod256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Fatal(err)
		return "", "", err
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethod256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Fatal(err)
		return "", "", err
	}
	return token, refreshToken, nil
}
func ValidateToken(signedToken string) (*SignedDetails, string, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		return nil, "", err
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok || !token.Valid {
		return nil, "", fmt.Errorf("Token is invalid")
	}
	return claims, "", nil
}
func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var updateObj primitive.D
	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj), bson.E{Key: "refresh_token", Value: signedRefreshToken}
	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: updatedAt})
	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.Update().SetUpsert(upsert)
	_, err := UserCollection.UpdateOne(
		ctx,
		filter,
		bson.D{Key: "$set", Value: updateObj},
		opt,
	)
	if err != nil {
		log.Panic(err)
		return
	}
	fmt.Println("Tokens updated in db")
}
