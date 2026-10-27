package controllers

import (
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"lab-reservation-system/database"
	"lab-reservation-system/models"
)

func TestReservationDetailAccess(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "details.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	previousDB := database.DB
	database.DB = db
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.DB = previousDB; sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Reservation{}); err != nil {
		t.Fatal(err)
	}
	group1, group2 := uint(1), uint(2)
	users := []models.User{
		{ID: 1, Username: "owner", Role: models.RoleStudent, GroupID: &group1},
		{ID: 2, Username: "other", Role: models.RoleStudent, GroupID: &group1},
		{ID: 3, Username: "advisor", Role: models.RoleAdvisor, GroupID: &group1},
		{ID: 4, Username: "other-advisor", Role: models.RoleAdvisor, GroupID: &group2},
		{ID: 5, Username: "admin", Role: models.RoleEquipmentAdmin},
		{ID: 6, Username: "sysadmin", Role: models.RoleSystemAdmin},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatal(err)
	}
	r := models.Reservation{ID: 1, UserID: 1, GroupID: 1, EquipmentID: 1, ExperimentName: "热处理", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), Status: models.ReservationPendingAdmin}
	if err := db.Create(&r).Error; err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		user   uint
		id     string
		status int
	}{
		{1, "1", 200}, {2, "1", 403}, {3, "1", 200}, {4, "1", 403}, {5, "1", 200}, {6, "1", 200}, {0, "1", 401}, {1, "999", 404}, {1, "invalid", 400}, {1, "0", 400},
	} {
		t.Run(fmt.Sprintf("user%d_id%s", test.user, test.id), func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set("user_id", test.user) })
			router.GET("/reservations/:id", (&ReservationController{}).Detail)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest("GET", "/reservations/"+test.id, nil))
			if response.Code != test.status {
				t.Fatalf("status %d, want %d: %s", response.Code, test.status, response.Body.String())
			}
		})
	}
}
