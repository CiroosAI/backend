package admins

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"project/database"
	"project/models"
	"project/utils"

	"gorm.io/gorm"
)

// Response models
type serverStatus struct {
	Status   bool `json:"status"`
	Database bool `json:"database"`
	Security bool `json:"security"`
}

type applicationsStatus struct {
	PendingWithdrawals int64 `json:"pending_withdrawals"`
	PendingForums      int64 `json:"pending_forums"`
}

type notificationItem struct {
	Notificated bool   `json:"notifycated"`
	Message     string `json:"message"`
	Time        string `json:"time"`
}

type notificationsPayload struct {
	PendingWithdrawals *[]notificationItem `json:"pending_withdrawals"`
	PendingForums      *[]notificationItem `json:"pending_forums"`
	NewUsers           *[]notificationItem `json:"new_users"`
}

type adminInfoResponse struct {
	Servers       serverStatus         `json:"servers"`
	Applications  applicationsStatus   `json:"applications"`
	Notifications notificationsPayload `json:"notifications"`
}

// GET /admins/info
func GetAdminInfo(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	// Servers health
	serverOK := true   // If this handler runs, server is up
	dbOK := pingDB(db) // Check DB connectivity with timeout
	securityOK := true // As requested, default to true

	// Applications: counts
	var pendingWithdrawals int64
	db.Model(&models.Withdrawal{}).Where("status = ?", "Pending").Count(&pendingWithdrawals)

	var pendingForums int64
	db.Model(&models.Forum{}).Where("status = ?", "Pending").Count(&pendingForums)

	// New users today
	now := time.Now()
	loc := now.Location()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	end := start.Add(24 * time.Hour)

	var newUsersToday int64
	db.Model(&models.User{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&newUsersToday)

	// Notifications: null when count is 0; otherwise provide a single item
	var notifs notificationsPayload

	if pendingWithdrawals > 0 {
		msg := fmt.Sprintf("%d penarikan menunggu persetujuan", pendingWithdrawals)
		items := []notificationItem{
			{Notificated: true, Message: msg, Time: time.Now().Format(time.RFC3339)},
		}
		notifs.PendingWithdrawals = &items
	} else {
		notifs.PendingWithdrawals = nil
	}

	if pendingForums > 0 {
		msg := fmt.Sprintf("%d postingan menunggu persetujuan", pendingForums)
		items := []notificationItem{
			{Notificated: true, Message: msg, Time: time.Now().Format(time.RFC3339)},
		}
		notifs.PendingForums = &items
	} else {
		notifs.PendingForums = nil
	}

	if newUsersToday > 0 {
		msg := fmt.Sprintf("%d pengguna baru terdaftar hari ini", newUsersToday)
		items := []notificationItem{
			// Keeping it simple per request: always false
			{Notificated: false, Message: msg, Time: time.Now().Format(time.RFC3339)},
		}
		notifs.NewUsers = &items
	} else {
		notifs.NewUsers = nil
	}

	resp := adminInfoResponse{
		Servers: serverStatus{
			Status:   serverOK,
			Database: dbOK,
			Security: securityOK,
		},
		Applications: applicationsStatus{
			PendingWithdrawals: pendingWithdrawals,
			PendingForums:      pendingForums,
		},
		Notifications: notifs,
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    resp,
	})
}

func pingDB(gdb *gorm.DB) bool {
	if gdb == nil {
		return false
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return false
	}
	return true
}
