package main

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
)

type user struct {
	name string
	room set
}

var (
	address        map[string]net.UDPAddr // address["userName"] = địa chỉ của user
	chatRooms      map[string]set         // chatRooms["roomName"] = chứa các user trong room
	availableUsers map[string]user        // availableUsers["userName"] = kiểm tra xem user có available hay ko
)

func main() {
	availableUsers = make(map[string]user, 1)
	chatRooms = make(map[string]set, 1)
	chatRooms["all"] = newSet()
	address = make(map[string]net.UDPAddr, 1)

	PORT := ":8080"

	s, err := net.ResolveUDPAddr("udp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer connection.Close()

	for {
		WaitForRequest(connection)
	}
}

func WaitForRequest(connection *net.UDPConn) {
	buffer := make([]byte, 1024)
	n, addr, _ := connection.ReadFromUDP(buffer)

	user := user{name: addr.String(), room: newSet()}
	// check if user exists
	if _, ok := availableUsers[addr.String()]; ok {
		// user exists
		message := string(buffer[0:n])

		// check if message is private
		if isPrivateMessage(message) {
			// get the user name
			pattern := `(\d+\.\d+\.\d+\.\d+:\d+)`
			re := regexp.MustCompile(pattern)
			userName := re.FindStringSubmatch(message)[1]
			add := address[userName]

			s := fmt.Sprintf("%s%s", addr.String(), message)
			log.Printf("Private message from %s to %s: %s", addr.String(), userName, message)
			connection.WriteToUDP([]byte(s), addr)
			connection.WriteToUDP([]byte(s), &add)
		} else if strings.Contains(message, "STOP") {
			// remove user from all rooms
			for room := range user.room.m {
				removeUserFromRoom(room, &user)
			}
			delete(availableUsers, user.name)
			delete(address, user.name)
			s := "user " + user.name + " has left the chat room\n"
			broadcastMessage(s, "all", connection)
		} else {
			s := fmt.Sprintf("%s: %s", addr.String(), message)
			broadcastMessage(s, "all", connection)
		}
	} else {
		// user does not exist
		insertUserToRoom("all", user)
		insertUserToUserList(user)
		insertUserToAddressList(user, *addr)

		s := "welcome " + user.name + " to the chat room\n"
		broadcastMessage(s, "all", connection)
	}
}

func broadcastMessage(message string, room string, connection *net.UDPConn) {
	log.Printf("Broadcasting message: %s to room: %s", message, room)
	if _, ok := chatRooms[room]; ok {
		for userName := range chatRooms[room].m {
			// send message to user
			addr := address[userName]
			connection.WriteToUDP([]byte(message), &addr)
		}
	}
}

func insertUserToAddressList(user user, addr net.UDPAddr) {
	address[user.name] = addr
}

func insertUserToRoom(room string, user user) {
	if _, ok := chatRooms[room]; ok {
		chatRooms[room].insert(user.name)
	} else {
		chatRooms[room] = newSet()
		chatRooms[room].insert(user.name)
	}
}

func insertUserToUserList(user user) {
	availableUsers[user.name] = user
}

func removeUserFromRoom(room string, user *user) {
	if _, ok := chatRooms[room]; ok {
		chatRooms[room].remove(user.name)
		user.room.remove(room)
	}
}

type set struct {
	m map[string]struct{}
}

func newSet() set {
	return set{make(map[string]struct{})}
}
func (s set) insert(value string) {
	s.m[value] = struct{}{}
}
func (s set) remove(value string) {
	delete(s.m, value)
}

func isPrivateMessage(input string) bool {
	// Regular expression to match the pattern "@string string"
	pattern := `@\d+\.\d+\.\d+\.\d+:\d+ \S+ \S+`
	re := regexp.MustCompile(pattern)

	// Check if the input matches the pattern
	if re.MatchString(input) {
		return true
	} else {
		return false
	}
}
