package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net"
	"os"
	"time"
)

var conn net.Conn //global variable so that we can use it in the various goroutines without passing it.

type Data struct {
	LastChangeTime   int64
	ServerSentTime   int64
	ClientRecvTime   int64
	ClientSendTime   int64
	KeyBaseValid     bool
	KeyBase          float64
	KeyModulus       float64
	SharedColorValid bool
	SharedColor      float64
	Message1Valid    bool
	Message1         [288]byte
	Message2Valid    bool
	Message2         [288]byte
	Message3Valid    bool
	Message3         [288]byte
	Message4Valid    bool
	Message4         [288]byte
}

func main() {
	var arg string                     //Stores IP address as entered by the user.
	connEst := make(chan int)          //use this to synchronize recvTCP() and sendTCP() after connection has been established.
	recvMsg1 := make(chan []byte, 288) //encrypted message length is 256 bytes + 32 bytes encryption
	recvMsg2 := make(chan []byte, 288)
	recvMsg3 := make(chan []byte, 288)
	recvMsg4 := make(chan []byte, 288)
	recvShared := make(chan float64)
	sendMsg1 := make(chan []byte, 288)
	sendMsg2 := make(chan []byte, 288)
	sendMsg3 := make(chan []byte, 288)
	sendMsg4 := make(chan []byte, 288)
	sendBase := make(chan float64)
	sendMod := make(chan float64)
	sendShared := make(chan float64)
	//status := make(chan bool)
	closed := make(chan int)
	passKey1 := make(chan float64)
	passKey2 := make(chan float64)
	passKey3 := make(chan float64)
	passKey4 := make(chan float64)
	serverSentTime := make(chan int64)
	clientRecvTime := make(chan int64)
	lastChangeTime := make(chan int64)

	var err error // define variable separately since we defined conn as a global variable and so we can't do it shorthand.
	//  p :=  make([]byte, 2048)

	go recvTCP()
	go sendTCP()
	go getKey()
	//  go chanBufferKey()
	//  go chanBufferTime()
	go decrypt1()
	go decrypt2()
	go decrypt3()
	go decrypt4()

	//Connect over 6LoWPAN - Client code copied from lab 2
	arg = os.Args[1]           // read IPv6 address from commandline
	arg = "[" + arg + "]:1024" // add port to IPv6 address
	//  fmt.Println(arg) //debugging
	conn, err = net.Dial("tcp6", arg)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	ticker := time.NewTicker(time.Second)
	close(connEst) //closing this channel will allow the other Go routines to progress.

	<-closed
	//use channels for synchronization? Need to wait forever.
} //End of Main function

func getKey() {
	var t int64       //will be used to get time in nanoseconds
	var modulus int64 //modulus for diffie hellman computation
	var base int64    //base for diffie hellman computation
	var a int64       //alice's (my) secret number
	var modulusBig *big.Int
	var baseBig *big.Int
	var aBig *big.Int
	var modulusFlt float64
	var baseFlt float64
	var aFlt float64
	var sharedA float64
	var sharedB float64
	var key float64
	var err error
	keyByte := make([]byte, 32)

	for { //this is a while(1) loop - repeat forever.
		//t = time.Now()UnixNano()
		modulusBig, err = rand.Prime(rand.Reader, 63) //use 53 for clean conversion into positive int64
		if err != nil {
			panic(err)
		}
		modulus = modulusBig.Int64() //Make sure p is of the type Int64
		modulusFlt = float64(modulus)

		baseBig, err = rand.Int(rand.Reader, big.NewInt(modulus)) //pick random base between 0 and modulus
		if err != nil {
			panic(err)
		}
		base = baseBig.Int64()
		baseFlt = float64(base)

		sendMod <- modulusFlt //send modulus to sendTCP via sendMod channel.
		sendBase <- baseFlt   //send base to sendTCP via sendBase channel.

		aBig, err = rand.Int(rand.Reader, big.NewInt(16)) // pick up to 16 to avoid overflow
		if err != nil {
			panic(err)
		}
		a = aBig.Int64()
		aFlt = float64(a)

		sharedA = math.Pow(baseFlt, aFlt)
		sharedA = math.Mod(sharedA, modulusFlt)

		sendShared <- sharedA //send shared value to sendTCP via sendShared channel.
		sharedB <- recvShared //receive sharedB velue from recvTCP via recvShared channel.

		sharedB = math.Pow(sharedB, aFlt)
		key = math.Mod(sharedB, modulusFlt)
		//format key for use in encrypt/decrypt - make it []byte
		binary.LittleEndian.PutUint64(keyByte, uint64(key))

		<-ticker.C //- allow this to update every second, per the ticker schedule/
		t = time.Now().UnixNano()
		lastChangeTime <- t
		//implement key by sending it to decrypt at the appropriate time, which will stall creation of new key
		passKey1 <- keyByte
		passKey2 <- keyByte
		passKey3 <- keyByte
		passKey4 <- keyByte

	} //end of for/while(1) loop
} //end func GetKey

