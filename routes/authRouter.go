package routes
import(
	"github.com/gin-gonic/gin"
	"golang-jwt-project/controllers""
)
func AuthRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.POST("users/signup",controller.Signup())
	incomingRoutes.POST("users/login",controller.Login())
}
