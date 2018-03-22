//
// Author: Vinhthuy Phan, 2018
//
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

//-----------------------------------------------------------------------------------
func teacher_gradesHandler(w http.ResponseWriter, r *http.Request, who string, uid int) {
	content, ext, decision := r.FormValue("content"), r.FormValue("ext"), r.FormValue("decision")
	changed := r.FormValue("changed")
	pid, _ := strconv.Atoi(r.FormValue("pid"))
	stid, _ := strconv.Atoi(r.FormValue("stid"))
	mesg := ""
	if changed == "True" {
		AddFeedbackSQL.Exec(uid, stid, content, time.Now())
		mesg = "Feedback saved to student's board."
		BoardsSem.Lock()
		defer BoardsSem.Unlock()
		b := &Board{
			Content:      content,
			Answer:       "",
			Attempts:     100,
			Ext:          ext,
			Pid:          pid,
			StartingTime: time.Now(),
		}
		Boards[stid] = append(Boards[stid], b)
	}
	score_id, current_merit, current_attempts := 0, 0, 0
	rows, _ := Database.Query("select id, merit, attempts from score where pid=? and stid=?", pid, stid)
	for rows.Next() {
		rows.Scan(&score_id, &current_merit, &current_attempts)
		break
	}
	rows.Close()
	merit, effort := 0, 0
	rows, _ = Database.Query("select merit, effort from problem where id=?", pid)
	for rows.Next() {
		rows.Scan(&merit, &effort)
		break
	}
	rows.Close()
	if decision == "correct" {
		if score_id == 0 {
			_, err := AddScoreSQL.Exec(pid, stid, merit, effort, 1)
			if err != nil {
				panic(err)
			}
		} else {
			_, err := UpdateScoreSQL.Exec(merit, effort, current_attempts+1, score_id)
			if err != nil {
				panic(err)
			}
		}
		mesg = "Problem graded correct.\n" + mesg
		MessageBoards[stid] = "Your submission was correct."

		next_pid, ok := NextProblem[int64(pid)]
		if ok {
			new_content, new_answer, new_ext, new_merit, new_effort, new_attempts := "", "", "", 0, 0, 0
			rows, _ = Database.Query("select content, answer, ext, merit, effort, attempts from problem where id=?", next_pid)
			for rows.Next() {
				rows.Scan(&new_content, &new_answer, &new_ext, &new_merit, &new_effort, &new_attempts)
				break
			}
			rows.Close()
			b := &Board{
				Content:      new_content,
				Answer:       new_answer,
				Attempts:     new_attempts,
				Ext:          new_ext,
				Pid:          int(next_pid),
				StartingTime: time.Now(),
			}
			Boards[stid] = append(Boards[stid], b)
			mesg = "Next problem added to student's board\n" + mesg
			MessageBoards[stid] += " You have a new problem on board."
		}
	} else {
		if score_id == 0 {
			_, err := AddScoreSQL.Exec(pid, stid, 0, effort, 1)
			if err != nil {
				panic(err)
			}
		} else {
			_, err := UpdateScoreSQL.Exec(current_merit, effort, current_attempts+1, score_id)
			if err != nil {
				panic(err)
			}
		}
		mesg = "Problem graded incorrect.\n" + mesg
		MessageBoards[stid] = "Your submission was not correct.  Try again."
	}
	fmt.Fprintf(w, mesg)
}

//-----------------------------------------------------------------------------------