package repository

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"noda/data/model"
	"noda/data/transfer"
	"noda/data/types"
	"regexp"
	"testing"
	"time"
)

const taskID = "f8d5b3a2-80f0-4460-bc40-2762141ffc06"

func TestTaskRepository_Save(t *testing.T) {
	defer beQuiet()()
	db, mock := newMock()
	defer db.Close()
	var (
		r        = NewTaskRepository(db)
		query    = regexp.QuoteMeta(`SELECT make_task ($1, $2, $3);`)
		creation = &transfer.TaskCreation{
			Title:       "task title",
			Description: "task description",
			Headline:    "task headline",
			Priority:    types.TaskPriorityMedium,
			Status:      types.TaskStatusIncomplete,
		}
		res string
		err error
	)

	t.Run("success", func(t *testing.T) {
		mock.
			ExpectQuery(query).
			WithArgs(userID, listID,
				fmt.Sprintf("ROW('%s', '%s', '%s', '%s', '%s', %s, %s)",
					creation.Title, creation.Headline, creation.Description, creation.Priority, creation.Status, "NULL", "NULL")).
			WillReturnRows(sqlmock.
				NewRows([]string{"make_task"}).
				AddRow(taskID))
		res, err = r.Save(userID, listID, creation)
		assert.Equal(t, taskID, res)
		assert.NoError(t, err)
	})

	t.Run("unexpected database error", func(t *testing.T) {
		mock.
			ExpectQuery(query).
			WillReturnError(&pq.Error{})
		res, err = r.Save(userID, listID, creation)
		assert.Error(t, err)
		assert.Equal(t, "", res)
	})
}

func TestTaskRepository_Duplicate(t *testing.T) {
	defer beQuiet()()
	db, mock := newMock()
	defer db.Close()
	var (
		r         = NewTaskRepository(db)
		query     = regexp.QuoteMeta(`SELECT duplicate_task ($1, $2);`)
		res       string
		err       error
		replicaID = uuid.New().String()
	)

	t.Run("success", func(t *testing.T) {
		mock.
			ExpectQuery(query).
			WithArgs(userID, taskID).
			WillReturnRows(sqlmock.
				NewRows([]string{"duplicate_task"}).
				AddRow(replicaID))
		res, err = r.Duplicate(userID, taskID)
		assert.Equal(t, replicaID, res)
		assert.NoError(t, err)
	})

	t.Run("unexpected database error", func(t *testing.T) {
		mock.
			ExpectQuery(query).
			WillReturnError(&pq.Error{})
		res, err = r.Duplicate(userID, taskID)
		assert.Error(t, err)
		assert.Equal(t, "", res)
	})
}

var taskTableColumns = []string{
	"task_id",
	"owner_id",
	"list_id",
	"position_in_list",
	"title",
	"headline",
	"description",
	"priority",
	"status",
	"is_pinned",
	"due_date",
	"remind_at",
	"completed_at",
	"created_at",
	"updated_at"}

func TestTaskRepository_FetchByID(t *testing.T) {
	defer beQuiet()()
	db, mock := newMock()
	defer db.Close()
	var (
		r     = NewTaskRepository(db)
		query = regexp.QuoteMeta(`SELECT fetch_task_by_id ($1, $2, $3);`)
		res   *model.Task
		err   error
		task  = &model.Task{
			ID:             uuid.MustParse(taskID),
			OwnerID:        uuid.MustParse(userID),
			ListID:         uuid.MustParse(listID),
			PositionInList: 1,
			Title:          "task title",
			Headline:       "task headline",
			Description:    "task description",
			Priority:       types.TaskPriorityHigh,
			Status:         types.TaskStatusComplete,
			IsPinned:       false,
			DueDate:        nil,
			RemindAt:       nil,
			CompletedAt:    nil,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
	)

	t.Run("success", func(t *testing.T) {
		mock.
			ExpectQuery(query).
			WithArgs(userID, listID, taskID).
			WillReturnRows(sqlmock.
				NewRows(taskTableColumns).
				AddRow(task.ID, task.OwnerID, task.ListID, task.PositionInList, task.Title, task.Headline, task.Description, task.Priority, task.Status, task.IsPinned, task.DueDate, task.RemindAt, task.CompletedAt, task.CreatedAt, task.UpdatedAt))
		res, err = r.FetchByID(userID, listID, taskID)
		assert.Equal(t, task, res)
		assert.NoError(t, err)
	})

	t.Run("unexpected database error", func(t *testing.T) {
		mock.
			ExpectQuery(query).
			WillReturnError(&pq.Error{})
		res, err = r.FetchByID(userID, listID, taskID)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}