package handlers

import (
	"bank-app/config"
	"bank-app/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GetUserByID(c *gin.Context) {
	paramID := c.Param("id")

	userID := c.MustGet("userID").(uint)

	if fmt.Sprintf("%d", userID) != paramID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	var user models.User
	// Preload accounts to get associated accounts for the user
	if err := config.DB.Preload("Accounts").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}