//Copy this four times, adjust channels as necessary.
func decrypt1() {
	key := make([]byte, 32)
	dataRaw := make([]byte, 288) //256 + block size of 32
	data := make([]byte, 288)
	iv := make([]byte, 32)    //iv is the length of the aes encryption block. We are using 256 bit or 32 byte.
	text := make([]byte, 256) //this is the length of our message, specified by our protocol to be 256 bytes.

	//run loop once to initialize all of the variables, remove shorthand definitions in the main code.
	dataRaw = <-recvMsg1 //receive encrypted message 1 from recvTCP via recvMsg1 channel.

	//update key if available, bypass if no new key
	select {
	case key = <-passKey1: //receive key from getKey() via passkey channel.
	default:
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	/*    if len(data) < aes.BlockSize {
	  return nil, errors.New("ciphertext too short")
	}
	*/
	iv = dataRaw[:aes.BlockSize]   //block size is 32 bytes
	text = dataRaw[aes.BlockSize:] //block size is 32 bytes
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err = base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}

	sendMsg1 <- data

	for { // run this loop forever

		data = <-recvMsg1 //receive encrypted message 1 from recvTCP via recvMsg1 channel.

		//update key if available, bypass if no new key
		select {
		case key = <-passKey1: //receive key from getKey() via passkey channel.
		default:
		}

		block, err = aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		/*    if len(data) < aes.BlockSize {
			  return nil, errors.New("ciphertext too short")
		  }
		*/
		iv = dataRaw[:aes.BlockSize]   //block size is 32 bytes
		text = dataRaw[aes.BlockSize:] //block size is 32 bytes
		cfb = cipher.NewCFBDecrypter(text, iv)
		cfb.XORKeyStream(text, text)
		data, err = base64.StdEncoding.DecodeString(string(text))
		if err != nil {
			return nil, err
		}

		sendMsg1 <- data

	} //end of for/while(1) loop

} //end of decryption function

//Copy this four times, adjust channels as necessary.
func decrypt2() {
	key := make([]byte, 32)
	dataRaw := make([]byte, 288) //256 + block size of 32
	data := make([]byte, 288)
	iv := make([]byte, 32)    //iv is the length of the aes encryption block. We are using 256 bit or 32 byte.
	text := make([]byte, 256) //this is the length of our message, specified by our protocol to be 256 bytes.

	//run loop once to initialize all of the variables, remove shorthand definitions in the main code.
	dataRaw = <-recvMsg2 //receive encrypted message 1 from recvTCP via recvMsg1 channel.

	//update key if available, bypass if no new key
	select {
	case key = <-passKey2: //receive key from getKey() via passkey channel.
	default:
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	/*    if len(data) < aes.BlockSize {
	  return nil, errors.New("ciphertext too short")
	}
	*/
	iv = dataRaw[:aes.BlockSize]   //block size is 32 bytes
	text = dataRaw[aes.BlockSize:] //block size is 32 bytes
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err = base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}

	sendMsg2 <- data

	for { // run this loop forever

		data = <-recvMsg2 //receive encrypted message 1 from recvTCP via recvMsg1 channel.

		//update key if available, bypass if no new key
		select {
		case key = <-passKey2: //receive key from getKey() via passkey channel.
		default:
		}

		block, err = aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		/*    if len(data) < aes.BlockSize {
			  return nil, errors.New("ciphertext too short")
		  }
		*/
		iv = dataRaw[:aes.BlockSize]   //block size is 32 bytes
		text = dataRaw[aes.BlockSize:] //block size is 32 bytes
		cfb = cipher.NewCFBDecrypter(text, iv)
		cfb.XORKeyStream(text, text)
		data, err = base64.StdEncoding.DecodeString(string(text))
		if err != nil {
			return nil, err
		}

		sendMsg2 <- data

	} //end of for/while(1) loop

} //end of decryption function

