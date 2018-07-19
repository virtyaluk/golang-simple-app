package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Message string `json:"message"`
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

type Joke struct {
	ID    int    `json:"id" binding:"required"`
	Likes int    `json:"likes"`
	Joke  string `json:"joke" binding:"required"`
}

// taken from here http://jokes.cc.com/funny-dark-humor
var jokes = []Joke{
	Joke{1, 0, "What's a Jewish mobster? 'I'm going to break the legs of your therapist.'"},
	Joke{2, 0, "I'll tell you how bad it was in Los Angeles during the riots -- people were actually flooding into Tijuana. That's how bad it was."},
	Joke{3, 0, "Am I dead, Angel? Because this must be Heaven. Of course, I am a machine and therefore do not experience death."},
	Joke{4, 0, "The number two cause of death among teenagers in America today are guns. You know the number one cause of death? Not having a gun."},
	Joke{5, 0, "I just watched a program about beavers. It was the best dam program I've ever seen."},
	Joke{6, 0, "In honor of the recent death of coach Joe Paterno, I think we should all take 12 years of silence."},
	Joke{7, 0, "Bill Gates is so rich he hired cancer to kill Steve Jobs."},
	Joke{8, 0, "My aunt died, God bless her, at a ripe old age of 104. We called her Aunt Tique."},
	Joke{9, 0, "My grandmother, 86 years old, just entered medical school. She's a cadaver, and she is living death to the fullest."},
	Joke{10, 0, "If at first you don't suceed, sky diving isn't for you."},
	Joke{11, 0, "I got a cousin who's a psychotic with low self-esteem. He only wants to assassinate the vice president."},
	Joke{12, 0, "If a schizophrenic threatens to commit suicide, is it considered a hostage situation?"},
}

var jwtMiddleWare *jwtmiddleware.JWTMiddleware

func main() {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			aud := os.Getenv("AUTH0_API_AUDIENCE")
			checkAudience := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)

			if !checkAudience {
				return token, errors.New("Invalid audience.")
			}
			// verify iss claim
			iss := os.Getenv("AUTH0_DOMAIN")
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)

			if !checkIss {
				return token, errors.New("Invalid issuer.")
			}

			cert, err := getPemCert(token)

			if err != nil {
				log.Fatalf("could not get cert: %+v", err)
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))

			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})
	jwtMiddleWare = jwtMiddleware
	// Set the router as the default one shipped with Gin
	router := gin.Default()

	// Serve the frontend
	router.Use(static.Serve("/", static.LocalFile("./views", true)))

	api := router.Group("/api")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
		api.GET("/jokes", authMiddleware(), JokeHandler)
		api.POST("/jokes/like/:jokeID", authMiddleware(), LikeJoke)
	}

	// Start the app
	router.Run(":3000")
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(os.Getenv("AUTH0_DOMAIN") + ".well-known/jwks.json")

	if err != nil {
		return cert, err
	}

	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	x5c := jwks.Keys[0].X5c

	for k, v := range x5c {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + v + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		return cert, errors.New("unable to find appropriate key")
	}

	return cert, nil
}

// authMiddleware intercepts the requests, and check for a valid jwt token
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the client secret key
		err := jwtMiddleWare.CheckJWT(c.Writer, c.Request)

		if err != nil {
			// Token not found
			fmt.Println(err)
			c.Abort()
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Writer.Write([]byte("Unauthorized"))

			return
		}
	}
}

// JokeHandler returns a list of jokes available (in memory)
func JokeHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	c.JSON(http.StatusOK, jokes)
}

func LikeJoke(c *gin.Context) {
	// Check joke ID is valid
	if jokeid, err := strconv.Atoi(c.Param("jokeID")); err == nil {
		// find joke and increment likes
		for i := 0; i < len(jokes); i++ {
			if jokes[i].ID == jokeid {
				jokes[i].Likes = jokes[i].Likes + 1
			}
		}

		c.JSON(http.StatusOK, &jokes)
	} else {
		// the jokes ID is invalid
		c.AbortWithStatus(http.StatusNotFound)
	}
}
