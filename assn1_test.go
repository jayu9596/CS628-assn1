package assn1

import (
	"reflect"
	"testing"

	"github.com/sarkarbidya/CS628-assn1/userlib"
)

// You can actually import other stuff if you want IN YOUR TEST
// HARNESS ONLY.  Note that this is NOT considered part of your
// solution, but is how you make sure your solution is correct.

func TestInitUser(t *testing.T) {
	t.Log("Initialization test")
	userlib.DebugPrint = true
	userlib.DebugPrint = false
	_, err1 := InitUser("", "")
	if err1 != nil {
		t.Log("Failed to initialize user")

	} else {
		t.Error("Initialized invalid user", err1)
	}

	// add more test cases here
}

func TestUserStorage(t *testing.T) {
	u1, err1 := GetUser("", "fubar")
	if err1 != nil {
		t.Log("Cannot load data for invalid user", u1)
	} else {
		t.Error("Data loaded for invalid user", err1)
	}

	// add more test cases here
}

func TestFileStoreLoadAppend(t *testing.T) {
	data1 := userlib.RandomBytes(4096)
	u1, err1 := InitUser("usernmae", "password")
	u1, err1 = GetUser("usernmae", "password")
	if err1 != nil {
		t.Error("Cannot load data for invalid user", u1)
	}
	//Store Load File TestCase
	ab1 := u1.StoreFile("file1", data1)
	if ab1 != nil {
		t.Error("Cannot store file file1", ab1)
	}
	data2, ab := u1.LoadFile("file1", 0)
	if ab != nil {
		t.Error("Cannot Load file file1", ab)
	}
	if !reflect.DeepEqual(data1, data2) {
		t.Error("data corrupted")
	} else {
		t.Log("data is not corrupted")
	}

	//Append File TestCase
	datanew1 := userlib.RandomBytes(4096)
	ab = u1.AppendFile("file1", datanew1)
	if ab != nil {
		t.Error("Cannot append to file file1", ab)
	}
	data2, ab = u1.LoadFile("file1", 1)
	if ab != nil {
		t.Error("Cannot Load file file1", ab)
	}
	if !reflect.DeepEqual(datanew1, data2) {
		t.Error("data corrupted")
	} else {
		t.Log("data is not corrupted")
	}

	// add test cases here
}

func TestFileShareReceive(t *testing.T) {
	// add test cases here

	data1 := userlib.RandomBytes(4096)
	u1, err1 := InitUser("usernmae", "password")
	u1, err1 = GetUser("usernmae", "password")
	u2, err1 := InitUser("usernmae1", "password1")
	u2, err1 = GetUser("usernmae1", "password1")
	ab1 := u1.StoreFile("file1", data1)
	if ab1 != nil || err1 != nil {
		t.Error("Cannot store file file1", ab1)
	}
	msg, err := u1.ShareFile("file1", "usernmae1")
	err = u2.ReceiveFile("abc", "usernmae", msg)
	data2, ab := u2.LoadFile("abc", 0)
	if ab != nil || err != nil {
		t.Error("Cannot Load file file1\n", ab)
	}
	if !reflect.DeepEqual(data1, data2) {
		t.Error("data corrupted")
	} else {
		t.Log("data is not corrupted")
	}
	err = u1.RevokeFile("file1")
	if err != nil {
		t.Error("Revoke Failed\n", ab)
	}
	data2, ab = u2.LoadFile("abc", 0)
	if ab != nil {
		t.Log("Test Functionality Failed\n")
	} else {
		t.Error("Test Case Passed", ab)
	}
	data2, ab = u1.LoadFile("file1", 0)
	if ab != nil || err != nil {
		t.Error("Cannot Load file file1\n", ab)
	}

}
