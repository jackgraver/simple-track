package apierr

import (
	"be-simpletracker/internal/database"

	"github.com/gin-gonic/gin"
)

// NotFoundMessage is the body of every "not found or not yours" answer. One message for both, so
// a response never tells a user whether someone else's row exists.
const NotFoundMessage = "not found"

// NotFoundOrInternal answers a failed lookup, update or delete: 404 {"error":"not found"} when
// the row is missing or belongs to someone else (database.IsNotFound), otherwise a logged 500
// with a fixed message.
func NotFoundOrInternal(c *gin.Context, err error) {
	if database.IsNotFound(err) {
		NotFound(c, NotFoundMessage)
		return
	}
	Internal(c, err)
}
