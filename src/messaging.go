//
// Author: Vinhthuy Phan, 2018
//
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type BulletinBoardMessage struct {
	Code           string
	I              int
	NextI          int
	PrevI          int
	PC             string
	P1             int
	P2             int
	ActiveProblems int
	BulletinItems  int
	Attendance     int
	Address        string
	Authenticated  bool
}

type AnswersBoardMessage struct {
	Counts  map[string]int
	Content string
}

//-----------------------------------------------------------------------------------
func view_answersHandler(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.FormValue("pid"))
	passcode := r.FormValue("pc")
	if _, ok := ActiveProblems[pid]; err == nil && ok && passcode == Passcode {
		t := template.New("")
		t, err := t.Parse(VIEW_ANSWERS_TEMPLATE)
		if err == nil {
			answers := ActiveProblems[pid].Answers
			counts := make(map[string]int)
			for i := 0; i < len(answers); i++ {
				counts[answers[i]]++
			}

			rows, _ := Database.Query("select content from problem where id=?", pid)
			defer rows.Close()
			content := ""
			for rows.Next() {
				rows.Scan(&content)
			}
			w.Header().Set("Content-Type", "text/html")
			t.Execute(w, &AnswersBoardMessage{Counts: counts, Content: content})
		} else {
			fmt.Println(err)
		}
	}
}

//-----------------------------------------------------------------------------------
func teacher_adds_bulletin_pageHandler(w http.ResponseWriter, r *http.Request, who string, uid int) {
	BulletinSem.Lock()
	defer BulletinSem.Unlock()
	BulletinBoard = append(BulletinBoard, r.FormValue("content"))
	fmt.Fprintf(w, "Content added to bulletin board")
}

//-----------------------------------------------------------------
func remove_bulletin_pageHandler(w http.ResponseWriter, r *http.Request) {
	BulletinSem.Lock()
	defer BulletinSem.Unlock()
	i, _ := strconv.Atoi(r.FormValue("i"))
	passcode := r.FormValue("pc")
	if passcode == Passcode && i >= 0 && i < len(BulletinBoard) {
		BulletinBoard = append(BulletinBoard[:i], BulletinBoard[i+1:]...)
		http.Redirect(w, r, "view_bulletin_board?i=0&pc="+passcode, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "view_bulletin_board?i="+r.FormValue("i")+"&pc="+passcode, http.StatusSeeOther)
	}
}

//-----------------------------------------------------------------------------------
func get_bulletin_board_data(i int, passcode string) *BulletinBoardMessage {
	if i >= len(BulletinBoard) {
		i = 0
	}
	// Get code and build page links
	code := ""
	if i >= 0 && i < len(BulletinBoard) {
		code = BulletinBoard[i]
	}

	// Get priority counts
	priority := []int{0, 0, 0}
	for j := 0; j < len(WorkingSubs); j++ {
		priority[WorkingSubs[j].Priority]++
	}
	next_i, prev_i := 0, 0
	if len(BulletinBoard) > 0 {
		next_i = (i + 1 + len(BulletinBoard)) % len(BulletinBoard)
		prev_i = (i - 1 + len(BulletinBoard)) % len(BulletinBoard)
	}
	data := &BulletinBoardMessage{
		Code:           code,
		I:              i,
		NextI:          next_i,
		PrevI:          prev_i,
		PC:             passcode,
		P1:             priority[1],
		P2:             priority[2],
		ActiveProblems: len(ActiveProblems),
		BulletinItems:  len(BulletinBoard),
		Attendance:     len(Students),
		Address:        Config.Address,
		Authenticated:  passcode == Passcode,
	}
	return data
}

//-----------------------------------------------------------------------------------
func bulletin_board_dataHandler(w http.ResponseWriter, r *http.Request) {
	data := get_bulletin_board_data(0, "")
	js, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(js)
}

//-----------------------------------------------------------------------------------
func view_bulletin_boardHandler(w http.ResponseWriter, r *http.Request) {
	i, err := strconv.Atoi(r.FormValue("i"))
	passcode := r.FormValue("pc")
	if err != nil {
		i = 0
	}

	temp := template.New("")
	t, err2 := temp.Parse(TEACHER_MESSAGING_TEMPLATE)
	if err2 != nil {
		log.Fatal(err2)
	}
	data := get_bulletin_board_data(i, passcode)
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

//-----------------------------------------------------------------------------------
// func student_messagesHandler(w http.ResponseWriter, r *http.Request) {
// 	stid, err := strconv.Atoi(r.FormValue("stid"))
// 	if err != nil {
// 		fmt.Fprintf(w, "Error")
// 	}
// 	_, ok := Students[stid]
// 	if ok {
// 		t := template.New("")
// 		t, err := t.Parse(STUDENT_MESSAGING_TEMPLATE)
// 		if err == nil {
// 			data := struct{ Message string }{Students[stid].Status}
// 			w.Header().Set("Content-Type", "text/html")
// 			t.Execute(w, data)
// 		} else {
// 			fmt.Println(err)
// 		}
// 	} else {
// 		fmt.Fprint(w, "Error")
// 	}
// }