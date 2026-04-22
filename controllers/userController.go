package controllers

import (
	"context"
	"fmt"
	helper "golang-jwt-project/helpers"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jdeepak1218/golang-jwt-project/database"
	"github.com/jdeepak1218/golang-jwt-project/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/bson"
	"go.mongodb.org/mongo-driver/mongo/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(pass string) string {
	hash,err := bcrypt.GenerateFromPassword([]byte(pass),bcrypt.DefaultCost())
	if err != nil{
		log.Panic(err)
	}
	return string(hash)
}
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword),[]byte(providedPassword))
	check := true
	msg := ""
	if err != nil{
		msg = fmt.Sprintf("email or password is incorrect")
		check = false
	}
	return check,msg
}
func Signup() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100 *time.Second)
		var user models.User
		if err := c.BindJSON(&user); err != nil{
			c.JSON(http.StatusBadRequest,gin.H{
				"error" : err.Error(),
			})
			return
		}
		validationErr := validate.Struct(user)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest,gin.H{
				"error" : validationErr.Error(),
			})
			return
		}
		count,err := userCollection.CountDocuments(ctx,bson.M{"email" : user.Email})
		defer cancel()
		if err != nil{
			log.Panic(err)
			c.JSON(http.StatusInternalServerError,gin.H{
				"error" : "Error occured while checking email"
			})
		}
		count,err := userCollection.CountDocuments(ctx,bson.M{"phone" : user.Phone})
		defer cancel()
		if err != nil{
			log.Panic(err)
			c.JSON(http.StatusInternalServerError,gin.H{
				"error" := "Error occured while checking phone number"
			})
		}
		password := HashPassword(*user.Password)
		user.Password = &password
		if count > 0{
			c.JSON(http.StatusInternalServerError,gin.H{
				"error" : "This email or phone number alread exist"
			})
		}
		user.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		user.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token,refreshToken,_ :=helper.GenerateAllTokens(*user.Email,*user.First_name,*user.Last_name,*user.User_type,*&user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken
		resultInsertionNumber,insertErr := userCollection.InsertOne(ctx,user)
		if insertErr != nil{
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError,gin.H{
				"error" : msg,
			})
			return
		}
		c.JSON(http.StatusOK,resultInsertionNumber)
	}
}
func Login() gin.HandlerFunc{
	return func(c *gin.Context){
		ctx,delete := context.WithTimeout(context.Background(),100 * time.Second)
		var user models.User
		var foundUser models.User
		if err := c.BindJSON(&user); err != nil{
			c.JSON(http.StatusBadRequest,gin.H{
				"error" : err.Error()
			})
			return
		}
		err := userCollection.FindOne(ctx,bson.M{"email":user.Email}).Decode(&foundUser)
		if err != nil{
			c.JSON(http.StatusUnauthorized,gin.H{
				"error" : "email or password is incorrect"
			})
			return
		}
		verifyPass,msg := VerifyPassword(*foundUser.Password,*user.Password)
		if verifyPass != true{
			c.JSON(http.StatusUnauthorized,gin.H{
				"message" : msg
			})
			return
		}
		token,refreshtoken,err := helper.GenerateAllTokens(
			*foundUser.Email,
			*foundUser.First_name,
			*foundUser.Last_name,
			*foundUser.User_type,
			*foundUser.User_id
		)
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{
				"error" : err.Error()
			})
			return
		}
		helper.UpdateAllTokens(token,refreshtoken,*foundUser.User_id)
		c.JSON(http.StatusOK,foundUser)
	}
}
func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context){
		err := helper.CheckUserType(c, "ADMIN")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		ctx,cancel := context.WithTimeout(context.Background(),100 * time.Second)
		defer cancel()
		recordPerPage,err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1{
			recordPerPage = 10
		}
		page,err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}
		startIdx := (page - 1) * recordPerPage
		matchStage := bson.D{{Key:"$match",Value:bson.D{}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{
            {Key: "_id",         Value: nil},
            {Key: "total_count", Value: bson.D{{Key: "$sum",  Value: 1}}},
            {Key: "data",        Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
        }}}
		projectStage := bson.D{{Key: "$project", Value: bson.D{
            {Key: "_id",         Value: 0},
            {Key: "total_count", Value: 1},
            {Key: "user_items",  Value: bson.D{{Key: "$slice",
                Value: []interface{}{"$data", startIndex, recordPerPage}}}},
        }}}
		result,err := userCollection.Aggregate(ctx,mongo.Pipeline{
			matchStage,
			groupStage,
			projectStage
		})
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{
				"error" : "error while loading users"
			})
			return
		}
		var allUsers []bson.M
		if err := result.All(ctx, &allUsers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, allUsers[0])
	}
}
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		var user models.User
		err := userCollection.FindOne(ctx,bson.M{"user_id" : userId}).Decode(&user)
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{
				"error" : err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK,user)
	}
}
