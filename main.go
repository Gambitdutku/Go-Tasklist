package main

import (
    "bufio"
    "database/sql"
    "fmt"
    "log"
    "os"
    "time"

    _ "github.com/go-sql-driver/mysql" // MySQL driver
)

var (
    currentTaskID int
    elapsedTimes   = make(map[int]time.Duration) // Store elapsed time for each task
    timer          *time.Timer
)

// Database connection function
func dbConnection() (*sql.DB, error) {
    dsn := "username:password@tcp(127.0.0.1:3306)/todo_app" // Update with your DB credentials
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    err = db.Ping()
    if err != nil {
        return nil, err
    }
    fmt.Println("Database connection established!")
    return db, nil
}

// Function to add a task
func addTask(db *sql.DB, title, description string) (int64, error) {
    query := "INSERT INTO tasks (title, description, status_id, elapsed_time) VALUES (?, ?, 1, 0)"
    result, err := db.Exec(query, title, description)
    if err != nil {
        return 0, err
    }
    taskID, err := result.LastInsertId()
    return taskID, err
}

// Function to get status string from the database
func getStatusString(db *sql.DB, statusID int) (string, error) {
    var status string
    query := "SELECT name FROM status WHERE id = ?"
    err := db.QueryRow(query, statusID).Scan(&status)
    if err != nil {
        return "", err
    }
    return status, nil
}

// Function to list tasks from the database
func listTasks(db *sql.DB) ([]string, error) {
    query := "SELECT id, title, description, status_id FROM tasks"
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tasks []string
    for rows.Next() {
        var id int
        var title string
        var description string
        var statusID int
        if err := rows.Scan(&id, &title, &description, &statusID); err != nil {
            return nil, err
        }

        // Fetch the status string from the database
        status, err := getStatusString(db, statusID)
        if err != nil {
            return nil, err
        }
        tasks = append(tasks, fmt.Sprintf("ID: %d | Title: %s | Description: %s | Status: %s", id, title, description, status))
    }
    return tasks, nil
}

// Function to update task status
func updateTaskStatus(db *sql.DB, taskID int, status int) error {
    query := "UPDATE tasks SET status_id = ? WHERE id = ?"
    _, err := db.Exec(query, status, taskID)
    return err
}

// Function to edit a task
func editTask(db *sql.DB, taskID int, newTitle, newDescription string) error {
    query := "UPDATE tasks SET title = ?, description = ? WHERE id = ?"
    _, err := db.Exec(query, newTitle, newDescription, taskID)
    return err
}

// Function to remove a task
func removeTask(db *sql.DB, taskID int) error {
    query := "DELETE FROM tasks WHERE id = ?"
    _, err := db.Exec(query, taskID)
    return err
}

// Start task timer
func startTask(taskID int) {
    if timer != nil {
        timer.Stop() // Stop any existing timer
    }

    currentTaskID = taskID
    elapsedTimes[taskID] = 0 // Reset elapsed time for this task

    timer = time.NewTimer(1 * time.Second)

    go func() {
        for {
            <-timer.C
            elapsedTimes[taskID] += 1 * time.Second
            // Here you might want to update the elapsed time in the database periodically
            timer.Reset(1 * time.Second)
        }
    }()
}

// Function to stop the task timer
func stopTask() {
    if timer != nil {
        timer.Stop() // Stop the timer
    }
}

// Main function
func main() {
    db, err := dbConnection()
    if err != nil {
        log.Fatal("Error connecting to the database:", err)
    }
    defer db.Close()

    scanner := bufio.NewScanner(os.Stdin)

    for {
        var choice int
        fmt.Println("1. Add Task")
        fmt.Println("2. List Tasks")
        fmt.Println("3. Start Task")
        fmt.Println("4. Stop Task")
        fmt.Println("5. Finish Task")
        fmt.Println("6. Edit Task")
        fmt.Println("7. Remove Task")
        fmt.Println("8. Exit")
        fmt.Print("Choose an option: ")

        _, err := fmt.Scan(&choice)
        if err != nil {
            log.Println("Invalid input. Please enter a number.")
            continue
        }

        switch choice {
        case 1:
            var title, description string
            fmt.Print("Enter task title: ")
            scanner.Scan()
            title = scanner.Text() // Read full line for title

            fmt.Print("Enter task description: ")
            scanner.Scan()
            description = scanner.Text() // Read full line for description

            taskID, err := addTask(db, title, description)
            if err != nil {
                log.Println("Error adding task:", err)
            } else {
                fmt.Printf("Task added with ID: %d\n", taskID)
            }

        case 2:
            tasks, err := listTasks(db)
            if err != nil {
                log.Println("Error listing tasks:", err)
            } else {
                fmt.Println("Tasks:")
                for _, task := range tasks {
                    fmt.Println(task)
                }
            }

        case 3:
            var taskID int
            fmt.Print("Enter task ID to start: ")
            fmt.Scan(&taskID)
            err := updateTaskStatus(db, taskID, 3) // Set status to In Progress
            if err != nil {
                log.Println("Error starting task:", err)
            } else {
                startTask(taskID)
                fmt.Printf("Task ID %d is now started.\n", taskID)
            }

        case 4:
            stopTask()
            fmt.Println("Task timer stopped.")

        case 5:
            var taskID int
            fmt.Print("Enter task ID to finish: ")
            fmt.Scan(&taskID)
            err := updateTaskStatus(db, taskID, 2) // Set status to Completed
            if err != nil {
                log.Println("Error finishing task:", err)
            } else {
                stopTask()
                fmt.Printf("Task ID %d is now finished.\n", taskID)
            }

        case 6:
            var taskID int
            var newTitle, newDescription string
            fmt.Print("Enter task ID to edit: ")
            fmt.Scan(&taskID)

            fmt.Print("Enter new task title: ")
            scanner.Scan()
            newTitle = scanner.Text() // Read full line for new title

            fmt.Print("Enter new task description: ")
            scanner.Scan()
            newDescription = scanner.Text() // Read full line for new description

            err := editTask(db, taskID, newTitle, newDescription)
            if err != nil {
                log.Println("Error editing task:", err)
            } else {
                fmt.Printf("Task ID %d has been updated.\n", taskID)
            }

        case 7:
            var taskID int
            fmt.Print("Enter task ID to remove: ")
            fmt.Scan(&taskID)
            err := removeTask(db, taskID)
            if err != nil {
                log.Println("Error removing task:", err)
            } else {
                fmt.Printf("Task ID %d has been removed.\n", taskID)
            }

        case 8:
            fmt.Println("Exiting...")
            return

        default:
            fmt.Println("Invalid option. Please try again.")
        }
    }
}
