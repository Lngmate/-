package services

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"lab-reservation-system/models"
)

func taskFixture(t *testing.T) (*gorm.DB, TaskRequest, time.Time) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "scheduler.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Equipment{}, &models.ExperimentTask{}, &models.Reservation{}, &models.Notification{}); err != nil {
		t.Fatal(err)
	}
	groupID := uint(1)
	user := models.User{Username: "student", GroupID: &groupID, CreditScore: 98}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	for i, policy := range []models.ApprovalPolicy{models.ApprovalNone, models.ApprovalDual, models.ApprovalSingle} {
		eq := models.Equipment{ID: uint(i + 1), Name: "设备", Location: "实验室", HourlyRate: 60, ApprovalType: policy, Status: models.StatusAvailable}
		if err := db.Create(&eq).Error; err != nil {
			t.Fatal(err)
		}
	}
	start := time.Now().Add(24 * time.Hour).Truncate(time.Minute)
	return db, TaskRequest{Name: "材料制备", EarliestStart: start, Deadline: start.Add(12 * time.Hour), Steps: []TaskStepInput{
		{EquipmentID: 1, Name: "加热", DurationMinutes: 120},
		{EquipmentID: 2, Name: "表征", DurationMinutes: 30, GapMinutes: 60},
		{EquipmentID: 3, Name: "分析", DurationMinutes: 30},
	}}, start
}

