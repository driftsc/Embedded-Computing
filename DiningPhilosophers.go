package main

import (
	"fmt"
	"math/rand"
	"time"
)

/*
This program creates 4 philosophers who alternate between eating an thinking.
They require a fork in both of their hands to eat.
The philosophers have to share the forks because there are only 4 forks,
and only one fork is between any two philosohers.

The philosophers are independent of each other, which is why they are represented
by go routines. The forks are shared resources which is why they are buffered channels.
The buffers are blocking and so only one philosopher can use a fork by putting his
identity into the fork buffer. Fork 0 sits in between philosopher 0 and philosopher 1,
fork 1 is between philosopher 1 and philosopher 2, and so on.

The philosophers will try to obtain the fork with the smaller number first.
This will prevent a deadlock condition.

If a philosopher cannot obtain a fork, he will either wait patiently or else
start thinking again for a brief period of time. After this time, he will attempt
to obtain the fork again. He will therefore ask for the fork more frequently than
a philosopher who has eaten, so he will be more likely to eat, alleviating starvation.

By waiting patiently or thinking for a brief period of time, the philosopher doesn't
engage in busy-waiting.
*/

func main() {
	//create forks and initialize them to be available (status 4).
	//Status 0 - 3 represents the philosopher that is using the fork.
	fork0 := make(chan int, 1)
	fork0 <- 4
	fork1 := make(chan int, 1)
	fork1 <- 4
	fork2 := make(chan int, 1)
	fork2 <- 4
	fork3 := make(chan int, 1)
	fork3 <- 4
	//forks(fork0, fork1, fork2, fork3)
	/*
		  Create a go routine for each philosopher, run each in parallel. Note that
			  philosopher0 uses fork0 and fork3, and sits between philosopher1 and philosopher3
			  philosopher1 uses fork0 and fork1, and sits between philosopher0 and philosopher2
			  philosopher2 uses fork1 and fork2, and sits between philosopher1 and philosopher3
			  philosopher3 uses fork2 and fork3, and sits between philosopher2 and philosopher0

			The order is aranged so that the philosopher picks up the lower number fork first before acquiring the higher number fork.
	*/
	go philosopher(fork0, fork3, "philosopher0", 1, 0, 3)
	go philosopher(fork0, fork1, "philosopher1", 0, 1, 2)
	go philosopher(fork1, fork2, "philosopher2", 1, 2, 3)
	go philosopher(fork2, fork3, "philosopher3", 2, 3, 0)

	//Continually read what each philosopher is doing.
	var input string
	fmt.Scanln(&input)
	fmt.Println(input)

}

//generalize the case so that I only have to write this function once. This is why there are so many input variables specifying the order of forks and philosophers.

func philosopher(fork0 chan int, fork1 chan int, name string, lower int, myNumber int, upper int) {
	//The philosopher is initialized. The for loop will run until the program is terminated.
	//initialize forkStatus1 and 2 variables.
	forkStatus1 := 0
	forkStatus2 := 0

	fmt.Println(name, "joined the table.")

	//The philosopher begins by thinking. The state variable is set to 0 = think, 1 = eat.
	//The previous state is initialized to 1 so that the program prints that the philosopher is thinking.
	//future cases will see if these are different (if the philosopher changes activities) so that it only prints new activities.
	previousState := 1
	state := 0

	for {

		//Logic for when the philosopher is thinking or wants to think.

		if state == 0 {
			//If the philosopher was just eating, then print that he is now thinking.
			if state != previousState {
				fmt.Println(name, "is thinking.")
				//update previous state so that the program knows that the philosopher is already thinking.
			}
			//Wait for the philosopher to finish thinking (5<t<10 seconds)
			time.Sleep(time.Millisecond * time.Duration(5000+rand.Intn(5001)))

			//This is the case when the philosopher eats again after he just ate (previous state = 1 and current state = 1)
		} else if previousState == 1 {
			time.Sleep(time.Millisecond * time.Duration(1000+rand.Intn(1001)))

		} else {

			/*
			   This is the logic for when the philosopher wants to eat.
			   collect fork status. Start with the lower number fork to prevent deadlock.
			   Fork0 is shared with lower Philosopher, so status of fork will either be lower (other guy is using it), myNumber, or 4 (available).
			   If in use, return status to the channel and wait for it to free up.
			*/

			forkStatus1 = <-fork0
			for forkStatus1 == lower {
				//lower number philosopher has fork. Wait half a second and then check the status again.
				fork0 <- forkStatus1
				time.Sleep(time.Millisecond * time.Duration(500))
				forkStatus1 = <-fork0
			}

			// if Fork is available, take control of it. Update internal fork status variable.
			if forkStatus1 == 4 {
				fork0 <- myNumber
				forkStatus1 = myNumber
			}
			//notice we don't need to do anything if the fork status is already in use by this philosopher. This case will never happen

			//at this point, we must have control of the first fork. Now let's get the second one.

			forkStatus2 = <-fork1
			for forkStatus2 == upper { //then the other philosopher has this fork and we need to wait. Return value to channel, wait, and then check it again.
				//wait half a second and then check the status again.
				fork1 <- forkStatus2
				time.Sleep(time.Millisecond * time.Duration(500))
				forkStatus2 = <-fork1
			}
			// if Fork is available, take control of it. Update internal fork status variable.
			if forkStatus2 == 4 {
				fork1 <- myNumber
				forkStatus2 = myNumber
			}

			//Now we have control of both forks. The philosopher can start eating. He will eat for 1 to 2 seconds.
			fmt.Println(name, "is eating.")

			time.Sleep(time.Millisecond * time.Duration(1000+rand.Intn(1001)))

			//relieve forks
			<-fork0
			fork0 <- 4
			<-fork1
			fork1 <- 4
		}
		//set previous state to show what the philosopher was just doing. Update state with what the philosopher will do next.
		previousState = state
		state = rand.Intn(2)
		// for troubleshooting, put this on next line:	fmt.Println(name, state)
	}

}
