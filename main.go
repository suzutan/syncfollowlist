package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	t "github.com/suzutan/syncfollows/internal/pkg/twitter"
)

type contextKey string

const contextClient contextKey = "client"
const contextOwnerID contextKey = "ownerID"
const contextListID contextKey = "listID"

func do(ctx context.Context) {

	client := ctx.Value(contextClient).(*twitter.Client)
	listID := ctx.Value(contextListID).(int64)
	ownerID := ctx.Value(contextOwnerID).(int64)

	cFollows := make(chan []int64)
	cListIDs := make(chan []int64)

	go func() {
		// get follows
		log.Print("fetch friend IDs")
		friendIDs, _, err := client.Friends.IDs(&twitter.FriendIDParams{
			Count: 5000,
		})
		if err != nil {
			log.Print(err)
			return
		}
		// adding ownerID to friendIDs
		follows := append(friendIDs.IDs, ownerID)
		cFollows <- follows
	}()

	go func() {
		// get follows list members
		log.Print("fetch List IDs")

		listMembers, _, err := client.Lists.Members(&twitter.ListsMembersParams{
			ListID: listID,
			Count:  5000,
		})
		if err != nil {
			log.Print(err)
			return
		}

		// map list members to IDs
		var listIDs []int64
		for _, member := range listMembers.Users {
			listIDs = append(listIDs, member.ID)
		}
		cListIDs <- listIDs
	}()

	follows := <-cFollows
	listIDs := <-cListIDs
	log.Printf("%d follows, %d list IDs", len(follows), len(listIDs))

	var addIDs = Int64ListDivide(follows, listIDs)
	var delIDs = Int64ListDivide(listIDs, follows)

	cAdd := make(chan int)
	cDel := make(chan int)

	go func() {
		//  add follows to list
		if len(addIDs) > 0 {
			res, err := client.Lists.MembersCreateAll(&twitter.ListsMembersCreateAllParams{
				ListID: listID,
				UserID: strings.Trim(strings.Join(strings.Fields(fmt.Sprint(addIDs)), ","), "[]"),
			})

			if err != nil {
				log.Print(err)
				cAdd <- 1
				return
			}
			if res.StatusCode == http.StatusOK {
				log.Printf("add success. count:%d\n", len(addIDs))
			} else {
				log.Printf("add failed. %s", res.Status)
			}
		} else {
			log.Print("addIds is 0, skip.")
		}
		cAdd <- 0

	}()

	go func() {
		// remove follows from list
		if len(delIDs) > 0 {
			res, err := client.Lists.MembersDestroyAll(&twitter.ListsMembersDestroyAllParams{
				ListID: listID,
				UserID: strings.Trim(strings.Join(strings.Fields(fmt.Sprint(delIDs)), ","), "[]"),
			})

			if err != nil {
				log.Print(err)
				cDel <- 1
				return
			}
			if res.StatusCode == http.StatusOK {
				log.Printf("delete success. count:%d\n", len(delIDs))
			} else {
				log.Printf("delete failed. %s", res.Status)
			}
		} else {
			log.Print("delIDs is 0, skip.")
		}
		cDel <- 0
	}()

	<-cAdd
	<-cDel

}

func run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	do(ctx)
	log.Printf("wait for %s\n", interval.String())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(c)
	defer close(c)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c:
			log.Print("Interrupt, stop")
			return
		case <-ticker.C:
			do(ctx)
			log.Printf("wait for %s\n", interval.String())
		}
	}
}

func main() {

	auth := &t.AuthConfig{
		ConsumerKey:       os.Getenv("CK"),
		ConsumerSecret:    os.Getenv("CS"),
		AccessToken:       os.Getenv("AT"),
		AccessTokenSecret: os.Getenv("ATS"),
	}
	listID, _ := strconv.ParseInt(os.Getenv("LIST_ID"), 10, 64)
	ownerID, _ := strconv.ParseInt(strings.Split(auth.AccessToken, "-")[0], 10, 64)
	client := t.New(auth)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = context.WithValue(ctx, contextClient, client)
	ctx = context.WithValue(ctx, contextListID, listID)
	ctx = context.WithValue(ctx, contextOwnerID, ownerID)

	run(ctx, 5*time.Minute)

}

// Int64ListDivide mainListを総当りし、divideListに存在しないrecordを抽出する
// [1,2,3] - [1,2] = [3]
// for i in [1,2,3]:
// 1: 1 in [1,2] -> true
// 2: 2 in [1,2] -> true
// 3: 3 in [1,2] -> false -> add 3 to sublist
// return [3]
func Int64ListDivide(mainList []int64, divideList []int64) []int64 {
	var result []int64
	for _, id := range mainList {
		if !int64Contains(divideList, id) {
			result = append(result, id)
		}
	}
	return result
}

func int64Contains(list []int64, target int64) bool {
	for _, id := range list {
		if id == target {
			return true
		}
	}
	return false
}
