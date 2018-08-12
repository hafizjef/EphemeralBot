package utils

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CatchErr catch error and print to log
func CatchErr(e error) {
	if e != nil {
		log.Println(e)
	}
}

// ServeError check for server related error and return message
func ServeError(e error, c *gin.Context) {
	if e != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": e.Error(),
		})
	}
}
