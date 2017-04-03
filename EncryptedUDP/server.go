package main

import (
  "fmt"
  "time"
  "math"
  "math/big"
  "crypto/rand"
  "encoding/json"
  "net"
  "bufio"
  "strings"
)


type Data struct {
  LastChangeTime int64
  ServerSentTime int64
  ClientRecvTime int64
  ClientSendTime int64
  KeyBaseValid bool
  KeyBase float64
  KeyModulus float64
  SharedColorValid bool
  SharedColor float64
  Message1Valid bool
  Message1 [288]byte
  Message2Valid bool
  Message2 [288]byte
  Message3Valid bool
  Message3 [288]byte
  Message4Valid bool
  Message4 [288]byte
}

var conn net.Conn//global variable so that we can use it in the various goroutines without passing it.

func main() {
  // listen on all interfaces
var err error
connEst := make(chan int) //use this to synchronize recvTCP() and sendTCP() after connection has been established.
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
serverRecvTime := make(chan int64)


ln, _ := net.Listen("tcp6", ":1024")
// accept connection on port
conn, _ = ln.Accept()

  go recvTCP()
  go sendTCP()
  go getKey()
  go calculateKeySwapTime()
  //go chanBuffer //KeySwap time
  //go chanBuffer //encryption key
  go encrypt1() //doing this the brute force way rather than trying to return values and organize. This will allow us to add several smaller keys together.
  go encrypt2()
  go encrypt3()
  go encrypt4()
//  go resultChecker1()
//  go resultChecker2()
//  go resultChecker3()
//  go resultChecker4()
  //Wait Forever - defer?
close(connEst) //closing this channel will allow the other Go routines to progress.

<- closed

} //End of function Main




