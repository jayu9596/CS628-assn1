package assn1

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.
//
import (

	// You neet to add with
	// go get github.com/sarkarbidya/CS628-assn1/userlib

	"crypto/rsa"
	"fmt"

	"github.com/sarkarbidya/CS628-assn1/userlib"

	// Life is much easier with json:  You are
	// going to want to use this so you can easily
	// turn complex structures into strings etc...

	"encoding/json"

	// Likewise useful for debugging etc
	"encoding/hex"

	// UUIDs are generated right based on the crypto RNG
	// so lets make life easier and use those too...
	//
	// You need to add with "go get github.com/google/uuid"
	"github.com/google/uuid"
	"github.com/jaydeep/userlib"

	// Useful for debug messages, or string manipulation for datastore keys
	"strings"

	// Want to import errors
	"errors"
)

// This serves two purposes: It shows you some useful primitives and
// it suppresses warnings for items not being imported
func someUsefulThings() {
	// Creates a random UUID
	f := uuid.New()
	userlib.DebugMsg("UUID as string:%v", f.String())

	// Example of writing over a byte of f
	f[0] = 10
	userlib.DebugMsg("UUID as string:%v", f.String())

	// takes a sequence of bytes and renders as hex
	h := hex.EncodeToString([]byte("fubar"))
	userlib.DebugMsg("The hex: %v", h)

	// Marshals data into a JSON representation
	// test
	// Will actually work with go structures as well
	d, _ := json.Marshal(f)
	userlib.DebugMsg("The json data: %v", string(d))
	var g uuid.UUID
	json.Unmarshal(d, &g)
	userlib.DebugMsg("Unmashaled data %v", g.String())

	// This creates an error type
	userlib.DebugMsg("Creation of error %v", errors.New(strings.ToTitle("This is an error")))

	// And a random RSA key.  In this case, ignoring the error
	// return value
	var key *userlib.PrivateKey
	key, _ = userlib.GenerateRSAKey()
	userlib.DebugMsg("Key is %v", key)
}

var configBlockSize = 4096 //Do not modify this variable

//setBlockSize - sets the global variable denoting blocksize to the passed parameter. This will be called only once in the beginning of the execution
func setBlockSize(blocksize int) {
	configBlockSize = blocksize
}

// Helper function: Takes the first 16 bytes and
// converts it into the UUID type
func bytesToUUID(data []byte) (ret uuid.UUID) {
	for x := range ret {
		ret[x] = data[x]
	}
	return
}

//User : User structure used to store the user information
type User struct {
	Username string
	Password string
	PrivKey  *rsa.PrivateKey
	// You can add other fields here if you want...
	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

// StoreFile : function used to create a  file
// It should store the file in blocks only if length
// of data []byte is a multiple of the blocksize; if
// this is not the case, StoreFile should return an error.
func (userdata *User) StoreFile(filename string, data []byte) (err error) {
}

//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need. The length of data []byte must be a multiple of
// the block size; if it is not, AppendFile must return an error.
// AppendFile : Function to append the file
func (userdata *User) AppendFile(filename string, data []byte) (err error) {

}

// LoadFile :This loads a block from a file in the Datastore.
//
// It should give an error if the file block is corrupted in any way.
// If there is no error, it must return exactly one block (of length blocksize)
// of data.
//
// LoadFile is also expected to be efficient. Reading a random block from the
// file should not fetch more than O(1) blocks from the Datastore.
func (userdata *User) LoadFile(filename string, offset int) (data []byte, err error) {
}

// ShareFile : Function used to the share file with other user
func (userdata *User) ShareFile(filename string, recipient string) (msgid string, err error) {
}

// ReceiveFile:Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
// ReceiveFile : function used to receive the file details from the sender
func (userdata *User) ReceiveFile(filename string, sender string, msgid string) error {
}

// RevokeFile : function used revoke the shared file access
func (userdata *User) RevokeFile(filename string) (err error) {
}

// This creates a sharing record, which is a key pointing to something
// in the datastore to share with the recipient.

// This enables the recipient to access the encrypted file as well
// for reading/appending.

// Note that neither the recipient NOR the datastore should gain any
// information about what the sender calls the file.  Only the
// recipient can access the sharing record, and only the recipient
// should be able to know the sender.
// You may want to define what you actually want to pass as a
// sharingRecord to serialized/deserialize in the data store.
type sharingRecord struct {
}

//GenerateUserKey : Returns a Unique Key based on username and password
func GenerateUserKey(string username, string password) (userKey []byte) {
	userKey := userlib.Argon2Key([]byte(password),
		[]byte(username),
		uint32(userlib.BlockSize))
	return userKey
}

// This creates a user.  It will only be called once for a user
// (unless the keystore and datastore are cleared during testing purposes)

// It should store a copy of the userdata, suitably encrypted, in the
// datastore and should store the user's public key in the keystore.

// The datastore may corrupt or completely erase the stored
// information, but nobody outside should be able to get at the stored
// User data: the name used in the datastore should not be guessable
// without also knowing the password and username.
// You are not allowed to use any global storage other than the
// keystore and the datastore functions in the userlib library.

// You can assume the user has a STRONG password

//InitUser : function used to create user
func InitUser(username string, password string) (userdataptr *User, err error) {
	privKey, err := userlib.GenerateRSAKey()
	userdata := new(User)
	userdata.Username = username
	userdata.Password = password
	userdata.PrivKey = privKey

	userlib.KeystoreSet(username, privKey.PublicKey)

	userKey := GenerateUserKey(username, password)
	userdataBytes, err := json.Marshal(userdata)
	if err != nil {
		fmt.Println(err)
		return
	}

	IV, err = json.Marshal(privKey.PublicKey)
	IV = IV[:userlib.BlockSize]

	userNameMac := username + "mac"

	cipherKey := GetEncryptedData(userKey, IV, []byte(username))
	cipherValue := GetEncryptedData(userKey, IV, []byte(userdataBytes))

	macKey := GetEncryptedData(userKey, IV, []byte(userNameMac))
	macValue := GenerateHMAC(userKey, cipherValue)

	userlib.DatastoreSet(string(cipherKey), cipherValue)
	userlib.DatastoreSet(string(macKey), macValue)

	return userdata, nil
}

// GetUser : This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
//GetUser : function used to get the user details
func GetUser(username string, password string) (userdataptr *User, err error) {
	userKey := GenerateUserKey(username, password)
	publicKey, ret := userlib.KeystoreGet(username)
	if !ret {
		return nil, nil
	}

	//Generate IV using Public Key of User & truncate to BlockSize
	IV, err = json.Marshal(publicKey)
	IV = IV[:userlib.BlockSize]

	userNameMac := username + "mac"

	cipherKey := GetEncryptedData(userKey, IV, []byte(username))
	macKey := GetEncryptedData(userKey, IV, []byte(userNameMac))

	macRetValue, ret := userlib.DatastoreGet(string(macKey))
	cipherRetValue, ret := userlib.DatastoreGet(string(cipherKey))

	//MAC compute & check
	macValue := GenerateHMAC(userKey, cipherRetValue)
	if !Equal(macValue, macRetValue) {
		fmt.Println("Data Corrupt!")
		return nil, nil
	}

	dataKey := GetEncryptedData(userKey, IV, []byte(username))
	dataValue, ret := userlib.DatastoreGet(string(dataKey))

	userdataBytes := GetDecryptedData(userKey, IV, dataValue)
	userdata := new(User)
	err = json.Unmarshal(userdataBytes, &userdata)
	if err != nil || userdata.Password != password {
		return nil, err
	}

	return userdata, nil
}
