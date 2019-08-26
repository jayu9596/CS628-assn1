package assn1

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.
//
import (

	// You neet to add with
	// go get github.com/sarkarbidya/CS628-assn1/userlib

	"crypto/rsa"
	"strconv"

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

//FileRecord : File Record structure used to store the file information
type FileRecord struct {
	Name  string
	Size  int
	Owner string
}

//ReverseBytes : Reverse a byte array
func ReverseBytes(key []byte) []byte {
	arr := make([]byte, len(key))
	copy(arr, key)
	for i := len(arr)/2 - 1; i >= 0; i-- {
		opp := len(arr) - 1 - i
		arr[i], arr[opp] = arr[opp], arr[i]
	}
	return arr
}

//GenerateUserKey : Returns a Unique Key based on username and password
func (userdata *User) GenerateUserKey() []byte {
	userKey := userlib.Argon2Key([]byte(userdata.Password), []byte(userdata.Username), uint32(userlib.BlockSize))
	return userKey
}

//GenerateFileKey : Returns a Unique Key based on username and password
func (userdata *User) GenerateFileKey(filename string) []byte {
	userKey := userdata.GenerateUserKey()
	salt := userlib.RandomBytes(userlib.BlockSize)
	fileKey := userlib.Argon2Key(userKey, salt, uint32(userlib.BlockSize))
	return fileKey
}

//GetFileKey : Retrieve File key after checking Integrity
func (userdata *User) GetFileKey(filename string) ([]byte, error) {
	userKey := userdata.GenerateUserKey()
	fileNameKey := filename + "key"
	fileNameMac := filename + "mac"
	pubKey, ret := userlib.KeystoreGet(userdata.Username)
	if !ret {
		return nil, errors.New("Error while retriving key")
	}
	IV, err := json.Marshal(pubKey)
	if err != nil {
		return nil, errors.New("Error while marshaling")
	}
	IV = IV[:userlib.BlockSize]

	cipherKey := GetEncryptedData(userKey, IV, []byte(fileNameKey))
	macKey := GetEncryptedData(userKey, IV, []byte(fileNameMac))
	macRetValue, ret := userlib.DatastoreGet(string(macKey))
	cipherRetValue, ret := userlib.DatastoreGet(string(cipherKey))

	macValue := GenerateHMAC(userKey, cipherRetValue)
	if !userlib.Equal(macValue, macRetValue) {
		return nil, errors.New("Data Corrupt")
	}

	cipherRetValue = GetDecryptedData(userKey, IV, cipherRetValue)
	return cipherRetValue, nil
}

//GetFile : Retrieve File DataStructure after checking Integrity
func (userdata *User) GetFile(filename string, fileKey []byte) (*FileRecord, error) {
	IV := ReverseBytes(fileKey)
	fileRecord := filename + "Record"
	fileRecordMac := filename + "RecordMac"
	cipherKey := GetEncryptedData(fileKey, IV, []byte(fileRecord))
	macKey := GetEncryptedData(fileKey, IV, []byte(fileRecordMac))

	macRetValue, ret := userlib.DatastoreGet(string(macKey))
	if !ret {
		return nil, errors.New("Data Corrupt")
	}

	cipherRetValue, ret := userlib.DatastoreGet(string(cipherKey))
	if !ret {
		return nil, errors.New("Data Corrupt")
	}

	macValue := GenerateHMAC(fileKey, cipherRetValue)
	if !userlib.Equal(macValue, macRetValue) {
		return nil, errors.New("Data Corrupt")
	}

	filedataBytes := GetDecryptedData(fileKey, IV, cipherRetValue)
	filedata := new(FileRecord)
	err := json.Unmarshal(filedataBytes, &filedata)
	if err != nil {
		return nil, errors.New("Error in UnMarshaling")
	}

	return filedata, nil
}

//GenerateHMAC : Returns hash value using key passed as parameter
func GenerateHMAC(Key []byte, cipherText []byte) []byte {
	mac := userlib.NewHMAC(Key)
	mac.Write(cipherText)
	macValue := mac.Sum(nil)
	return macValue
}

//GetEncryptedData : Returns encrypted value using key passed as parameter
func GetEncryptedData(Key []byte, IV []byte, cipherText []byte) []byte {
	encryptedCipherText := make([]byte, len(cipherText))
	cipherEncStream := userlib.CFBEncrypter(Key, IV)
	cipherEncStream.XORKeyStream(encryptedCipherText, cipherText)
	return encryptedCipherText
}

//GetDecryptedData : Returns decrypted value using key passed as parameter
func GetDecryptedData(Key []byte, IV []byte, cipherText []byte) []byte {
	encryptedCipherText := make([]byte, len(cipherText))
	cipherEncStream := userlib.CFBDecrypter(Key, IV)
	cipherEncStream.XORKeyStream(encryptedCipherText, cipherText)
	return encryptedCipherText
}

// StoreFile : function used to create a  file
// It should store the file in blocks only if length
// of data []byte is a multiple of the blocksize; if
// this is not the case, StoreFile should return an error.
func (userdata *User) StoreFile(filename string, data []byte) (err error) {
	if len(data)%configBlockSize != 0 && len(data) > 0 {
		return errors.New("File size is not a mulltiple of Blocksize")
	}

	fileKey := userdata.GenerateFileKey(filename)
	filedata := new(FileRecord)
	filedata.Name = filename
	filedata.Size = 0
	filedata.Owner = userdata.Username

	//Generate IV using Public Key of User & truncate to BlockSize
	publicKey, ret := userlib.KeystoreGet(userdata.Username)
	if !ret {
		return errors.New("Error in retriving public key")
	}
	IV, err := json.Marshal(publicKey)
	if err != nil {
		return err
	}
	IV = IV[:userlib.BlockSize]

	//Store Filekey for Owner on DataStore
	fileNameKey := filedata.Name + "key"
	fileNameMac := filedata.Name + "mac"
	userKey := userdata.GenerateUserKey()
	cipherKey := GetEncryptedData(userKey, IV, []byte(fileNameKey))
	macKey := GetEncryptedData(userKey, IV, []byte(fileNameMac))

	cipherValue := GetEncryptedData(userKey, IV, fileKey)
	macValue := GenerateHMAC(userKey, cipherValue)

	userlib.DatastoreSet(string(cipherKey), cipherValue)
	userlib.DatastoreSet(string(macKey), macValue)

	//Store FileRecord of a file on DataStore
	IV = ReverseBytes(fileKey)
	filedataBytes, err := json.Marshal(filedata)
	if err != nil {
		return errors.New("Error in Marshaling")
	}
	fileRecord := filedata.Name + "Record"
	fileRecordMac := filedata.Name + "RecordMac"
	cipherKey = GetEncryptedData(fileKey, IV, []byte(fileRecord))
	macKey = GetEncryptedData(fileKey, IV, []byte(fileRecordMac))

	cipherValue = GetEncryptedData(fileKey, IV, filedataBytes)
	macValue = GenerateHMAC(fileKey, cipherValue)

	userlib.DatastoreSet(string(cipherKey), cipherValue)
	userlib.DatastoreSet(string(macKey), macValue)

	return userdata.AppendFile(filedata.Name, data)
}

//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need. The length of data []byte must be a multiple of
// the block size; if it is not, AppendFile must return an error.
// AppendFile : Function to append the file
func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	if len(data)%configBlockSize != 0 && len(data) > 0 {
		return errors.New("File size is not a mulltiple of Blocksize")
	}

	fileKey, err := userdata.GetFileKey(filename)
	if err != nil {
		return errors.New("file does not exist")
	}

	filedata, err := userdata.GetFile(filename, fileKey)
	if err != nil {
		return errors.New("file does not exist / Corrupt Data")
	}

	fileKeyString := string(fileKey)
	IV := ReverseBytes(fileKey)
	length := len(data) / configBlockSize
	offset := filedata.Size

	for i := 0; offset < filedata.Size+length; offset, i = offset+1, i+1 {
		//Store filedata offset wise
		filedataKey := fileKeyString + strconv.Itoa(offset)
		cipherKey := GetEncryptedData(fileKey, IV, []byte(filedataKey))
		cipherValue := GetEncryptedData(fileKey, IV, data[i*configBlockSize:(i+1)*configBlockSize])
		userlib.DatastoreSet(string(cipherKey), cipherValue)

		//Store filedata mac
		filedataMac := filedataKey + "mac"
		macValue := GenerateHMAC(fileKey, cipherValue)
		macKey := GetEncryptedData(fileKey, IV, []byte(filedataMac))
		userlib.DatastoreSet(string(macKey), macValue)
	}

	//Update File data structure
	filedata.Size = offset
	filedataBytes, err := json.Marshal(filedata)
	if err != nil {
		return errors.New("Error in Marshaling")
	}
	fileRecord := filedata.Name + "Record"
	fileRecordMac := filedata.Name + "RecordMac"
	cipherKey := GetEncryptedData(fileKey, IV, []byte(fileRecord))
	macKey := GetEncryptedData(fileKey, IV, []byte(fileRecordMac))

	cipherValue := GetEncryptedData(fileKey, IV, filedataBytes)
	macValue := GenerateHMAC(fileKey, cipherValue)

	userlib.DatastoreSet(string(cipherKey), cipherValue)
	userlib.DatastoreSet(string(macKey), macValue)

	return nil
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
	fileKey, err := userdata.GetFileKey(filename)
	if err != nil {
		return nil, errors.New("file does not exist")
	}
	filedata, err := userdata.GetFile(filename, fileKey)
	if err != nil {
		return nil, errors.New("file does not exist / Data Corrupt")
	}
	if filedata.Size >= offset {
		return nil, errors.New("Offset more than filesize")
	}
	IV := ReverseBytes(fileKey)
	fileKeyString := string(fileKey)
	filedataKey := fileKeyString + strconv.Itoa(offset)
	cipherKey := GetEncryptedData(fileKey, IV, []byte(filedataKey))
	cipherRetValue, ret := userlib.DatastoreGet(string(cipherKey))
	if !ret {
		return nil, errors.New("Error in retriving filedata at some offset")
	}
	filedataMac := filedataKey + "mac"
	macKey := GetEncryptedData(fileKey, IV, []byte(filedataMac))
	macRetValue, ret := userlib.DatastoreGet(string(macKey))
	if !ret {
		return nil, errors.New("Error in retriving mac of filedata at some offset")
	}
	macValue := GenerateHMAC(fileKey, cipherRetValue)
	if !userlib.Equal(macValue, macRetValue) {
		return nil, errors.New("Data Corrupt")
	}
	return GetDecryptedData(fileKey, IV, cipherRetValue), nil
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
func GenerateUserKey(username string, password string) []byte {
	userKey := userlib.Argon2Key([]byte(password), []byte(username), uint32(userlib.BlockSize))
	return userKey
}

//GenerateHMAC : Returns hash value using key passed as parameter
func GenerateHMAC(Key []byte, cipherText []byte) []byte {
	mac := userlib.NewHMAC(Key)
	mac.Write(cipherText)
	macValue := mac.Sum(nil)
	return macValue
}

//GetEncryptedData : Returns encrypted value using key passed as parameter
func GetEncryptedData(Key []byte, IV []byte, cipherText []byte) []byte {
	encryptedCipherText := make([]byte, len(cipherText))
	cipherEncStream := userlib.CFBEncrypter(Key, IV)
	cipherEncStream.XORKeyStream(encryptedCipherText, cipherText)
	return encryptedCipherText
}

//GetDecryptedData : Returns decrypted value using key passed as parameter
func GetDecryptedData(Key []byte, IV []byte, cipherText []byte) []byte {
	encryptedCipherText := make([]byte, len(cipherText))
	cipherEncStream := userlib.CFBDecrypter(Key, IV)
	cipherEncStream.XORKeyStream(encryptedCipherText, cipherText)
	return encryptedCipherText
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

	userKey := userdata.GenerateUserKey()
	userdataBytes, err := json.Marshal(userdata)
	if err != nil {
		return nil, err
	}

	IV, err := json.Marshal(privKey.PublicKey)
	if err != nil {
		return nil, err
	}
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
	userdata := new(User)
	userdata.Username = username
	userdata.Password = password
	userKey := userdata.GenerateUserKey()
	publicKey, ret := userlib.KeystoreGet(username)
	if !ret {
		return nil, nil
	}

	//Generate IV using Public Key of User & truncate to BlockSize
	IV, err := json.Marshal(publicKey)
	if err != nil {
		return nil, err
	}
	IV = IV[:userlib.BlockSize]

	userNameMac := username + "mac"

	cipherKey := GetEncryptedData(userKey, IV, []byte(username))
	macKey := GetEncryptedData(userKey, IV, []byte(userNameMac))

	macRetValue, ret := userlib.DatastoreGet(string(macKey))
	if !ret {
		return nil, err
	}

	cipherRetValue, ret := userlib.DatastoreGet(string(cipherKey))
	if !ret {
		return nil, err
	}

	//MAC compute & check
	macValue := GenerateHMAC(userKey, cipherRetValue)
	if !userlib.Equal(macValue, macRetValue) {
		return nil, errors.New("Data Corrupt")
	}

	dataKey := GetEncryptedData(userKey, IV, []byte(username))
	dataValue, ret := userlib.DatastoreGet(string(dataKey))
	if !ret {
		return nil, err
	}

	userdataBytes := GetDecryptedData(userKey, IV, dataValue)
	err = json.Unmarshal(userdataBytes, &userdata)
	if err != nil || userdata.Password != password {
		return nil, err
	}

	return userdata, nil
}
