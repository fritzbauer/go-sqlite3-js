// -*- coding: utf-8 -*-
// Copyright 2020 The Matrix.org Foundation C.I.C.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"database/sql"
	"log"

	_ "github.com/matrix-org/go-sqlite3-js/sqlite3_js"
)

var c chan struct{}

func init() {
	c = make(chan struct{})
}

func rowsAffected(msg string, res sql.Result) {
	if res == nil {
		return
	}
	ra, err := res.RowsAffected()
	log.Printf("%s : rows affected: %d %s", msg, ra, err)
}

func main() {
	var db *sql.DB
	var err error
	if db, err = sql.Open("sqlite3_js", "test.db"); err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("create table foo(id int PRIMARY KEY, name string)")
	if err != nil {
		log.Fatal(err)
	}

	res, err := db.Exec("insert into foo values(42, 'meaning of life')")
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected("After insert on db", res)

	res, err = db.Exec("insert into foo values(42, 'meaning of life')")
	if err == nil {
		log.Fatal("Expected err returned from primary key conflict INSERT but got nil")
	}

	res, err = db.Exec("insert into foo values(43, 'meaning of life')")
	if err != nil {
		log.Fatal(err)
	}

	res, err = db.Exec("UPDATE foo SET name='mol' WHERE name='meaning of life'")
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected("After updating 2 rows", res)

	_, err = db.Exec("insert into foo values(?, ?)", 31337, "so leet")
	if err != nil {
		log.Fatal(err)
	}

	var stmt *sql.Stmt
	if stmt, err = db.Prepare("insert into foo values(?, ?)"); err != nil {
		log.Fatal(err)
	}
	res, err = stmt.Exec(12345678, "monotonic")
	rowsAffected("After insert on stmt", res)

	var txn *sql.Tx
	if txn, err = db.Begin(); err != nil {
		log.Fatal(err)
	}
	if stmt, err = db.Prepare("insert into foo values(?, ?)"); err != nil {
		log.Fatal(err)
	}
	stmt = txn.Stmt(stmt)
	stmt.Exec(666, "not happening")
	txn.Rollback()

	if txn, err = db.Begin(); err != nil {
		log.Fatal(err)
	}
	if stmt, err = db.Prepare("insert into foo values(?, ?)"); err != nil {
		log.Fatal(err)
	}
	stmt = txn.Stmt(stmt)
	res, err = stmt.Exec(999, "happening")
	rowsAffected("After insert on stmt in txn", res)
	txn.Commit()

	if stmt, err = db.Prepare("select * from foo"); err != nil {
		log.Fatal(err)
	}
	var id int
	var name string
	stmt.QueryRow().Scan(&id, &name)
	log.Printf("Got first row: %d, %s", id, name)
	stmt.Close()

	rows, err := db.Query("select * from foo")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Got row: %d, %s", id, name)
	}

	<-c
}
