package server

import (
	"database/sql"
	"fmt"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

var sqlTable = `
DROP TABLE If Exists Exercise;
DROP TABLE If Exists Warmup;
DROP TABLE If Exists Warmdown;
DROP TABLE If Exists Interval;
DROP TABLE If Exists Repeat;

CREATE TABLE IF NOT EXISTS Exercise (
	id integer PRIMARY KEY,
	name varchar(255),
    comment text DEFAULT "",
	CONSTRAINT uc_ExerciseID UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS Warmup (
	id integer PRIMARY KEY,
	position tinyint DEFAULT 0,
	effort_type varchar(32) DEFAULT "distance",
    effort text NOT NULL,
	repeatID integer,
	exerciseID integer
);

CREATE TABLE IF NOT EXISTS Warmdown (
	id integer PRIMARY KEY,
	position tinyint DEFAULT 0,
	effort_type varchar(32) DEFAULT "distance",
    effort text NOT NULL,
	repeatID integer,
	exerciseID integer
);

CREATE TABLE IF NOT EXISTS Interval (
	id integer PRIMARY KEY,
	position tinyint DEFAULT 0,
	laps tinyint NOT NULL,
    length INTEGER NOT NULL,
	percentage tinyint NOT NULL,
	rest text,
	effort_type varchar(32) DEFAULT "distance",
	effort text DEFAULT "", -- storing time in there
	repeatID integer,
	exerciseID integer
);

Create Table IF NOT EXISTS Repeat  (
	id integer PRIMARY KEY,
	repeat tinyint,
	position tinyint DEFAULT 0,
	exerciseID integer
);
`

//TODO: remove
var aSample = `
	DELETE FROM Exercise;
	DELETE FROM Warmup;
	DELETE FROM Warmdown;
	DELETE FROM Interval;
	DELETE FROM Repeat;

	INSERT INTO Exercise(name) VALUES("Pyramids Short");

	INSERT INTO Warmup(effort_type, effort, position, exerciseID) VALUES("distance", "5km very easy around", 0, 1);
	INSERT INTO Warmdown(effort_type, effort, position, exerciseID) VALUES("time", "15 mn footing", 2, 1);

	-- INSERT INTO Repeat(Repeat, position, exerciseID) VALUES(5, 1, 1);
	-- INSERT INTO Interval(laps, length, percentage, rest, effort_type, repeatID) VALUES(6, 1000, 90, "400m active", "distance", 1);

	INSERT INTO Interval(position, laps, length, percentage, rest, effort_type, exerciseID) VALUES(1, 1, 400, 100, "repos arret", "distance", 1);
`

func DBConnect() (err error) {
	// TODO: proper sqlite location
	DB, err = sql.Open("sqlite3", "/tmp/a.db")
	if err != nil {
		return
	}

	_, err = DB.Exec(sqlTable)
	if err != nil {
		return
	}

	_, err = DB.Exec(aSample)
	return
}

func sqlTX(query string, args ...interface{}) (res sql.Result, err error) {
	tx, err := DB.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(query)
	if err != nil {
		return

	}
	defer stmt.Close()
	res, err = stmt.Exec(args...)
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}

func getSteps(t string, id int, steps *[]Step) (err error) {
	var getWarmupSQL = fmt.Sprintf(`SELECT id, position, effort, effort_type FROM Warmup WHERE %s=?`, t)
	var getWarmdownSQL = fmt.Sprintf(`SELECT id, position, effort, effort_type FROM Warmdown WHERE %s=?`, t)
	var getIntervalSQL = fmt.Sprintf(`SELECT id, position, laps, length, percentage, rest, effort_type, effort FROM Interval WHERE %s=?`, t)

	rows, err := DB.Query(getWarmupSQL, id)
	if err != nil {
		if err != sql.ErrNoRows {
			return
		}
	}
	for rows.Next() {
		var step = Step{
			Type: "warmup",
		}
		err = rows.Scan(&step.ID, &step.Position, &step.Effort, &step.EffortType)
		*steps = append(*steps, step)
	}

	rows, err = DB.Query(getWarmdownSQL, id)
	if err != nil {
		if err != sql.ErrNoRows {
			return
		}
	}
	for rows.Next() {
		var step = Step{
			Type: "warmdown",
		}
		err = rows.Scan(&step.ID, &step.Position, &step.Effort, &step.EffortType)
		*steps = append(*steps, step)
	}

	rows, err = DB.Query(getIntervalSQL, id)
	if err != nil {
		if err != sql.ErrNoRows {
			return
		}
	}
	for rows.Next() {
		var step = Step{
			Type: "interval",
		}
		err = rows.Scan(&step.ID, &step.Position, &step.Laps, &step.Length, &step.Percentage, &step.Rest, &step.EffortType, &step.Effort)
		*steps = append(*steps, step)
	}
	return
}

func getProgram(exerciseID int) (exercise Exercise, err error) {
	var getExerciseSQL = `SELECT name, comment FROM Exercise WHERE id=?`
	var getRepeatSQL = `SELECT id, position, repeat FROM Repeat WHERE exerciseID=?`
	var steps []Step
	var repeatStep []Step

	exercise = Exercise{
		ID: exerciseID,
	}

	err = DB.QueryRow(getExerciseSQL, exerciseID).Scan(
		&exercise.Name,
		&exercise.Comment)
	if err != nil {
		if err != sql.ErrNoRows {
			return
		}
		err = &error404{"ProgramNotFound"}
		return
	}

	err = getSteps("exerciseID", exercise.ID, &steps)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	step := Step{
		Type: "repeat",
	}
	repeat := Repeat{}
	err = DB.QueryRow(getRepeatSQL, exercise.ID).Scan(
		&repeat.ID,
		&step.Position,
		&repeat.Repeat)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Printf("repeat error2: %s\n", err.Error())
			return
		} else {
			err = nil
		}
	} else {
		err = getSteps("repeatID", exercise.ID, &repeatStep)
		if err != nil {
			if err != sql.ErrNoRows {
				return
			} else {
				err = nil
			}
		}
	}
	repeat.Steps = repeatStep
	step.Repeat = repeat
	if len(repeat.Steps) != 0 {
		steps = append(steps, step)
	}
	exercise.Steps = steps

	sort.Sort(exercise)

	return
}