//Copy this four times, adjust channels as necessary.
func decrypt3() {
	key := make([]byte, 32)
	dataRaw := make([]byte, 288) //256 + block size of 32
	data := make([]byte, 288)
	iv := make([]byte, 32)    //iv is the length of the aes encryption block. We are using 256 bit or 32 byte.
	text := make([]byte, 256) //this is the length of our message, specified by our protocol to be 256 bytes.

	//run loop once to initialize all of the variables, remove shorthand definitions in the main code.
	dataRaw = <-recvMsg3 //receive encrypted message 1 from recvTCP via recvMsg1 channel.

	//update key if available, bypass if no new key
	select {
	case key = <-passKey3: //receive key from getKey() via passkey channel.
	default:
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	/*    if len(data) < aes.BlockSize {
	  return nil, errors.New("ciphertext too short")
	}
	*/
	iv = dataRaw[:aes.BlockSize]   //block size is 32 bytes
	text = dataRaw[aes.BlockSize:] //block size is 32 bytes
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err = base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}

	sendMsg3 <- data

	for { // run this loop forever

		data = <-recvMsg3 //receive encrypted message 1 from recvTCP via recvMsg1 channel.

		//update key if available, bypass if no new key
		select {
		case key = <-passKey3: //receive key from getKey() via passkey channel.
		default:
		}

		block, err = aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		/*    if len(data) < aes.BlockSize {
			  return nil, errors.New("ciphertext too short")
		  }
		*/
		iv = dataRaw[:aes.BlockSize]   //block size is 32 bytes
		text = dataRaw[aes.BlockSize:] //block size is 32 bytes
		cfb = cipher.NewCFBDecrypter(text, iv)
		cfb.XORKeyStream(text, text)
		data, err = base64.StdEncoding.DecodeString(string(text))
		if err != nil {
			return nil, err
		}

		sendMsg3 <- data

	} //end of for/while(1) loop

} //end of decryption function

//Copy this four times, adjust channels as necessary.
func decrypt4() {
	key := make([]byte, 32)
	dataRaw := make([]byte, 288) //256 + block size of 32
	data := make([]byte, 288)
	iv := make([]byte, 32)    //iv is the length of the aes encryption block. We are using 256 bit or 32 byte.
	text := make([]byte, 256) //this is the length of our message, specified by our protocol to be 256 bytes.

	//run loop once to initialize all of the variables, remove shorthand definitions in the main code.
	dataRaw = <-recvMsg4 //receive encrypted message 1 from recvTCP via recvMsg1 channel.

	//update key if available, bypass if no new key
	select {
	case key = <-passKey4: //receive key from getKey() via passkey channel.
	default:
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	/*    if len(data) < aes.BlockSize {
	  return nil, errors.New("ciphertext too short")
	}
	*/
	iv = dataRaw[:aes.BlockSize]   //block size is 32 bytes
	text = dataRaw[aes.BlockSize:] //block size is 32 bytes
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err = base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}

	sendMsg4 <- data

	for { // run this loop forever

		data = <-recvMsg4 //receive encrypted message 1 from recvTCP via recvMsg1 channel.

		//update key if available, bypass if no new key
		select {
		case key = <-passKey4: //receive key from getKey() via passkey channel.
		default:
		}

		block, err = aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		/*    if len(data) < aes.BlockSize {
			  return nil, errors.New("ciphertext too short")
		  }
		*/
		iv = dataRaw[:aes.BlockSize]   //block size is 32 bytes
		text = dataRaw[aes.BlockSize:] //block size is 32 bytes
		cfb = cipher.NewCFBDecrypter(text, iv)
		cfb.XORKeyStream(text, text)
		data, err = base64.StdEncoding.DecodeString(string(text))
		if err != nil {
			return nil, err
		}

		sendMsg4 <- data

	} //end of for/while(1) loop

} //end of decryption function

