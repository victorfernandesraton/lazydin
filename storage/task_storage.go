package storage

import "database/sql"

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
    )
`
const createTask = `INSERT INTO tasks (status) VALUES (?) RETURNING id, status, created_at, updated_at;`

const updateTask = `UPDATE tasks SET status = ? SET updated_at = CURRENT_TIMESTAMP WHERE id = ? RETURNING id, status, created_at, updated_at;`

const selectTaskById = `SELECT id, status, created_at, updated_at FROM tasks WHERE id = ?;`

const selectAllTasks = `SELECT id, status, created_at, updated_at FROM tasks;`

const getPostsByTaskId = `SELECT url, content, author_url , created_at, updated_at FROM task_post INNER JOIN posts ON task_post.post_url = posts.url WHERE task_id = ?;`

const getPostsByTasks = `SELECT post.url as url, post.content as content, post.author_url as author_url , post.created_at as created_at, post.updated_at as updated_at FROM task_post INNER JOIN posts ON task_post.post_url = posts.url INNER JOIN tasks on task_post.task_id = task.id and task.status > 1`

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
