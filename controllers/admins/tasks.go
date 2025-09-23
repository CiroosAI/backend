package admins

import (
	"encoding/json"
	"net/http"
	"project/database"
	"project/models"
	"project/utils"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// GET /api/admin/tasks
func TaskListHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	// 1) Ambil semua tasks
	var tasks []models.Task
	if err := db.Find(&tasks).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// 2) Hitung total_claimed global (jumlah seluruh user_tasks)
	var totalClaimed int64
	if err := db.Model(&models.UserTask{}).Count(&totalClaimed).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// 3) Hitung jumlah klaim per task (GROUP BY task_id)
	type taskCount struct {
		TaskID uint
		Cnt    int64
	}
	var counts []taskCount
	var taskIDs []uint
	for _, t := range tasks {
		taskIDs = append(taskIDs, t.ID)
	}
	countMap := make(map[uint]int64, len(taskIDs))
	if len(taskIDs) > 0 {
		if err := db.
			Table("user_tasks").
			Select("task_id, COUNT(*) as cnt").
			Where("task_id IN ?", taskIDs).
			Group("task_id").
			Scan(&counts).Error; err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
				Success: false,
				Message: "Terjadi kesalahan sistem, silakan coba lagi",
			})
			return
		}
		for _, c := range counts {
			countMap[c.TaskID] = c.Cnt
		}
	}

	// 4) Hitung total_paid = sum(reward * jumlah klaim task)
	var totalPaid float64
	for _, t := range tasks {
		if c, ok := countMap[t.ID]; ok && c > 0 {
			totalPaid += t.Reward * float64(c)
		}
	}

	// 5) Bangun response list task + total_claimed per task
	type TaskWithStats struct {
		models.Task
		TotalClaimed int64 `json:"total_claimed"`
	}
	taskItems := make([]TaskWithStats, 0, len(tasks))
	for _, t := range tasks {
		tc := countMap[t.ID]
		taskItems = append(taskItems, TaskWithStats{
			Task:         t,
			TotalClaimed: tc,
		})
	}

	// Bungkus dalam objek data sesuai kebutuhan
	type TaskListData struct {
		TotalClaimed int64           `json:"total_claimed"`
		TotalPaid    float64         `json:"total_paid"`
		Tasks        []TaskWithStats `json:"tasks"`
	}
	data := TaskListData{
		TotalClaimed: totalClaimed,
		TotalPaid:    totalPaid,
		Tasks:        taskItems,
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    data,
	})
}

type TaskRequest struct {
	Name                  string  `json:"name"`
	Reward                float64 `json:"reward"`
	RequiredLevel         int     `json:"required_level"`
	RequiredActiveMembers int     `json:"required_active_members"`
	Status                string  `json:"status"`
}

// POST /api/admin/tasks
func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	task := models.Task{
		Name:                  req.Name,
		Reward:                req.Reward,
		RequiredLevel:         req.RequiredLevel,
		RequiredActiveMembers: int64(req.RequiredActiveMembers),
		Status:                req.Status,
	}

	db := database.DB
	if err := db.Create(&task).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Message: "Tugas berhasil di tambahkan",
		Data:    task,
	})
}

// PUT /api/admin/tasks/:id
func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	db := database.DB

	var task models.Task
	if err := db.First(&task, taskID).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
			Success: false,
			Message: "Tugas tidak ditemukan",
		})
		return
	}

	task.Name = req.Name
	task.Reward = req.Reward
	task.RequiredLevel = req.RequiredLevel
	task.RequiredActiveMembers = int64(req.RequiredActiveMembers)
	task.Status = req.Status

	if err := db.Save(&task).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Tugas berhasil di perbarui",
		Data:    task,
	})
}

// GET /api/admin/user-tasks
func UserTasksHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	// Pagination (optional)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Build base queries with joins
	query := db.
		Table("user_tasks AS ut").
		Joins("JOIN users u ON ut.user_id = u.id").
		Joins("JOIN tasks t ON ut.task_id = t.id")

	countQuery := db.
		Table("user_tasks AS ut").
		Joins("JOIN users u ON ut.user_id = u.id").
		Joins("JOIN tasks t ON ut.task_id = t.id")

	// Apply search on users.name or users.number
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("u.name LIKE ? OR u.number LIKE ?", like, like)
		countQuery = countQuery.Where("u.name LIKE ? OR u.number LIKE ?", like, like)
	}

	// Total data (untuk pagination)
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal menghitung total data",
		})
		return
	}

	// Struktur untuk hasil scan join
	type rowScan struct {
		ID        uint
		UserID    uint
		UserName  string
		Phone     string
		TaskID    uint
		TaskName  string
		Reward    float64
		ClaimedAt time.Time
	}

	var rows []rowScan
	if err := query.
		Select(`
			ut.id,
			ut.user_id,
			u.name AS user_name,
 			u.number AS phone,
			ut.task_id,
			t.name AS task_name,
			t.reward AS reward,
			ut.claimed_at
		`).
		Order("ut.claimed_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// Response DTO
	type UserTaskResponse struct {
		ID        uint    `json:"id"`
		UserID    uint    `json:"user_id"`
		UserName  string  `json:"user_name"`
		Phone     string  `json:"phone"`
		TaskID    uint    `json:"task_id"`
		TaskName  string  `json:"task_name"`
		Reward    float64 `json:"reward"`
		ClaimedAt string  `json:"claimed_at"`
	}

	items := make([]UserTaskResponse, 0, len(rows))
	for _, r := range rows {
		items = append(items, UserTaskResponse{
			ID:        r.ID,
			UserID:    r.UserID,
			UserName:  r.UserName,
			Phone:     r.Phone,
			TaskID:    r.TaskID,
			TaskName:  r.TaskName,
			Reward:    r.Reward,
			ClaimedAt: r.ClaimedAt.Format(time.RFC3339),
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    items,
	})
}