func recvTCP() {
  data := make([]btye, 2048) //set up data variable so that we can read from TCP
  var recvTime int64 //set up variable to read system time in nanoseconds
  var i int
  var allData Data //this is data formatted back into a struct.
  var err error
  var delay int64


  <- connEst //Channel used to synchronize the Go routines so that we can establish a connection before trying to send data.

  for { //execute this code forever.

    //read and write over the network
    _, err = bufio.NewReader(conn).Read(data)
    if err != nil {
      fmt.Printf("Some error %v\n", err)
    }
    recvTime = time.Now()UnixNano()//getTime()


    //decode data into the struct that it should be.
    err = json.Unmarshal(data, &allData)

    //do the appropriate things with the data.
    serverSentTime <- allData.ServerSentTime
    clientRecvTime <- allData.ClientRecvTime
    lastChangeTime <- allData.LastChangeTime
    serverRecvTime <- recvTime


          if allData.SharedColorValid == true {
            sendShared <- allData.SharedColor
          }

          if allData.Message1Valid == true {
            //send data to be decrypted via gochannel
            sendMsg1 <- allData.Message1
            //don't worry about resetting status because it will be reset with new data packet.
          }

          if allData.Message2Valid == true {
            //send data to be decrypted via gochannel
            sendMsg2 <- allData.Message2
          }

          if allData.Message3Valid == true {
            //send data to be decrypted via gochannel
            sendMsg3 <- allData.Message3
          }

          if allData.Message4Valid == true {
            //send data to be decrypted via gochannel
            sendMsg4 <- allData.Message4
          }


    }




func sendTCP() {
  data := make([]byte, 2048)
  var dataRaw Data
  var err error
  serverTime

  <- connEst //Synchronization - wait until after the connection has been established

  for {

  select {
  case: dataRaw.KeyBase = <- sendBase
    dataRaw.KeyBaseValid = true
  default:
  }

  select {
  case: dataRaw.KeyModulus = <- sendMod
    dataRaw.KeyBaseValid = true
  default:
  }

  select {
  case: dataRaw.SharedColor = <- sendShared
    dataRaw.SharedColorValid = true
  default:
  }

  select {
  case: dataRaw.Message1 = <- sendMsg1
    dataRaw.Message1Valid = true
  default:
  }

  select {
  case: dataRaw.Message2 = <- sendMsg2
    dataRaw.Message2Valid = true
  default:
  }

  select {
  case: dataRaw.Message3 = <- sendMsg3
    dataRaw.Message3Valid = true
  default:
  }

  select {
  case: dataRaw.Message4 = <- sendMsg4
    dataRaw.Message4Valid = true
  default:
  }

  serverTime = time.Now()UnixNano()//getTime()
  data, err = json.Marshal(dataRaw) //package data up
  //set struct back to zero so that we have an accurate idea of what data is actually being sent.
  dataRaw.KeyBaseValid = false
  dataRaw.SharedColorValid = false
  dataRaw.Message1Valid = false
  dataRaw.Message2Valid = false
  dataRaw.Message3Valid = false
  dataRaw.Message4Valid = false

  _,err = conn.Write(data) //send data to server.
  if err != nil {
    panic(err)
  }



}//end while loop




}




//This function allows us to get a key using the Diffie Hellman Key Exchanhge method.
func getKey() {
  var t int64
  var a int64
  var aBig *big.Int
  var aFlt float64
  var sharedA float64
  var sharedB float64
  var key float64
  var err error
  keyByte := make([]byte, 32)
  var baseFlt float64
  var modFlt float64



  for {
    baseFlt <- sendBase
    modFlt <- sendMod
    aBig, err = rand.Int(rand.Reader, big.NewInt(16)) //pick up to 16 to avoid overflow.
    if err != nil {
      panic(err)
    }
    aFlt = float64(a)
    sharedA = math.Pow(baseFlt, aFlt)
    sharedA = math.Mod(sharedA, modFlt)
    sendShared <- sharedA
    sharedB <- recvShared

    sharedB = math.Pow(sharedB, aFlt)
    key = math.Mod(sharedB, modulusFlt)
    //format key for use in encrypt/decrypt - make it []byte
    binary.LittleEndian.PutUint64(keyByte, uint64(key))
/*
    <-ticker.C //- allow this to update every second, per the ticker schedule/
    t = time.Now()UnixNano()

    passKey1 <- keyByte
    passKey2 <- keyByte
    passKey3 <- keyByte
    passKey4 <- keyByte
*/

  } //end For loop
} //end getKey()



func encrypt1() {
  data := make([]byte, 256)
  returnedLong := make([]byte, 288)
  returned := make([]byte, 256)
  ciphertext := make([]byte, 288)
  key := make([]byte, 32)
  rand.Read(data)
  key <- passKey1 //we won't be able to start until we get a key.

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    b := base64.StdEncoding.EncodeToString(data)
  //  ciphertext := make([]byte, aes.BlockSize+len(b))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }
    cfb := cipher.NewCFBEncrypter(block, iv)
    cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

    sendMsg1 <- ciphertext
    returnedLong = <- recvMsg1
    returned = returnedLong[256:]

    if bytes.Equal(data,returned) == false {
      fmt.Printf("Missed Message")
    }

    for {
      rand.Read(data)
      select{
      case: key <- passKey1
      default:
      }

      block, err = aes.NewCipher(key)
      if err != nil {
          return nil, err
      }
      b = base64.StdEncoding.EncodeToString(data)
    //  ciphertext := make([]byte, aes.BlockSize+len(b))
      iv = ciphertext[:aes.BlockSize]
      if _, err := io.ReadFull(rand.Reader, iv); err != nil {
          return nil, err
      }
      cfb = cipher.NewCFBEncrypter(block, iv)
      cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

      sendMsg1 <- ciphertext
      returnedLong = <- recvMsg1
      returned = returnedLong[256:]

      if bytes.Equal(data,returned) == false {
        fmt.Printf("Missed Message")
      }


    }//end of while loop

}//end of encryption block.





func encrypt2() {
  data := make([]byte, 256)
  returnedLong := make([]byte, 288)
  returned := make([]byte, 256)
  ciphertext := make([]byte, 288)
  key := make([]byte, 32)
  rand.Read(data)
  key <- passKey2 //we won't be able to start until we get a key.

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    b := base64.StdEncoding.EncodeToString(data)
  //  ciphertext := make([]byte, aes.BlockSize+len(b))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }
    cfb := cipher.NewCFBEncrypter(block, iv)
    cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

    sendMsg2 <- ciphertext
    returnedLong = <- recvMsg2
    returned = returnedLong[256:]

    if bytes.Equal(data,returned) == false {
      fmt.Printf("Missed Message")
    }

    for {
      rand.Read(data)
      select{
      case: key <- passKey2
      default:
      }

      block, err = aes.NewCipher(key)
      if err != nil {
          return nil, err
      }
      b = base64.StdEncoding.EncodeToString(data)
    //  ciphertext := make([]byte, aes.BlockSize+len(b))
      iv = ciphertext[:aes.BlockSize]
      if _, err := io.ReadFull(rand.Reader, iv); err != nil {
          return nil, err
      }
      cfb = cipher.NewCFBEncrypter(block, iv)
      cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

      sendMsg2 <- ciphertext
      returnedLong = <- recvMsg2
      returned = returnedLong[256:]

      if bytes.Equal(data,returned) == false {
        fmt.Printf("Missed Message")
      }


    }//end of while loop

}//end of encryption block.





func encrypt3() {
  data := make([]byte, 256)
  returnedLong := make([]byte, 288)
  returned := make([]byte, 256)
  ciphertext := make([]byte, 288)
  key := make([]byte, 32)
  rand.Read(data)
  key <- passKey3 //we won't be able to start until we get a key.

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    b := base64.StdEncoding.EncodeToString(data)
  //  ciphertext := make([]byte, aes.BlockSize+len(b))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }
    cfb := cipher.NewCFBEncrypter(block, iv)
    cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

    sendMsg3 <- ciphertext
    returnedLong = <- recvMsg3
    returned = returnedLong[256:]

    if bytes.Equal(data,returned) == false {
      fmt.Printf("Missed Message")
    }

    for {
      rand.Read(data)
      select{
      case: key <- passKey3
      default:
      }

      block, err = aes.NewCipher(key)
      if err != nil {
          return nil, err
      }
      b = base64.StdEncoding.EncodeToString(data)
    //  ciphertext := make([]byte, aes.BlockSize+len(b))
      iv = ciphertext[:aes.BlockSize]
      if _, err = io.ReadFull(rand.Reader, iv); err != nil {
          return nil, err
      }
      cfb = cipher.NewCFBEncrypter(block, iv)
      cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

      sendMsg3 <- ciphertext
      returnedLong = <- recvMsg3
      returned = returnedLong[256:]

      if bytes.Equal(data,returned) == false {
        fmt.Printf("Missed Message")
      }


    }//end of while loop

}//end of encryption block.





func encrypt4() {
  data := make([]byte, 256)
  returnedLong := make([]byte, 288)
  returned := make([]byte, 256)
  ciphertext := make([]byte, 288)
  key := make([]byte, 32)
  rand.Read(data)
  key <- passKey4 //we won't be able to start until we get a key.

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    b := base64.StdEncoding.EncodeToString(data)
  //  ciphertext := make([]byte, aes.BlockSize+len(b))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }
    cfb := cipher.NewCFBEncrypter(block, iv)
    cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

    sendMsg4 <- ciphertext
    returnedLong = <- recvMsg4
    returned = returnedLong[256:]

    if bytes.Equal(data,returned) == false {
      fmt.Printf("Missed Message")
    }

    for {
      rand.Read(data)
      select{
      case: key <- passKey4
      default:
      }

      block, err = aes.NewCipher(key)
      if err != nil {
          return nil, err
      }
      b = base64.StdEncoding.EncodeToString(data)
    //  ciphertext := make([]byte, aes.BlockSize+len(b))
      iv = ciphertext[:aes.BlockSize]
      if _, err = io.ReadFull(rand.Reader, iv); err != nil {
          return nil, err
      }
      cfb = cipher.NewCFBEncrypter(block, iv)
      cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

      sendMsg4 <- ciphertext
      returnedLong = <- recvMsg4
      returned = returnedLong[256:]

      if bytes.Equal(data,returned) == false {
        fmt.Printf("Missed Message")
      }


    }//end of while loop

}//end of encryption block.



/*

http://stackoverflow.com/questions/18817336/golang-encrypting-a-string-with-aes-and-base64
http://stackoverflow.com/questions/17260107/int16-to-byte-array
https://systembash.com/a-simple-go-tcp-server-and-tcp-client/
https://gobyexample.com/non-blocking-channel-operations


*/