func previewRequest(t *testing.T, db *gorm.DB, req TaskRequest) TaskRequest {
	t.Helper()
	plan, err := PlanTask(db, 1, req, false, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range plan.Steps {
		req.StartTimes = append(req.StartTimes, step.StartTime)
	}
	return req
}

func TestPlanTaskAvoidsBusyIntervalsAndHonorsGaps(t *testing.T) {
	db, req, start := taskFixture(t)
	busy := []models.Reservation{
		{EquipmentID: 1, UserID: 1, GroupID: 1, ExperimentName: "已有预约", StartTime: start.Add(time.Hour), EndTime: start.Add(2 * time.Hour), Status: models.ReservationPendingAdmin},
		{EquipmentID: 1, UserID: 1, GroupID: 1, ExperimentName: "后续预约", StartTime: start.Add(2 * time.Hour), EndTime: start.Add(3 * time.Hour), Status: models.ReservationActive},
		{EquipmentID: 2, UserID: 1, GroupID: 1, ExperimentName: "已取消", StartTime: start, EndTime: start.Add(12 * time.Hour), Status: models.ReservationCancelled},
	}
	if err := db.Create(&busy).Error; err != nil {
		t.Fatal(err)
	}
	plan, err := PlanTask(db, 1, req, false, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !plan.StartTime.Equal(start.Add(3*time.Hour)) || !plan.Steps[1].StartTime.Equal(start.Add(6*time.Hour)) {
		t.Fatalf("unexpected schedule: %+v", plan)
	}
	if plan.TotalCost != 180 {
		t.Fatalf("cost = %v", plan.TotalCost)
	}
	// A slot ending exactly at an existing booking's start remains usable.
	req.Steps[0].DurationMinutes = 60
	plan, err = PlanTask(db, 1, req, false, time.Now())
	if err != nil || !plan.StartTime.Equal(start) {
		t.Fatalf("boundary rejected: %+v, %v", plan, err)
	}
}

func TestSubmitTaskKeepsApprovalPoliciesAndLinks(t *testing.T) {
	db, req, _ := taskFixture(t)
	task, err := SubmitTask(db, 1, previewRequest(t, db, req))
	if err != nil {
		t.Fatal(err)
	}
	want := []models.ReservationStatus{models.ReservationApproved, models.ReservationPendingAdvisor, models.ReservationPendingAdmin}
	for i, r := range task.Reservations {
		if r.TaskID == nil || *r.TaskID != task.ID || r.TaskStep != i+1 || r.Status != want[i] || len(r.CheckinCode) != 6 {
			t.Fatalf("invalid linked reservation: %+v", r)
		}
	}
	var stored models.ExperimentTask
	if err := db.Preload("Reservations").First(&stored, task.ID).Error; err != nil || len(stored.Reservations) != 3 {
		t.Fatalf("reload: %+v %v", stored, err)
	}
}

func TestSubmitTaskStalePreviewCreatesNothing(t *testing.T) {
	db, req, _ := taskFixture(t)
	req = previewRequest(t, db, req)
	r := models.Reservation{EquipmentID: 2, UserID: 1, GroupID: 1, ExperimentName: "被抢约", StartTime: req.StartTimes[1], EndTime: req.StartTimes[1].Add(time.Hour), Status: models.ReservationApproved}
	if err := db.Create(&r).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := SubmitTask(db, 1, req); !errors.Is(err, ErrScheduleChanged) {
		t.Fatalf("want stale preview, got %v", err)
	}
	assertCount(t, db, &models.ExperimentTask{}, 0)
	assertCount(t, db, &models.Reservation{}, 1)
}

func TestSubmitTaskRollsBackOnLateWriteFailure(t *testing.T) {
	db, req, _ := taskFixture(t)
	req = previewRequest(t, db, req)
	// Fail the notification write after the task and all reservations were inserted.
	if err := db.Callback().Create().Before("gorm:create").Register("fail_notification", func(tx *gorm.DB) {
		if tx.Statement.Table == "notifications" {
			tx.AddError(errors.New("injected storage failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := SubmitTask(db, 1, req); err == nil {
		t.Fatal("expected failure")
	}
	assertCount(t, db, &models.ExperimentTask{}, 0)
	assertCount(t, db, &models.Reservation{}, 0)
}

func TestConcurrentTaskSubmissionsOnlyOneSucceeds(t *testing.T) {
	db, req, _ := taskFixture(t)
	req = previewRequest(t, db, req)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := SubmitTask(db, 1, req); results <- err }()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrScheduleChanged) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("successes = %d", successes)
	}
	assertCount(t, db, &models.ExperimentTask{}, 1)
	assertCount(t, db, &models.Reservation{}, 3)
}

func TestTaskConstraints(t *testing.T) {
	for _, kind := range []string{"deadline", "maintenance", "credit", "past", "duration", "gap", "missing_preview"} {
		t.Run(kind, func(t *testing.T) {
			db, req, start := taskFixture(t)
			switch kind {
			case "deadline":
				req.Deadline = start.Add(time.Hour)
			case "maintenance":
				db.Model(&models.Equipment{}).Where("id = 2").Update("status", models.StatusMaintenance)
			case "credit":
				db.Model(&models.User{}).Where("id = 1").Update("credit_score", 50)
			case "past":
				req.EarliestStart = time.Now().Add(-time.Hour)
			case "duration":
				req.Steps[1].DurationMinutes = 0
			case "gap":
				req.Steps[0].GapMinutes = 20
			case "missing_preview":
				if _, err := SubmitTask(db, 1, req); err == nil {
					t.Fatal("missing preview accepted")
				}
				return
			}
			if _, err := PlanTask(db, 1, req, false, time.Now()); err == nil {
				t.Fatal("constraint was ignored")
			}
		})
	}
}

func TestRepeatedEquipmentInTask(t *testing.T) {
	db, req, _ := taskFixture(t)
	req.Steps[2].EquipmentID = 1
	task, err := SubmitTask(db, 1, previewRequest(t, db, req))
	if err != nil || len(task.Reservations) != 3 {
		t.Fatalf("repeated equipment: %v", err)
	}
}

func assertCount(t *testing.T, db *gorm.DB, model interface{}, want int64) {
	t.Helper()
	var count int64
	if err := db.Model(model).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("%T count = %d, want %d", model, count, want)
	}
}
