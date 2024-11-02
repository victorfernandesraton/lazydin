package storage

import (
	"database/sql"

	"github.com/victorfernandesraton/lazydin/domain"
)

const createTaskTableQuery = `
    CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        status INTEGER DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS task_post (
        task_id INTEGER,
        post_url TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(task_id) REFERENCES tasks(id),
        FOREIGN KEY(post_url) REFERENCES posts(url)
    );
`
const createTask = `INSERT INTO tasks (status) VALUES (?) RETURNING id, status, created_at, updated_at;`

const createTaskPost = `INSERT INTO task_post (task_id, post_url) VALUES (?, ?) RETURNING task_id, post_url, created_at, updated_at;`
const updateTask = `UPDATE tasks SET status = ? SET updated_at = CURRENT_TIMESTAMP WHERE id = ? RETURNING id, status, created_at, updated_at;`

const selectTaskById = `SELECT id, status, created_at, updated_at FROM tasks WHERE id = ?;`

const selectAllTasks = `SELECT id, status, created_at, updated_at FROM tasks;`

const selectDistinctPostsByTaskId = `SELECT DISTINCT post_url FROM task_post WHERE task_id = ?;`

const selectDistinctTaskPostsByPostUrl = `SELECT DISTINCT task_id FROM task_post WHERE post_url = ?;`

type TaskStorage struct {
	db *sql.DB
}

func NewTaskStorage(db *sql.DB) *TaskStorage {
	return &TaskStorage{db: db}
}

func (ts *TaskStorage) CreateTable() error {
	_, err := ts.db.Exec(createTaskTableQuery)
	return err
}

func (ts *TaskStorage) CreateTask() (*domain.Task, error) {
	var task domain.Task
	err := ts.db.QueryRow(createTask, 1).
		Scan(&task.Id, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (ts *TaskStorage) CreateTaskPost(taskId int64, postUrl string) (*domain.TaskPost, error) {
	var task domain.TaskPost
	err := ts.db.QueryRow(createTaskPost, taskId, postUrl).Scan(&task.TaskId, &task.PostUrl, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (ts *TaskStorage) UpdateTaskById(id int64, status int) (*domain.Task, error) {
	var task domain.Task
	err := ts.db.QueryRow(updateTask, status, id).
		Scan(&task.Id, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (ts *TaskStorage) GetTaskById(id int64) (*domain.Task, error) {
	var task domain.Task
	err := ts.db.QueryRow(selectTaskById, id).
		Scan(&task.Id, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (ts *TaskStorage) GetAllTasks() ([]domain.Task, error) {
	var tasks []domain.Task
	rows, err := ts.db.Query(selectAllTasks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var task domain.Task
		err = rows.Scan(&task.Id, &task.Status, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (ts *TaskStorage) GetDistinctPostsByTaskId(taskId int64) ([]string, error) {
	rows, err := ts.db.Query(selectDistinctPostsByTaskId, taskId)
	if err != nil {
		return nil, err
	}
	var urls []string
	defer rows.Close()
	for rows.Next() {
		var url string
		err = rows.Scan(&url)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func (ts *TaskStorage) GetDistinctTaskPostsByPostUrl(postUrl string) ([]int64, error) {
	rows, err := ts.db.Query(selectDistinctTaskPostsByPostUrl, postUrl)
	if err != nil {
		return nil, err
	}
	var taskIds []int64
	defer rows.Close()
	for rows.Next() {
		var taskId int64
		err = rows.Scan(&taskId)
		if err != nil {
			return nil, err
		}
		taskIds = append(taskIds, taskId)
	}
	return taskIds, nil
}