func recvTCP() {
	data := make([]btye, 2048) //set up data variable so that we can read from TCP
	var ClientRecvTime int64   //set up variable to read system time in nanoseconds
	var i int
	var allData Data //this is data formatted back into a struct.
	var err error
	var delay int64

	<-connEst //Channel used to synchronize the Go routines so that we can establish a connection before trying to send data.

	for { //execute this code forever.

		//read and write over the network
		_, err = bufio.NewReader(conn).Read(data)
		if err != nil {
			fmt.Printf("Some error %v\n", err)
		}
		ClientRecvTime = time.Now().UnixNano() //getTime()

		//decode data into the struct that it should be.
		err = json.Unmarshal(data, &allData)

		//do the appropriate things with the data.

		//this device only sends the key base and modulus. we don't do anything with the read data.

		if allData.SharedColorValid == true {
			sendShared <- allData.SharedColor
		}

		if allData.Message1Valid == true {
			//send data to be decrypted via gochannel
			recvMsg1 <- allData.Message1
			//don't worry about resetting status because it will be reset with new data packet.
		}

		if allData.Message2Valid == true {
			//send data to be decrypted via gochannel
			recvMsg2 <- allData.Message2
		}

		if allData.Message3Valid == true {
			//send data to be decrypted via gochannel
			recvMsg3 <- allData.Message3
		}

		if allData.Message4Valid == true {
			//send data to be decrypted via gochannel
			recvMsg4 <- allData.Message4
		}

	} // end of for/while(1) loop
	close(closed) //when this terminates, so will the main program.
} //end of func recvTCP

func sendTCP() {

	data := make([]byte, 2048)
	var dataRaw Data
	var err error

	<-connEst //Synchronization - wait until after the connection has been established

	//put together a package of data.
	dataRaw.KeyBaseValid = true
	dataRaw.KeyBase = <-sendBase
	dataRaw.KeyModulus = <-sendMod
	dataRaw.SharedColorValid = true
	dataRaw.SharedColor = <-sendShared
	ClientSendTime := time.Now().UnixNano() //getTime()
	data, err = json.Marshal(dataRaw)
	//set struct back to zero
	dataRaw.KeyBaseValid = false
	dataRaw.SharedColorValid = false

	_, err = conn.Write(data)
	if err != nil {
		panic(err)
	}

	for {
		//get timing info from latest packet.
		dataRaw.ServerSentTime <- serverSentTime
		dataRaw.ClientRecvTime <- clientRecvTime
		//gather all of your data.

		select {
		case dataRaw.LastChangeTime = <-lastChangeTime:
		default:
		}

		select {
		case dataRaw.KeyBase = <-sendBase:
			dataRaw.KeyBaseValid = true
		default:
		}

		select {
		case dataRaw.KeyModulus = <-sendMod:
			dataRaw.KeyBaseValid = true
		default:
		}

		select {
		case dataRaw.SharedColor = <-sendShared:
			dataRaw.SharedColorValid = true
		default:
		}

		select {
		case dataRaw.Message1 = <-sendMsg1:
			dataRaw.Message1Valid = true
		default:
		}

		select {
		case dataRaw.Message2 = <-sendMsg2:
			dataRaw.Message2Valid = true
		default:
		}

		select {
		case dataRaw.Message3 = <-sendMsg3:
			dataRaw.Message3Valid = true
		default:
		}

		select {
		case dataRaw.Message4 = <-sendMsg4:
			dataRaw.Message4Valid = true
		default:
		}

		ClientSendTime = time.Now().UnixNano() //getTime()
		data, err = json.Marshal(dataRaw)      //package data up
		//set struct back to zero so that we have an accurate idea of what data is actually being sent.
		dataRaw.KeyBaseValid = false
		dataRaw.SharedColorValid = false
		dataRaw.Message1Valid = false
		dataRaw.Message2Valid = false
		dataRaw.Message3Valid = false
		dataRaw.Message4Valid = false

		_, err = conn.Write(data) //send data to server.
		if err != nil {
			panic(err)
		}

	} //end while loop

} //end of func sendTCP

/*

http://stackoverflow.com/questions/18817336/golang-encrypting-a-string-with-aes-and-base64
http://stackoverflow.com/questions/17260107/int16-to-byte-array
https://systembash.com/a-simple-go-tcp-server-and-tcp-client/
https://gobyexample.com/non-blocking-channel-operations


*/