func getAllPrograms() (exercises []Exercise, err error) {
	var getAllExercises = `SELECT ID from Exercise`
	rows, err := DB.Query(getAllExercises)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return
	}

	for rows.Next() {
		e := Exercise{}
		err = rows.Scan(&e.ID)
		if err != nil {
			return
		}
		e, err = getProgram(e.ID)
		if err != nil {
			return
		}
		exercises = append(exercises, e)
	}
	return
}

func addProgram(exercise Exercise) (res sql.Result, err error) {
	sql := `insert or replace into Exercise (ID, name, comment) values (?, ?, ?);`
	res, err = sqlTX(sql, exercise.ID, exercise.Name, exercise.Comment)
	if err != nil {
		fmt.Println(err.Error())
	}

	res, err = sqlTX(`delete from Warmup where exerciseID=?`, exercise.ID)
	if err != nil {
		fmt.Println(err.Error())
	}
	res, err = sqlTX(`delete from Warmdown where exerciseID=?`, exercise.ID)
	if err != nil {
		fmt.Println(err.Error())
	}
	res, err = sqlTX(`delete from Interval where exerciseID=?`, exercise.ID)
	if err != nil {
		fmt.Println(err.Error())
	}

	for position, step := range exercise.Steps {

		if step.Type == "warmup" {
			sql = `insert into Warmup (id, effort_type, effort, position, exerciseID) values (?, ?, ?, ?, ?);`
			res, err = sqlTX(sql, step.ID, step.EffortType, step.Effort, position, exercise.ID)
			if err != nil {
				fmt.Println(err.Error())
			}
		} else if step.Type == "warmdown" {
			sql = `insert into Warmdown (id, effort_type, effort, position, exerciseID) values (?, ?, ?, ?, ?);`
			res, err = sqlTX(sql, step.ID, step.EffortType, step.Effort, position, exercise.ID)
			if err != nil {
				fmt.Println(err.Error())
			}

		} else if step.Type == "interval" {
			sql = ``
			res, err = sqlTX(`
			insert into Interval (
			id, position, laps,
			length, percentage, rest,
			effort_type, effort, exerciseID) values
			(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				step.ID, position, step.Laps,
				step.Length, step.Percentage, step.Rest,
				step.EffortType, step.Effort, exercise.ID)
			if err != nil {
				fmt.Println(err.Error())
			}

		}
	}
	return
}
